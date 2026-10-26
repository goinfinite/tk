package tkValueObject

import "testing"

func TestNewActivityRecordCode(test *testing.T) {
	test.Run("ValidActivityRecordCode", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"LoginFailed", false},
			{"LoginSuccessful", false},
			{"AccountCreated", false},
			{"AccountDeleted", false},
			{"AccountPasswordUpdated", false},
			{"AccountApiKeyUpdated", false},
			{"AccountQuotaUpdated", false},
			{"UnauthorizedAccess", false},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordCode(testCase.inputValue)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	test.Run("InvalidActivityRecordCode", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"", true},
			{"a", true},
			{1000, true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordCode(testCase.inputValue)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})
}
