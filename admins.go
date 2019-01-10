package hudorBot

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
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

func adminGroups(conn *redis.Client, userID int, page int) (groups []minimalGroup, pageCount int, err error) {
	key := adminKey(userID)
	groupIDs, err := conn.SMembers(key).Result()
	if err != nil {
		return nil, 0, err
	}

	perPage := 6
	pagedGroups, pageCount := paginator(groupIDs, perPage, page)

	titleChan := make(chan minimalGroup, perPage)
	errChan := make(chan error, perPage)

	for _, id := range pagedGroups {
		go func(id string, gChan chan<- minimalGroup, errChan chan<- error) {
			gp, _ := strconv.ParseInt(id, 10, 64)
			gpKey := groupKey(gp)
			title, err := conn.HGet(gpKey, "title").Result()
			if err != nil {
				errChan <- err
			} else {
				gChan <- minimalGroup{
					ID:    gp,
					Title: title,
				}
			}
		}(id, titleChan, errChan)
	}

	groups = make([]minimalGroup, 0, len(pagedGroups))
	var errHappend error

	for i := 0; i < len(pagedGroups); i++ {
		select {
		case title := <-titleChan:
			groups = append(groups, title)
			break
		case err := <-errChan:
			errHappend = err
			break
		}
	}

	close(errChan)
	close(titleChan)
	if errHappend != nil {
		return nil, 0, errHappend
	}
	return groups, pageCount, nil
}
