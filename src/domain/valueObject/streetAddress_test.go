package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewStreetAddress(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput StreetAddress
			expectError    bool
		}{
			{"123 Main St", StreetAddress("123 Main St"), false},
			{"456 Oak Avenue", StreetAddress("456 Oak Avenue"), false},
			{"10 Downing Street, London", StreetAddress("10 Downing Street, London"), false},
			{"A1B2C3", StreetAddress("A1B2C3"), false},
			{"1st Avenue", StreetAddress("1st Avenue"), false},
			{"Avenida Paulista, 123 - Prédio A - Bloco B - Suíte C", StreetAddress("Avenida Paulista, 123 - Prédio A - Bloco B - Suíte C"), false},
			// Trimmed inputs
			{" street", StreetAddress("street"), false}, // trimmed to "street"
			{"street ", StreetAddress("street"), false}, // trimmed to "street"
			{" true ", StreetAddress("true"), false},    // trimmed to "true"
			{"789 Elm St.", StreetAddress("789 Elm St."), false},
			{"street.", StreetAddress("street."), false},
			// Invalid: ends with invalid char
			{"street,", StreetAddress(""), true},
			// Invalid: too short
			{"A1B", StreetAddress(""), true},
			{"1", StreetAddress(""), true},
			// Invalid: starts with invalid char
			{".street", StreetAddress(""), true},
			// Invalid: contains invalid chars
			{"street@home", StreetAddress(""), true},
			{"street#home", StreetAddress(""), true},
			// Invalid: too long
			{"A" + strings.Repeat("a", 513) + "1", StreetAddress(""), true},
			// Invalid: non-convertible types
			{[]string{"123 Main St"}, StreetAddress(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewStreetAddress(testCase.inputValue)
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
			inputValue     StreetAddress
			expectedOutput string
		}{
			{StreetAddress("123 Main St"), "123 Main St"},
			{StreetAddress("456 Oak Avenue"), "456 Oak Avenue"},
			{StreetAddress("789 Elm St."), "789 Elm St."},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
