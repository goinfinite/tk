package tkValueObject

import "testing"

func TestNewX509Fingerprint(t *testing.T) {
	t.Run("ValidFingerprint", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509Fingerprint
			expectError    bool
		}{
			{
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0",
				X509Fingerprint("A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0"),
				false,
			},
			{
				"1234567890abcdef1234567890abcdef12345678",
				X509Fingerprint("1234567890abcdef1234567890abcdef12345678"),
				false,
			},
			{
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2",
				X509Fingerprint(
					"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2",
				),
				false,
			},
			{
				"12:34:56:78:90:AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78",
				X509Fingerprint("1234567890ABCDEF1234567890ABCDEF12345678"),
				false,
			},
			{
				"12 34 56 78 90 AB CD EF 12 34 56 78 90 AB CD EF 12 34 56 78",
				X509Fingerprint("1234567890ABCDEF1234567890ABCDEF12345678"),
				false,
			},
			{"", X509Fingerprint(""), true},
			{
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9",
				X509Fingerprint(""),
				true,
			},
			{
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C",
				X509Fingerprint(""),
				true,
			},
			{"InvalidFingerprint", X509Fingerprint(""), true},
			{123, X509Fingerprint(""), true},
			{nil, X509Fingerprint(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509Fingerprint(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
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
			inputValue     X509Fingerprint
			expectedOutput string
		}{
			{
				X509Fingerprint("A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0"),
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0",
			},
			{
				X509Fingerprint("1234567890abcdef1234567890abcdef12345678"),
				"1234567890abcdef1234567890abcdef12345678",
			},
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
