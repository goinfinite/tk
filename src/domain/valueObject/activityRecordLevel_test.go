package tkValueObject

import "testing"

func TestNewActivityRecordLevel(t *testing.T) {
	t.Run("ValidActivityRecordLevel", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"DEBUG", false},
			{"INFO", false},
			{"WARN", false},
			{"ERROR", false},
			{"SEC", false},
			{"debug", false}, // case insensitive
			{"info", false},
			{"warn", false},
			{"error", false},
			{"sec", false},
			{"SECURITY", false}, // alias
			{"WARNING", false},  // alias
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordLevel(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("InvalidActivityRecordLevel", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"LOG", true},
			{"MSG", true},
			{"RECORD", true},
			{"FYI", true},
			{"", true},
			{123, true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordLevel(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})
}
