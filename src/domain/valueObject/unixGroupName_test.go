package tkValueObject

import (
	"testing"
)

func TestNewUnixGroupName(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixGroupName
			expectError    bool
		}{
			{"root", UnixGroupName("root"), false},
			{"www-data", UnixGroupName("www-data"), false},
			{"dev", UnixGroupName("dev"), false},
			{"sudo", UnixGroupName("sudo"), false},
			{"_group", UnixGroupName("_group"), false},
			{"group123", UnixGroupName("group123"), false},
			{"group_name", UnixGroupName("group_name"), false},
			{"group-name", UnixGroupName("group-name"), false},
			{"a", UnixGroupName("a"), false},
			{"groupwith32characterslong12345", UnixGroupName("groupwith32characterslong12345"), false}, // 32 chars (max length)
			// Invalid group names
			{"", UnixGroupName(""), true},
			{"1group", UnixGroupName(""), true},                                              // starts with number
			{"-group", UnixGroupName(""), true},                                              // starts with dash
			{"group!", UnixGroupName(""), true},                                              // special char
			{"group@domain", UnixGroupName(""), true},                                        // special char
			{"group with space", UnixGroupName(""), true},                                    // space
			{"groupwith40characters123456789012345678901234567890", UnixGroupName(""), true}, // too long
			{123, UnixGroupName(""), true},
			{[]string{"root"}, UnixGroupName(""), true},
			{nil, UnixGroupName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixGroupName(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixGroupName
			expectedOutput string
		}{
			{UnixGroupName("root"), "root"},
			{UnixGroupName("www-data"), "www-data"},
			{UnixGroupName("dev"), "dev"},
			{UnixGroupName("sudo"), "sudo"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
