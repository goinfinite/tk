package tkValueObject

import (
	"os"
	"testing"
)

func TestNewUnixFilePermissions(t *testing.T) {
	t.Run("NewUnixFilePermissions", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFilePermissions
			expectError    bool
		}{
			{"644", UnixFilePermissions("644"), false},
			{"755", UnixFilePermissions("755"), false},
			{"0644", UnixFilePermissions("0644"), false},
			{"0755", UnixFilePermissions("0755"), false},
			{"000", UnixFilePermissions("000"), false},
			{"777", UnixFilePermissions("777"), false},
			{"4755", UnixFilePermissions("4755"), false},
			{"2755", UnixFilePermissions("2755"), false},
			{"1777", UnixFilePermissions("1777"), false},
			{644, UnixFilePermissions("644"), false},
			{"", UnixFilePermissions(""), true},
			{"0", UnixFilePermissions(""), true},
			{"64", UnixFilePermissions(""), true},
			{"64444", UnixFilePermissions(""), true},
			{"888", UnixFilePermissions(""), true},
			{"abc", UnixFilePermissions(""), true},
			{"-644", UnixFilePermissions(""), true},
			{true, UnixFilePermissions(""), true},
			{[]string{"644"}, UnixFilePermissions(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFilePermissions(testCase.inputValue)
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

	t.Run("ToFileModeMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePermissions
			expectedOutput os.FileMode
			expectError    bool
		}{
			{"644", 0o644, false},
			{"0644", 0o644, false},
			{"000", 0, false},
			{"777", 0o777, false},
			{"4755", os.ModeSetuid | 0o755, false},
			{"2755", os.ModeSetgid | 0o755, false},
			{"1777", os.ModeSticky | 0o777, false},
			{"7770", os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o770, false},
			{"", 0, true},
			{"888", 0, true},
			{"64444", 0, true},
			{"abc", 0, true},
			{"-644", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := testCase.inputValue.ToFileMode()
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
			inputValue     UnixFilePermissions
			expectedOutput string
		}{
			{UnixFilePermissions("644"), "644"},
			{UnixFilePermissions("0755"), "0755"},
			{UnixFilePermissions("4755"), "4755"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
