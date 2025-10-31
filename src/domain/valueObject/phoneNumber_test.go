package tkValueObject

import (
	"testing"
)

func TestNewPhoneNumber(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput PhoneNumber
			expectError    bool
		}{
			{"1234567890", PhoneNumber("1234567890"), false},
			{"(123) 456-7890", PhoneNumber("1234567890"), false},
			{"1234567890", PhoneNumber("1234567890"), false}, // int converted to string
			{"+1-234-567-8901", PhoneNumber("12345678901"), false},
			{"12345", PhoneNumber("12345"), false},                       // Minimum digits
			{"1234567890123456", PhoneNumber("1234567890123456"), false}, // Maximum digits
			{"555-1234", PhoneNumber("5551234"), false},
			{"true", PhoneNumber(""), true}, // bool converted to string, but no digits
			{"1234", PhoneNumber(""), true},
			{"abc", PhoneNumber(""), true},
			{"1-2-3", PhoneNumber(""), true},
			{"12345678901234567", PhoneNumber(""), true},
			{"1-2-3-4-5-6-7-8-9-0-1-2-3-4-5-6-7", PhoneNumber(""), true},
			{[]string{"1234567890"}, PhoneNumber(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewPhoneNumber(testCase.inputValue)
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
			inputValue     PhoneNumber
			expectedOutput string
		}{
			{PhoneNumber("1234567890"), "1234567890"},
			{PhoneNumber("5551234"), "5551234"},
			{PhoneNumber("1234567890123456"), "1234567890123456"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
