package main

import (
	"fmt"
	"math"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// RegisterRoutes configures the authenticated API endpoints.
func RegisterRoutes(se *core.ServeEvent) {
	se.Router.POST("/api/bid/{id}", handleBid).Bind(apis.RequireAuth())
	se.Router.POST("/api/change-tokens", chaneUsersAmount).Bind(apis.RequireAuth())
	se.Router.POST("/api/set-validated/{user}", setVerified).Bind(apis.RequireAuth())
	se.Router.POST("/api/resolve-auction/{id}", resolveAuction).Bind(apis.RequireAuth())
	se.Router.POST("/api/seen-notifications/{id}", seenNotification).Bind(apis.RequireAuth())
	se.Router.POST("/api/seen-notifications", seenNotifications).Bind(apis.RequireAuth())
	se.Router.POST("/api/clear-tokens", clearTokens).Bind(apis.RequireAuth())
	se.Router.POST("/api/add-to-favourites/{id}", addToFavourites).Bind(apis.RequireAuth())
	se.Router.POST("/api/remove-from-favourites/{id}", removeFromFavourites).Bind(apis.RequireAuth())

}

// addToFavourites adds an auction to the user's favourites.
func addToFavourites(e *core.RequestEvent) error {
	auctionId := e.Request.PathValue("id")
	if auctionId == "" {
		return e.BadRequestError("Auction ID is required", nil)
	}
	auctionRecord, err := e.App.FindRecordById("auctions", auctionId)
	if err != nil {
		return e.BadRequestError("Could not retrieve auction", err)
	}
	auctionRecord.Set("favourites+", e.Auth.Id)
	if err := e.App.Save(auctionRecord); err != nil {
		return e.BadRequestError("Could not save auction", err)
	}
	return e.JSON(200, map[string]interface{}{
		"success": true,
	})
}

// removeFromFavourites removes an auction from the user's favourites.
func removeFromFavourites(e *core.RequestEvent) error {
	auctionId := e.Request.PathValue("id")
	if auctionId == "" {
		return e.BadRequestError("Auction ID is required", nil)
	}
	auctionRecord, err := e.App.FindRecordById("auctions", auctionId)
	if err != nil {
		return e.BadRequestError("Could not retrieve auction", err)
	}
	auctionRecord.Set("favourites-", e.Auth.Id)
	if err := e.App.Save(auctionRecord); err != nil {
		return e.BadRequestError("Could not save auction", err)
	}
	return e.JSON(200, map[string]interface{}{
		"success": true,
	})

}

// resolveAuction marks an auction result as resolved.
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
		return e.BadRequestError("Record is already resolved!", nil)
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

// setVerified updates the validated flag for a user.
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

// chaneUsersAmount adjusts tokens for one or more users.
func chaneUsersAmount(e *core.RequestEvent) error {
	var data struct {
		UserIds []string `json:"userIds"`
		Amount  int      `json:"amount"`
		Reason  string   `json:"reason"`
	}

	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Invalid data", err)
	}

	if !checkIfUserIsInRole(e.Auth, "manager") {
		return e.UnauthorizedError("Unauthorized", nil)

	}
	if len(data.UserIds) == 0 {
		return e.BadRequestError("UserIds are required", nil)
	}
	if data.Amount == 0 {
		return e.BadRequestError("Amount cannot be 0", nil)
	}
	message := ""
	if data.Reason != "" {
		message = data.Reason
	} else {
		if data.Amount > 0 {
			message = "Token top-up"
		} else {
			message = "Token deduction"
		}
	}
	return e.App.RunInTransaction(func(tx core.App) error {
		for _, userId := range data.UserIds {
			user, err := tx.FindRecordById("users", userId)
			if err != nil {
				return e.BadRequestError("User not found", err)
			}

			newAmount := user.GetInt("tokens") + data.Amount

			user.Set("tokens", newAmount)
			if err := tx.Save(user); err != nil {
				return e.BadRequestError("Error saving user", err)
			}

			if err := createTransactionRecord(tx, user.Id, data.Amount, message, e.Auth.Id); err != nil {
				return e.BadRequestError("Failed to create transaction record", err)
			}

		}
		for _, r := range data.UserIds {
			notifyUser(r, fmt.Sprintf("Your tokens have been updated by %d. Reason: %s", data.Amount, message))
		}
		return e.JSON(200, map[string]interface{}{
			"success": true,
		})
	})
}

// handleBid validates and records a bid on an auction.
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
	settings, err := GetSettings(e.App)
	if err != nil {
		return e.BadRequestError("Error getting settings", err)
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
		if err != nil {
			return e.BadRequestError("Error checking existing bids", err)
		}
		existinBidForCompare := 0
		if len(existingBids) > 0 {
			existinBidForCompare = existingBids[0].GetInt("amount")
		}
		// 5. Check user balance
		availableTokens := 0
		if e.Auth.Id == auction.GetString("winner") {
			availableTokens = user.GetInt("tokens") - user.GetInt("reservedTokens") + existinBidForCompare
		} else {
			availableTokens = user.GetInt("tokens") - user.GetInt("reservedTokens")
		}
		e.App.Logger().Debug("Bid tokens", "user", user.GetInt("tokens"), "res", user.GetInt("reservedTokens"), "exBid", existinBidForCompare, "all", availableTokens)
		if bidData.Amount > availableTokens {
			return e.BadRequestError("Insufficient tokens", nil)
		}

		// 6. Get existing bid

		// 7. Create or update bid
		var bidRecord *core.Record
		if len(existingBids) == 0 {
			collection, err := tx.FindCachedCollectionByNameOrId("bids")
			if err != nil {
				return e.BadRequestError("Error creating bid", err)
			}
			bidRecord = core.NewRecord(collection)
			bidRecord.Set("auction", auctionId)
			bidRecord.Set("user", e.Auth.Id)
		} else {
			bidRecord = existingBids[0]
		}

		// 8. Update bid and user records
		bidRecord.Set("amount", bidData.Amount)
		bidRecord.Set("timestamp", time.Now().Unix())

		//Reset previsou winner tokens
		previsousWinnerId := auction.GetString("winner")
		if previsousWinnerId != "" && previsousWinnerId != e.Auth.Id {
			previsousWinner, err := tx.FindRecordById("users", previsousWinnerId)
			if err != nil {
				return e.BadRequestError("Error finding previous winner", err)
			}
			previsousWinner.Set("reservedTokens", previsousWinner.GetInt("reservedTokens")-auction.GetInt("currentBid"))
			if err := tx.Save(previsousWinner); err != nil {
				return e.BadRequestError("Error saving previous winner", err)
			}
			// Notify previous winner
			notifyUser(previsousWinner.Id, fmt.Sprintf("Your bid was outbid by %d tokens", bidData.Amount))
		}
		tokensToReserve := 0
		if e.Auth.Id == auction.GetString("winner") {
			tokensToReserve = bidData.Amount - existinBidForCompare
		} else {
			tokensToReserve = bidData.Amount
		}
		user.Set("reservedTokens", user.GetInt("reservedTokens")+tokensToReserve)
		auction.Set("currentBid", bidData.Amount)
		auction.Set("winner", e.Auth.Id)
		if settings.EnableFloatingEndOfAuction {
			newEndTime := time.Now().UTC().Add(time.Minute * time.Duration(settings.FloatingEndOfAuctionMinutes))
			if newEndTime.After(auction.GetDateTime("endTime").Time()) {
				auction.Set("endTime", newEndTime)
			}
		}

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

// seenNotifications marks all notifications for the current user as seen.
func seenNotifications(e *core.RequestEvent) error {
	notifications, err := e.App.FindRecordsByFilter("notifications", "user = {:userId}", "", 0, 0, dbx.Params{"userId": e.Auth.Id})
	if err != nil {
		return e.BadRequestError("Error finding notifications", err)
	}
	for _, notification := range notifications {
		notification.Set("seen", true)
		if err := e.App.Save(notification); err != nil {
			return e.BadRequestError("Error saving notification", err)
		}
	}
	return e.JSON(200, map[string]interface{}{
		"success": true,
	})

}

// seenNotification marks a single notification as seen.
func seenNotification(e *core.RequestEvent) error {
	notificationId := e.Request.PathValue("id")
	if notificationId == "" {
		return e.BadRequestError("Notification ID is required", nil)
	}
	notification, err := e.App.FindRecordById("notifications", notificationId)
	if err != nil {
		return e.BadRequestError("Could not retrieve notification", err)
	}
	notification.Set("seen", true)
	if err := e.App.Save(notification); err != nil {
		return e.BadRequestError("Could not save notification", err)
	}
	return e.JSON(200, map[string]interface{}{
		"success": true,
	})
}

// clearTokens removes a percentage of tokens from all users.
func clearTokens(e *core.RequestEvent) error {
	var data struct {
		//UserIds    []string `json:"userIds"`
		Percentage int `json:"percentage"`
	}

	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Invalid data", err)
	}
	if data.Percentage < 0 || data.Percentage > 100 {
		return e.BadRequestError("Percentage must be between 0 and 100", nil)
	}
	if !checkIfUserIsInRole(e.Auth, "manager") {
		return e.UnauthorizedError("Unauthorized", nil)
	}

	return e.App.RunInTransaction(func(tx core.App) error {
		changeData := []ChangeTokens{}
		userRecords, err := tx.FindAllRecords("users")
		if err != nil {
			return e.BadRequestError("Error finding users", err)
		}
		for _, userRecord := range userRecords {
			if userRecord.GetInt("reservedTokens") > 0 {
				return e.BadRequestError("User has reserved tokens", nil)
			}
			currentTokens := userRecord.GetInt("tokens")
			removedAmountFloat := float64(currentTokens) * (float64(data.Percentage) / 100)
			removedAmount := int(math.Ceil(removedAmountFloat))
			userRecord.Set("tokens", currentTokens-removedAmount)
			if err := tx.Save(userRecord); err != nil {
				return e.BadRequestError("Error saving user", err)
			}
			if err := createTransactionRecord(tx, userRecord.Id, -removedAmount, "Token percentage removal", e.Auth.Id); err != nil {
				return e.BadRequestError("Failed to create transaction record", err)
			}
			changeData = append(changeData, ChangeTokens{userRecord.Id, -removedAmount})

		}
		for _, r := range changeData {
			notifyUser(r.User, fmt.Sprintf("Your tokens have been updated by %d", r.Amount))
		}
		return e.JSON(200, map[string]interface{}{
			"success": true,
			"changes": changeData,
		})
	})
}
