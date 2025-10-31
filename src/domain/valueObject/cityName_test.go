package tkValueObject

import (
	"testing"
)

func TestNewCityName(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CityName
			expectError    bool
		}{
			// cSpell:disable
			{"são paulo", CityName("São Paulo"), false},
			{"são luís", CityName("São Luís"), false},
			{"São José dos Campos", CityName("São José Dos Campos"), false},
			{"New York", CityName("New York"), false},
			{"Södertälje", CityName("Södertälje"), false},
			{"vila bela da santíssima trindade", CityName("Vila Bela Da Santíssima Trindade"), false},
			{"uru", CityName("Uru"), false},
			{"Drøbak", CityName("Drøbak"), false},
			{"Ny-Ålesund", CityName("Ny-Ålesund"), false},
			{"Mönsterås", CityName("Mönsterås"), false},
			{"Åre", CityName("Åre"), false},
			{"Ålesund", CityName("Ålesund"), false},
			{"Skanör-Falsterbo", CityName("Skanör-Falsterbo"), false},
			{"Århus", CityName("Århus"), false},
			{"München", CityName("München"), false},
			{"Bègles", CityName("Bègles"), false},
			{"Villeneuve-d'Ascq", CityName("Villeneuve-D'ascq"), false},
			{"Saint-Étienne-de-Villeréal", CityName("Saint-Étienne-De-Villeréal"), false},
			// cSpell:enable
			// Invalid inputs
			{"", CityName(""), true},
			{" ", CityName(""), true},
			{"<script>alert('xss')</script>", CityName(""), true},
			{"rm -rf /", CityName(""), true},
			{"@nDr3A5_", CityName(""), true},
			{"a", CityName(""), true}, // too short
			{"a123456789012345678901234567890123456789012345678901234567890", CityName(""), true}, // too long
			{123, CityName(""), true},
			{[]string{"city"}, CityName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCityName(testCase.inputValue)
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
			inputValue     CityName
			expectedOutput string
		}{
			{CityName("São Paulo"), "São Paulo"},
			{CityName("New York"), "New York"},
			{CityName("London"), "London"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
