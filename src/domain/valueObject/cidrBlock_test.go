package tkValueObject

import (
	"testing"
)

func TestNewCidrBlock(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CidrBlock
			expectError    bool
		}{
			{"192.168.1.0/24", CidrBlock("192.168.1.0/24"), false},
			{"10.0.0.0/8", CidrBlock("10.0.0.0/8"), false},
			{"172.16.0.0/12", CidrBlock("172.16.0.0/12"), false},
			{"0.0.0.0/0", CidrBlock("0.0.0.0/0"), false},
			{"::1/128", CidrBlock("::1/128"), false},
			{"2001:db8::/32", CidrBlock("2001:db8::/32"), false},
			{"::/0", CidrBlock("::/0"), false},
			// Invalid CIDR blocks
			{"192.168.1.256/24", CidrBlock(""), true},
			{"192.168.1.0/33", CidrBlock(""), true},
			{"300.0.0.0/8", CidrBlock(""), true},
			{"2001:db8::/130", CidrBlock(""), true},
			{"invalid", CidrBlock(""), true},
			{"", CidrBlock(""), true},
			{123, CidrBlock(""), true},
			{true, CidrBlock(""), true},
			{[]string{"192.168.1.0/24"}, CidrBlock(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCidrBlock(testCase.inputValue)
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
			inputValue     CidrBlock
			expectedOutput string
		}{
			{CidrBlock("192.168.1.0/24"), "192.168.1.0/24"},
			{CidrBlock("::1/128"), "::1/128"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsIpv4Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     CidrBlock
			expectedOutput bool
		}{
			{CidrBlock("192.168.1.0/24"), true},
			{CidrBlock("10.0.0.0/8"), true},
			{CidrBlock("0.0.0.0/0"), true},
			{CidrBlock("::1/128"), false},
			{CidrBlock("2001:db8::/32"), false},
			{CidrBlock("::/0"), false},
			{CidrBlock(""), false},
			{CidrBlock("invalid"), false},
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
			inputValue     CidrBlock
			expectedOutput bool
		}{
			{CidrBlock("192.168.1.0/24"), false},
			{CidrBlock("10.0.0.0/8"), false},
			{CidrBlock("0.0.0.0/0"), false},
			{CidrBlock("::1/128"), true},
			{CidrBlock("2001:db8::/32"), true},
			{CidrBlock("::/0"), true},
			{CidrBlock(""), true},
			{CidrBlock("invalid"), true},
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
			inputValue     CidrBlock
			expectedOutput bool
		}{
			{CidrBlock("192.168.1.0/24"), true},
			{CidrBlock("10.0.0.0/8"), true},
			{CidrBlock("172.16.0.0/12"), true},
			{CidrBlock("8.8.8.0/24"), false},    // Google DNS (public)
			{CidrBlock("0.0.0.0/0"), false},     // All IPv4 addresses
			{CidrBlock("fd00::/8"), true},       // Private IPv6
			{CidrBlock("2001:db8::/32"), false}, // Documentation IPv6 (not private)
			{CidrBlock(""), false},
			{CidrBlock("invalid"), false},
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
			inputValue     CidrBlock
			expectedOutput bool
		}{
			{CidrBlock("192.168.1.0/24"), false},
			{CidrBlock("10.0.0.0/8"), false},
			{CidrBlock("172.16.0.0/12"), false},
			{CidrBlock("8.8.8.0/24"), true},    // Google DNS (public)
			{CidrBlock("0.0.0.0/0"), true},     // All IPv4 addresses
			{CidrBlock("fd00::/8"), false},     // Private IPv6
			{CidrBlock("2001:db8::/32"), true}, // Documentation IPv6 (not private)
			{CidrBlock(""), true},
			{CidrBlock("invalid"), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsPublic()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
