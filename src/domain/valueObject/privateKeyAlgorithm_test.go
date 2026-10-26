package tkValueObject

import (
	"testing"
)

func TestNewPrivateKeyAlgorithm(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput PrivateKeyAlgorithm
			expectError    bool
		}{
			{"RSA", PrivateKeyAlgorithm("RSA"), false},
			{"ECDSA", PrivateKeyAlgorithm("ECDSA"), false},
			{"DSA", PrivateKeyAlgorithm("DSA"), false},
			{"Ed25519", PrivateKeyAlgorithm("Ed25519"), false},
			{"rsa", PrivateKeyAlgorithm("RSA"), false},
			{"ecdsa", PrivateKeyAlgorithm("ECDSA"), false},
			{"dsa", PrivateKeyAlgorithm("DSA"), false},
			{"ed25519", PrivateKeyAlgorithm("Ed25519"), false},
			{"ED25519", PrivateKeyAlgorithm("Ed25519"), false},
			{"invalid", PrivateKeyAlgorithm(""), true},
			{"", PrivateKeyAlgorithm(""), true},
			{"AES", PrivateKeyAlgorithm(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPrivateKeyAlgorithm(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("NonStringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput PrivateKeyAlgorithm
			expectError    bool
		}{
			{123, PrivateKeyAlgorithm(""), true},
			{true, PrivateKeyAlgorithm(""), true},
			{false, PrivateKeyAlgorithm(""), true},
			{[]string{"RSA"}, PrivateKeyAlgorithm(""), true},
			{nil, PrivateKeyAlgorithm(""), true},
			{map[string]string{"algorithm": "RSA"}, PrivateKeyAlgorithm(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPrivateKeyAlgorithm(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     PrivateKeyAlgorithm
			expectedOutput string
		}{
			{PrivateKeyAlgorithm("RSA"), "RSA"},
			{PrivateKeyAlgorithm("ECDSA"), "ECDSA"},
			{PrivateKeyAlgorithm("DSA"), "DSA"},
			{PrivateKeyAlgorithm("Ed25519"), "Ed25519"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})
}
