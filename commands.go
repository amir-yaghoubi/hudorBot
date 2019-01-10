package hudorbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

func newCommandHandler(conn *redis.Client, bot *tgbotapi.BotAPI) *commandHandler {
	return &commandHandler{
		redis: conn,
		bot:   bot,
	}
}

type commandHandler struct {
	redis *redis.Client
	bot   *tgbotapi.BotAPI
}

func (c *commandHandler) hudor(message *tgbotapi.Message) {
	log := logrus.WithFields(logrus.Fields{
		"cmd":  "hudor",
		"from": message.From.ID,
		"chat": message.Chat.ID,
	})

	if message.Chat.IsSuperGroup() {
		creator, err := groupCreator(c.redis, message.Chat.ID)
		if err != nil {
			log.Fatal(err)
		}

		if message.From.ID != creator {
			log.Info("this hudor command came from regular user. [>>skip]")
			msg := hudorCanOnlySendFromCreator(message.Chat.ID)
			msg.ReplyToMessageID = message.MessageID

			if _, err := c.bot.Send(msg); err != nil {
				log.Error(err)
			}
			return
		}

		admins, err := c.bot.GetChatAdministrators(message.Chat.ChatConfig())
		if err != nil {
			log.Error(err)
			errorMessage := errorHappenedDuringProcess(message.Chat.ID)
			if _, err = c.bot.Send(errorMessage); err != nil {
				log.Error(err)
			}
			return
		}

		isActive, err := isGroupActive(c.redis, message.Chat.ID)
		if err != nil {
			log.Fatal(err)
		}

		for _, admin := range admins {
			if admin.User.ID == c.bot.Self.ID && admin.IsAdministrator() {

				if !admin.CanRestrictMembers {
					log.Info("cannot activate this group, Ban users permission is missing")
					msg := errorPermissionRequired(message.Chat.ID)
					msg.ReplyToMessageID = message.MessageID
					if _, err := c.bot.Send(msg); err != nil {
						log.Error(err)
					}

					if isActive { // deactivate group if it's active when permission is missing
						log.Info("group is active, perform deactivating group action")
						if err := changeGroupActiveStatus(c.redis, message.Chat.ID, false); err != nil {
							log.Fatal(err)
						}
					}
					return
				}

				if !isActive {
					log.Info("everything is good to go, activating bot in this group")
					if err := changeGroupActiveStatus(c.redis, message.Chat.ID, true); err != nil {
						log.Fatal(err)
					}
					msg := hudorActivated(message.Chat.ID)
					msg.ReplyToMessageID = message.MessageID
					if _, err := c.bot.Send(msg); err != nil {
						log.Error(err)
					}
				} else {
					log.Info("bot already activated")
					msg := hodurAlreadyIsActive(message.Chat.ID)
					msg.ReplyToMessageID = message.MessageID
					if _, err := c.bot.Send(msg); err != nil {
						log.Error(err)
					}
				}

				return
			}
		}

		log.Info("we cannot activate this group, our bot is not an admin")
		msg := errorBotIsNotAdmin(message.Chat.ID)
		msg.ReplyToMessageID = message.MessageID
		if _, err = c.bot.Send(msg); err != nil {
			log.Error(err)
		}

		if isActive { // deactive group if it's active and bot is not an administrator
			log.Info("group is active, perform deactivating group action")
			if err := changeGroupActiveStatus(c.redis, message.Chat.ID, false); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		msg := hodurOnlyActiveInSuperGroups(message.Chat.ID)
		msg.ReplyToMessageID = message.MessageID
		if _, err := c.bot.Send(msg); err != nil {
			log.Error(err)
		}
	}
}

func (c *commandHandler) settings(message *tgbotapi.Message) {
	log := logrus.WithFields(logrus.Fields{
		"cmd":  "settings",
		"from": message.From.ID,
		"chat": message.Chat.ID,
	})

	if message.Chat.IsSuperGroup() {
		settings, err := findGroupByID(c.redis, message.Chat.ID)
		if err != nil {
			log.Fatal(err)
		} else if settings == nil {
			log.Warn("this group does not exist in redis")
		}

		var whitelistBots []string
		if settings != nil {
			wlKey := whiteListKey(message.Chat.ID)
			whitelistBots, err = c.redis.SMembers(wlKey).Result()
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Info("sending group informations to chat")
		msg := groupInformations(message.Chat.ID, settings, whitelistBots)
		msg.ReplyToMessageID = message.MessageID
		if _, err := c.bot.Send(msg); err != nil {
			log.Error(err)
		}
		return
	}

	if message.Chat.IsPrivate() {
		state, err := getState(c.redis, message.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		if state.IsSelection() {
			keyboard, err := groupSelectionsKeyboard(c.redis, state.Page, message.From.ID)
			if err != nil {
				log.Fatal(err)
			}

			msg := selectGroupState(message.Chat.ID, keyboard)
			if _, err := c.bot.Send(msg); err != nil {
				log.Error(err)
			}
			return
		}

		if state.IsSettingsOrAbove() {
			if !state.IsSettings() {
				err := setStateToSettings(c.redis, message.From.ID, state.GroupID)
				if err != nil {
					log.Fatal(err)
				}
			}
			settings, err := findGroupByID(c.redis, state.GroupID)
			if err != nil {
				log.Fatal(err)
			}

			keyboard := createKeyboardForSettings(settings)
			msg := settingsState(message.Chat.ID, settings, &keyboard)
			if _, err := c.bot.Send(msg); err != nil {
				log.Error(err)
			}
		}
		return
	}
}

func (c *commandHandler) groups(message *tgbotapi.Message) {
	log := logrus.WithFields(logrus.Fields{
		"cmd":  "groups",
		"from": message.From.ID,
		"chat": message.Chat.ID,
	})

	if message.Chat.IsPrivate() {

		state, err := getState(c.redis, message.From.ID)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("user present in state: %s", state.ID)

		if !state.IsSelection() {
			log.Info("reseting state to selection")
			if newState, err := setStateToSelection(c.redis, message.From.ID); err != nil {
				log.Fatal(err)
			} else {
				state = newState
			}
		}

		keyboard, err := groupSelectionsKeyboard(c.redis, state.Page, message.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		msg := selectGroupState(message.Chat.ID, keyboard)
		if _, err := c.bot.Send(msg); err != nil {
			log.Error(err)
		}
	}
}

func (c *commandHandler) HandleAnswers(message tgbotapi.Message) {
	log := logrus.WithFields(logrus.Fields{
		"from": message.From.ID,
		"chat": message.Chat.ID,
		"text": message.Text,
	})

	if message.Chat.IsPrivate() {
		state, err := getState(c.redis, message.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		if state.IsSetLimit() {
			log.Infof("possible setLimit answer")

			var msg tgbotapi.MessageConfig

			limit, err := strconv.ParseInt(message.Text, 10, 64)
			if err != nil || limit < 0 || limit > 10 {
				log.Info("user entered invalid limit")
				msg = invalidWarnLimit(message.Chat.ID)
			} else {
				log.Infof("new limit value is valid, value: %d", limit)

				log.Info("authorizing admin")

				aKey := adminKey(message.From.ID)
				isValid, err := c.redis.SIsMember(aKey, state.GroupID).Result()
				if err != nil {
					log.Fatal(err)
				}

				if !isValid {
					log.Infof("this user is not admin of group: %d anymore", state.GroupID)
					msg = userIsNoLongerAdmin(message.Chat.ID)
				} else {
					log.Info("admin is authorized")
					if err := changeGroupWarnLimit(c.redis, state.GroupID, limit); err != nil {
						log.Fatal(err)
					}
					if err := setStateToSettings(c.redis, message.From.ID, state.GroupID); err != nil {
						log.Fatal(err)
					}

					msg = warnLimitChanged(message.Chat.ID, limit)
				}
			}

			if _, err := c.bot.Send(msg); err != nil {
				log.Error(err)
			}

			return
		}
	}
}

func (c *commandHandler) Handle(message tgbotapi.Message) {
	cmd := message.Command()
	log := logrus.WithFields(logrus.Fields{
		"cmd":  cmd,
		"from": message.From.ID,
		"chat": message.Chat.ID,
	})

	command := message.CommandWithAt()
	if strings.IndexRune(command, '@') > -1 && strings.Index(command, c.bot.Self.UserName) < 0 {
		log.Info("skip command, it's not direct command to us")
		return
	}

	switch cmd {
	case "hudor":
		c.hudor(&message)
		break
	case "settings":
		c.settings(&message)
		break
	case "groups":
		c.groups(&message)
		break
	case "help":
		break
	default:
		log.Info("unknown command")
	}
}

func (c *commandHandler) pageCallback(callback *tgbotapi.CallbackQuery, pageString string) {
	log := logrus.WithFields(logrus.Fields{
		"chat":     callback.From.ID,
		"page":     pageString,
		"callback": "page",
	})

	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	err = setStatePage(c.redis, callback.From.ID, int(page))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("user state page changed")

	keyboard, err := groupSelectionsKeyboard(c.redis, int(page), callback.From.ID)
	if err != nil {
		log.Fatal(err)
	}
	updatedKeyboard := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      callback.Message.Chat.ID,
			MessageID:   callback.Message.MessageID,
			ReplyMarkup: keyboard,
		},
	}
	if _, err := c.bot.Send(updatedKeyboard); err != nil {
		log.Error(err)
	} else {
		log.Info("callback message updated with new paged groups")
	}

	text := fmt.Sprintf("صفحه %d", page)
	response := tgbotapi.NewCallback(callback.ID, text)
	if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
		log.Error(err)
	} else {
		log.Info("callback process finished")
	}
}

func (c *commandHandler) selectCallback(callback *tgbotapi.CallbackQuery, groupID string) {
	log := logrus.WithFields(logrus.Fields{
		"chat":     callback.From.ID,
		"groupID":  groupID,
		"callback": "select",
	})

	state, err := getState(c.redis, callback.From.ID)
	if err != nil {
		log.Fatal(err)
	}

	if !state.IsSelection() {
		log.Info("skip processing callback, user is not in selection state")
		response := tgbotapi.NewCallback(callback.ID, "برای تغییر گروه از دکمه بازگشت استفاده کنید.")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	aKey := adminKey(callback.From.ID)
	isValid, err := c.redis.SIsMember(aKey, groupID).Result()
	if err != nil {
		log.Fatal(err)
	}

	if !isValid {
		log.Warn("attempt to select group that's no longer related to this user")
		response := tgbotapi.NewCallback(callback.ID, "گروه مورد نظر یافت نشد!")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	gID, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		gID = 0
	}

	settings, err := findGroupByID(c.redis, gID)
	if err != nil {
		log.Fatal(err)
	} else if settings == nil {
		log.Warn("group does not exists")
		response := tgbotapi.NewCallback(callback.ID, "گروه مورد نظر یافت نشد!")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	err = setStateToSettings(c.redis, callback.From.ID, gID)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("successfully moved user state from selection to settings")

	keyboard := createKeyboardForSettings(settings)
	msg := settingsState(callback.Message.Chat.ID, settings, &keyboard)
	editMsg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      msg.ChatID,
			MessageID:   callback.Message.MessageID,
			ReplyMarkup: &keyboard,
		},
		Text: msg.Text,
	}
	if _, err := c.bot.Send(editMsg); err != nil {
		log.Error(err)
	}

	text := fmt.Sprintf("گروه %s انتخاب شد", settings.Title)
	response := tgbotapi.NewCallback(callback.ID, text)
	if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
		log.Error(err)
	} else {
		log.Info("callback process finished")
	}
}

func (c *commandHandler) navigateCallback(callback *tgbotapi.CallbackQuery, to string) {
	log := logrus.WithFields(logrus.Fields{
		"chat":       callback.From.ID,
		"navigateTo": to,
		"callback":   "navigate",
	})

	state, err := getState(c.redis, callback.From.ID)
	if err != nil {
		log.Fatal(err)
	}

	if !state.canSetStateTo(to) {
		log.Warnf("cannot navigate back from %s to %s", state.ID, to)
		response := tgbotapi.NewCallback(callback.ID, "امکان بازگشت وجود ندارد")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	if to == "selection" {
		state, err := setStateToSelection(c.redis, callback.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("changed state to selection")
		keyboard, err := groupSelectionsKeyboard(c.redis, state.Page, callback.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		msg := selectGroupState(callback.Message.Chat.ID, keyboard)
		editMsgCfg := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      msg.ChatID,
				MessageID:   callback.Message.MessageID,
				ReplyMarkup: keyboard,
			},
			Text: msg.Text,
		}

		if _, err := c.bot.Send(editMsgCfg); err != nil {
			log.Error(err)
		}

		response := tgbotapi.NewCallback(callback.ID, "بازگشت به "+state.StateFa())
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}

	}

	if to == "setLimit" {
		_, err := setStateToSetLimit(c.redis, callback.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		msg := pleaseProvideLimit(callback.Message.Chat.ID)
		editMsgCfg := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:    msg.ChatID,
				MessageID: callback.Message.MessageID,
			},
			Text: msg.Text,
		}

		if _, err := c.bot.Send(editMsgCfg); err != nil {
			log.Error(err)
		}

		response := tgbotapi.NewCallback(callback.ID, "")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
	}
}

func (c *commandHandler) changeActiveStatus(callback *tgbotapi.CallbackQuery, status string) {
	log := logrus.WithFields(logrus.Fields{
		"chat":     callback.From.ID,
		"changeTo": status,
		"callback": "gActive",
	})

	state, err := getState(c.redis, callback.From.ID)
	if err != nil {
		log.Fatal(err)
	}

	if !state.IsSettings() {
		log.Warn("cannot modify group active status when state is not settings")
		response := tgbotapi.NewCallback(callback.ID, "امکان تغییر وضعیت گروه وجود ندارد")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	isActive, err := strconv.ParseBool(status)
	if err != nil {
		isActive = false
	}

	var isBotCanOperate = false

	// if group is not active we should check for bot permissions
	if isActive {
		chatConfig := tgbotapi.ChatConfig{
			ChatID: state.GroupID,
		}
		admins, err := c.bot.GetChatAdministrators(chatConfig)
		if err != nil {
			log.Error(err)
			return
		}

		for _, admin := range admins {
			if admin.User.ID == c.bot.Self.ID && admin.IsAdministrator() && admin.CanRestrictMembers {
				isBotCanOperate = true
			}
		}
	} else {
		isBotCanOperate = true
	}

	if !isBotCanOperate {
		text := fmt.Sprintf("🔏 نیاز به دسترسی مورد نیاز برای هودور می‌باشد. 🔏")
		response := tgbotapi.NewCallback(callback.ID, text)
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	if err := changeGroupActiveStatus(c.redis, state.GroupID, isActive); err != nil {
		log.Fatal(err)
	}

	log.Infof("isActive successfully changed to %t", isActive)

	group, err := findGroupByID(c.redis, state.GroupID)
	if err != nil {
		log.Fatal(err)
	} else if group == nil {
		log.Error("group does not exist")
		return
	}

	keyboard := createKeyboardForSettings(group)
	editMsgCfg := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      callback.Message.Chat.ID,
			MessageID:   callback.Message.MessageID,
			ReplyMarkup: &keyboard,
		},
	}

	if _, err := c.bot.Send(editMsgCfg); err != nil {
		log.Error(err)
	} else {
		log.Info("settings message updated")
	}

	text := fmt.Sprintf("وضعیت: %s", group.IsActiveFa())
	response := tgbotapi.NewCallback(callback.ID, text)
	if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
		log.Error(err)
	} else {
		log.Info("callback process finished")
	}
}

func (c *commandHandler) changeShowWarn(callback *tgbotapi.CallbackQuery, showWarn string) {
	log := logrus.WithFields(logrus.Fields{
		"chat":     callback.From.ID,
		"changeTo": showWarn,
		"callback": "gShowWarn",
	})

	state, err := getState(c.redis, callback.From.ID)
	if err != nil {
		log.Fatal(err)
	}

	if !state.IsSettings() {
		log.Warn("cannot modify group showWarn status when state is not settings")
		response := tgbotapi.NewCallback(callback.ID, "امکان تغییر وضعیت نمایش اخطار وجود ندارد")
		if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
			log.Error(err)
		} else {
			log.Info("callback process finished")
		}
		return
	}

	show, err := strconv.ParseBool(showWarn)
	if err != nil {
		show = false
	}

	if err := changeGroupShowWarnStatus(c.redis, state.GroupID, show); err != nil {
		log.Fatal(err)
	}

	log.Infof("showWarn successfully changed to %t", show)

	group, err := findGroupByID(c.redis, state.GroupID)
	if err != nil {
		log.Fatal(err)
	} else if group == nil {
		log.Error("group does not exist")
		return
	}

	keyboard := createKeyboardForSettings(group)
	editMsgCfg := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      callback.Message.Chat.ID,
			MessageID:   callback.Message.MessageID,
			ReplyMarkup: &keyboard,
		},
	}

	if _, err := c.bot.Send(editMsgCfg); err != nil {
		log.Error(err)
	} else {
		log.Info("settings message updated")
	}

	text := fmt.Sprintf("نمایش اخطار: %s", group.ShowWarnFa())
	response := tgbotapi.NewCallback(callback.ID, text)
	if _, err = c.bot.AnswerCallbackQuery(response); err != nil {
		log.Error(err)
	} else {
		log.Info("callback process finished")
	}
}

func (c *commandHandler) HandleCallback(callback tgbotapi.CallbackQuery) {
	log := logrus.WithFields(logrus.Fields{
		"chat": callback.From.ID,
		"data": callback.Data,
	})
	log.Info("received callback query")

	data := strings.Split(callback.Data, ":")
	if len(data) != 2 {
		log.Error("invalid data format")
		return
	}

	switch data[0] {
	case "page":
		log.Info("callback routed to pageCallback")
		c.pageCallback(&callback, data[1])
		break
	case "select":
		log.Info("callback routed to selectCallback")
		c.selectCallback(&callback, data[1])
		break
	case "navigate":
		log.Info("callback routed to navigateCallback")
		c.navigateCallback(&callback, data[1])
		break
	case "gActive":
		log.Info("callback routed to changeActiveStatus")
		c.changeActiveStatus(&callback, data[1])
		break
	case "gShowWarn":
		log.Info("callback routed to changeShowWarn")
		c.changeShowWarn(&callback, data[1])
		break
	}
}
