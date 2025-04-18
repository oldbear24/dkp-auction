package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// add up queries...
		_, err := app.DB().Update("auctions", dbx.Params{"state": "ongoing"}, dbx.NewExp("state = ''")).Execute()
		if err != nil {
			return err
		}
		return nil
	}, func(app core.App) error {
		// add down queries...
		return nil
	})
}
