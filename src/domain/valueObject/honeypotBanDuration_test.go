package tkValueObject

import "testing"

func TestNewHoneypotBanDuration(t *testing.T) {
	testCaseStructs := []struct {
		inputValue  any
		expectError bool
	}{
		{"24h", false},
		{"1h", false},
		{"30m", false},
		{"invalid", true},
		{"", true},
		{"0s", true},
		{"0m", true},
		{"0h", true},
		{"-1h", true},
		{"-30m", true},
		{123, true},
	}

	for _, testCase := range testCaseStructs {
		_, err := NewHoneypotBanDuration(testCase.inputValue)
		if testCase.expectError && err == nil {
			t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
		}
		if !testCase.expectError && err != nil {
			t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
		}
	}
}
