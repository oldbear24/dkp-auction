package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

			bidRecords, err := tx.FindRecordsByFilter("bids", "auction = {:auctionId}", "", 0, 0, dbx.Params{"auctionId": record.Id})
			if err != nil {
				return err
			}
			bidExists := len(bidRecords) > 0

			record.Set("state", "finished")

			coll, err := tx.FindCachedCollectionByNameOrId("auctionsResult")
			if err != nil {
				return err
			}
			resultRecord := core.NewRecord(coll)
			resultRecord.Set("auction", record.Id)
			if bidExists {
				winnerRecordIndex := -1
				for i := range bidRecords {
					if winnerRecordIndex < 0 || bidRecords[i].GetInt("amount") > bidRecords[winnerRecordIndex].GetInt("amount") {
						winnerRecordIndex = i
					}
				}
				record.Set("winner", bidRecords[winnerRecordIndex].GetString("user"))
				for _, bidRecord := range bidRecords {
					userRecord, err := tx.FindRecordById("users", bidRecord.GetString("user"))
					if err != nil {
						return err
					}

					if bidRecord.Id == bidRecords[winnerRecordIndex].Id {
						userRecord.Set("reservedTokens", userRecord.GetInt("reservedTokens")-bidRecord.GetInt("amount"))
						userRecord.Set("tokens", userRecord.GetInt("tokens")-bidRecord.GetInt("amount"))
						if err := createTransactionRecord(tx, userRecord.Id, -bidRecord.GetInt("amount"), "Win in auction", ""); err != nil {
							return err
						}
						notifyUser(userRecord.Id, "You won the auction")
					} else {
						userRecord.Set("reservedTokens", userRecord.GetInt("reservedTokens")-bidRecord.GetInt("amount"))
					}
					err = tx.Save(userRecord)
					if err != nil {
						return err
					}
				}
			}
			err = tx.Save(record)
			if err != nil {
				return err
			}
			if err := tx.Save(resultRecord); err != nil {
				return err
			}
			notifyRole("manager", "Auction has ended")
		}
		return nil
	})
}

func updateUserNames(app *pocketbase.PocketBase) error {
	settingsArray, err := app.FindAllRecords("settings")
	if err != nil {
		return err
	}
	if len(settingsArray) == 0 {
		return errors.New("settings error does not exists")
	}
	settings := settingsArray[0]
	if !settings.GetBool("nameSynchronization") {
		return nil
	}
	if settings.GetString("synchronizationType") == "tlgh" {
		baseUrl := settings.GetString("synchronizationUrl")
		clientId := settings.GetString("synchronizationClient")
		pass := settings.GetString("synchronizationPassword")
		guildId := settings.GetString("synchronizationDiscordGuildId")
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
