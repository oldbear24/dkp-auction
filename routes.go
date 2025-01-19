package main

import (
	"slices"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func RegisterRoutes(se *core.ServeEvent) {
	se.Router.POST("/api/bid/{id}", handleBid).Bind(apis.RequireAuth())
	se.Router.POST("/api/change-tokens", chaneUsersAmount).Bind(apis.RequireAuth())
	se.Router.POST("/api/set-verified/{user}", setVerified).Bind(apis.RequireAuth())

}

func setVerified(e *core.RequestEvent) error {
	var data struct {
		Verified bool `json:"verified"`
	}
	if condition := e.Auth == nil; condition {
		return e.UnauthorizedError("Unauthorized", nil)
	}
	if condition := !slices.Contains(e.Auth.GetStringSlice("roles"), "manager"); condition {
		return e.UnauthorizedError("Unauthorized", nil)
	}
	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Invalid data", err)
	}
	user, err := e.App.FindRecordById("users", e.Request.PathValue("user"))
	if err != nil {
		return e.BadRequestError("User not found", err)
	}
	user.Set("verified", data.Verified)
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

	if e.Auth == nil {
		return e.UnauthorizedError("Unauthorized", nil)
	}
	if !slices.Contains(e.Auth.GetStringSlice("roles"), "manager") {
		return e.UnauthorizedError("Unauthorized", nil)

	}
	return e.App.RunInTransaction(func(tx core.App) error {
		for _, userId := range data.UserIds {
			user, err := tx.FindRecordById("users", userId)
			if err != nil {
				return e.BadRequestError("User not found", err)
			}
			newAmount := user.GetInt("tokens") + data.Amount
			if newAmount < 0 {
				newAmount = 0

			}
			user.Set("tokens", newAmount)
			if err := tx.Save(user); err != nil {
				return e.BadRequestError("Error saving user", err)
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
			return e.BadRequestError("Bid too low", nil)
		}

		// 5. Check user balance
		availableTokens := user.GetInt("tokens") - user.GetInt("reservedTokens")
		if bidData.Amount > availableTokens {
			return e.BadRequestError("Insufficient tokens", nil)
		}

		// 6. Get existing bid
		existingBids, err := tx.FindRecordsByFilter(
			"bids",
			"auction = {:auctionId} && user = {:userId}",
			"",
			1,
			0,
			dbx.Params{"auctionId": auctionId, "userId": e.Auth.Id},
		)
		if err != nil {
			return e.BadRequestError("Error checking existing bids", err)
		}

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
		//auction.Set("highestBidder", e.Auth.Id)

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
