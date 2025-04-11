package tkValueObject

import (
	"testing"
)

func TestNewIpAddress(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput IpAddress
			expectError    bool
		}{
			// Valid IPv4 addresses
			{"192.168.1.1", IpAddress("192.168.1.1"), false},
			{"10.0.0.1", IpAddress("10.0.0.1"), false},
			{"172.16.0.1", IpAddress("172.16.0.1"), false},
			// Valid IPv6 addresses
			{"::1", IpAddress("::1"), false},
			{"2001:db8::1", IpAddress("2001:db8::1"), false},
			// Invalid IP addresses
			{"192.168.1.256", IpAddress(""), true},
			{"300.0.0.1", IpAddress(""), true},
			{"123.456.78.90", IpAddress(""), true},
			{"abcd::12345", IpAddress(""), true},
			// Empty string
			{"", IpAddress(""), true},
			// Non-string input
			{123, IpAddress(""), true},
			{true, IpAddress(""), true},
			{[]string{"192.168.1.1"}, IpAddress(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewIpAddress(testCase.inputValue)
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
			inputValue     IpAddress
			expectedOutput string
		}{
			{IpAddress("192.168.1.1"), "192.168.1.1"},
			{IpAddress("::1"), "::1"},
			{IpAddress(""), ""},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("SystemConstant", func(t *testing.T) {
		expectedValue := "127.0.0.1"
		actualValue := IpAddressSystem.String()
		if actualValue != expectedValue {
			t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [IpAddressSystem]", actualValue, expectedValue)
		}
	})
}
