package hudorbot

import (
	"fmt"
	"strconv"
)

func userInfoKey(userID int) string {
	return fmt.Sprintf("user:%d", userID)
}

type userInfo struct {
	ID           int
	FirstName    string
	LastName     string
	UserName     string
	LanguageCode string
}

func (u *userInfo) Map() map[string]interface{} {
	return map[string]interface{}{
		"ID":           strconv.FormatInt(int64(u.ID), 10),
		"FirstName":    u.FirstName,
		"LastName":     u.LastName,
		"UserName":     u.UserName,
		"LanguageCode": u.LanguageCode,
	}
}
