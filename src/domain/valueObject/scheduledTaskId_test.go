package tkValueObject

import "testing"

func TestNewScheduledTaskId(t *testing.T) {
	t.Run("ValidScheduledTaskId", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ScheduledTaskId
			expectError    bool
		}{
			{"0", ScheduledTaskId(0), false},
			{int(0), ScheduledTaskId(0), false},
			{uint64(42), ScheduledTaskId(42), false},
			{ScheduledTaskId(42), ScheduledTaskId(42), false},
			{float64(42), ScheduledTaskId(42), false},
			// Invalid ids
			{"-1", ScheduledTaskId(0), true},
			{int(-1), ScheduledTaskId(0), true},
			{"invalid", ScheduledTaskId(0), true},
			{[]string{"1"}, ScheduledTaskId(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewScheduledTaskId(testCase.inputValue)
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

	t.Run("Uint64Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     ScheduledTaskId
			expectedOutput uint64
		}{
			{ScheduledTaskId(0), 0},
			{ScheduledTaskId(42), 42},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     ScheduledTaskId
			expectedOutput string
		}{
			{ScheduledTaskId(0), "0"},
			{ScheduledTaskId(42), "42"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
