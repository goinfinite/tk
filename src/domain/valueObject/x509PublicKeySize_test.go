package tkValueObject

import "testing"

func TestNewX509PublicKeySize(t *testing.T) {
	t.Run("ValidPublicKeySize", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PublicKeySize
			expectError    bool
		}{
			{256, X509PublicKeySize(256), false},
			{384, X509PublicKeySize(384), false},
			{521, X509PublicKeySize(521), false},
			{1024, X509PublicKeySize(1024), false},
			{2048, X509PublicKeySize(2048), false},
			{3072, X509PublicKeySize(3072), false},
			{4096, X509PublicKeySize(4096), false},
			{8192, X509PublicKeySize(8192), false},
			{"2048", X509PublicKeySize(2048), false},
			{"4096", X509PublicKeySize(4096), false},
			{uint16(2048), X509PublicKeySize(2048), false},
			{0, X509PublicKeySize(0), true},
			{128, X509PublicKeySize(0), true},
			{512, X509PublicKeySize(0), true},
			{1536, X509PublicKeySize(0), true},
			{16384, X509PublicKeySize(0), true},
			{"invalid", X509PublicKeySize(0), true},
			{-1, X509PublicKeySize(0), true},
			{nil, X509PublicKeySize(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PublicKeySize(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
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

	t.Run("Uint16Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509PublicKeySize
			expectedOutput uint16
		}{
			{X509PublicKeySize(2048), 2048},
			{X509PublicKeySize(4096), 4096},
			{X509PublicKeySize(256), 256},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint16()

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
			inputValue     X509PublicKeySize
			expectedOutput string
		}{
			{X509PublicKeySize(2048), "2048"},
			{X509PublicKeySize(4096), "4096"},
			{X509PublicKeySize(256), "256"},
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
