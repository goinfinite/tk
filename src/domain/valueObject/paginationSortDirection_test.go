package tkValueObject

import (
	"testing"
)

func TestNewPaginationSortDirection(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput PaginationSortDirection
			expectError    bool
		}{
			{"asc", PaginationSortDirection("asc"), false},
			{"desc", PaginationSortDirection("desc"), false},
			{"ASC", PaginationSortDirection("asc"), false}, // Case insensitive
			{"DESC", PaginationSortDirection("desc"), false},
			// Invalid pagination sort directions
			{"", PaginationSortDirection(""), true},
			{"ascending", PaginationSortDirection(""), true},
			{"descending", PaginationSortDirection(""), true},
			{"invalid", PaginationSortDirection(""), true},
			{123, PaginationSortDirection(""), true},
			{true, PaginationSortDirection(""), true},
			{[]string{"asc"}, PaginationSortDirection(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPaginationSortDirection(testCase.inputValue)
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
			inputValue     PaginationSortDirection
			expectedOutput string
		}{
			{PaginationSortDirection("asc"), "asc"},
			{PaginationSortDirection("desc"), "desc"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
