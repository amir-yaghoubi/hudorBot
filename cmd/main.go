package main

import (
	"errors"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	bot "github.com/amir-yaghoobi/hudorBot"
	"github.com/spf13/viper"

	"github.com/sirupsen/logrus"
)

func loadConfigurations() (*bot.HudorConfig, error) {
	v := viper.New()

	v.SetConfigName("config")

	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.hudor")
	v.AddConfigPath("/etc/hudor/")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	logrus.Infof("configuration file: %q", v.ConfigFileUsed())

	config := &bot.HudorConfig{}

	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	config.Clean()

	if len(config.TelegramToken) == 0 {
		return nil, errors.New("please provide valid telegramToken in config file")
	}

	return config, nil
}

func main() {
	config, err := loadConfigurations()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("%#v\n", *config)

	tgBot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		logrus.Fatal(err)
	}
	tgBot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tgBot.GetUpdatesChan(u)
	if err != nil {
		logrus.Fatal(err)
	}

	startTime := time.Now()

	logrus.Infof("bot %q started at %s\n", tgBot.Self.UserName, startTime.Format(time.RFC3339))
	service := bot.NewBotService(config, tgBot)
	service.Start(updates)
}
