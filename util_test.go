package main

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// TestCheckIfUserIsInRole verifies role membership checks.
func TestCheckIfUserIsInRole(t *testing.T) {
	app := newTestApp(t)

	user := createTestUser(t, app, "role@example.com", []string{"admin", "member"})

	if !checkIfUserIsInRole(user, "admin") {
		t.Fatalf("expected user to have admin role")
	}
	if checkIfUserIsInRole(user, "lootCouncil") {
		t.Fatalf("expected user to not have lootCouncil role")
	}
}

// TestCreateTransactionRecord verifies that transaction records are persisted.
func TestCreateTransactionRecord(t *testing.T) {
	app := newTestApp(t)

	user := createTestUser(t, app, "player@example.com", []string{"member"})
	author := createTestUser(t, app, "author@example.com", []string{"admin"})

	if err := createTransactionRecord(app.App, user.Id, 25, "loot transfer", author.Id); err != nil {
		t.Fatalf("createTransactionRecord returned error: %v", err)
	}

	record, err := app.FindFirstRecordByData("transactions", "note", "loot transfer")
	if err != nil {
		t.Fatalf("failed to find transaction record: %v", err)
	}

	if got := record.GetString("user"); got != user.Id {
		t.Fatalf("expected transaction user %q, got %q", user.Id, got)
	}
	if got := record.GetString("author"); got != author.Id {
		t.Fatalf("expected transaction author %q, got %q", author.Id, got)
	}
	if got := record.GetInt("amount"); got != 25 {
		t.Fatalf("expected transaction amount 25, got %d", got)
	}
}

// createTestUser inserts a user record for use in tests.
func createTestUser(t *testing.T, app *pocketbase.PocketBase, email string, roles []string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("failed to find users collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.SetEmail(email)
	record.SetPassword("password-123")
	record.Set("role", roles)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save user record: %v", err)
	}

	return record
}
