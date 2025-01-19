package main

import (
	"slices"

	"github.com/pocketbase/pocketbase/core"
)

func checkIfUserIsInRole(record *core.Record, role string) bool {
	return slices.Contains(record.GetStringSlice("roles"), role)
}
