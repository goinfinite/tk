package tkValueObject

import "testing"

func TestNewX509Locality(test *testing.T) {
	test.Run("ValidLocality", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509Locality
			expectError    bool
		}{
			{"San Francisco", X509Locality("San Francisco"), false},
			{"New York", X509Locality("New York"), false},
			{"London", X509Locality("London"), false},
			{"Saint-Denis", X509Locality("Saint-Denis"), false},
			{"O'Fallon", X509Locality("O'Fallon"), false},
			{"A", X509Locality("A"), false},
			{"São Paulo", X509Locality("Sao Paulo"), false},
			{"Montréal", X509Locality("Montreal"), false},
			{"München", X509Locality("Munchen"), false},
			{"Zürich", X509Locality("Zurich"), false},
			{"", X509Locality(""), true},
			{
				"ThisLocalityNameIsWayTooLongAndExceedsTheMaximumAllowedLengthOf128CharactersWhichIsTheStandardLimitForX509CertificateLocalityFields",
				X509Locality(""),
				true,
			},
			{"Invalid123", X509Locality(""), true},
			{"Invalid@City", X509Locality(""), true},
			{"Invalid\nCity", X509Locality(""), true},
			{123, X509Locality(""), true},
			{nil, X509Locality(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509Locality(testCase.inputValue)

			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				subtest.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509Locality
			expectedOutput string
		}{
			{X509Locality("San Francisco"), "San Francisco"},
			{X509Locality("London"), "London"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
