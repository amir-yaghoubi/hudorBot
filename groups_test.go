package bot

import (
	"testing"
)

func TestNewGroupSettings(t *testing.T) {
	testCases := []struct {
		input  map[string]string
		output groupSettings
	}{
		{
			input:  map[string]string{},
			output: groupSettings{},
		},
		{
			input: map[string]string{
				"isActive":    "0",
				"showWarn":    "0",
				"limit":       "5",
				"creator":     "18854411",
				"title":       "test1",
				"description": "desc test1",
			},
			output: groupSettings{
				IsActive:    false,
				ShowWarn:    false,
				Limit:       5,
				Creator:     18854411,
				Title:       "test1",
				Description: "desc test1",
			},
		},
		{
			input: map[string]string{
				"isActive": "1",
				"showWarn": "1",
				"limit":    "wrong",
			},
			output: groupSettings{
				IsActive: true,
				ShowWarn: true,
				Limit:    0,
				Creator:  0,
			},
		},
	}

	for _, tc := range testCases {
		o := newGroupSettings(tc.input)
		if o.IsActive != tc.output.IsActive {
			t.Errorf("excpected isActive to be %t but receive %t\n", tc.output.IsActive, o.IsActive)
		}
		if o.ShowWarn != tc.output.ShowWarn {
			t.Errorf("excpected ShowWarn to be %t but recieve %t\n", tc.output.ShowWarn, o.ShowWarn)
		}
		if o.Limit != tc.output.Limit {
			t.Errorf("excpected Limit to be %d but receive %d\n", tc.output.Limit, o.Limit)
		}
		if o.Creator != tc.output.Creator {
			t.Errorf("excpected Creator to be %d but receive %d\n", tc.output.Creator, o.Creator)
		}
		if o.Title != tc.output.Title {
			t.Errorf("excpected Title to be %s but receive %s\n", tc.output.Title, o.Title)
		}
		if o.Description != tc.output.Description {
			t.Errorf("excpected Description to be %s but receive %s\n", tc.output.Description, o.Description)
		}
	}
}

func TestMap(t *testing.T) {
	testCases := []struct {
		input  groupSettings
		output map[string]string
	}{
		{
			input: groupSettings{
				IsActive:    false,
				ShowWarn:    false,
				Limit:       0,
				Creator:     -10,
				Title:       "",
				Description: "",
			},
			output: map[string]string{
				"isActive":    "false",
				"showWarn":    "false",
				"limit":       "0",
				"creator":     "-10",
				"title":       "",
				"description": "",
			},
		},
		{
			input: groupSettings{},
			output: map[string]string{
				"isActive":    "false",
				"showWarn":    "false",
				"limit":       "0",
				"creator":     "0",
				"title":       "",
				"description": "",
			},
		},
		{
			input: groupSettings{
				IsActive:    true,
				ShowWarn:    true,
				Limit:       2,
				Creator:     189510000,
				Title:       "test",
				Description: "test desc",
			},
			output: map[string]string{
				"isActive":    "true",
				"showWarn":    "true",
				"limit":       "2",
				"creator":     "189510000",
				"title":       "test",
				"description": "test desc",
			},
		},
	}

	for _, tc := range testCases {
		o := tc.input.Map()
		if o["isActive"] != tc.output["isActive"] {
			t.Errorf("excpected")
		}
		if o["isActive"] != tc.output["isActive"] {
			t.Errorf("excpected isActive to be %s but receive %s\n", tc.output["isActive"], o["isActive"])
		}
		if o["showWarn"] != tc.output["showWarn"] {
			t.Errorf("excpected ShowWarn to be %s but recieve %s\n", tc.output["showWarn"], o["showWarn"])
		}
		if o["limit"] != tc.output["limit"] {
			t.Errorf("excpected Limit to be %s but receive %s\n", tc.output["limit"], o["limit"])
		}
		if o["creator"] != tc.output["creator"] {
			t.Errorf("excpected Creator to be %s but receive %s\n", tc.output["creator"], o["creator"])
		}
		if o["title"] != tc.output["title"] {
			t.Errorf("excpected Title to be %s but receive %s\n", tc.output["title"], o["title"])
		}
		if o["description"] != tc.output["description"] {
			t.Errorf("excpected Description to be %s but receive %s\n", tc.output["description"], o["description"])
		}
	}
}

func TestWhitelistKey(t *testing.T) {
	testCases := []struct {
		input  int64
		output string
	}{
		{
			100000,
			"allowed:100000",
		},
		{
			-100000,
			"allowed:-100000",
		},
		{
			0,
			"allowed:0",
		},
	}

	for _, tc := range testCases {
		output := whiteListKey(tc.input)
		if output != tc.output {
			t.Errorf("excpected whitelist key to be %s but received %s", tc.output, output)
		}
	}
}

func TestGroupKey(t *testing.T) {
	testCases := []struct {
		input  int64
		output string
	}{
		{
			100000,
			"group:100000",
		},
		{
			-100000,
			"group:-100000",
		},
		{
			0,
			"group:0",
		},
	}

	for _, tc := range testCases {
		output := groupKey(tc.input)
		if output != tc.output {
			t.Errorf("excpected group settings key to be %s but received %s", tc.output, output)
		}
	}
}
