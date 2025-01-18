package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"hidden": false,
			"id": "select1466534506",
			"maxSelect": 4,
			"name": "role",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"member",
				"lootCouncil",
				"admin",
				"manager"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"hidden": false,
			"id": "select1466534506",
			"maxSelect": 3,
			"name": "role",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"member",
				"lootCouncil",
				"admin"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
