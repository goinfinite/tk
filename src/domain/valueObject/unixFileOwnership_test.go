package tkValueObject

import (
	"testing"
)

func TestNewUnixFileOwnership(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFileOwnership
			expectError    bool
		}{
			{"dev:dev", UnixFileOwnership("dev:dev"), false},
			{"sudo:sudo", UnixFileOwnership("sudo:sudo"), false},
			{"root:root", UnixFileOwnership("root:root"), false},
			{"www-data:www-data", UnixFileOwnership("www-data:www-data"), false},
			{"www-data:root", UnixFileOwnership("www-data:root"), false},
			{"www-data:dev", UnixFileOwnership("www-data:dev"), false},
			{"dev:www-data", UnixFileOwnership("dev:www-data"), false},
			{"dev:root", UnixFileOwnership("dev:root"), false},
			{"root:dev", UnixFileOwnership("root:dev"), false},
			{"root:www-data", UnixFileOwnership("root:www-data"), false},
			// Invalid file ownerships
			{"", UnixFileOwnership(""), true},
			{1000, UnixFileOwnership(""), true},
			{true, UnixFileOwnership(""), true},
			{":dev", UnixFileOwnership(""), true},
			{"dev:", UnixFileOwnership(""), true},
			{"dev:dev:dev", UnixFileOwnership(""), true},
			{"dev/dev", UnixFileOwnership(""), true},
			{"invalid!:invalid!", UnixFileOwnership(""), true}, // invalid username/group
			{[]string{"dev:dev"}, UnixFileOwnership(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFileOwnership(testCase.inputValue)
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
			inputValue     UnixFileOwnership
			expectedOutput string
		}{
			{UnixFileOwnership("dev:dev"), "dev:dev"},
			{UnixFileOwnership("root:root"), "root:root"},
			{UnixFileOwnership("www-data:www-data"), "www-data:www-data"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadUsernameMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFileOwnership
			expectedOutput UnixUsername
			expectError    bool
		}{
			{UnixFileOwnership("dev:dev"), UnixUsername("dev"), false},
			{UnixFileOwnership("root:root"), UnixUsername("root"), false},
			{UnixFileOwnership("www-data:www-data"), UnixUsername("www-data"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadUsername()
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadGroupNameMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFileOwnership
			expectedOutput UnixGroupName
			expectError    bool
		}{
			{UnixFileOwnership("dev:dev"), UnixGroupName("dev"), false},
			{UnixFileOwnership("root:root"), UnixGroupName("root"), false},
			{UnixFileOwnership("www-data:www-data"), UnixGroupName("www-data"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadGroupName()
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
