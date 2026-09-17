package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewScheduledTaskOutput(t *testing.T) {
	t.Run("ValidScheduledTaskOutput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ScheduledTaskOutput
			expectError    bool
		}{
			{"task completed", ScheduledTaskOutput("task completed"), false},
			{"  task completed  ", ScheduledTaskOutput("task completed"), false},
			{"", ScheduledTaskOutput(""), false},
			{strings.Repeat("a", 65536), ScheduledTaskOutput(strings.Repeat("a", 65536)), false},
			{strings.Repeat("a", 65537), ScheduledTaskOutput(strings.Repeat("a", 65536)), false},
			{123, ScheduledTaskOutput("123"), false},
			{true, ScheduledTaskOutput("true"), false},
			{[]string{"output"}, ScheduledTaskOutput(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewScheduledTaskOutput(testCase.inputValue)
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
			inputValue     ScheduledTaskOutput
			expectedOutput string
		}{
			{ScheduledTaskOutput("task completed"), "task completed"},
			{ScheduledTaskOutput(""), ""},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
