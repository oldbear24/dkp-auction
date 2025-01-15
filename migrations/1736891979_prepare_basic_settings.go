package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// add up queries...
		app.Settings().Meta.AppName = "Auction DKP"
		return nil
	}, func(app core.App) error {
		// add down queries...

		return nil
	})
}
