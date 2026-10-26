package tkValueObject

import (
	"testing"
)

func TestNewAccessTokenType(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput AccessTokenType
			expectError    bool
		}{
			{"sessionToken", AccessTokenTypeSessionToken, false},
			{"secretKey", AccessTokenTypeSecretKey, false},
			// Invalid inputs
			{"invalid", AccessTokenType(""), true},
			{"", AccessTokenType(""), true},
			{"SESSIONTOKEN", AccessTokenType(""), true}, // Case sensitive
			{123, AccessTokenType(""), true},
			{true, AccessTokenType(""), true},
			{[]string{"sessionToken"}, AccessTokenType(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewAccessTokenType(testCase.inputValue)
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
			inputValue     AccessTokenType
			expectedOutput string
		}{
			{AccessTokenTypeSessionToken, "sessionToken"},
			{AccessTokenTypeSecretKey, "secretKey"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
