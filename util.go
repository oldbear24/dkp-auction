package main

import (
	"database/sql"
	"errors"
	"slices"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func checkIfUserIsInRole(record *core.Record, role string) bool {
	return slices.Contains(record.GetStringSlice("role"), role)
}

func createTransactionRecord(app core.App, userId string, amount int, note string, authorId string) error {
	coll, err := app.FindCachedCollectionByNameOrId("transactions")
	if err != nil {
		return err
	}
	record := core.NewRecord(coll)
	record.Set("user", userId)
	record.Set("amount", amount)
	record.Set("note", note)
	record.Set("author", authorId)
	if err := app.Save(record); err != nil {
		return err
	}
	return nil
}

func GetSettings(app core.App) (*Settings, error) {
	settings := &Settings{}
	err := app.DB().Select("*").From("settings").Limit(1).One(&settings)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no settings are found, create a default row then re-query
			if _, err := app.DB().Insert("settings", dbx.Params{}).Execute(); err != nil {
				return nil, err
			}
			// try to read the inserted settings
			settings = &Settings{}
			if err := app.DB().Select("*").From("settings").Limit(1).One(&settings); err != nil {
				return nil, err
			}
			return settings, nil
		}
		return nil, err
	}
	return settings, nil
}

func TranslateRarity(rarity int) string {
	switch rarity {
	case 3:
		return "common"
	case 4:
		return "uncommon"
	case 5:
		return "rare"
	case 6:
		return "rare_t2"
	case 10:
		return "epic"
	case 11:
		return "epic_t2"
	case 12:
		return "epic_t3"

	default:
		return ""
	}
}
