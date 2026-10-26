package tkValueObject

import (
	"testing"
)

func TestNewCountryCode(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CountryCode
			expectError    bool
		}{
			{"BR", CountryCode("BR"), false},
			{"US", CountryCode("US"), false},
			{"jp", CountryCode("JP"), false},   // gets uppercased
			{"ru ", CountryCode("RU"), false},  // gets trimmed
			{"  CN", CountryCode("CN"), false}, // gets trimmed and uppercased
			{"GB", CountryCode("GB"), false},
			{"DE", CountryCode("DE"), false},
			// Invalid inputs
			{"", CountryCode(""), true},
			{"XX", CountryCode(""), true}, // invalid country code
			{nil, CountryCode(""), true},
			{true, CountryCode(""), true},
			{float64(32), CountryCode(""), true},
			{100, CountryCode(""), true},
			{"<script>alert('xss')</script>", CountryCode(""), true},
			{"rm -rf /", CountryCode(""), true},
			{"@nDr3A5_", CountryCode(""), true},
			{[]string{"BR"}, CountryCode(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCountryCode(testCase.inputValue)
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
			inputValue     CountryCode
			expectedOutput string
		}{
			{CountryCode("BR"), "BR"},
			{CountryCode("US"), "US"},
			{CountryCode("GB"), "GB"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadCountryNameMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     CountryCode
			expectedOutput string
			expectError    bool
		}{
			{CountryCode("BR"), "Brazil", false},
			{CountryCode("US"), "United States", false},
			{CountryCode("GB"), "United Kingdom", false},
			{CountryCode("DE"), "Germany", false},
			{CountryCode("FR"), "France", false},
			// Invalid country code
			{CountryCode("XX"), "", true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadCountryName()
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
