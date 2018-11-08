package main

import (
	"os"
	"time"

	"github.com/amir-yaghoobi/telegramBotRemover"
	"github.com/go-redis/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

/*
	TODO Cleanup after detecting remove message

*/

// Add goimport on save 👍
func getEnv(env string) string {
	val := os.Getenv(env)
	if val == "" {
		logrus.Fatalf("you have to set %s enviroment!\n", env)
	}
	return val
}

func connectToRedis(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		DB:       0,
		Addr:     addr,
		Password: "",
	})
	return client
}

func main() {
	botToken := getEnv("TG_TOKEN")

	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logrus.Fatal(err)
	}
	tgBot.Debug = false

	rDB := connectToRedis("localhost:6379")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tgBot.GetUpdatesChan(u)
	if err != nil {
		logrus.Fatal(err)
	}

	startTime := time.Now()

	logrus.Infof("Bot %q started at %s\n", tgBot.Self.UserName, startTime.Format(time.RFC3339))
	service := bot.NewBotService(rDB, tgBot)
	service.Start(updates)

	// TODO graceful
}
