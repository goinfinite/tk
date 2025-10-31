package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewUnixUsername(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixUsername
			expectError    bool
		}{
			{"a", UnixUsername("a"), false},
			{"a_1", UnixUsername("a_1"), false},
			{"_abc-123", UnixUsername("_abc-123"), false},
			{"b-c_d-e", UnixUsername("b-c_d-e"), false},
			{"valid_username_with_32_chars_x", UnixUsername("valid_username_with_32_chars_x"), false}, // 32 chars (max length)
			{"A_B", UnixUsername("a_b"), false},                                                       // Test lowercase conversion
			{"test$", UnixUsername("test$"), false},                                                   // Special case ending with $
			// Invalid unix usernames
			{"", UnixUsername(""), true},
			{"/1invalid_start_with_digit", UnixUsername(""), true},
			{"1invalid", UnixUsername(""), true},
			{"-invalid-start-with-dash", UnixUsername(""), true},
			{"invalid_character@", UnixUsername(""), true},
			{"toolongname_with_more_than_32_characters_long", UnixUsername(""), true},
			{123, UnixUsername(""), true},       // "123" invalid
			{true, UnixUsername("true"), false}, // Convertible from bool
			{[]string{"user"}, UnixUsername(""), true},
			{strings.Repeat("a", 33), UnixUsername(""), true}, // Too long
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixUsername(testCase.inputValue)
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
			inputValue     UnixUsername
			expectedOutput string
		}{
			{UnixUsername("a"), "a"},
			{UnixUsername("a_1"), "a_1"},
			{UnixUsername("valid_name"), "valid_name"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
