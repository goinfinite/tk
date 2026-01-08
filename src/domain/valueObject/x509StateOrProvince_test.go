package tkValueObject

import "testing"

func TestNewX509StateOrProvince(t *testing.T) {
	t.Run("ValidStateOrProvince", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509StateOrProvince
			expectError    bool
		}{
			{"California", X509StateOrProvince("California"), false},
			{"New York", X509StateOrProvince("New York"), false},
			{"Ontario", X509StateOrProvince("Ontario"), false},
			{"Baden-Wurttemberg", X509StateOrProvince("Baden-Wurttemberg"), false},
			{"Ile-de-France", X509StateOrProvince("Ile-de-France"), false},
			{"A", X509StateOrProvince("A"), false},
			{
				"São Paulo",
				X509StateOrProvince("Sao Paulo"),
				false,
			},
			{
				"Québec",
				X509StateOrProvince("Quebec"),
				false,
			},
			{
				"Åland",
				X509StateOrProvince("Aland"),
				false,
			},
			{"", X509StateOrProvince(""), true},
			{
				"ThisStateOrProvinceNameIsWayTooLongAndExceedsTheMaximumAllowedLengthOf128CharactersWhichIsTheStandardLimitForX509CertificateStateFields",
				X509StateOrProvince(""),
				true,
			},
			{"Invalid123", X509StateOrProvince(""), true},
			{"Invalid@State", X509StateOrProvince(""), true},
			{"Invalid\nState", X509StateOrProvince(""), true},
			{123, X509StateOrProvince(""), true},
			{nil, X509StateOrProvince(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509StateOrProvince(testCase.inputValue)

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
			inputValue     X509StateOrProvince
			expectedOutput string
		}{
			{X509StateOrProvince("California"), "California"},
			{X509StateOrProvince("Ontario"), "Ontario"},
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
