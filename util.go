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
	if app.Save(record) != nil {
		return err
	}
	return nil
}

func GetSettings(app core.App) (*Settings, error) {
	settings := &Settings{}
	err := app.DB().Select("*").From("settings").Limit(1).One(&settings)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no settings are found, create default settings
			settings = &Settings{
				NameSynchronization:           false,
				SynchronizationType:           "",
				SynchronizationUrl:            "",
				SynchronizationClient:         "",
				SynchronizationPassword:       "",
				SynchronizationDiscordGuildId: "",
				EnableFloatingEndOfAuction:    false,
				FloatingEndOfAuctionMinutes:   0,
			}
			if _, err := app.DB().Insert("settings", dbx.Params{}).Execute(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}

		return nil, err
	}
	return settings, nil
}
