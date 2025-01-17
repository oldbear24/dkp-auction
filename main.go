package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	_ "github.com/oldbear24/dkp-auction/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
	"golang.org/x/sync/errgroup"
)

func main() {
	app := pocketbase.New()
	// Add your custom routes or hooks here

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
	app.OnRecordCreate("auctions").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("state") == "" {
			e.Record.Set("state", "ongoing")
		}
		return e.Next()
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
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
