package tkValueObject

import (
	"testing"
)

func TestNewFqdn(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Fqdn
			expectError    bool
		}{
			{"example.com", Fqdn("example.com"), false},
			{"sub.example.com", Fqdn("sub.example.com"), false},
			{"*.example.com", Fqdn("*.example.com"), false},
			{"my-site.co.uk", Fqdn("my-site.co.uk"), false},
			{"sub-domain.example.org", Fqdn("sub-domain.example.org"), false},
			{"EXAMPLE.COM", Fqdn("example.com"), false}, // gets lowercased
			// Invalid inputs
			{"-example.com", Fqdn(""), true},
			{"example-.com", Fqdn(""), true},
			{"example.c", Fqdn(""), true},
			{"example..com", Fqdn(""), true},
			{"*example.com", Fqdn(""), true},
			{"192.168.1.1", Fqdn(""), true}, // IP address
			{"", Fqdn(""), true},
			{123, Fqdn(""), true},
			{true, Fqdn(""), true},
			{[]string{"example.com"}, Fqdn(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewFqdn(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Fqdn
			expectedOutput string
		}{
			{Fqdn("example.com"), "example.com"},
			{Fqdn("sub.example.com"), "sub.example.com"},
			{Fqdn("*.example.com"), "*.example.com"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
