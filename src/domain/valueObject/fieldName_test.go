package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewFieldName(test *testing.T) {
	test.Run("ValidFieldName", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput FieldName
		}{
			{"name", FieldName("name")},
			{"createdAt", FieldName("createdAt")},
			{"account_id", FieldName("account_id")},
			{"account-name", FieldName("account-name")},
			{"account.name", FieldName("account.name")},
			{"a", FieldName("a")},
			{strings.Repeat("a", 64), FieldName(strings.Repeat("a", 64))},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewFieldName(testCase.inputValue)
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

	test.Run("InvalidFieldName", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectedErr string
		}{
			{nil, "FieldNameMustBeString"},
			{"", "FieldNameCannotBeEmpty"},
			{"   ", "FieldNameCannotBeEmpty"},
			{"1name", "InvalidFieldName"},
			{"_name", "InvalidFieldName"},
			{"na me", "InvalidFieldName"},
			{"na*me", "InvalidFieldName"},
			{"na?me", "InvalidFieldName"},
			{"na:me", "InvalidFieldName"},
			{"na/me", "InvalidFieldName"},
			{strings.Repeat("a", 65), "FieldNameTooBig"},
			{123, "InvalidFieldName"},
			{map[string]any{}, "FieldNameMustBeString"},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewFieldName(testCase.inputValue)
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
		fieldName := FieldName("createdAt")
		if fieldName.String() != "createdAt" {
			subtest.Errorf("UnexpectedStringValue: %s", fieldName.String())
		}
	})
}
