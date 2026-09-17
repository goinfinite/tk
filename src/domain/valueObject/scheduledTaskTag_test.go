package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewScheduledTaskTag(t *testing.T) {
	t.Run("ValidScheduledTaskTag", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ScheduledTaskTag
			expectError    bool
		}{
			{"backup", ScheduledTaskTag("backup"), false},
			{"daily-backup", ScheduledTaskTag("daily-backup"), false},
			{"Tag_1", ScheduledTaskTag("tag_1"), false},
			{"BACKUP", ScheduledTaskTag("backup"), false},
			// Invalid tags
			{"", ScheduledTaskTag(""), true},
			{"1invalid", ScheduledTaskTag(""), true},
			{"-invalid", ScheduledTaskTag(""), true},
			{"invalid tag", ScheduledTaskTag(""), true},
			{"invalid@tag", ScheduledTaskTag(""), true},
			{strings.Repeat("a", 258), ScheduledTaskTag(""), true},
			{123, ScheduledTaskTag(""), true},
			{[]string{"backup"}, ScheduledTaskTag(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewScheduledTaskTag(testCase.inputValue)
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
			inputValue     ScheduledTaskTag
			expectedOutput string
		}{
			{ScheduledTaskTag("backup"), "backup"},
			{ScheduledTaskTag("daily-backup"), "daily-backup"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
