package bot

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

func whiteListKey(groupID int64) string {
	return fmt.Sprintf("whitelist:%d", groupID)
}

func groupKey(groupID int64) string {
	return fmt.Sprintf("group:%d", groupID)
}

// TODO isBotApproved function

func changeGroupActiveStatus(conn *redis.Client, chatID int64, isActive bool) error {
	gpKey := groupKey(chatID)
	return conn.HSet(gpKey, "isActive", isActive).Err()
}

func groupCreator(conn *redis.Client, chatID int64) (creator int, err error) {
	gpKey := groupKey(chatID)
	creator, err = conn.HGet(gpKey, "creator").Int()
	return creator, err
}

func isGroupActive(conn *redis.Client, chatID int64) (bool, error) {
	gpKey := groupKey(chatID)
	isActiveString, err := conn.HGet(gpKey, "isActive").Result()
	if err != nil {
		return false, err
	}

	isActive, err := strconv.ParseBool(isActiveString)
	if err != nil {
		isActive = false
	}

	return isActive, nil
}

func findGroupByID(conn *redis.Client, chatID int64) (*groupSettings, error) {
	gpKey := groupKey(chatID)
	gHash, err := conn.HGetAll(gpKey).Result()
	if err != nil {
		return nil, err
	}
	if len(gHash) == 0 {
		return nil, nil
	}

	groupSettings := newGroupSettings(gHash)
	return groupSettings, nil
}

func newGroupSettings(groupSetting map[string]string) *groupSettings {
	isActive, err := strconv.ParseBool(groupSetting["isActive"])
	if err != nil {
		isActive = false
	}

	showWarn, err := strconv.ParseBool(groupSetting["showWarn"])
	if err != nil {
		showWarn = false
	}

	limit, err := strconv.ParseInt(groupSetting["limit"], 10, 64)
	if err != nil {
		limit = 0
	}

	creator, err := strconv.ParseInt(groupSetting["creator"], 10, 64)
	if err != nil {
		creator = 0
	}

	gp := &groupSettings{
		Limit:       limit,
		IsActive:    isActive,
		ShowWarn:    showWarn,
		Creator:     int(creator),
		Title:       groupSetting["title"],
		Description: groupSetting["description"],
	}
	return gp
}

type minimalGroup struct {
	ID    int64
	Title string
}

type groupSettings struct {
	IsActive    bool
	ShowWarn    bool
	Limit       int64
	Creator     int
	Title       string
	Description string
}

func (g *groupSettings) IsActiveFa() string {
	if g.IsActive {
		return "❇️ فعال ❇️"
	}
	return "🚫 غیر فعال 🚫"
}

func (g *groupSettings) ShowWarnFa() string {
	if g.ShowWarn {
		return "❇️ فعال ❇️"
	}
	return "🚫 غیر فعال 🚫"
}

func (g *groupSettings) Map() map[string]interface{} {
	return map[string]interface{}{
		"isActive":    strconv.FormatBool(g.IsActive),
		"showWarn":    strconv.FormatBool(g.ShowWarn),
		"limit":       strconv.FormatInt(g.Limit, 10),
		"creator":     strconv.FormatInt(int64(g.Creator), 10),
		"title":       g.Title,
		"description": g.Description,
	}
}
