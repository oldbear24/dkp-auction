package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// add up queries...
		settings := app.Settings()
		settings.Meta.AppName = "Auction DKP"
		return app.Save(settings)
	}, func(app core.App) error {
		// add down queries...

		return nil
	})
}
