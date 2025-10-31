package tkValueObject

import (
	"testing"
)

func TestNewUnixUserId(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixUserId
			expectError    bool
		}{
			{0, UnixUserId(0), false},
			{1, UnixUserId(1), false},
			{65535, UnixUserId(65535), false},
			{"0", UnixUserId(0), false},
			{"455", UnixUserId(455), false},
			{"65365", UnixUserId(65365), false},
			{uint64(1000), UnixUserId(1000), false},
			// Invalid user IDs
			{-1, UnixUserId(0), true},
			{-10000, UnixUserId(0), true},
			{"", UnixUserId(0), true},
			{"abc", UnixUserId(0), true},
			{true, UnixUserId(0), true},
			{[]string{"0"}, UnixUserId(0), true},
			{nil, UnixUserId(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixUserId(testCase.inputValue)
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
			inputValue     UnixUserId
			expectedOutput string
		}{
			{UnixUserId(0), "0"},
			{UnixUserId(1), "1"},
			{UnixUserId(455), "455"},
			{UnixUserId(65365), "65365"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("Uint64Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixUserId
			expectedOutput uint64
		}{
			{UnixUserId(0), 0},
			{UnixUserId(1), 1},
			{UnixUserId(455), 455},
			{UnixUserId(65365), 65365},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
