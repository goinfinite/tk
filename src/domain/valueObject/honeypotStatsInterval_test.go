package tkValueObject

import "testing"

func TestNewHoneypotStatsInterval(t *testing.T) {
	testCaseStructs := []struct {
		inputValue  any
		expectError bool
	}{
		{"30m", false},
		{"1h", false},
		{"1h30m", false},
		{"24h", false},
		{"9999999999h", true},
		{"invalid", true},
		{"", true},
		{"0", false},
		{123, true},
	}

	for _, testCase := range testCaseStructs {
		_, err := NewHoneypotStatsInterval(testCase.inputValue)
		if testCase.expectError && err == nil {
			t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
		}
		if !testCase.expectError && err != nil {
			t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
		}
	}
}
