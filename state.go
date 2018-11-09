package bot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const stateExpiry = time.Hour * 6

func stateKey(userID int) string {
	return fmt.Sprintf("state:%d", userID)
}

func NewState(stateMap map[string]string) *State {
	page, err := strconv.ParseInt(stateMap["page"], 10, 64)
	if page < 1 || err != nil {
		page = 1
	}

	groupID, err := strconv.ParseInt(stateMap["groupID"], 10, 64)
	if err != nil {
		groupID = 0
	}

	return &State{
		ID:      stateMap["id"],
		Page:    int(page),
		GroupID: groupID,
	}
}

type State struct {
	ID      string
	Page    int
	GroupID int64
}

func (s *State) IsSelection() bool {
	return s.ID == "selection"
}

func (s *State) IsSettings() bool {
	return s.ID == "settings"
}

func (s *State) StateFa() string {
	switch s.ID {
	case "selection":
		return "انتخاب گروه"
	case "settings":
		return "تنظیمات گروه"
	}
	return ""
}

func (s *State) canBackTo(to string) bool {
	switch s.ID {
	case "selection":
		return false
	case "settings":
		return to == "selection"
	}

	return false
}

func (s *State) Map() map[string]string {
	page := strconv.FormatInt(int64(s.Page), 10)
	groupID := strconv.FormatInt(s.GroupID, 10)
	return map[string]string{
		"id":      s.ID,
		"page":    page,
		"groupID": groupID,
	}
}

func getState(conn *redis.Client, userID int) (*State, error) {
	sKey := stateKey(userID)

	stateMap, err := conn.HGetAll(sKey).Result()
	if err != nil {
		return nil, err
	}

	if len(stateMap) < 1 {
		initState := map[string]interface{}{
			"id":   "selection",
			"page": 1,
		}
		pipe := conn.Pipeline()
		pipe.HMSet(sKey, initState)
		pipe.Expire(sKey, stateExpiry)
		if _, err := pipe.Exec(); err != nil {
			return nil, err
		}
		stateMap["id"] = "selection"
		stateMap["page"] = "1"
	}

	state := NewState(stateMap)
	return state, nil
}

func setStateToSelection(conn *redis.Client, userID int) (*State, error) {
	sKey := stateKey(userID)

	pipe := conn.Pipeline()
	pipe.HSet(sKey, "id", "selection")
	pipe.HDel(sKey, "groupID")
	pipe.Expire(sKey, stateExpiry)
	stateMap := pipe.HGetAll(sKey)
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}
	state := NewState(stateMap.Val())
	return state, err
}

func setStateToSettings(conn *redis.Client, userID int, groupID int64) error {
	skey := stateKey(userID)
	state := map[string]interface{}{
		"id":      "settings",
		"groupID": groupID,
	}

	pipe := conn.Pipeline()
	pipe.HMSet(skey, state)
	pipe.Expire(skey, stateExpiry)
	_, err := pipe.Exec()
	return err
}

func setStatePage(conn *redis.Client, userID int, page int) error {
	sKey := stateKey(userID)
	pipe := conn.Pipeline()
	pipe.HSet(sKey, "page", page)
	pipe.Expire(sKey, stateExpiry)
	_, err := pipe.Exec()
	return err
}

func groupSelectionsKeyboard(conn *redis.Client, state *State, userID int) (*tgbotapi.InlineKeyboardMarkup, error) {
	groups, pageCount, err := adminGroups(conn, userID, state.Page)
	if err != nil {
		return nil, err
	}

	prevPage := state.Page - 1
	nextPage := state.Page + 1
	if state.Page >= pageCount {
		nextPage = -1
	}
	if state.Page == 1 {
		prevPage = -1
	}
	keyboard := createKeyboardForGroupSelections(groups, int(prevPage), int(nextPage))
	return &keyboard, nil
}

func paginator(ids []string, perPage int, page int) (paged []string, pageCount int) {
	if page < 1 {
		page = 1
	}
	length := len(ids)

	if length > perPage {
		pageCount = length / perPage
		if length%perPage > 0 {
			pageCount++
		}
		if page > pageCount {
			page = pageCount
		}
		from := (page - 1) * perPage
		to := page * perPage

		if page == pageCount {
			to = from + (length - from)
		}
		return ids[from:to], pageCount
	}
	return ids, 1
}

func createKeyboardForGroupSelections(groups []minimalGroup, prevPage int, nextPage int) tgbotapi.InlineKeyboardMarkup {
	rowsCount := len(groups) / 2
	if len(groups)%2 > 0 {
		rowsCount++
	}

	buttons := make([][]tgbotapi.InlineKeyboardButton, rowsCount)

	for row := range buttons {
		items := 2
		if row+1 == len(buttons) {
			items = len(groups) - row*2
		}
		buttons[row] = make([]tgbotapi.InlineKeyboardButton, items)
		for i := range buttons[row] {
			index := row*2 + i
			data := fmt.Sprintf("select:%d", groups[index].ID)
			buttons[row][i] = tgbotapi.NewInlineKeyboardButtonData(groups[index].Title, data)
		}
	}

	pageButtons := make([]tgbotapi.InlineKeyboardButton, 0)
	if prevPage > 0 {
		data := fmt.Sprintf("page:%d", prevPage) // ⏩
		b := tgbotapi.NewInlineKeyboardButtonData("صفحه قبل ⏪", data)
		pageButtons = append(pageButtons, b)
	}
	if nextPage > 0 {
		data := fmt.Sprintf("page:%d", nextPage)
		b := tgbotapi.NewInlineKeyboardButtonData("⏩ صفحه بعد", data)
		pageButtons = append(pageButtons, b)
	}

	buttons = append(buttons, pageButtons)
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func createKeyboardForSettings(group *groupSettings) tgbotapi.InlineKeyboardMarkup {
	// ------------- Toggle buttons -------------
	var activeText string
	var activeData string
	var showWarnText string
	var showWarnData string
	if group.IsActive {
		activeText = "✳️ فعال ✳️"
	} else {
		activeText = "🚫 غیر فعال 🚫"
	}
	activeData = fmt.Sprintf("gActive:%t", !group.IsActive)

	if group.ShowWarn {
		showWarnText = "✳️ نمایش اخطار ✳️"
	} else {
		showWarnText = "🚫 عدم نمایش اخطار 🚫"
	}
	showWarnData = fmt.Sprintf("gShowWarn:%t", !group.ShowWarn)

	activeButton := tgbotapi.NewInlineKeyboardButtonData(activeText, activeData)
	showWarnButton := tgbotapi.NewInlineKeyboardButtonData(showWarnText, showWarnData)
	toggleButtons := tgbotapi.NewInlineKeyboardRow(showWarnButton, activeButton)

	// ------------- Second row -------------
	botListButton := tgbotapi.NewInlineKeyboardButtonData("لیست ربات‌ها", "gBot:1")
	changeLimit := tgbotapi.NewInlineKeyboardButtonData("تغییر تعداد اخطارها", "gSetLimit:")
	secondRow := tgbotapi.NewInlineKeyboardRow(changeLimit, botListButton)

	// ------------- Navigation row -------------
	backButton := tgbotapi.NewInlineKeyboardButtonData("بازگشت ⏪", "navigate:selection")
	navRow := tgbotapi.NewInlineKeyboardRow(backButton)

	return tgbotapi.NewInlineKeyboardMarkup(toggleButtons, secondRow, navRow)
}
