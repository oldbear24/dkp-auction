package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
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
					"hidden": false,
					"id": "bool3670437562",
					"name": "nameSynchronization",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
				},
				{
					"hidden": false,
					"id": "select422795190",
					"maxSelect": 1,
					"name": "synchronizationType",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "select",
					"values": [
						"tlgh"
					]
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text2788972646",
					"max": 0,
					"min": 0,
					"name": "synchronizationUrl",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text832374637",
					"max": 0,
					"min": 0,
					"name": "synchronizationClient",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text3142868301",
					"max": 0,
					"min": 0,
					"name": "synchronizationPassword",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				}
			],
			"id": "pbc_2769025244",
			"indexes": [],
			"listRule": null,
			"name": "settings",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)

	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_2769025244")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
