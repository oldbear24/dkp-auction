package main

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

func RegisterApiRoutes(se *core.ServeEvent) {
	se.Router.POST("/api/app/change-tokens", changeTokensApi).Bind(validateApiTokenMiddleware()).Bind(setUserAuthMiddleware())

}

func changeTokensApi(e *core.RequestEvent) error {

	return chaneUsersAmount(e)
}

func validateApiToken(app core.App, token string) (string, error) {
	record, err := app.FindFirstRecordByData("apiKeys", "apiKey", token)
	if err != nil {
		return "", err
	}
	return record.Id, nil
}

func validateApiTokenMiddleware() *hook.Handler[*core.RequestEvent] {
	return &hook.Handler[*core.RequestEvent]{
		Id: "validate-api-token",
		Func: func(e *core.RequestEvent) error {
			e.App.Logger().Debug("Validating API token for request", "path", e.Request.URL.Path)
			token := e.Request.Header.Get("api-token")
			if token == "" {
				return e.UnauthorizedError("API toke was not provided.", nil)
			}
			tokenID, err := validateApiToken(e.App, token)
			if err != nil {
				return e.UnauthorizedError(err.Error(), nil)
			}
			e.App.Logger().Debug("API token validated", "path", e.Request.URL.Path, "tokenID", tokenID)

			return e.Next()
		},
		Priority: 10,
	}
}

func setUserAuthMiddleware() *hook.Handler[*core.RequestEvent] {
	return &hook.Handler[*core.RequestEvent]{
		Id: "set-auth-user",
		Func: func(e *core.RequestEvent) error {
			discordId := e.Request.Header.Get("discord-user-id")
			userRecord, err := e.App.FindFirstRecordByData("users", "discordId", discordId)
			if err != nil {
				return e.UnauthorizedError("User not found", err)
			}
			e.Auth = userRecord
			e.App.Logger().Debug("User set for request", "path", e.Request.URL.Path, "user", userRecord)
			return e.Next()
		},
		Priority: 20,
	}
}
