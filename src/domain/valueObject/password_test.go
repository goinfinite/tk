package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewPassword(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Password
			expectError    bool
		}{
			{"Valid123!", Password("Valid123!"), false},
			{"Abc1!", Password("Abc1!"), false},
			{"Password123!", Password("Password123!"), false},
			{"MySecurePass123!@#", Password("MySecurePass123!@#"), false},
			{"12345a!", Password("12345a!"), false},                                               // Minimum length
			{"A1!" + strings.Repeat("a", 125), Password("A1!" + strings.Repeat("a", 125)), false}, // Maximum length approx
			// Invalid: too short
			{"Abc1", Password(""), true},
			{"A1!", Password(""), true},
			{"123!", Password(""), true},
			// Invalid: too long
			{"A1!" + strings.Repeat("a", 126), Password(""), true},
			// Invalid: missing letter
			{"123456!", Password(""), true},
			// Invalid: missing number
			{"Password!", Password(""), true},
			// Invalid: missing special
			{"Password123", Password(""), true},
			// Invalid: non-string
			{123, Password(""), true},
			{true, Password(""), true},
			{[]string{"Valid123!"}, Password(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPassword(testCase.inputValue)
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
			inputValue     Password
			expectedOutput string
		}{
			{Password("Valid123!"), "Valid123!"},
			{Password("Password123!"), "Password123!"},
			{Password("MySecurePass123!@#"), "MySecurePass123!@#"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
