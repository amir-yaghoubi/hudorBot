package hudorBot

import (
	"testing"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
)

func TestAdminKey(t *testing.T) {
	testCases := []struct {
		input  int
		output string
	}{
		{
			100000,
			"admin:100000",
		},
		{
			-100000,
			"admin:-100000",
		},
		{
			0,
			"admin:0",
		},
	}

	for _, tc := range testCases {
		output := adminKey(tc.input)
		if output != tc.output {
			t.Errorf("excpected admin key to be %s but received %s", tc.output, output)
		}
	}
}

func TestFindCreator(t *testing.T) {
	testCases := []struct {
		input  []tgbotapi.ChatMember
		output *tgbotapi.User
	}{
		{
			input:  []tgbotapi.ChatMember{},
			output: nil,
		},
		{
			input: []tgbotapi.ChatMember{
				{
					Status: "creator",
					User: &tgbotapi.User{
						ID:        100,
						FirstName: "group",
						LastName:  "creator",
						UserName:  "group_creator",
					},
				},
			},
			output: &tgbotapi.User{
				ID:        100,
				FirstName: "group",
				LastName:  "creator",
				UserName:  "group_creator",
			},
		},
		{
			input: []tgbotapi.ChatMember{
				{
					Status: "member",
					User: &tgbotapi.User{
						ID:        100,
						FirstName: "group member",
						LastName:  "member",
						UserName:  "group_member",
					},
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		output := findCreator(tc.input)
		if !cmp.Equal(output, tc.output) {
			t.Errorf("excpected to creator be %#v but instead received %#v", tc.output, output)
		}
	}
}
