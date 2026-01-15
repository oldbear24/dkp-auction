package main

import (
	"testing"

	_ "github.com/oldbear24/dkp-auction/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func TestTranslateRarity(t *testing.T) {
	cases := []struct {
		name     string
		value    int
		expected string
	}{
		{"common", 3, "common"},
		{"uncommon", 4, "uncommon"},
		{"rare", 5, "rare"},
		{"rare_t2", 6, "rare_t2"},
		{"epic", 10, "epic"},
		{"epic_t2", 11, "epic_t2"},
		{"epic_t3", 12, "epic_t3"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := TranslateRarity(tc.value); got != tc.expected {
				t.Fatalf("TranslateRarity(%d) = %q, expected %q", tc.value, got, tc.expected)
			}
		})
	}
}

func TestTranslateRarityUnknown(t *testing.T) {
	if got := TranslateRarity(99); got != "" {
		t.Fatalf("TranslateRarity(99) = %q, expected empty string", got)
	}
}

func TestGetSettingsReturnsStoredRecord(t *testing.T) {
	app := newTestApp(t)

	insertSettingsRecord(t, app)

	settings, err := GetSettings(app.App)
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}

	if settings == nil {
		t.Fatalf("expected settings result, got nil")
	}

	if !settings.NameSynchronization || settings.SynchronizationType != "tlgh" || settings.SynchronizationUrl != "http://example.com" || settings.SynchronizationClient != "client-id" || settings.SynchronizationPassword != "secret" || settings.SynchronizationDiscordGuildId != "guild-id" || !settings.EnableFloatingEndOfAuction || settings.FloatingEndOfAuctionMinutes != 15 || !settings.EnableTLDBAdapterSync || settings.TldbAdapterUrl != "http://tldb" {
		t.Fatalf("GetSettings returned unexpected values: %#v", settings)
	}
}

func TestGetSettingsInsertsDefaultRowWhenMissing(t *testing.T) {
	app := newTestApp(t)

	settings, err := GetSettings(app.App)
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if settings == nil {
		t.Fatalf("expected settings to be created, got nil")
	}

	var count int
	if err := app.DB().Select("count(*)").From("settings").Row(&count); err != nil {
		t.Fatalf("failed counting settings rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected a single inserted default row, got %d", count)
	}
}

func newTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{HideStartBanner: true, DefaultDataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap pocketbase: %v", err)
	}

	if err := app.RunAllMigrations(); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		if err := app.ResetBootstrapState(); err != nil {
			t.Fatalf("failed to reset pocketbase state: %v", err)
		}
	})

	return app
}

func insertSettingsRecord(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("settings")
	if err != nil {
		t.Fatalf("failed to find settings collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("nameSynchronization", true)
	record.Set("synchronizationType", "tlgh")
	record.Set("synchronizationUrl", "http://example.com")
	record.Set("synchronizationClient", "client-id")
	record.Set("synchronizationPassword", "secret")
	record.Set("synchronizationDiscordGuildId", "guild-id")
	record.Set("enableFloatingEndOfAuction", true)
	record.Set("floatingEndOfAuctionMinutes", 15)
	record.Set("enableTLDBAdapterSync", true)
	record.Set("tldbAdapterUrl", "http://tldb")

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save settings record: %v", err)
	}
}
