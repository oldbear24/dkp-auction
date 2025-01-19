package main

import (
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
		}
		return nil
	})
}
