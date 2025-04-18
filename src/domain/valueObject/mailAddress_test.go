package tkValueObject

import (
	"testing"
)

func TestNewMailAddress(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput MailAddress
			expectError    bool
		}{
			{"user@example.com", MailAddress("user@example.com"), false},
			{"john.doe@example.co.uk", MailAddress("john.doe@example.co.uk"), false},
			{"user+tag@example.com", MailAddress("user+tag@example.com"), false},
			{"first.last@subdomain.example.com", MailAddress("first.last@subdomain.example.com"), false},
			{"user@127.0.0.1", MailAddress("user@127.0.0.1"), false},
			// IPv6 format not supported by mail.ParseAddress
			{"user@[IPv6:2001:db8::1]", MailAddress(""), true},
			{"\"quoted\"@example.com", MailAddress("\"quoted\"@example.com"), false},
			{" user@example.com ", MailAddress("user@example.com"), false}, // Trimmed
			// Invalid email addresses
			{"", MailAddress(""), true},
			{"plainaddress", MailAddress(""), true},
			{"@missingusername.com", MailAddress(""), true},
			{"user@", MailAddress(""), true},
			{"user@.com", MailAddress(""), true},
			{"user@domain@domain.com", MailAddress(""), true},
			{".user@domain.com", MailAddress(""), true},
			{"user@domain..com", MailAddress(""), true},
			{"user@domain.com.", MailAddress(""), true},
			// Non-string inputs
			{123, MailAddress(""), true},
			{true, MailAddress(""), true},
			{[]string{"user@example.com"}, MailAddress(""), true},
			{nil, MailAddress(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewMailAddress(testCase.inputValue)
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
			inputValue     MailAddress
			expectedOutput string
		}{
			{MailAddress("user@example.com"), "user@example.com"},
			{MailAddress("john.doe@example.co.uk"), "john.doe@example.co.uk"},
			{MailAddress("user+tag@example.com"), "user+tag@example.com"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
