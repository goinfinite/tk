package tkValueObject

import "testing"

func TestNewX509KeyUsage(t *testing.T) {
	t.Run("ValidKeyUsage", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509KeyUsage
			expectError    bool
		}{
			{
				"digitalSignature",
				X509KeyUsageDigitalSignature,
				false,
			},
			{
				"contentCommitment",
				X509KeyUsageContentCommitment,
				false,
			},
			{
				"keyEncipherment",
				X509KeyUsageKeyEncipherment,
				false,
			},
			{
				"dataEncipherment",
				X509KeyUsageDataEncipherment,
				false,
			},
			{"keyAgreement", X509KeyUsageKeyAgreement, false},
			{"keyCertSign", X509KeyUsageKeyCertSign, false},
			{"cRLSign", X509KeyUsageCRLSign, false},
			{"encipherOnly", X509KeyUsageEncipherOnly, false},
			{"decipherOnly", X509KeyUsageDecipherOnly, false},
			{"", X509KeyUsage(""), true},
			{"invalidUsage", X509KeyUsage(""), true},
			{"nonRepudiation", X509KeyUsage(""), true},
			{123, X509KeyUsage(""), true},
			{nil, X509KeyUsage(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509KeyUsage(testCase.inputValue)

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
			inputValue     X509KeyUsage
			expectedOutput string
		}{
			{X509KeyUsageDigitalSignature, "digitalSignature"},
			{X509KeyUsageKeyEncipherment, "keyEncipherment"},
			{X509KeyUsageKeyCertSign, "keyCertSign"},
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
