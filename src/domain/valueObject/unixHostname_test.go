package tkValueObject

import (
	"testing"
)

func TestNewUnixHostname(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixHostname
			expectError    bool
		}{
			{"localhost", UnixHostname("localhost"), false},
			{"example.com", UnixHostname("example.com"), false},
			{"sub.domain.com", UnixHostname("sub.domain.com"), false},
			{"123-abc.com", UnixHostname("123-abc.com"), false},
			{"my-hostname", UnixHostname("my-hostname"), false},
			{"hostname123", UnixHostname("hostname123"), false},
			{"host-name-123", UnixHostname("host-name-123"), false},
			{"xn--d1acj3b", UnixHostname("xn--d1acj3b"), false},
			{"xn--bcher-kva.example", UnixHostname("xn--bcher-kva.example"), false},
			{"example.co.uk", UnixHostname("example.co.uk"), false},
			{"EXAMPLE.COM", UnixHostname("example.com"), false}, // should be lowercased
			{"a", UnixHostname("a"), false},
			{"host.with.many.subdomains.example.com", UnixHostname("host.with.many.subdomains.example.com"), false},
			// Invalid hostnames
			{"", UnixHostname(""), true},
			{"UNION SELECT * FROM USERS", UnixHostname(""), true},
			{"/path\n/path", UnixHostname(""), true},
			{"?param=value", UnixHostname(""), true},
			{"/path/'; DROP TABLE users; --", UnixHostname(""), true},
			{"-hostname", UnixHostname(""), true}, // starts with dash
			{"hostname-", UnixHostname(""), true}, // ends with dash
			{"host..name", UnixHostname(""), true}, // double dot
			{"host name", UnixHostname(""), true}, // space
			{"host!name", UnixHostname(""), true}, // special char
			{"123", UnixHostname("123"), false}, // numeric string is valid hostname
			{"true", UnixHostname("true"), false}, // boolean string is valid hostname
			{[]string{"localhost"}, UnixHostname(""), true},
			{nil, UnixHostname(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixHostname(testCase.inputValue)
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
			inputValue     UnixHostname
			expectedOutput string
		}{
			{UnixHostname("localhost"), "localhost"},
			{UnixHostname("example.com"), "example.com"},
			{UnixHostname("sub.domain.com"), "sub.domain.com"},
			{UnixHostname("123-abc.com"), "123-abc.com"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
