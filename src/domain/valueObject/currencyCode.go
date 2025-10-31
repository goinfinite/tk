package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	CurrencyCodeAED     CurrencyCode = "AED"
	CurrencyCodeAUD     CurrencyCode = "AUD"
	CurrencyCodeBRL     CurrencyCode = "BRL"
	CurrencyCodeCAD     CurrencyCode = "CAD"
	CurrencyCodeCHF     CurrencyCode = "CHF"
	CurrencyCodeCLP     CurrencyCode = "CLP"
	CurrencyCodeCNY     CurrencyCode = "CNY"
	CurrencyCodeCOP     CurrencyCode = "COP"
	CurrencyCodeCZK     CurrencyCode = "CZK"
	CurrencyCodeDKK     CurrencyCode = "DKK"
	CurrencyCodeEUR     CurrencyCode = "EUR"
	CurrencyCodeGBP     CurrencyCode = "GBP"
	CurrencyCodeHKD     CurrencyCode = "HKD"
	CurrencyCodeHUF     CurrencyCode = "HUF"
	CurrencyCodeIDR     CurrencyCode = "IDR"
	CurrencyCodeILS     CurrencyCode = "ILS"
	CurrencyCodeINR     CurrencyCode = "INR"
	CurrencyCodeJPY     CurrencyCode = "JPY"
	CurrencyCodeKRW     CurrencyCode = "KRW"
	CurrencyCodeMXN     CurrencyCode = "MXN"
	CurrencyCodeMYR     CurrencyCode = "MYR"
	CurrencyCodeNOK     CurrencyCode = "NOK"
	CurrencyCodeNZD     CurrencyCode = "NZD"
	CurrencyCodePEN     CurrencyCode = "PEN"
	CurrencyCodePHP     CurrencyCode = "PHP"
	CurrencyCodePLN     CurrencyCode = "PLN"
	CurrencyCodeRON     CurrencyCode = "RON"
	CurrencyCodeRUB     CurrencyCode = "RUB"
	CurrencyCodeSAR     CurrencyCode = "SAR"
	CurrencyCodeSEK     CurrencyCode = "SEK"
	CurrencyCodeSGD     CurrencyCode = "SGD"
	CurrencyCodeTHB     CurrencyCode = "THB"
	CurrencyCodeTRY     CurrencyCode = "TRY"
	CurrencyCodeTWD     CurrencyCode = "TWD"
	CurrencyCodeUSD     CurrencyCode = "USD"
	CurrencyCodeXCD     CurrencyCode = "XCD"
	CurrencyCodeZAR     CurrencyCode = "ZAR"
	CurrencyCodeDefault CurrencyCode = CurrencyCodeUSD
)

type CurrencyCode string

func NewCurrencyCode(value any) (currencyCode CurrencyCode, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return currencyCode, errors.New("CurrencyCodeMustBeString")
	}

	stringValue = strings.ToUpper(stringValue)

	currencyCode = CurrencyCode(stringValue)
	switch currencyCode {
	case CurrencyCodeAED, CurrencyCodeAUD, CurrencyCodeBRL, CurrencyCodeCAD,
		CurrencyCodeCHF, CurrencyCodeCLP, CurrencyCodeCNY, CurrencyCodeCOP,
		CurrencyCodeCZK, CurrencyCodeDKK, CurrencyCodeEUR, CurrencyCodeGBP,
		CurrencyCodeHKD, CurrencyCodeHUF, CurrencyCodeIDR, CurrencyCodeILS,
		CurrencyCodeINR, CurrencyCodeJPY, CurrencyCodeKRW, CurrencyCodeMXN,
		CurrencyCodeMYR, CurrencyCodeNOK, CurrencyCodeNZD, CurrencyCodePEN,
		CurrencyCodePHP, CurrencyCodePLN, CurrencyCodeRON, CurrencyCodeRUB,
		CurrencyCodeSAR, CurrencyCodeSEK, CurrencyCodeSGD, CurrencyCodeTHB,
		CurrencyCodeTRY, CurrencyCodeTWD, CurrencyCodeUSD, CurrencyCodeXCD,
		CurrencyCodeZAR:
		return currencyCode, nil
	default:
		return currencyCode, errors.New("InvalidCurrencyCode")
	}
}

func (vo CurrencyCode) String() string {
	return string(vo)
}

func (vo CurrencyCode) ReadCurrencyName() (string, error) {
	availableCurrencyCodesWithNameMap := map[CurrencyCode]string{
		"AED": "United Arab Emirates Dirham",
		"AUD": "Australian Dollar",
		"BRL": "Brazilian Real",
		"CAD": "Canadian Dollar",
		"CHF": "Swiss Franc",
		"CLP": "Chilean Peso",
		"CNY": "Chinese Yuan",
		"COP": "Colombian Peso",
		"CZK": "Czech Koruna",
		"DKK": "Danish Krone",
		"EUR": "Euro",
		"GBP": "British Pound Sterling",
		"HKD": "Hong Kong Dollar",
		"HUF": "Hungarian Forint",
		"IDR": "Indonesian Rupiah",
		"ILS": "Israeli New Shekel",
		"INR": "Indian Rupee",
		"JPY": "Japanese Yen",
		"KRW": "South Korean Won",
		"MXN": "Mexican Peso",
		"MYR": "Malaysian Ringgit",
		"NOK": "Norwegian Krone",
		"NZD": "New Zealand Dollar",
		"PEN": "Peruvian Sol",
		"PHP": "Philippine Peso",
		"PLN": "Polish Złoty",
		"RON": "Romanian Leu",
		"RUB": "Russian Ruble",
		"SAR": "Saudi Riyal",
		"SEK": "Swedish Krona",
		"SGD": "Singapore Dollar",
		"THB": "Thai Baht",
		"TRY": "Turkish Lira",
		"TWD": "New Taiwan Dollar",
		"USD": "United States Dollar",
		"XCD": "East Caribbean Dollar",
		"ZAR": "South African Rand",
	}

	currencyName, exists := availableCurrencyCodesWithNameMap[vo]
	if !exists {
		return "", errors.New("InvalidCurrencyCode")
	}

	return currencyName, nil
}
