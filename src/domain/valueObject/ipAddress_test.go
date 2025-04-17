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
			{"192.168.1.1", IpAddress("192.168.1.1"), false},
			{"10.0.0.1", IpAddress("10.0.0.1"), false},
			{"172.16.0.1", IpAddress("172.16.0.1"), false},
			{"::1", IpAddress("::1"), false},
			{"2001:db8::1", IpAddress("2001:db8::1"), false},
			// Invalid IP Addresses
			{"192.168.1.256", IpAddress(""), true},
			{"300.0.0.1", IpAddress(""), true},
			{"123.456.78.90", IpAddress(""), true},
			{"abcd::12345", IpAddress(""), true},
			{123, IpAddress(""), true},
			{true, IpAddress(""), true},
			{[]string{"192.168.1.1"}, IpAddress(""), true},
			{"", IpAddress(""), true},
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
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("LocalConstant", func(t *testing.T) {
		expectedValue := "127.0.0.1"
		actualValue := IpAddressLocal.String()
		if actualValue != expectedValue {
			t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [IpAddressLocal]", actualValue, expectedValue)
		}
	})

	t.Run("IsLocalMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     IpAddress
			expectedOutput bool
		}{
			{IpAddress("127.0.0.1"), true},
			{IpAddress("192.168.1.1"), false},
			{IpAddress("::1"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsLocal()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsIpv4Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     IpAddress
			expectedOutput bool
		}{
			{IpAddress("192.168.1.1"), true},
			{IpAddress("::1"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsIpv4()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsIpv6Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     IpAddress
			expectedOutput bool
		}{
			{IpAddress("192.168.1.1"), false},
			{IpAddress("::1"), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsIpv6()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsPrivateMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     IpAddress
			expectedOutput bool
		}{
			{IpAddress("192.168.1.1"), true},
			{IpAddress("::1"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsPrivate()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsPublicMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     IpAddress
			expectedOutput bool
		}{
			{IpAddress("192.168.1.1"), false},
			{IpAddress("::1"), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsPublic()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ToCidrBlockMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     IpAddress
			expectedOutput CidrBlock
		}{
			{IpAddress("192.168.1.1"), CidrBlock("192.168.1.1/32")},
			{IpAddress("::1"), CidrBlock("::1/128")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ToCidrBlock()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
