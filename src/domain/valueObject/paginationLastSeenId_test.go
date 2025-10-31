package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewPaginationLastSeenId(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput PaginationLastSeenId
			expectError    bool
		}{
			{"abc", PaginationLastSeenId("abc"), false},
			{"123", PaginationLastSeenId("123"), false},
			{"a-b_c", PaginationLastSeenId("a-b_c"), false},
			{"test-id", PaginationLastSeenId("test-id"), false},
			{strings.Repeat("a", 256), PaginationLastSeenId(strings.Repeat("a", 256)), false},
			{123, PaginationLastSeenId("123"), false},
			{true, PaginationLastSeenId("true"), false},
			// Invalid pagination last seen IDs
			{"", PaginationLastSeenId(""), true},
			{strings.Repeat("a", 257), PaginationLastSeenId(""), true},
			{"invalid@id", PaginationLastSeenId(""), true},
			{"with space", PaginationLastSeenId(""), true},
			{"with.dot", PaginationLastSeenId(""), true},
			{[]string{"abc"}, PaginationLastSeenId(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPaginationLastSeenId(testCase.inputValue)
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
			inputValue     PaginationLastSeenId
			expectedOutput string
		}{
			{PaginationLastSeenId("abc"), "abc"},
			{PaginationLastSeenId("123"), "123"},
			{PaginationLastSeenId("test-id"), "test-id"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
