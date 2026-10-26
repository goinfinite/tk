package tkValueObject

import "testing"

func TestNewActivityRecordLevel(test *testing.T) {
	test.Run("ValidActivityRecordLevel", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"DEBUG", false},
			{"INFO", false},
			{"ERROR", false},
			{"debug", false}, // case insensitive
			{"info", false},
			{"SECURITY", false},
			{"WARNING", false},
			{"error", false},
			{"sec", false},  // alias
			{"warn", false}, // alias
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordLevel(testCase.inputValue)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	test.Run("InvalidActivityRecordLevel", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})
}
