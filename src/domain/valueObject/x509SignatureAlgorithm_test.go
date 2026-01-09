package tkValueObject

import "testing"

func TestNewX509SignatureAlgorithm(t *testing.T) {
	t.Run("ValidSignatureAlgorithm", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509SignatureAlgorithm
			expectError    bool
		}{
			{
				"SHA256WithRSA",
				X509SignatureAlgorithmSHA256WithRSA,
				false,
			},
			{
				"SHA384WithRSA",
				X509SignatureAlgorithmSHA384WithRSA,
				false,
			},
			{
				"SHA512WithRSA",
				X509SignatureAlgorithmSHA512WithRSA,
				false,
			},
			{
				"ECDSAWithSHA256",
				X509SignatureAlgorithmECDSAWithSHA256,
				false,
			},
			{
				"ECDSAWithSHA384",
				X509SignatureAlgorithmECDSAWithSHA384,
				false,
			},
			{
				"ECDSAWithSHA512",
				X509SignatureAlgorithmECDSAWithSHA512,
				false,
			},
			{"Ed25519", X509SignatureAlgorithmEd25519, false},
			{"", X509SignatureAlgorithm(""), true},
			{"InvalidAlgorithm", X509SignatureAlgorithm(""), true},
			{"MD5WithRSA", X509SignatureAlgorithm(""), true},
			{"SHA1WithRSA", X509SignatureAlgorithm(""), true},
			{123, X509SignatureAlgorithm(""), true},
			{nil, X509SignatureAlgorithm(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509SignatureAlgorithm(
				testCase.inputValue,
			)

			if testCase.expectError && err == nil {
				t.Errorf(
					"MissingExpectedError: [%v]",
					testCase.inputValue,
				)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf(
					"UnexpectedError: '%s' [%v]",
					err.Error(),
					testCase.inputValue,
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

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509SignatureAlgorithm
			expectedOutput string
		}{
			{X509SignatureAlgorithmSHA256WithRSA, "SHA256WithRSA"},
			{X509SignatureAlgorithmECDSAWithSHA256, "ECDSAWithSHA256"},
			{X509SignatureAlgorithmEd25519, "Ed25519"},
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
