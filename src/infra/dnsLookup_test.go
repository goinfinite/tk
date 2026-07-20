package tkInfra

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewDnsLookupDefaults(t *testing.T) {
	t.Run("ZeroQueryTimeoutSecsDefaults", func(t *testing.T) {
		lookup := NewDnsLookup(DnsLookupSettings{QueryTimeoutSecs: 0})
		if lookup.queryTimeoutSecs != dnsLookupQueryTimeoutSecsDefault {
			t.Errorf(
				"ZeroTimeoutNotDefaulted: expected %d, got %d",
				dnsLookupQueryTimeoutSecsDefault, lookup.queryTimeoutSecs,
			)
		}
	})

	t.Run("ZeroDialTimeoutMsDefaults", func(t *testing.T) {
		lookup := NewDnsLookup(DnsLookupSettings{DialTimeoutMs: 0})
		if lookup.dialTimeoutMs != dnsLookupDialTimeoutMsDefault {
			t.Errorf(
				"ZeroDialTimeoutNotDefaulted: expected %d, got %d",
				dnsLookupDialTimeoutMsDefault, lookup.dialTimeoutMs,
			)
		}
	})

	t.Run("EmptyPrimaryResolverDefaults", func(t *testing.T) {
		lookup := NewDnsLookup(DnsLookupSettings{})
		if lookup.primaryResolver != dnsLookupPrimaryResolverDefault {
			t.Errorf(
				"EmptyPrimaryResolverNotDefaulted: expected %s, got %s",
				dnsLookupPrimaryResolverDefault, lookup.primaryResolver,
			)
		}
	})

	t.Run("EmptySecondaryResolverDefaults", func(t *testing.T) {
		lookup := NewDnsLookup(DnsLookupSettings{})
		if lookup.secondaryResolver != dnsLookupSecondaryResolverDefault {
			t.Errorf(
				"EmptySecondaryResolverNotDefaulted: expected %s, got %s",
				dnsLookupSecondaryResolverDefault, lookup.secondaryResolver,
			)
		}
	})
}

func TestDnsLookupExecute(t *testing.T) {
	validHostname, err := tkValueObject.NewUnixHostname("example.com")
	if err != nil {
		t.Fatalf("CreateValidHostnameFailed: %v", err)
	}

	ptrIpAddress, err := tkValueObject.NewIpAddress("8.8.8.8")
	if err != nil {
		t.Fatalf("CreatePtrIpAddressFailed: %v", err)
	}
	ptrHostname := tkValueObject.UnixHostname(ptrIpAddress.String())

	dnsLookup := NewDnsLookup(DnsLookupSettings{})

	testCases := []struct {
		name       string
		hostname   tkValueObject.UnixHostname
		recordType *tkValueObject.DnsRecordType
	}{
		{"DefaultRecordType", validHostname, nil},
		{"RecordTypeA", validHostname, &tkValueObject.DnsRecordTypeA},
		{"RecordTypeAAAA", validHostname, &tkValueObject.DnsRecordTypeAAAA},
		{"RecordTypeMX", validHostname, &tkValueObject.DnsRecordTypeMX},
		{"RecordTypeTXT", validHostname, &tkValueObject.DnsRecordTypeTXT},
		{"RecordTypeNS", validHostname, &tkValueObject.DnsRecordTypeNS},
		{"RecordTypeCNAME", validHostname, &tkValueObject.DnsRecordTypeCNAME},
		{"RecordTypePTR", ptrHostname, &tkValueObject.DnsRecordTypePTR},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			results, err := dnsLookup.Execute(testCase.hostname, testCase.recordType)
			if err != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.name)
			}
			if len(results) == 0 {
				t.Errorf("NoResultsReturned: %s", testCase.name)
			}
		})
	}
}
