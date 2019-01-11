package hudorbot

import (
	"testing"
)

func TestStateKey(t *testing.T) {
	testCases := []struct {
		input  int
		output string
	}{
		{
			input:  100,
			output: "state:100",
		},
		{
			input:  -100,
			output: "state:-100",
		},
		{
			input:  0,
			output: "state:0",
		},
	}

	for _, tc := range testCases {
		output := stateKey(tc.input)
		if output != tc.output {
			t.Errorf("excpected state key to be %s but received %s", tc.output, output)
		}
	}
}
