package hudorbot

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func membersKey(groupID int64, userID int) string {
	return fmt.Sprintf("group:%d:%d", groupID, userID)
}

func incrementMemberWarns(conn *redis.Client, chatID int64, userID int) (int64, error) {
	warnKey := membersKey(chatID, userID)
	incPipe := conn.Pipeline()
	incWarn := incPipe.Incr(warnKey)
	incPipe.Expire(warnKey, time.Hour*24*7)
	_, err := incPipe.Exec()
	return incWarn.Val(), err
}
