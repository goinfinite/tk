package tkValueObject

import (
	"testing"
)

func TestNewUnixGroupId(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixGroupId
			expectError    bool
		}{
			{0, UnixGroupId(0), false},
			{1, UnixGroupId(1), false},
			{65535, UnixGroupId(65535), false},
			{"0", UnixGroupId(0), false},
			{"455", UnixGroupId(455), false},
			{"65365", UnixGroupId(65365), false},
			{uint64(1000), UnixGroupId(1000), false},
			// Invalid group IDs
			{-1, UnixGroupId(0), true},
			{-10000, UnixGroupId(0), true},
			{"", UnixGroupId(0), true},
			{"abc", UnixGroupId(0), true},
			{true, UnixGroupId(0), true},
			{[]string{"0"}, UnixGroupId(0), true},
			{nil, UnixGroupId(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixGroupId(testCase.inputValue)
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
			inputValue     UnixGroupId
			expectedOutput string
		}{
			{UnixGroupId(0), "0"},
			{UnixGroupId(1), "1"},
			{UnixGroupId(455), "455"},
			{UnixGroupId(65365), "65365"},
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
			inputValue     UnixGroupId
			expectedOutput uint64
		}{
			{UnixGroupId(0), 0},
			{UnixGroupId(1), 1},
			{UnixGroupId(455), 455},
			{UnixGroupId(65365), 65365},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
