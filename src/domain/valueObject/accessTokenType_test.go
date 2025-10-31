package tkValueObject

import (
	"testing"
)

func TestNewAccessTokenType(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
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
			inputValue     AccessTokenType
			expectedOutput string
		}{
			{AccessTokenTypeSessionToken, "sessionToken"},
			{AccessTokenTypeSecretKey, "secretKey"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
