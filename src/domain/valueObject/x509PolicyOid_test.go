package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewX509PolicyOID(test *testing.T) {
	test.Run("ValidPolicyOID", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PolicyOID
			expectError    bool
		}{
			{"1.2.840.113549", X509PolicyOID("1.2.840.113549"), false},
			{
				"2.5.29.32.0",
				X509PolicyOID("2.5.29.32.0"),
				false,
			},
			{
				"1.2.840.113549.1.1.11",
				X509PolicyOID("1.2.840.113549.1.1.11"),
				false,
			},
			{"1.2", X509PolicyOID("1.2"), false},
			{
				"2.16.840.1.114412.1.1",
				X509PolicyOID("2.16.840.1.114412.1.1"),
				false,
			},
			{"", X509PolicyOID(""), true},
			{"1", X509PolicyOID(""), true},
			{"123", X509PolicyOID(""), true},
			{"1.2.3.A", X509PolicyOID(""), true},
			{"1.2.3.4.5.", X509PolicyOID(""), true},
			{".1.2.3", X509PolicyOID(""), true},
			{"invalid.oid", X509PolicyOID(""), true},
			{strings.Repeat("1.", 128) + "1", X509PolicyOID(""), true},
			{123, X509PolicyOID(""), true},
			{nil, X509PolicyOID(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PolicyOID(testCase.inputValue)

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
			inputValue     X509PolicyOID
			expectedOutput string
		}{
			{X509PolicyOID("1.2.840.113549"), "1.2.840.113549"},
			{X509PolicyOID("2.5.29.32.0"), "2.5.29.32.0"},
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
