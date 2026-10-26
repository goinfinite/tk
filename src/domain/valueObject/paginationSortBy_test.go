package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewPaginationSortBy(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput PaginationSortBy
			expectError    bool
		}{
			{"name", PaginationSortBy("name"), false},
			{"created_at", PaginationSortBy("created_at"), false},
			{"id", PaginationSortBy("id"), false},
			{"field.with.dots", PaginationSortBy("field.with.dots"), false},
			{"field-with-dashes", PaginationSortBy("field-with-dashes"), false},
			{"field with spaces", PaginationSortBy("field with spaces"), false},
			{"a", PaginationSortBy("a"), false},
			{strings.Repeat("a", 256), PaginationSortBy(strings.Repeat("a", 256)), false},
			{123, PaginationSortBy("123"), false},
			{true, PaginationSortBy("true"), false},
			// Invalid pagination sort by
			{"", PaginationSortBy(""), true},
			{strings.Repeat("a", 257), PaginationSortBy(""), true},
			{"invalid@sort", PaginationSortBy(""), true},
			{"with/slash", PaginationSortBy(""), true},
			{"with\nnewline", PaginationSortBy(""), true},
			{[]string{"name"}, PaginationSortBy(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPaginationSortBy(testCase.inputValue)
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
			inputValue     PaginationSortBy
			expectedOutput string
		}{
			{PaginationSortBy("name"), "name"},
			{PaginationSortBy("created_at"), "created_at"},
			{PaginationSortBy("field with spaces"), "field with spaces"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
