package tkValueObject

import "testing"

func TestNewX509ExtendedKeyUsage(t *testing.T) {
	t.Run("ValidExtendedKeyUsage", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509ExtendedKeyUsage
			expectError    bool
		}{
			{"serverAuth", X509ExtendedKeyUsageServerAuth, false},
			{"clientAuth", X509ExtendedKeyUsageClientAuth, false},
			{"codeSigning", X509ExtendedKeyUsageCodeSigning, false},
			{
				"emailProtection",
				X509ExtendedKeyUsageEmailProtection,
				false,
			},
			{"timeStamping", X509ExtendedKeyUsageTimeStamping, false},
			{"ocspSigning", X509ExtendedKeyUsageOCSPSigning, false},
			{"", X509ExtendedKeyUsage(""), true},
			{"invalidUsage", X509ExtendedKeyUsage(""), true},
			{"serverAuthentication", X509ExtendedKeyUsage(""), true},
			{123, X509ExtendedKeyUsage(""), true},
			{nil, X509ExtendedKeyUsage(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509ExtendedKeyUsage(testCase.inputValue)

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
			inputValue     X509ExtendedKeyUsage
			expectedOutput string
		}{
			{X509ExtendedKeyUsageServerAuth, "serverAuth"},
			{X509ExtendedKeyUsageClientAuth, "clientAuth"},
			{X509ExtendedKeyUsageCodeSigning, "codeSigning"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v'", actualOutput, testCase.expectedOutput)
			}
		}
	})
}
