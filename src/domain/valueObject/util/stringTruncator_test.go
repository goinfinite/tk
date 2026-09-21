package tkVoUtil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSafeTruncateString(t *testing.T) {
	testCases := []struct {
		input          string
		maxBytes       int
		expectedOutput string
	}{
		{"command output", 4096, "command output"},
		{"", 4096, ""},
		{strings.Repeat("a", 4096), 4096, strings.Repeat("a", 4096)},
		{strings.Repeat("a", 4097), 4096, strings.Repeat("a", 4096)},
		{strings.Repeat("a", 4093) + "€", 4096, strings.Repeat("a", 4093) + "€"},
		{strings.Repeat("a", 4095) + "€", 4096, strings.Repeat("a", 4095)},
		{strings.Repeat("a", 65535) + "é", 65536, strings.Repeat("a", 65535)},
		{"é", 1, ""},
		{"command output", 0, ""},
		{"command output", -1, ""},
	}

	for _, testCase := range testCases {
		actualOutput := SafeTruncateString(testCase.input, testCase.maxBytes)

		if actualOutput != testCase.expectedOutput {
			t.Errorf(
				"UnexpectedOutput: '%s' vs '%s' for input '%s'",
				actualOutput, testCase.expectedOutput, testCase.input,
			)
		}

		if !utf8.ValidString(actualOutput) {
			t.Errorf("InvalidUtf8Output: '%s'", actualOutput)
		}
	}
}
