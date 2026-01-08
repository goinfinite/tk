package tkValueObject

import "testing"

func TestNewX509SignatureValue(t *testing.T) {
	t.Run("ValidSignatureValue", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509SignatureValue
			expectError    bool
		}{
			{
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2C3D4E5F6",
				X509SignatureValue(
					"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2C3D4E5F6",
				),
				false,
			},
			{
				"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
				X509SignatureValue(
					"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
				),
				false,
			},
			{
				"MEUCIQDxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==",
				X509SignatureValue(
					"MEUCIQDxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==",
				),
				false,
			},
			{
				"MEUCIQDxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==\n",
				X509SignatureValue(
					"MEUCIQDxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==",
				),
				false,
			},
			{
				"MEUCIQDxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==\r\n",
				X509SignatureValue(
					"MEUCIQDxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==",
				),
				false,
			},
			{
				"MEUCIQD\nxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==",
				X509SignatureValue(
					"MEUCIQD\nxzE8zF4hR6Q+K7L8M9N0O1P2Q3R4S5T6U7V8W9X0Y1Z2A3B4C5D6E7F8G9H==",
				),
				false,
			},
			{"", X509SignatureValue(""), true},
			{
				"TooShortSignature",
				X509SignatureValue(""),
				true,
			},
			{"Invalid@Signature", X509SignatureValue(""), true},
			{123, X509SignatureValue(""), true},
			{nil, X509SignatureValue(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509SignatureValue(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf(
					"MissingExpectedError: [%v]",
					testCase.inputValue,
				)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf(
					"UnexpectedError: '%s' [%v]",
					err.Error(), testCase.inputValue,
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
			inputValue     X509SignatureValue
			expectedOutput string
		}{
			{
				X509SignatureValue(
					"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2C3D4E5F6",
				),
				"A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2C3D4E5F6",
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
