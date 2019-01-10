package hudorbot

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func warnUser(chatID int64, current int64, limit int64) tgbotapi.MessageConfig {
	text := fmt.Sprintf("⚠️ اخطار %d از %d ⚠️\nامکان افزدون ربات تنها برای ادمین گروه فعال می‌باشد.", current, limit)
	return tgbotapi.NewMessage(chatID, text)
}

func superGroupIntroduction(chatID int64) tgbotapi.MessageConfig {
	text := `سلام 👋
	من هودورم، وظیفه من محافظت 🛡 از گروه‌ها در برابر ربات‌های اسپمر هست.
	برای این که هودور آماده رزم ⚔️ بشه، نیاز به اجازه شما داره.
	هودور رو ادمین گروه کنین و دسترسی *Ban users* رو بهش بدین بعد با دستور /hudor صداش بزنین.
	راستی هودور فقط از سازنده گروه دستور میگیره`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}

func botAddedToWhitelist(chatID int64, messageID int, username string) tgbotapi.MessageConfig {
	text := fmt.Sprintf("🤖 بات @%s به لیست ربات‌های مجاز به فعالیت افزوده شد. ✅", username)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableNotification = true
	msg.ReplyToMessageID = messageID

	return msg
}

func botCannotOperateWithoutCreator(chatID int64) tgbotapi.MessageConfig {
	text := `⛔️ فعالیت در این گروه امکان پذیر نمی‌باشد. ⛔️
	دلیل: سازنده گروه باید در گروه حضور داشته باشد.`

	return tgbotapi.NewMessage(chatID, text)
}

func errorHappenedDuringProcess(chatID int64) tgbotapi.MessageConfig {
	text := `❌ اوه شت 😱😱 
	متاسفانه خطایی رخ داده و نتونستم درخواست رو پردازش کنم.`
	return tgbotapi.NewMessage(chatID, text)
}

func hudorCanOnlySendFromCreator(chatID int64) tgbotapi.MessageConfig {
	text := "🔐 هودور فقط از سازنده گروه حرف شنوی داره 🔐"
	return tgbotapi.NewMessage(chatID, text)
}

func errorPermissionRequired(chatID int64) tgbotapi.MessageConfig {
	text := "⛔️ هودور نیاز به اجازه شما داره، دسترسی *Ban users* رو به هودور بدین تا بتونه شروع کنه! ⛔️"
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	return msg
}

func errorBotIsNotAdmin(chatID int64) tgbotapi.MessageConfig {
	text := `⚠️ دست و پای هودور بسته هست ⛓، هودور رو ادمین گروه کنین و دسترسی *Ban users* رو بهش بدین تا از غل و زنجیر آزاد بشه ⚠️`
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	return msg
}

func hudorActivated(chatID int64) tgbotapi.MessageConfig {
	text := `❇️ هودور با موفقیت فعال شد ❇️
	💎 نکات 💎
	1️⃣ جهت نمایش تنظیمات گروه دستور /settings را ارسال نمایید
	2️⃣ سازنده گروه می‌تواند تنظیمات گروه را از طریق چت خصوصی تغییر دهد
	
	از گروه بدون ربات‌های مزاحم لذت ببرین 😎`
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	return msg
}

func hodurAlreadyIsActive(chatID int64) tgbotapi.MessageConfig {
	text := "🛡⚔️ هودور فعال می‌باشد ⚔️🛡"
	return tgbotapi.NewMessage(chatID, text)
}

func hodurOnlyActiveInSuperGroups(chatID int64) tgbotapi.MessageConfig {
	text := `من فقط می‌تونم توی سوپرگروه ها فعالیت کنم ☹️😞
	اگه می‌خوای بیشتر راجبم بدونی دستور /help رو بزن تا برات بگم`
	return tgbotapi.NewMessage(chatID, text)
}

func groupInformations(chatID int64, group *groupSettings, bots []string) tgbotapi.MessageConfig {
	var text string
	if group == nil {
		text = "⚠️ در حال حاضر اطلاعاتی از این گروه در دست نیست ⚠️"
	} else {
		var whitelistedBots string

		if len(bots) == 0 {
			whitelistedBots = "🔘 هیچ رباتی مجاز به فعالیت نیست 🔘"
		} else {
			var botLimit int
			if len(bots) > 20 {
				botLimit = 20
			} else {
				botLimit = len(bots)
			}

			for _, bot := range bots[:botLimit] {
				whitelistedBots += "▪️ @" + bot + "\n"
			}

			if len(bots) > 20 {
				whitelistedBots += fmt.Sprintf("و %d بات دیگر", len(bots)-20)
			}
		}

		text = fmt.Sprintf(`🔹 گروه: %s
🔹 وضعیت فعالیت: %s
🔹 نمایش اخطار: %s
🔹 تعداد اخطارها قبل از حذف کاربر: %d بار
🔹 بات‌های مجاز به فعالیت:
%s`, group.Title, group.IsActiveFa(), group.ShowWarnFa(), group.Limit, whitelistedBots)
	}

	return tgbotapi.NewMessage(chatID, text)
}

func selectGroupState(chatID int64, keyboard *tgbotapi.InlineKeyboardMarkup) tgbotapi.MessageConfig {
	text := "💢 گروه مورد نظر خود را انتخاب کنید 💢"

	msg := tgbotapi.NewMessage(chatID, text)

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	return msg
}

func settingsState(chatID int64, settings *groupSettings, keyboard *tgbotapi.InlineKeyboardMarkup) tgbotapi.MessageConfig {
	text := fmt.Sprintf(`گروه: 🔰 %s 🔰
	تعداد اخطارها قبل از بن کاربر: %d بار`, settings.Title, settings.Limit)

	msg := tgbotapi.NewMessage(chatID, text)

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	return msg
}

func pleaseProvideLimit(chatID int64) tgbotapi.MessageConfig {
	text := "لطفا یک عدد بین ۱ تا ۱۰ را وارد کنید."
	return tgbotapi.NewMessage(chatID, text)
}

func invalidWarnLimit(chatID int64) tgbotapi.MessageConfig {
	text := `⚠️ مقدار وارد شده صحیح نمی‌باشد ⚠️
لطفا یک عدد بین ۱ تا ۱۰ وارد نمایید.`

	return tgbotapi.NewMessage(chatID, text)
}

func warnLimitChanged(chatID int64, newLimit int64) tgbotapi.MessageConfig {
	text := fmt.Sprintf("تعداد اخطار‌های گروه مورد نظر به %d تغییر پیدا کرد. ✅", newLimit)
	return tgbotapi.NewMessage(chatID, text)
}

func userIsNoLongerAdmin(chatID int64) tgbotapi.MessageConfig {
	text := "🚫 متاسفانه شما دیگر ادمین گروه انتخابی نیستید! 🚫"
	return tgbotapi.NewMessage(chatID, text)
}
