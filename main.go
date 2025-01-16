package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	_ "github.com/oldbear24/dkp-auction/migrations"
	"github.com/pocketbase/dbx"
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
		se.Router.POST("/api/bid/:id", func(e *core.RequestEvent) error {
			app.Logger().Debug("Bid event", "eventData", e)
			bodyData := make([]byte, e.Request.ContentLength)
			_, err := e.Request.Body.Read(bodyData)
			if err != nil {
				return e.BadRequestError("Error reading request body", err)
			}
			defer e.Request.Body.Close()
			bidData := BidStruct{}
			err = json.Unmarshal(bodyData, &bidData)
			if err != nil {
				return e.BadRequestError("Error parsing request body", err)
			}

			biddedAmount := bidData.Amount
			if e.Auth == nil {
				return e.UnauthorizedError("Unauthorized", nil)
			}
			auctionId := e.Request.URL.Query().Get("id")
			if auctionId == "" {
				return e.BadRequestError("Auction ID is required", nil)
			}
			app.RunInTransaction(func(tx core.App) error {
				auctionRecord, err := tx.FindRecordById("auctions", auctionId)
				if err != nil {
					app.Logger().Error("Error finding auction", "error", err, "id", auctionId)
					return e.BadRequestError("Auction not found", err)
				}
				if auctionRecord.GetString("state") != "ongoing" {
					return e.BadRequestError("Auction has ended", nil)
				}
				user, err := tx.FindRecordById("users", e.Auth.Id)
				if err != nil {
					app.Logger().Error("Error finding user", "error", err)
					return e.BadRequestError("User not found", err)
				}
				var biddedOnAuction int = 0
				bidRecordExists := false
				bidRecord, err := tx.FindRecordsByFilter("bids", "user = {:auctionId} && user = {:userId}", "", 1, 0, dbx.Params{"auctionId": auctionId, "userId": e.Auth.Id})
				if err == nil {
					bidRecordExists = true
					biddedOnAuction = bidRecord[0].GetInt("amount")
				}
				auctionBid := auctionRecord.GetInt("currentBid")
				if auctionBid == 0 {
					auctionBid = auctionRecord.GetInt("startingBid")
				}

				userTokens := user.GetInt("tokens")
				userReservedTokens := user.GetInt("reservedTokens")

				userTokens = userTokens - userReservedTokens
				if biddedAmount > userTokens {
					return e.BadRequestError("Insufficient tokens", nil)
				}
				fullAuctionBid := auctionBid + biddedOnAuction
				if biddedAmount <= fullAuctionBid {
					return e.BadRequestError("Bid sum should be greater than current bid", nil)
				}
				auctionRecord.Set("currentBid", fullAuctionBid)
				if bidRecordExists {
					bidRecord[0].Set("amount", biddedAmount)
					err = tx.Save(bidRecord[0])
					if err != nil {
						return e.BadRequestError("Error updating bid", err)
					}
				} else {
					coll, err := tx.FindCollectionByNameOrId("bids")
					if err != nil {
						app.Logger().Error("Error finding collection", "error", err)
						return e.BadRequestError("Error finding collection", err)
					}
					newBidRecord := core.NewRecord(coll)
					newBidRecord.Set("amount", biddedAmount)
					newBidRecord.Set("auctionId", auctionId)
					newBidRecord.Set("userId", e.Auth.Id)
					err = tx.Save(newBidRecord)
					if err != nil {
						return e.BadRequestError("Error creating bid", err)
					}
					user.Set("reservedTokens", userReservedTokens+biddedAmount)
					err = tx.Save(user)
					if err != nil {
						return e.BadRequestError("Error updating user", err)
					}
				}
				return nil
			})

			return e.Next()
		})
		return se.Next()
	})
	app.OnRecordAuthWithOAuth2Request("users").BindFunc(func(e *core.RecordAuthWithOAuth2RequestEvent) error {

		app.Logger().Debug("Discord auth event", "eventData", e)
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
