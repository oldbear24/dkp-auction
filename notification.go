package main

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// notificationChannel buffers notification payloads for async processing.
var notificationChannel = make(chan notification, 1024)

// startNotificationWorker consumes queued notifications and persists them.
func startNotificationWorker(app core.App) {
	for {
		n := <-notificationChannel
		userIds := []string{}
		if n.IsRole {
			users, err := app.FindRecordsByFilter("users", "role:each ?= {:role}", "", 0, 0, dbx.Params{"role": n.UserIdOrRole})
			if err != nil {
				app.Logger().Error("Error finding users by role", "error", err)
				continue
			}
			for _, user := range users {
				userIds = append(userIds, user.Id)
			}
		} else {
			userIds = append(userIds, n.UserIdOrRole)
		}
		for _, userId := range userIds {
			if err := createNotification(app, userId, n.Message); err != nil {
				app.Logger().Error("Error creating notification", "error", err)
			}
		}
	}
}

// notifyUser enqueues a notification for a specific user.
func notifyUser(userId string, message string) {
	notificationChannel <- notification{UserIdOrRole: userId, Message: message, IsRole: false}
}

// notifyRole enqueues a notification for all users with a role.
func notifyRole(role string, message string) {
	notificationChannel <- notification{UserIdOrRole: role, Message: message, IsRole: true}
}

// createNotification writes a notification record for the given user.
func createNotification(app core.App, userId string, message string) error {
	coll, err := app.FindCachedCollectionByNameOrId("notifications")
	if err != nil {
		return err
	}
	record := core.NewRecord(coll)
	record.Set("user", userId)
	record.Set("text", message)
	if err := app.Save(record); err != nil {
		return err
	}
	return nil
}

type notification struct {
	UserIdOrRole string
	Message      string
	IsRole       bool
}
