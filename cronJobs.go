package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func finishAuction(app *pocketbase.PocketBase) error {
	records, err := app.FindRecordsByFilter("auctions", "state = 'ongoing' && endTime < @now", "", 0, 0, nil)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil

	}
	return app.RunInTransaction(func(tx core.App) error {
		for _, record := range records {
			record.Set("state", "finished")

			coll, err := tx.FindCachedCollectionByNameOrId("auctionsResult")
			if err != nil {
				return err
			}
			resultRecord := core.NewRecord(coll)
			resultRecord.Set("auction", record.Id)
			userId := ""
			if record.GetString("winner") != "" {
				userRecord, err := tx.FindRecordById("users", record.GetString("winner"))
				if err != nil {
					return err
				}

				userRecord.Set("reservedTokens", userRecord.GetInt("reservedTokens")-record.GetInt("currentBid"))
				userRecord.Set("tokens", userRecord.GetInt("tokens")-record.GetInt("currentBid"))
				if err := createTransactionRecord(tx, userRecord.Id, -record.GetInt("currentBid"), "Win in auction", ""); err != nil {
					return err
				}
				if err := tx.Save(userRecord); err != nil {
					return err
				}
				userId = userRecord.Id
			}

			err = tx.Save(record)
			if err != nil {
				return err
			}
			if err := tx.Save(resultRecord); err != nil {
				return err
			}
			if userId != "" {
				notifyUser(userId, "You won the auction")
			}
			notifyRole("manager", "Auction has ended")
		}
		return nil
	})
}

func updateUserNames(app *pocketbase.PocketBase) error {
	settings, err := GetSettings(app)
	if err != nil {
		return err
	}
	if settings.SynchronizationType == "tlgh" {
		baseUrl := settings.SynchronizationUrl
		clientId := settings.SynchronizationClient
		pass := settings.SynchronizationPassword
		guildId := settings.SynchronizationDiscordGuildId
		var loginData struct {
			Identity string `json:"identity"`
			Password string `json:"password"`
		}
		loginData.Identity = clientId
		loginData.Password = pass
		data, err := json.Marshal(loginData)
		if err != nil {
			return err
		}
		dataReader := bytes.NewReader(data)
		loginPath, err := url.JoinPath(baseUrl, "/api/collections/users/auth-with-password")
		if err != nil {
			return err
		}
		resp, err := http.Post(loginPath, "application/json", dataReader)

		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var loginResponse struct {
			Token string `json:"token"`
		}
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &loginResponse); err != nil {
			return err
		}
		var bearer = "Bearer " + loginResponse.Token
		nickUrl, err := url.JoinPath(baseUrl, "/api/tlgh/get-nicknames/", guildId)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("GET", nickUrl, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", bearer)
		client := &http.Client{}

		resp2, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		type User struct {
			Id       string `json:"id"`
			Nickname string `json:"nickname"`
		}
		usersResponse := []User{}

		body2, err := io.ReadAll(resp2.Body)

		if err != nil {
			return err
		}
		err = json.Unmarshal(body2, &usersResponse)
		if err != nil {
			return err
		}
		records, err := app.FindRecordsByFilter("users", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			for _, user := range usersResponse {
				if user.Id == record.GetString("discordId") {
					record.Set("name", user.Nickname)
					if err := app.Save(record); err != nil {
						return err
					}
					break
				}

			}
		}
	}

	return nil
}

func runTokenHealthCheck(app *pocketbase.PocketBase) error {
	records, err := app.FindAllRecords("users")
	if err != nil {
		return err
	}
	isError := false
	userResults := []TokenHealtCheckUser{}
	for _, record := range records {
		userTokens := record.GetInt("tokens")
		transactionRecords, err := app.FindRecordsByFilter("transactions", "user = {:userId}", "", 0, 0, dbx.Params{"userId": record.Id})
		if err != nil {
			return err
		}
		transactionTokens := 0

		for _, transactionRecord := range transactionRecords {
			transactionTokens += transactionRecord.GetInt("amount")
		}
		differece := userTokens - transactionTokens
		state := "ok"
		if differece != 0 {
			state = "error"
			isError = true
		}
		userResults = append(userResults, TokenHealtCheckUser{
			State:             state,
			User:              record.Id,
			UserTokens:        userTokens,
			TransactionTokens: transactionTokens,
			Differece:         differece,
		})
	}

	state := "ok"
	if isError {
		state = "error"
	}
	healthCheck := TokenHealtCheck{
		State:       state,
		UserResults: userResults,
	}
	coll, err := app.FindCachedCollectionByNameOrId("tokenHealthChecks")
	if err != nil {
		return err
	}

	record := core.NewRecord(coll)
	jsonData, err := json.Marshal(healthCheck)
	if err != nil {
		return err
	}
	record.Set("state", healthCheck.State)
	record.Set("result", string(jsonData))
	if err := app.Save(record); err != nil {
		return err
	}
	if isError {
		notifyRole("admin", "Token health check failed")
	}
	return nil
}
