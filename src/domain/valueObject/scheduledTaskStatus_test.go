package tkValueObject

import "testing"

func TestNewScheduledTaskStatus(t *testing.T) {
	t.Run("ValidScheduledTaskStatus", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ScheduledTaskStatus
			expectError    bool
		}{
			{"pending", ScheduledTaskStatusPending, false},
			{"running", ScheduledTaskStatusRunning, false},
			{"completed", ScheduledTaskStatusCompleted, false},
			{"failed", ScheduledTaskStatusFailed, false},
			{"cancelled", ScheduledTaskStatusCancelled, false},
			{"timeout", ScheduledTaskStatusTimeout, false},
			{"RUNNING", ScheduledTaskStatusRunning, false},
			{"  running  ", ScheduledTaskStatusRunning, false},
			// Invalid statuses
			{"", ScheduledTaskStatus(""), true},
			{"invalid", ScheduledTaskStatus(""), true},
			{"done", ScheduledTaskStatus(""), true},
			{123, ScheduledTaskStatus(""), true},
			{[]string{"running"}, ScheduledTaskStatus(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewScheduledTaskStatus(testCase.inputValue)
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
			inputValue     ScheduledTaskStatus
			expectedOutput string
		}{
			{ScheduledTaskStatus("pending"), "pending"},
			{ScheduledTaskStatus("completed"), "completed"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
