package tkInfra

import (
	"testing"
)

func TestShellEscape(t *testing.T) {
	shellEscape := ShellEscape{}

	t.Run("Quote", func(t *testing.T) {
		testCaseStructs := []struct {
			input          string
			expectedOutput string
		}{
			{
				"",
				"''",
			},
			{
				"simple",
				"simple",
			},
			{
				"no-special",
				"no-special",
			},
			{
				"with space",
				"'with space'",
			},
			{
				"with'quote",
				"'with'\"'\"'quote'",
			},
			{
				"special!@#",
				"'special!@#'",
			},
			{
				"mix ed'chars and spaces",
				"'mix ed'\"'\"'chars and spaces'",
			},
			{
				"hello 世界",
				"'hello 世界'",
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := shellEscape.Quote(testCase.input)
			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"QuoteMismatch: input='%s', expected='%s', actual='%s'",
					testCase.input, testCase.expectedOutput, actualOutput,
				)
			}
		}
	})

	t.Run("StripUnsafe", func(t *testing.T) {
		testCaseStructs := []struct {
			input          string
			expectedOutput string
		}{
			{
				"safe string",
				"safe string",
			},
			{
				"string\x00with\x01null",
				"stringwithnull",
			},
			{
				"printable\t\n\r",
				"printable",
			},
			{
				"mix\x7f\x80ed",
				"mixed",
			},
			{
				"café",
				"café",
			},
			{
				"hello 世界",
				"hello 世界",
			},
			{
				string([]byte{0xff, 0xfe, 0xfd}), // InvalidUtf8
				"",
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := shellEscape.StripUnsafe(testCase.input)
			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"StripUnsafeMismatch: input='%s', expected='%s', actual='%s'",
					testCase.input, testCase.expectedOutput, actualOutput,
				)
			}
		}
	})
}
