package tkValueObject

import "testing"

func TestNewScheduledTaskStatus(test *testing.T) {
	test.Run("ValidScheduledTaskStatus", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
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
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
