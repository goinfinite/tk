package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewUnixCommand(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixCommand
			expectError    bool
		}{
			{"curl https://google.com", UnixCommand("curl https://google.com"), false},
			{"mv file1 file2", UnixCommand("mv file1 file2"), false},
			{"os vhost get", UnixCommand("os vhost get"), false},
			{"os services create-installable -n php", UnixCommand("os services create-installable -n php"), false},
			{"ls -la", UnixCommand("ls -la"), false},
			{"echo hello", UnixCommand("echo hello"), false},
			{"date | md5sum | awk '{print $1}'", UnixCommand("date | md5sum | awk '{print $1}'"), false},
			// Convertible types
			{123, UnixCommand("123"), false},   // int to string
			{true, UnixCommand("true"), false}, // bool to string
			// Edge cases
			{"ab", UnixCommand("ab"), false},                                           // Minimum length
			{strings.Repeat("a", 4096), UnixCommand(strings.Repeat("a", 4096)), false}, // Maximum length
			// Invalid: too short
			{"a", UnixCommand(""), true},
			{"", UnixCommand(""), true},
			// Invalid: too long
			{strings.Repeat("a", 4097), UnixCommand(""), true},
			// Invalid: non-convertible types
			{[]string{"ls"}, UnixCommand(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixCommand(testCase.inputValue)
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
			inputValue     UnixCommand
			expectedOutput string
		}{
			{UnixCommand("curl https://google.com"), "curl https://google.com"},
			{UnixCommand("mv file1 file2"), "mv file1 file2"},
			{UnixCommand("ls -la"), "ls -la"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
