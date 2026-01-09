package tkValueObject

import "testing"

func TestNewX509VersionNumber(t *testing.T) {
	t.Run("ValidVersionNumber", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509VersionNumber
			expectError    bool
		}{
			{1, X509VersionNumber(1), false},
			{2, X509VersionNumber(2), false},
			{3, X509VersionNumber(3), false},
			{"1", X509VersionNumber(1), false},
			{"2", X509VersionNumber(2), false},
			{"3", X509VersionNumber(3), false},
			{uint8(1), X509VersionNumber(1), false},
			{uint8(2), X509VersionNumber(2), false},
			{uint8(3), X509VersionNumber(3), false},
			{0, X509VersionNumber(0), true},
			{4, X509VersionNumber(0), true},
			{"0", X509VersionNumber(0), true},
			{"4", X509VersionNumber(0), true},
			{-1, X509VersionNumber(0), true},
			{"invalid", X509VersionNumber(0), true},
			{nil, X509VersionNumber(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509VersionNumber(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf(
					"MissingExpectedError: [%v]",
					testCase.inputValue,
				)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf(
					"UnexpectedError: '%s' [%v]",
					err.Error(), testCase.inputValue,
				)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	t.Run("Uint8Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509VersionNumber
			expectedOutput uint8
		}{
			{X509VersionNumber(1), 1},
			{X509VersionNumber(2), 2},
			{X509VersionNumber(3), 3},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint8()

			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509VersionNumber
			expectedOutput string
		}{
			{X509VersionNumber(1), "1"},
			{X509VersionNumber(2), "2"},
			{X509VersionNumber(3), "3"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
