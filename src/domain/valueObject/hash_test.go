package tkValueObject

import (
	"testing"
)

func TestNewHash(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Hash
			expectError    bool
		}{
			{"84412CB56723EE2B08D680D78D0D8A46", Hash("84412CB56723EE2B08D680D78D0D8A46"), false},
			{"0340b99427817c02f343941e984c9bb2", Hash("0340b99427817c02f343941e984c9bb2"), false},
			{"d8b5fdce85438fb7cb4b343b6d9c812e351e253d", Hash("d8b5fdce85438fb7cb4b343b6d9c812e351e253d"), false},
			{"7df07c76c7dbf395c07c224959bc19a62e84cc5f9db24398523f4961beedb673", Hash("7df07c76c7dbf395c07c224959bc19a62e84cc5f9db24398523f4961beedb673"), false},
			{"341666654e2340d961374f0fcfb67e06e752b7b2ca34186163745fc62f82c0470b507e087a4ffee8dabc576b779ab43efada95b88024fdddb393cef09cf7e1fb", Hash("341666654e2340d961374f0fcfb67e06e752b7b2ca34186163745fc62f82c0470b507e087a4ffee8dabc576b779ab43efada95b88024fdddb393cef09cf7e1fb"), false},
			{"a1b2c3d4e5f678901234567890abcdef1234567890abcdef1234567890abcdef", Hash("a1b2c3d4e5f678901234567890abcdef1234567890abcdef1234567890abcdef"), false},
			// Invalid hashes
			{"", Hash(""), true},
			{"     ", Hash(""), true},
			{"Kf81h", Hash(""), true},
			{"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4gQ3JhcyBhbGlxdWV0IGRpYW0gaWQgcGxhY2VyYXQgZGFwaWJ1cy4gQ3VyYWJpdHVyIGVsZWlmZW5kIG1hdHRpcyB1cm5hIG5vbiB2dWxwdXRhdGUuIFN1c3BlbmRpc3NlIHBvdGVudGkuIE51bmMgZGlnbmlzc2ltIG5pc2wgdml0YWUgbnVsb@invalid", Hash(""), true},
			{123, Hash(""), true},
			{true, Hash(""), true},
			{[]string{"hash"}, Hash(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewHash(testCase.inputValue)
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
			inputValue     Hash
			expectedOutput string
		}{
			{Hash("84412CB56723EE2B08D680D78D0D8A46"), "84412CB56723EE2B08D680D78D0D8A46"},
			{Hash("0340b99427817c02f343941e984c9bb2"), "0340b99427817c02f343941e984c9bb2"},
			{Hash("d8b5fdce85438fb7cb4b343b6d9c812e351e253d"), "d8b5fdce85438fb7cb4b343b6d9c812e351e253d"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
