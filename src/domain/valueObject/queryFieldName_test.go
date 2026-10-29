package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewQueryFieldName(test *testing.T) {
	test.Run("ValidQueryFieldName", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput QueryFieldName
		}{
			{"name", QueryFieldName("name")},
			{"createdAt", QueryFieldName("createdAt")},
			{"account_id", QueryFieldName("account_id")},
			{"account-name", QueryFieldName("account-name")},
			{"account.name", QueryFieldName("account.name")},
			{"a", QueryFieldName("a")},
			{strings.Repeat("a", 256), QueryFieldName(strings.Repeat("a", 256))},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewQueryFieldName(testCase.inputValue)
			if err != nil {
				subtest.Errorf(
					"UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue,
				)
				continue
			}
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("InvalidQueryFieldName", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectedErr string
		}{
			{nil, "QueryFieldNameMustBeString"},
			{"", "QueryFieldNameCannotBeEmpty"},
			{"   ", "QueryFieldNameCannotBeEmpty"},
			{"1name", "InvalidQueryFieldName"},
			{"_name", "InvalidQueryFieldName"},
			{"na me", "InvalidQueryFieldName"},
			{"na*me", "InvalidQueryFieldName"},
			{"na?me", "InvalidQueryFieldName"},
			{"na:me", "InvalidQueryFieldName"},
			{"na/me", "InvalidQueryFieldName"},
			{strings.Repeat("a", 257), "QueryFieldNameTooBig"},
			{123, "InvalidQueryFieldName"},
			{map[string]any{}, "QueryFieldNameMustBeString"},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewQueryFieldName(testCase.inputValue)
			if err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
				continue
			}
			if err.Error() != testCase.expectedErr {
				subtest.Errorf(
					"UnexpectedError: '%s' vs '%s' [%v]",
					err.Error(), testCase.expectedErr, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		queryFieldName := QueryFieldName("createdAt")
		if queryFieldName.String() != "createdAt" {
			subtest.Errorf("UnexpectedStringValue: %s", queryFieldName.String())
		}
	})
}
