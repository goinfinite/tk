package tkValueObject

import "testing"

func TestNewX509PublicKeyAlgorithm(t *testing.T) {
	t.Run("ValidPublicKeyAlgorithm", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PublicKeyAlgorithm
			expectError    bool
		}{
			{"RSA", X509PublicKeyAlgorithmRSA, false},
			{"rsa", X509PublicKeyAlgorithmRSA, false},
			{"ECDSA", X509PublicKeyAlgorithmECDSA, false},
			{"ecdsa", X509PublicKeyAlgorithmECDSA, false},
			{"Ed25519", X509PublicKeyAlgorithmEd25519, false},
			{"ed25519", X509PublicKeyAlgorithmEd25519, false},
			{"DSA", X509PublicKeyAlgorithmDSA, false},
			{"dsa", X509PublicKeyAlgorithmDSA, false},
			{"", X509PublicKeyAlgorithm(""), true},
			{"InvalidAlgorithm", X509PublicKeyAlgorithm(""), true},
			{"AES", X509PublicKeyAlgorithm(""), true},
			{123, X509PublicKeyAlgorithm(""), true},
			{nil, X509PublicKeyAlgorithm(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PublicKeyAlgorithm(
				testCase.inputValue,
			)

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

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509PublicKeyAlgorithm
			expectedOutput string
		}{
			{X509PublicKeyAlgorithmRSA, "RSA"},
			{X509PublicKeyAlgorithmECDSA, "ECDSA"},
			{X509PublicKeyAlgorithmEd25519, "Ed25519"},
			{X509PublicKeyAlgorithmDSA, "DSA"},
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
