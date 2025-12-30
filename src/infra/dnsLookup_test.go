package tkInfra

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestDnsLookup(t *testing.T) {
	validHostname, err := tkValueObject.NewUnixHostname("example.com")
	if err != nil {
		t.Fatalf("CreateValidHostnameFailed: %v", err)
	}

	invalidHostname, err := tkValueObject.NewUnixHostname("invalid..hostname")
	if err == nil {
		t.Fatal("InvalidHostnameShouldFail")
	}

	ptrIpAddress, err := tkValueObject.NewIpAddress("8.8.8.8")
	if err != nil {
		t.Fatalf("CreatePtrIpAddressFailed: %v", err)
	}

	testCases := []struct {
		name        string
		hostname    tkValueObject.UnixHostname
		recordType  *tkValueObject.DnsRecordType
		expectError bool
	}{
		{
			"DefaultRecordType",
			validHostname,
			nil,
			false,
		},
		{
			"ValidHostnameRecordTypeA",
			validHostname,
			&tkValueObject.DnsRecordTypeA,
			false,
		},
		{
			"ValidHostnameRecordTypeAAAA",
			validHostname,
			&tkValueObject.DnsRecordTypeAAAA,
			false,
		},
		{
			"ValidHostnameRecordTypeMX",
			validHostname,
			&tkValueObject.DnsRecordTypeMX,
			false,
		},
		{
			"ValidHostnameRecordTypeTXT",
			validHostname,
			&tkValueObject.DnsRecordTypeTXT,
			false,
		},
		{
			"ValidHostnameRecordTypeNS",
			validHostname,
			&tkValueObject.DnsRecordTypeNS,
			false,
		},
		{
			"ValidHostnameRecordTypeCNAME",
			validHostname,
			&tkValueObject.DnsRecordTypeCNAME,
			false,
		},
		{
			"ValidIpAddressRecordTypePTR",
			tkValueObject.UnixHostname(ptrIpAddress.String()),
			&tkValueObject.DnsRecordTypePTR,
			false,
		},
		{
			"InvalidHostname",
			invalidHostname,
			&tkValueObject.DnsRecordTypeA,
			true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dnsLookup := NewDnsLookup(testCase.hostname, testCase.recordType)

			if dnsLookup == nil {
				t.Error("DnsLookupInstanceIsNull")
			}

			if dnsLookup.hostname.String() != testCase.hostname.String() {
				t.Errorf(
					"HostnameMismatch: expected %s, got %s",
					testCase.hostname.String(), dnsLookup.hostname.String(),
				)
			}

			expectedRecordType := tkValueObject.DnsRecordTypeDefault
			if testCase.recordType != nil {
				expectedRecordType = *testCase.recordType
			}
			if dnsLookup.recordType != expectedRecordType {
				t.Errorf(
					"RecordTypeMismatch: expected %s, got %s",
					expectedRecordType.String(), dnsLookup.recordType.String(),
				)
			}

			results, err := dnsLookup.Execute()

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: %s", testCase.name)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.name)
			}
			if !testCase.expectError && len(results) == 0 {
				t.Errorf("NoResultsReturned: %s", testCase.name)
			}
			if testCase.expectError && len(results) != 0 {
				t.Errorf("ExpectedNoResultsOnError: got %d [%s]", len(results), testCase.name)
			}
		})
	}
}
