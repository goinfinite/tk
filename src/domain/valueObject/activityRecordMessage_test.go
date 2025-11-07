package tkValueObject

import "testing"

func TestNewActivityRecordMessage(t *testing.T) {
	t.Run("ValidActivityRecordMessage", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"Something went wrong with respective scheduled task", false},
			{"Unable to install marketplace item", false},
			{"Error with install PHP", false},
			{"", false},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordMessage(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("InvalidActivityRecordMessage", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{123, false},
			{true, false},
			{[]string{"msg"}, true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordMessage(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})
}
