package tkValueObject

import "testing"

func TestNewX509SerialNumber(t *testing.T) {
	t.Run("ValidSerialNumber", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509SerialNumber
			expectError    bool
		}{
			{
				"1A2B3C4D5E6F",
				X509SerialNumber("1A2B3C4D5E6F"),
				false,
			},
			{
				"abcdef1234567890",
				X509SerialNumber("abcdef1234567890"),
				false,
			},
			{
				"00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD",
				X509SerialNumber("00112233445566778899AABBCCDD"),
				false,
			},
			{
				"01 23 45 67 89 AB CD EF",
				X509SerialNumber("0123456789ABCDEF"),
				false,
			},
			{"1", X509SerialNumber("1"), false},
			{
				"1234567890ABCDEF1234567890ABCDEF12345678",
				X509SerialNumber("1234567890ABCDEF1234567890ABCDEF12345678"),
				false,
			},
			{"", X509SerialNumber(""), true},
			{
				"1234567890ABCDEF1234567890ABCDEF123456789A",
				X509SerialNumber(""),
				true,
			},
			{"InvalidSerial", X509SerialNumber(""), true},
			{"GHIJKL", X509SerialNumber(""), true},
			{"G123", X509SerialNumber(""), true},
			{nil, X509SerialNumber(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509SerialNumber(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf(
					"MissingExpectedError: [%v]",
					testCase.inputValue,
				)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf(
					"UnexpectedError: '%s' [%v]",
					err.Error(),
					testCase.inputValue,
				)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509SerialNumber
			expectedOutput string
		}{
			{X509SerialNumber("1A2B3C4D5E6F"), "1A2B3C4D5E6F"},
			{X509SerialNumber("abcdef1234567890"), "abcdef1234567890"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
