package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
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

func getTLDBItems(app *pocketbase.PocketBase) error {
	settings, err := GetSettings(app)
	if err != nil {
		return err
	}
	if !settings.EnableTLDBAdapterSync {
		app.Logger().Info("TLDB Adapter Sync is disabled in settings")
		return nil
	}
	httpClient := &http.Client{}

	url, err := url.JoinPath(settings.TldbAdapterUrl, "/api/data")
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "PocketBase/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch TLDB items, status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response TLDBAdapterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	coll, err := app.FindCachedCollectionByNameOrId("items")
	if err != nil {
		return err
	}
	for _, item := range response.Items {
		if strings.HasPrefix(item.Name, "Extract:") || strings.HasPrefix(item.Name, "Precious Lithograph:") {
			continue // Skip items that are extracts
		}
		formatedId := strings.ReplaceAll(item.Id, "_", "")

		record, err := app.FindRecordById("items", formatedId)
		if err != nil {
			record = core.NewRecord(coll)
		}
		record.Set("id", formatedId)
		record.Set("name", item.Name)
		record.Set("rarity", TranslateRarity(item.Rarity))
		iconUrlPart := strings.ToLower(item.Icon)

		url := "https://cdn.tldb.info/db/images/ags/v35/256/" + iconUrlPart + ".png"
		reqImg, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		reqImg.Header.Add("User-Agent", "PocketBase/1.0")
		respImg, err := httpClient.Do(reqImg)
		if err != nil {
			return err
		}
		if respImg.StatusCode == http.StatusOK {
			defer respImg.Body.Close()
			imgData, err := io.ReadAll(respImg.Body)
			if err != nil {
				return err
			}

			f, err := filesystem.NewFileFromBytes(imgData, formatedId+".png")
			if err != nil {
				return err
			}

			record.Set("icon", f)
		} else {
			app.Logger().Error("Failed to fetch item icon.", "IconName", item.Name, "StatusCode", respImg.StatusCode, "Url", url)
		}

		if err := app.Save(record); err != nil {
			return err
		}
	}
	// Notify that items have been updated
	notifyRole("admin", "Items have been updated from TLDB")

	return nil
}
