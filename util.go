package main

import (
	"slices"

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
