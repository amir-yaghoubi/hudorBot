package bot

import (
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
		log.Info("sending group informations to chat")
		msg := groupInformations(message.Chat.ID, settings)
		msg.ReplyToMessageID = message.MessageID
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

	switch cmd {
	case "hudor":
		c.hudor(&message)
		break
	case "settings":
		c.settings(&message)
		break
	case "groups":
		break
	case "help":
		break
	default:
		log.Info("unknown command")
	}
}
