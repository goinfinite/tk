package tkValueObject

import (
	"testing"
)

func TestNewZipCode(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput ZipCode
			expectError    bool
		}{
			// cSpell:disable
			{"12345 ", ZipCode("12345"), false},
			{"12345-678", ZipCode("12345678"), false},
			{"12345678", ZipCode("12345678"), false},
			{"12345-6789", ZipCode("123456789"), false},
			{" 123456789", ZipCode("123456789"), false},
			// cSpell:enable
			{"1234567890", ZipCode("1234567890"), false},
			{"01234", ZipCode("01234"), false},
			// Invalid zip codes
			{"", ZipCode(""), true},
			{" ", ZipCode(""), true},
			{"a", ZipCode(""), true},
			{"<script>alert('xss')</script>", ZipCode(""), true},
			{"rm -rf /", ZipCode(""), true},
			{"@nDr3A5_", ZipCode(""), true},
			{"12", ZipCode(""), true}, // too short
			{"12345678901", ZipCode(""), true}, // too long
			{123, ZipCode("123"), false}, // valid numeric
			{true, ZipCode(""), true},
			{[]string{"12345"}, ZipCode(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewZipCode(testCase.inputValue)
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
			inputValue     ZipCode
			expectedOutput string
		}{
			{ZipCode("12345"), "12345"},
			{ZipCode("12345678"), "12345678"},
			{ZipCode("01234"), "01234"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
