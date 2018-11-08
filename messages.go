package bot

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func warnUser(chatId int64, current int64, limit int64) tgbotapi.MessageConfig {
	text := fmt.Sprintf("⚠️ اخطار %d از %d ⚠️\nامکان افزدون ربات تنها برای ادمین گروه فعال می‌باشد.", current, limit)
	return tgbotapi.NewMessage(chatId, text)
}

func superGroupIntroduction(chatID int64) tgbotapi.MessageConfig {
	text := `سلام 👋
	به منظور شروع فعالیت بات، ابتدا بات را ادمین کرده و دسترسی حذف کاربر را به آن بدهید.

	⚒ تنظیمات پیشفرض بات:
	1️⃣ بن کاربر بعد از ۳ بار افزودن ربات
	2️⃣ حذف خودکار پیام‌های ارسالی توسط ربات‌ها: ✅ فعال
	3️⃣ ربات‌های مجاز به فعالیت: ⛔️ هیچکدام
	4️⃣ وضعیت فعالیت: 🔴 غیرفعال`

	return tgbotapi.NewMessage(chatID, text)
}

func botAddedToWhitelist(chatID int64, messageID int, username string) tgbotapi.MessageConfig {
	text := fmt.Sprintf(`🤖 بات @%s به لیست ربات‌های مجاز به فعالیت افزوده شد. ✅`, username)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableNotification = true
	msg.ReplyToMessageID = messageID

	return msg
}
