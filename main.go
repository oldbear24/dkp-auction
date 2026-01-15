package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/oldbear24/dkp-auction/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
	"golang.org/x/sync/errgroup"
)

// main configures and starts the PocketBase application.
func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), true))

		return se.Next()
	})

	// Add your custom routes or hooks here
	app.Cron().MustAdd("finishAuctions", "* * * * *", func() {
		if err := finishAuction(app); err != nil {
			app.Logger().Error("finishAuction error", "error", err)
		}
	})
	app.Cron().MustAdd("updateUserNames", "0 3 * * *", func() {
		if err := updateUserNames(app); err != nil {
			app.Logger().Error("updateUserNames error", "error", err)
		}
	})
	app.Cron().MustAdd("runTokenHealthCheck", "0 1 * * *", func() {
		if err := runTokenHealthCheck(app); err != nil {
			app.Logger().Error("runTokenHealthCheck error", "error", err)
		}
	})
	app.Cron().MustAdd("getTLDBItems", "0 2 * * 6", func() {
		if err := getTLDBItems(app); err != nil {
			app.Logger().Error("getTLDBItems error", "error", err)
		}
	})
	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// register a global middleware

		se.Router.BindFunc(func(e *core.RequestEvent) error {
			if strings.HasPrefix(e.Request.URL.Path, "/api") {
				timestamp := e.Request.Header.Get("timestamp")
				if timestamp != "" {
					e.Response.Header().Set("timestamp", timestamp)
				}
			}
			return e.Next()
		})
		RegisterRoutes(se)
		RegisterApiRoutes(se)
		return se.Next()
	})
	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {

		e.App.Logger().Debug("Discord auth event", "eventData", e)
		return e.Next()
	})
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("role") == "" {
			e.Record.Set("role", "member")
		}
		return e.Next()
	})
	app.OnRecordCreate("apiKeys").BindFunc(func(e *core.RecordEvent) error {
		e.Record.Set("apiKey", security.RandomString(250))
		return e.Next()
	})
	app.OnRecordAfterUpdateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		var data struct {
			UserId string `json:"userId"`
		}
		data.UserId = e.Record.Id
		notify(e.App, "manage_users_update", data)
		return e.Next()
	})
	app.OnRecordCreate("auctions").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("state") == "" {
			e.Record.Set("state", "ongoing")
		}
		if e.Record.GetString("mainImage") == "" {
			rec, err := e.App.FindFirstRecordByData("items", "name", e.Record.GetString("itemName"))
			if err == nil {
				itemImageKey := rec.BaseFilesPath() + "/" + rec.GetString("icon")
				fsys, err := app.NewFilesystem()
				if err != nil {
					return err
				}
				defer fsys.Close()
				r, err := fsys.GetReader(itemImageKey)
				if err != nil {
					return err
				}
				defer r.Close()
				content := new(bytes.Buffer)
				_, err = io.Copy(content, r)
				if err != nil {
					return err
				}
				fileData, err := io.ReadAll(content)
				if err != nil {
					return err
				}
				f, err := filesystem.NewFileFromBytes(fileData, rec.GetString("icon"))
				if err != nil {
					return err
				}
				e.Record.Set("mainImage", f)
				e.Record.Set("rarity", rec.Get("rarity"))
			}
		}
		return e.Next()
	})
	app.OnRecordCreate("settings").BindFunc(func(e *core.RecordEvent) error {
		err := e.App.DB().Select("id").From("settings").One(nil)
		if err == nil {
			return errors.New("settings record already exists")
		}
		return e.Next()
	})
	go startNotificationWorker(app.App)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
// notify sends a subscription message to connected clients.
func notify(app core.App, subscription string, data any) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := subscriptions.Message{
		Name: subscription,
		Data: rawData,
	}

	group := new(errgroup.Group)

	chunks := app.SubscriptionsBroker().ChunkedClients(300)

	for _, chunk := range chunks {
		group.Go(func() error {
			for _, client := range chunk {
				if !client.HasSubscription(subscription) {
					continue
				}

				client.Send(message)
			}

			return nil
		})
	}

	return group.Wait()
}
