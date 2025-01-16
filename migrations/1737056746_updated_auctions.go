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
		if err := collection.Fields.AddMarshaledJSONAt(8, []byte(`{
			"hidden": false,
			"id": "select2744374011",
			"maxSelect": 1,
			"name": "state",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"ongoing",
				"finished"
			]
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"cascadeDelete": false,
			"collectionId": "_pb_users_auth_",
			"hidden": false,
			"id": "relation217473038",
			"maxSelect": 1,
			"minSelect": 0,
			"name": "winner",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "relation"
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
		collection.Fields.RemoveById("select2744374011")

		// remove field
		collection.Fields.RemoveById("relation217473038")

		return app.Save(collection)
	})
}
