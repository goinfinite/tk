package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewScheduledTaskName(t *testing.T) {
	t.Run("ValidScheduledTaskName", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ScheduledTaskName
			expectError    bool
		}{
			{"Backup Task", ScheduledTaskName("Backup Task"), false},
			{"daily-backup", ScheduledTaskName("daily-backup"), false},
			{"Task_1", ScheduledTaskName("Task_1"), false},
			// Invalid names
			{"", ScheduledTaskName(""), true},
			{"1invalid", ScheduledTaskName(""), true},
			{"-invalid", ScheduledTaskName(""), true},
			{"invalid@name", ScheduledTaskName(""), true},
			{strings.Repeat("a", 770), ScheduledTaskName(""), true},
			{123, ScheduledTaskName(""), true},
			{[]string{"Backup Task"}, ScheduledTaskName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewScheduledTaskName(testCase.inputValue)
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
			inputValue     ScheduledTaskName
			expectedOutput string
		}{
			{ScheduledTaskName("Backup Task"), "Backup Task"},
			{ScheduledTaskName("daily-backup"), "daily-backup"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
