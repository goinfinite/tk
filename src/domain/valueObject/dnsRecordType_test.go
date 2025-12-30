package tkValueObject

import (
	"testing"
)

func TestNewDnsRecordType(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput DnsRecordType
			expectError    bool
		}{
			{"A", DnsRecordTypeA, false},
			{"AAAA", DnsRecordTypeAAAA, false},
			{"MX", DnsRecordTypeMX, false},
			{"TXT", DnsRecordTypeTXT, false},
			{"NS", DnsRecordTypeNS, false},
			{"CNAME", DnsRecordTypeCNAME, false},
			{"PTR", DnsRecordTypePTR, false},
			{"a", DnsRecordTypeA, false},
			{"Mx", DnsRecordTypeMX, false},
			// Invalid inputs
			{"INVALID", DnsRecordType(""), true},
			{[]uint{123}, DnsRecordType(""), true},
			{"", DnsRecordType(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewDnsRecordType(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})
}
