package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

func NewCommandHandler(conn *redis.Client, bot *tgbotapi.BotAPI) *commandHandler {
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
			keyboard, err := groupSelectionsKeyboard(c.redis, state, message.From.ID)
			if err != nil {
				log.Fatal(err)
			}

			msg := selectGroupState(message.Chat.ID, keyboard)
			if _, err := c.bot.Send(msg); err != nil {
				log.Error(err)
			}
			return
		}

		if state.IsSettings() {
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

		// MOCK DATA
		// for i := 1; i < 50; i++ {
		// 	k := groupKey(int64(i))
		// 	g := groupSettings{
		// 		Creator:  i * 200,
		// 		IsActive: false,
		// 		ShowWarn: false,
		// 		Limit:    99,
		// 		Title:    fmt.Sprintf("گروه شماره %d", i),
		// 	}
		// 	c.redis.HMSet(k, g.Map())
		// 	k = adminKey(98299621)
		// 	c.redis.SAdd(k, i)
		// }
		// MOCK DATA

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

		keyboard, err := groupSelectionsKeyboard(c.redis, state, message.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		msg := selectGroupState(message.Chat.ID, keyboard)
		if _, err := c.bot.Send(msg); err != nil {
			log.Error(err)
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

func (c *commandHandler) pageCallback(callback *tgbotapi.CallbackQuery, page string) {
	log := logrus.WithFields(logrus.Fields{
		"chat":     callback.From.ID,
		"page":     page,
		"callback": "page",
	})

	key := stateKey(callback.From.ID)
	err := c.redis.HSet(key, "page", page).Err()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("user state page changed")
	p, _ := strconv.ParseInt(page, 10, 64)
	if p < 1 {
		p = 1
	}
	groups, pageCount, err := adminGroups(c.redis, callback.From.ID, int(p))
	if err != nil {
		log.Fatal(err)
	}

	prevPage := p - 1
	nextPage := p + 1
	if int(p) >= pageCount {
		nextPage = -1
	}
	if p == 1 {
		prevPage = -1
	}

	keyboard := createKeyboardForGroupSelections(groups, int(prevPage), int(nextPage))
	updatedKeyboard := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      callback.Message.Chat.ID,
			MessageID:   callback.Message.MessageID,
			ReplyMarkup: &keyboard,
		},
	}
	if _, err := c.bot.Send(updatedKeyboard); err != nil {
		log.Error(err)
	} else {
		log.Info("callback message updated with new paged groups")
	}

	text := fmt.Sprintf("صفحه %s", page)
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

	if !state.canBackTo(to) {
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
		keyboard, err := groupSelectionsKeyboard(c.redis, state, callback.From.ID)
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
	}
}
