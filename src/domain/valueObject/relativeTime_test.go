package tkValueObject

import (
	"testing"
)

func TestNewRelativeTime(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput RelativeTime
			expectError    bool
		}{
			{"5 minutes", RelativeTime("5 minutes"), false},
			{"1 hour ago", RelativeTime("1 hour ago"), false},
			{"2 days from now", RelativeTime("2 days from now"), false},
			{"30 seconds", RelativeTime("30 seconds"), false},
			{"1.5 hours ago", RelativeTime("1.5 hours ago"), false},
			{"7 days", RelativeTime("7 days"), false},
			{"1 week ago", RelativeTime("1 week ago"), false},
			{"6 months from now", RelativeTime("6 months from now"), false},
			{"2 years", RelativeTime("2 years"), false},
			// Short forms
			{"5m", RelativeTime("5m"), false},
			{"1h ago", RelativeTime("1h ago"), false},
			{"2d from now", RelativeTime("2d from now"), false},
			{"30s", RelativeTime("30s"), false},
			{"1w ago", RelativeTime("1w ago"), false},
			{"6M from now", RelativeTime("6M from now"), false},
			{"2y", RelativeTime("2y"), false},
			// Invalid formats
			{"", RelativeTime(""), true},
			{"invalid", RelativeTime(""), true},
			{"5", RelativeTime(""), true},
			{"minutes ago", RelativeTime(""), true},
			{"5 invalid", RelativeTime(""), true},
			{"5 minutes invalid", RelativeTime(""), true},
			{"abc minutes", RelativeTime(""), true},
			// Non-string
			{123, RelativeTime(""), true},
			{true, RelativeTime(""), true},
			{[]string{"5 minutes"}, RelativeTime(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewRelativeTime(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     RelativeTime
			expectedOutput string
		}{
			{RelativeTime("5 minutes"), "5 minutes"},
			{RelativeTime("1 hour ago"), "1 hour ago"},
			{RelativeTime("2 days from now"), "2 days from now"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
