package tkValueObject

import (
	"testing"
)

func TestNewUnixAbsoluteGlobPath(t *testing.T) {
	t.Run("NewUnixAbsoluteGlobPath", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixAbsoluteGlobPath
			expectError    bool
		}{
			{"/var/www", UnixAbsoluteGlobPath("/var/www"), false},
			{"/var/home/*/Downloads", UnixAbsoluteGlobPath("/var/home/*/Downloads"), false},
			{"/var/log/?.log", UnixAbsoluteGlobPath("/var/log/?.log"), false},
			{"/etc/[a-z]onf/*", UnixAbsoluteGlobPath("/etc/[a-z]onf/*"), false},
			{"/tmp/*[0-9]", UnixAbsoluteGlobPath("/tmp/*[0-9]"), false},
			{"relative/*", UnixAbsoluteGlobPath("/relative/*"), false},
			{"", UnixAbsoluteGlobPath(""), true},
			{"/var/www;", UnixAbsoluteGlobPath(""), true},
			{"/var/home/|x/*", UnixAbsoluteGlobPath(""), true},
			{"/var/home/a b/*", UnixAbsoluteGlobPath("/var/home/a b/*"), false},
			{"/var/../etc/*", UnixAbsoluteGlobPath(""), true},
			{"/var/./x/*", UnixAbsoluteGlobPath("/var/./x/*"), false},
			{"/~user/*", UnixAbsoluteGlobPath(""), true},
			{"/var/\x01/*", UnixAbsoluteGlobPath(""), true},
			{true, UnixAbsoluteGlobPath("/true"), false},
			{[]string{"/var/*"}, UnixAbsoluteGlobPath(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixAbsoluteGlobPath(testCase.inputValue)
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
			inputValue     UnixAbsoluteGlobPath
			expectedOutput string
		}{
			{UnixAbsoluteGlobPath("/var/www"), "/var/www"},
			{UnixAbsoluteGlobPath("/var/home/*/Downloads"), "/var/home/*/Downloads"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
