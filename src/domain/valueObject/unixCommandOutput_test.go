package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewUnixCommandOutput(test *testing.T) {
	test.Run("ValidUnixCommandOutput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixCommandOutput
			expectError    bool
		}{
			{"command output", UnixCommandOutput("command output"), false},
			{"", UnixCommandOutput(""), false},
			{strings.Repeat("a", 4096), UnixCommandOutput(strings.Repeat("a", 4096)), false},
			{strings.Repeat("a", 4097), UnixCommandOutput(strings.Repeat("a", 4096)), false},
			{strings.Repeat("a", 4093) + "€", UnixCommandOutput(strings.Repeat("a", 4093) + "€"), false},
			{strings.Repeat("a", 4095) + "€", UnixCommandOutput(strings.Repeat("a", 4095)), false},
			{123, UnixCommandOutput("123"), false},
			{true, UnixCommandOutput("true"), false},
			{[]string{"output"}, UnixCommandOutput(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixCommandOutput(testCase.inputValue)
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
			inputValue     UnixCommandOutput
			expectedOutput string
		}{
			{UnixCommandOutput("command output"), "command output"},
			{UnixCommandOutput(""), ""},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
