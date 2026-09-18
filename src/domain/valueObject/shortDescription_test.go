package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewShortDescription(t *testing.T) {
	t.Run("ValidShortDescription", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ShortDescription
			expectError    bool
		}{
			{"A valid description.", ShortDescription("A valid description."), false},
			{"ab", ShortDescription("ab"), false},
			{strings.Repeat("a", 2048), ShortDescription(strings.Repeat("a", 2048)), false},
			// Invalid descriptions
			{"", ShortDescription(""), true},
			{"a", ShortDescription(""), true},
			{"invalid\nnewline", ShortDescription(""), true},
			{"invalid\ttab", ShortDescription(""), true},
			{"invalid\u0085description", ShortDescription(""), true},
			{strings.Repeat("a", 2049), ShortDescription(""), true},
			{123, ShortDescription("123"), false},
			{[]string{"description"}, ShortDescription(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewShortDescription(testCase.inputValue)
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
			inputValue     ShortDescription
			expectedOutput string
		}{
			{ShortDescription("A valid description."), "A valid description."},
			{ShortDescription("ab"), "ab"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
