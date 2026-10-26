package tkValueObject

import "testing"

func TestNewX509PublicKeyAlgorithm(test *testing.T) {
	test.Run("ValidPublicKeyAlgorithm", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				subtest.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
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
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
