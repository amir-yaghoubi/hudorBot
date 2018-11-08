package bot

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

func whiteListKey(groupID int64) string {
	return fmt.Sprintf("allowed:%d", groupID)
}

func groupKey(groupID int64) string {
	return fmt.Sprintf("group:%d", groupID)
}

func deactivateGroup(conn *redis.Client, chatID int64) error {
	gpKey := groupKey(chatID)
	return conn.HSet(gpKey, "isActive", "false").Err()
}

func findGroupByID(conn *redis.Client, chatID int64) (*groupSettings, error) {
	// TODO check for empty groupHash and send back error?
	gpKey := groupKey(chatID)
	gHash, err := conn.HGetAll(gpKey).Result()
	if err != nil {
		return nil, err
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

type groupSettings struct {
	IsActive    bool
	ShowWarn    bool
	Limit       int64
	Creator     int
	Title       string
	Description string
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
