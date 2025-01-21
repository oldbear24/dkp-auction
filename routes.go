package main

import (
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func RegisterRoutes(se *core.ServeEvent) {
	se.Router.POST("/api/bid/{id}", handleBid).Bind(apis.RequireAuth())
	se.Router.POST("/api/change-tokens", chaneUsersAmount).Bind(apis.RequireAuth())
	se.Router.POST("/api/set-validated/{user}", setVerified).Bind(apis.RequireAuth())
	se.Router.POST("/api/resolve-auction/{id}", resolveAuction).Bind(apis.RequireAuth())

}
func resolveAuction(e *core.RequestEvent) error {
	if !checkIfUserIsInRole(e.Auth, "manager") {
		return e.UnauthorizedError("Unauthorized", nil)
	}
	auctionResolveId := e.Request.PathValue("id")
	if auctionResolveId == "" {
		return e.BadRequestError("Auction result ID is required", nil)
	}
	record, err := e.App.FindRecordById("auctionsResult", auctionResolveId)
	if err != nil {
		return e.BadRequestError("Could not retrieve record", err)
	}
	if record.GetBool("resolved") {
		e.BadRequestError("Record is already resolved!", nil)
	}
	record.Set("resolved", true)
	record.Set("resolvedBy", e.Auth.Id)
	if err := e.App.Save(record); err != nil {
		return e.BadRequestError("Could not save record", err)
	}
	return e.JSON(200, map[string]interface{}{
		"success": true,
	})
}
func setVerified(e *core.RequestEvent) error {
	var data struct {
		Validated bool `json:"validated"`
	}
	if !checkIfUserIsInRole(e.Auth, "manager") {
		return e.UnauthorizedError("Unauthorized", nil)
	}
	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Invalid data", err)
	}
	user, err := e.App.FindRecordById("users", e.Request.PathValue("user"))
	if err != nil {
		return e.BadRequestError("User not found", err)
	}
	user.Set("validated", data.Validated)
	if err := e.App.Save(user); err != nil {
		return e.BadRequestError("Error saving user", err)
	}
	return e.JSON(200, map[string]interface{}{
		"success": true,
	})
}

func chaneUsersAmount(e *core.RequestEvent) error {
	var data struct {
		UserIds []string `json:"userIds"`
		Amount  int      `json:"amount"`
	}

	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Invalid data", err)
	}

	if !checkIfUserIsInRole(e.Auth, "manager") {
		return e.UnauthorizedError("Unauthorized", nil)

	}
	return e.App.RunInTransaction(func(tx core.App) error {
		for _, userId := range data.UserIds {
			user, err := tx.FindRecordById("users", userId)
			if err != nil {
				return e.BadRequestError("User not found", err)
			}
			if data.Amount < 0 {
				return e.BadRequestError("Amount cannot be lower than 0", nil)
			}

			newAmount := user.GetInt("tokens") + data.Amount

			user.Set("tokens", newAmount)
			if err := tx.Save(user); err != nil {
				return e.BadRequestError("Error saving user", err)
			}
			if err := createTransactionRecord(tx, user.Id, data.Amount, "Token top-up", e.Auth.Id); err != nil {
				return e.BadRequestError("Failed to create transaction record", err)
			}

		}
		return e.JSON(200, map[string]interface{}{
			"success": true,
		})
	})
}

func handleBid(e *core.RequestEvent) error {
	var bidData struct {
		Amount int `json:"amount"`
	}

	if err := e.BindBody(&bidData); err != nil {
		return e.BadRequestError("Invalid bid data", err)
	}

	if e.Auth == nil {
		return e.UnauthorizedError("Unauthorized", nil)
	}

	auctionId := e.Request.PathValue("id")
	if auctionId == "" {
		return e.BadRequestError("Auction ID is required", nil)
	}

	return e.App.RunInTransaction(func(tx core.App) error {
		// 1. Get auction
		auction, err := tx.FindRecordById("auctions", auctionId)
		if err != nil {
			return e.NotFoundError("Auction not found", err)
		}
		if auction.GetDateTime("endTime").Before(types.NowDateTime()) {
			return e.BadRequestError("Auction has ended", nil)

		}
		// 2. Validate auction state
		if auction.GetString("state") != "ongoing" {
			return e.BadRequestError("Auction is not active", nil)
		}

		// 3. Get user
		user, err := tx.FindRecordById("users", e.Auth.Id)
		if err != nil {
			return e.NotFoundError("User not found", err)
		}

		// 4. Validate bid amount
		currentBid := auction.GetInt("currentBid")
		startingBid := auction.GetInt("startingBid")
		minBid := max(startingBid, currentBid+1)

		if bidData.Amount < minBid {
			return e.BadRequestError("Bid is too low", nil)
		}
		existingBids, err := tx.FindRecordsByFilter(
			"bids",
			"auction = {:auctionId} && user = {:userId}",
			"",
			1,
			0,
			dbx.Params{"auctionId": auctionId, "userId": e.Auth.Id},
		)
		existinBidForCompare := 0
		if len(existingBids) > 0 {
			existinBidForCompare = existingBids[0].GetInt("amount")
		}
		if err != nil {
			return e.BadRequestError("Error checking existing bids", err)
		}
		// 5. Check user balance
		availableTokens := user.GetInt("tokens") - (user.GetInt("reservedTokens") - existinBidForCompare)
		e.App.Logger().Debug("Bid tokens", "user", user.GetInt("tokens"), "res", user.GetInt("reservedTokens"), "exBid", existinBidForCompare, "all", availableTokens)
		if bidData.Amount > availableTokens {
			return e.BadRequestError("Insufficient tokens", nil)
		}

		// 6. Get existing bid

		// 7. Create or update bid
		var bidRecord *core.Record
		if len(existingBids) > 0 {
			bidRecord = existingBids[0]
			previousAmount := bidRecord.GetInt("amount")
			user.Set("reservedTokens", user.GetInt("reservedTokens")-previousAmount)
		} else {
			collection, err := tx.FindCachedCollectionByNameOrId("bids")
			if err != nil {
				return e.BadRequestError("Error creating bid", err)
			}
			bidRecord = core.NewRecord(collection)
			bidRecord.Set("auction", auctionId)
			bidRecord.Set("user", e.Auth.Id)
		}

		// 8. Update bid and user records
		bidRecord.Set("amount", bidData.Amount)
		bidRecord.Set("timestamp", time.Now().Unix())

		user.Set("reservedTokens", user.GetInt("reservedTokens")+bidData.Amount)
		auction.Set("currentBid", bidData.Amount)
		auction.Set("winner", e.Auth.Id)

		// 9. Save all changes
		if err := tx.Save(bidRecord); err != nil {
			return e.BadRequestError("Error saving bid", err)
		}
		if err := tx.Save(user); err != nil {
			return e.BadRequestError("Error updating user tokens", err)
		}
		if err := tx.Save(auction); err != nil {
			return e.BadRequestError("Error updating auction", err)
		}

		return e.JSON(200, map[string]interface{}{
			"success": true,
			"bid":     bidRecord,
		})
	})
}
