package bot

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func adminKey(userID int) string {
	return fmt.Sprintf("admin:%d", userID)
}

func findCreator(admins []tgbotapi.ChatMember) *tgbotapi.User {
	for _, admin := range admins {
		if admin.IsCreator() {
			return admin.User
		}
	}
	return nil
}
