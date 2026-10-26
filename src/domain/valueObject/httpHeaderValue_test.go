package tkValueObject

import "testing"

func TestNewHttpHeaderValue(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput HttpHeaderValue
			expectError    bool
		}{
			{"", HttpHeaderValue(""), false},
			{"shared-secret", HttpHeaderValue("shared-secret"), false},
			{"secret value", HttpHeaderValue("secret value"), false},
			{
				"value\twith\ttabs",
				HttpHeaderValue("value\twith\ttabs"),
				false,
			},
			{"abc=def", HttpHeaderValue("abc=def"), false},
			{"quote\"comma,", HttpHeaderValue("quote\"comma,"), false},
			{"café", HttpHeaderValue("café"), false},
			{123, HttpHeaderValue("123"), false},
			{"line\nbreak", HttpHeaderValue(""), true},
			{"carriage\rreturn", HttpHeaderValue(""), true},
			{"nul\x00", HttpHeaderValue(""), true},
			{"bell\x07", HttpHeaderValue(""), true},
			{"del\x7f", HttpHeaderValue(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewHttpHeaderValue(
				testCase.inputValue,
			)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf(
					"UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue,
				)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     HttpHeaderValue
			expectedOutput string
		}{
			{HttpHeaderValue("shared-secret"), "shared-secret"},
			{HttpHeaderValue("secret value"), "secret value"},
			{HttpHeaderValue(""), ""},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})
}
