package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1337428601")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"hidden": false,
			"id": "select3082862150",
			"maxSelect": 1,
			"name": "rarity",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"common",
				"uncommon",
				"rare",
				"rare_t2",
				"epic",
				"epic_t2",
				"epic_t3",
				"heroic",
				"artifact"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1337428601")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("select3082862150")

		return app.Save(collection)
	})
}
