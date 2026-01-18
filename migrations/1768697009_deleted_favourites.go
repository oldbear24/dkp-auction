package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_2176316817")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	}, func(app core.App) error {
		jsonData := `{
			"createRule": "@request.auth.id=user",
			"deleteRule": "@request.auth.id=user",
			"fields": [
				{
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"cascadeDelete": true,
					"collectionId": "pbc_1337428601",
					"hidden": false,
					"id": "relation3739547027",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "auction",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": true,
					"collectionId": "_pb_users_auth_",
					"hidden": false,
					"id": "relation2375276105",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "user",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				}
			],
			"id": "pbc_2176316817",
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_Bsdsu3X1ig` + "`" + ` ON ` + "`" + `favourites` + "`" + ` (\n  ` + "`" + `user` + "`" + `,\n  ` + "`" + `auction` + "`" + `\n)"
			],
			"listRule": "",
			"name": "favourites",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": "@request.auth.id=user"
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
