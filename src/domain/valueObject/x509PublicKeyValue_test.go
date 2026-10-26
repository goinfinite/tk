package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewX509PublicKeyValue(test *testing.T) {
	validKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyWmVf7k3NpNvz1234567890
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/+/+/+/==`

	test.Run("ValidPublicKeyValue", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PublicKeyValue
			expectError    bool
		}{
			{
				validKey,
				X509PublicKeyValue(validKey),
				false,
			},
			{
				"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
					"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789A=",
				X509PublicKeyValue(
					"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
						"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789A=",
				),
				false,
			},
			{"", X509PublicKeyValue(""), true},
			{"TooShort", X509PublicKeyValue(""), true},
			{
				"Invalid@Key!#$%^&*" + strings.Repeat("A", 90),
				X509PublicKeyValue(""),
				true,
			},
			{
				strings.Repeat("A", 1048577),
				X509PublicKeyValue(""),
				true,
			},
			{123, X509PublicKeyValue(""), true},
			{nil, X509PublicKeyValue(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PublicKeyValue(testCase.inputValue)

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
			inputValue     X509PublicKeyValue
			expectedOutput string
		}{
			{X509PublicKeyValue(validKey), validKey},
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
