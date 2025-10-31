package tkValueObject

import (
	"testing"
)

func TestNewCurrencyCode(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CurrencyCode
			expectError    bool
		}{
			// Valid currency codes
			{"BRL", CurrencyCode("BRL"), false},
			{"USD", CurrencyCode("USD"), false},
			{"EUR", CurrencyCode("EUR"), false},
			{"GBP", CurrencyCode("GBP"), false},
			{"AUD", CurrencyCode("AUD"), false},
			{"XCD", CurrencyCode("XCD"), false},
			{"AED", CurrencyCode("AED"), false},
			{"CAD", CurrencyCode("CAD"), false},
			{"CHF", CurrencyCode("CHF"), false},
			{"CLP", CurrencyCode("CLP"), false},
			{"CNY", CurrencyCode("CNY"), false},
			{"COP", CurrencyCode("COP"), false},
			{"CZK", CurrencyCode("CZK"), false},
			{"DKK", CurrencyCode("DKK"), false},
			{"HKD", CurrencyCode("HKD"), false},
			{"HUF", CurrencyCode("HUF"), false},
			{"IDR", CurrencyCode("IDR"), false},
			{"ILS", CurrencyCode("ILS"), false},
			{"INR", CurrencyCode("INR"), false},
			{"JPY", CurrencyCode("JPY"), false},
			{"KRW", CurrencyCode("KRW"), false},
			{"MXN", CurrencyCode("MXN"), false},
			{"MYR", CurrencyCode("MYR"), false},
			{"NOK", CurrencyCode("NOK"), false},
			{"NZD", CurrencyCode("NZD"), false},
			{"PEN", CurrencyCode("PEN"), false},
			{"PHP", CurrencyCode("PHP"), false},
			{"PLN", CurrencyCode("PLN"), false},
			{"RON", CurrencyCode("RON"), false},
			{"RUB", CurrencyCode("RUB"), false},
			{"SAR", CurrencyCode("SAR"), false},
			{"SEK", CurrencyCode("SEK"), false},
			{"SGD", CurrencyCode("SGD"), false},
			{"THB", CurrencyCode("THB"), false},
			{"TRY", CurrencyCode("TRY"), false},
			{"TWD", CurrencyCode("TWD"), false},
			{"ZAR", CurrencyCode("ZAR"), false},
			// Case insensitive
			{"usd", CurrencyCode("USD"), false},
			{"eur", CurrencyCode("EUR"), false},
			// Invalid currency codes
			{"XXX", CurrencyCode(""), true},
			{"INVALID", CurrencyCode(""), true},
			{"", CurrencyCode(""), true},
			{123, CurrencyCode(""), true},
			{true, CurrencyCode(""), true},
			{[]string{"USD"}, CurrencyCode(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCurrencyCode(testCase.inputValue)
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
			inputValue     CurrencyCode
			expectedOutput string
		}{
			{CurrencyCode("USD"), "USD"},
			{CurrencyCode("EUR"), "EUR"},
			{CurrencyCode("JPY"), "JPY"},
			{CurrencyCode("CNY"), "CNY"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadCurrencyNameMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     CurrencyCode
			expectedOutput string
			expectError    bool
		}{
			{CurrencyCode("BRL"), "Brazilian Real", false},
			{CurrencyCode("USD"), "United States Dollar", false},
			{CurrencyCode("EUR"), "Euro", false},
			{CurrencyCode("GBP"), "British Pound Sterling", false},
			{CurrencyCode("AUD"), "Australian Dollar", false},
			{CurrencyCode("XCD"), "East Caribbean Dollar", false},
			{CurrencyCode("AED"), "United Arab Emirates Dirham", false},
			{CurrencyCode("CAD"), "Canadian Dollar", false},
			{CurrencyCode("CHF"), "Swiss Franc", false},
			{CurrencyCode("CLP"), "Chilean Peso", false},
			{CurrencyCode("CNY"), "Chinese Yuan", false},
			{CurrencyCode("COP"), "Colombian Peso", false},
			{CurrencyCode("CZK"), "Czech Koruna", false},
			{CurrencyCode("DKK"), "Danish Krone", false},
			{CurrencyCode("HKD"), "Hong Kong Dollar", false},
			{CurrencyCode("HUF"), "Hungarian Forint", false},
			{CurrencyCode("IDR"), "Indonesian Rupiah", false},
			{CurrencyCode("ILS"), "Israeli New Shekel", false},
			{CurrencyCode("INR"), "Indian Rupee", false},
			{CurrencyCode("JPY"), "Japanese Yen", false},
			{CurrencyCode("KRW"), "South Korean Won", false},
			{CurrencyCode("MXN"), "Mexican Peso", false},
			{CurrencyCode("MYR"), "Malaysian Ringgit", false},
			{CurrencyCode("NOK"), "Norwegian Krone", false},
			{CurrencyCode("NZD"), "New Zealand Dollar", false},
			{CurrencyCode("PEN"), "Peruvian Sol", false},
			{CurrencyCode("PHP"), "Philippine Peso", false},
			{CurrencyCode("PLN"), "Polish Złoty", false},
			{CurrencyCode("RON"), "Romanian Leu", false},
			{CurrencyCode("RUB"), "Russian Ruble", false},
			{CurrencyCode("SAR"), "Saudi Riyal", false},
			{CurrencyCode("SEK"), "Swedish Krona", false},
			{CurrencyCode("SGD"), "Singapore Dollar", false},
			{CurrencyCode("THB"), "Thai Baht", false},
			{CurrencyCode("TRY"), "Turkish Lira", false},
			{CurrencyCode("TWD"), "New Taiwan Dollar", false},
			{CurrencyCode("ZAR"), "South African Rand", false},
			// Invalid
			{CurrencyCode("XXX"), "", true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := testCase.inputValue.ReadCurrencyName()
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
}
