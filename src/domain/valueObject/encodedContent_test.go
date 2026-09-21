package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewEncodedContent(t *testing.T) {
	t.Run("ValidEncodedContent", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput EncodedContent
			expectError    bool
		}{
			{"aGVsbG8=", EncodedContent("aGVsbG8="), false},
			{"dGVzdA==", EncodedContent("dGVzdA=="), false},
			{"YWJj", EncodedContent("YWJj"), false},
			// Invalid content
			{"", EncodedContent(""), true},
			{"not base64!", EncodedContent(""), true},
			{strings.Repeat("A", 10485764), EncodedContent(""), true},
			{"a", EncodedContent(""), true},
			{123, EncodedContent(""), true},
			{[]string{"aGVsbG8="}, EncodedContent(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewEncodedContent(testCase.inputValue)
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

	t.Run("ReadDecodedContentMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     EncodedContent
			expectedOutput string
			expectError    bool
		}{
			{EncodedContent("aGVsbG8="), "hello", false},
			{EncodedContent("dGVzdA=="), "test", false},
			{EncodedContent("YWJj"), "abc", false},
			{EncodedContent(""), "", false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := testCase.inputValue.ReadDecodedContent()
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
			inputValue     EncodedContent
			expectedOutput string
		}{
			{EncodedContent("aGVsbG8="), "aGVsbG8="},
			{EncodedContent("YWJj"), "YWJj"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
