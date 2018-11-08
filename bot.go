package bot

import (
	"strconv"

	"github.com/go-redis/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

func NewBotService(redis *redis.Client, bot *tgbotapi.BotAPI) *BotService {
	return &BotService{
		redis: redis,
		bot:   bot,
	}
}

type BotService struct {
	redis *redis.Client
	bot   *tgbotapi.BotAPI
}

// initGroup will set default settings for group
// and add this group to creator's group list
func (s *BotService) initGroup(message tgbotapi.Message) *groupSettings {
	admins, err := s.bot.GetChatAdministrators(message.Chat.ChatConfig())
	if err != nil {
		logrus.Warnf("cannot retrieve chat: %d administrators! err: %v\n", message.Chat.ID, err)
		return nil
	}

	introduction := superGroupIntroduction(message.Chat.ID)
	_, err = s.bot.Send(introduction)
	if err != nil {
		logrus.Warnf("cannot send introduction message into chat: %d, error: %v\n", message.Chat.ID, err)
	}

	creator := findCreator(admins)
	if creator == nil {
		logrus.Warnf("chat: %d does not have any creators 🤤!!\n", message.Chat.ID)
		return nil
	}

	gpKey := groupKey(message.Chat.ID)
	adminKey := adminKey(creator.ID)

	settings := groupSettings{
		IsActive:    false,
		ShowWarn:    true,
		Limit:       3,
		Creator:     creator.ID,
		Title:       message.Chat.Title,
		Description: message.Chat.Description,
	}

	pipe := s.redis.Pipeline()
	pipe.SAdd(adminKey, message.Chat.ID)
	pipe.HMSet(gpKey, settings.Map())
	_, err = pipe.Exec()
	if err != nil {
		logrus.Fatal(err)
	}

	return &settings
}

func (s *BotService) kickUser(chatID int64, userID int) (Ok bool, err error) {
	kickCfg := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			UserID: userID,
			ChatID: chatID,
		},
		UntilDate: 400,
	}
	response, err := s.bot.KickChatMember(kickCfg)
	if response.ErrorCode == 400 {
		return false, nil
	}
	return response.Ok, err
}

func (s *BotService) deleteMessage(chatID int64, messageID int) (ok bool, err error) {
	msg := tgbotapi.NewDeleteMessage(chatID, messageID)
	response, err := s.bot.DeleteMessage(msg)
	if response.ErrorCode == 400 {
		return false, nil
	}

	return response.Ok, err
}

func (s *BotService) processNewUsers(message tgbotapi.Message, users []tgbotapi.User) {
	log := logrus.WithFields(logrus.Fields{
		"chat": message.Chat.ID,
		"user": message.From.ID,
	})

	wlKey := whiteListKey(message.Chat.ID)

	groupSettings, err := findGroupByID(s.redis, message.Chat.ID)
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		if user.ID == s.bot.Self.ID {
			settings := s.initGroup(message)
			if settings != nil {
				groupSettings = settings
				log.Info("initilized group with default settings")
			}
			continue
		}
	}

	for _, user := range users {
		if user.ID == s.bot.Self.ID {
			continue
		}
		log := logrus.WithFields(logrus.Fields{
			"chat": message.Chat.ID,
			"user": message.From.ID,
			"bot":  user.ID,
		})

		if message.From.ID == groupSettings.Creator {
			added, err := s.redis.SAdd(wlKey, user.ID).Result()
			if err != nil {
				log.Fatal(err)
			}
			if added > 0 {
				log.Info("bot added to whitelist (added by creator)")

				msg := botAddedToWhitelist(message.Chat.ID, message.MessageID, user.UserName)
				_, err := s.bot.Send(msg)
				if err != nil {
					log.Warn("cannot send the message into group")
				}
			}
			continue
		}

		if !groupSettings.IsActive {
			continue
		}

		isApproved, err := s.redis.SIsMember(wlKey, user.ID).Result()
		if err != nil {
			log.Fatal(err)
		}

		if isApproved {
			log.Info("whitelisted bot added to chat")
			continue
		}

		log.Info("spam bot detected, trying to remove it")
		ok, err := s.kickUser(message.Chat.ID, user.ID)
		if err != nil {
			log.Warn(err)
			continue
		}

		if !ok {
			log.Warn("cannot kick spammer bot! permission required")
			err := deactivateGroup(s.redis, message.Chat.ID)
			if err != nil {
				log.Fatal(err)
			}
			log.Info("deactived group")
			continue
		}

		log.Info("spammer bot successfully removed from chat")

		usrWarns, err := incrementMemberWarns(s.redis, message.Chat.ID, message.From.ID)
		if err != nil {
			log.Fatal(err)
		}

		if usrWarns >= groupSettings.Limit {
			log.Info("user reached to their warning limitations")
			ok, err := s.kickUser(message.Chat.ID, message.From.ID)
			if err != nil {
				log.Warn(err)
				continue
			}
			if !ok {
				log.Warn("cannot ban spammer user")
				err := deactivateGroup(s.redis, message.Chat.ID)
				if err != nil {
					log.Fatal(err)
				}
				log.Info("deactived group")
				continue
			}

			log.Info("banned the spammer user")

			warnKey := membersKey(message.Chat.ID, message.From.ID)
			if s.redis.Del(warnKey).Err() != nil {
				log.Fatal(err)
			}
		} else if groupSettings.ShowWarn {
			warnText := warnUser(message.Chat.ID, usrWarns, groupSettings.Limit)
			_, err := s.bot.Send(warnText)
			if err != nil {
				log.Warnf("cannot send message in supergroup! err: %s", err)
			}
		}
	}
}

func (s *BotService) processLeftUser(message tgbotapi.Message, leftChatMember tgbotapi.User) {
	log := logrus.WithFields(logrus.Fields{
		"chat":           message.Chat.ID,
		"leftChatMember": leftChatMember.ID,
	})

	// TODO clean up after supergroup creator lefted !!!!
	if s.bot.Self.ID == leftChatMember.ID {
		log.Info("our bot removed from group, starting clean up process")
		gpKey := groupKey(message.Chat.ID)
		wlKey := whiteListKey(message.Chat.ID)
		creatorStr, err := s.redis.HGet(gpKey, "creator").Result()
		if err != nil {
			log.Fatal(err)
		}

		pipe := s.redis.Pipeline()
		pipe.Del(gpKey)
		pipe.Del(wlKey)

		admin, err := strconv.ParseInt(creatorStr, 10, 64)
		if err == nil {
			adminKey := adminKey(int(admin))
			pipe.SRem(adminKey, message.Chat.ID)
		}

		if _, err := pipe.Exec(); err != nil {
			log.Fatal(err)
		} else {
			log.Info("group successfully cleaned up")
		}
	}

	if leftChatMember.IsBot {
		wlKey := whiteListKey(message.Chat.ID)
		_, err := s.redis.SRem(wlKey, leftChatMember.ID).Result()
		if err != nil {
			log.Fatal(err)
		}
		log.Info("bot removed from group. srem from whitelist if exists")
	}
}

func (s *BotService) processBotMessage(message tgbotapi.Message) {
	log := logrus.WithFields(logrus.Fields{
		"bot":     message.From.ID,
		"chat":    message.Chat.ID,
		"message": message.MessageID,
	})

	isActive, err := isGroupActive(s.redis, message.Chat.ID)
	if err != nil {
		log.Fatal(err)
	}

	if !isActive {
		log.Info("skip processing bot message due to inActive group")
		return
	}

	wlKey := whiteListKey(message.Chat.ID)
	isApproved, err := s.redis.SIsMember(wlKey, message.From.ID).Result()
	if err != nil {
		log.Fatal(err)
	}

	if !isApproved {
		log.Info("message from unauthorized bot detected")
		ok, err := s.kickUser(message.Chat.ID, message.From.ID)
		if err != nil {
			log.Error(err)
		} else if !ok {
			log.Warn("cannot kick spammer bot! permission required")
			err := deactivateGroup(s.redis, message.Chat.ID)
			if err != nil {
				log.Fatal(err)
			}
			log.Info("deactived group")
			return
		}

		log.Infof("unauthorized bot removed from group")

		ok, err = s.deleteMessage(message.Chat.ID, message.MessageID)
		if err != nil {
			log.Error(err)
		} else if !ok {
			log.Warn("cannot delete the message from group")
		} else {
			log.Info("deleted message from unauthorized bot")
		}
	}
}

func (s *BotService) Start(updates <-chan tgbotapi.Update) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsSuperGroup() {
			newChatMembers := update.Message.NewChatMembers
			if newChatMembers != nil {
				go s.processNewUsers(*update.Message, *newChatMembers)
			}

			leftChatMember := update.Message.LeftChatMember
			if leftChatMember != nil {
				go s.processLeftUser(*update.Message, *leftChatMember)
			}

			if update.Message.From.IsBot {
				go s.processBotMessage(*update.Message)
			}
		}

		if update.Message.Chat.IsPrivate() {
			// TODO process commands
		}

	}
}
