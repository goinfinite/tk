package tkInfra

import (
	"context"
	"fmt"
	"testing"
	"time"

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

	t.Run("ProvidedPrimaryResolverHonored", func(t *testing.T) {
		customPrimaryIpAddress, err := tkValueObject.NewIpAddress("1.1.1.1")
		if err != nil {
			t.Fatalf("CreateCustomPrimaryIpAddressFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver: customPrimaryIpAddress,
		})

		if lookup.primaryResolver != customPrimaryIpAddress {
			t.Errorf(
				"ProvidedPrimaryResolverIgnored: expected %s, got %s",
				customPrimaryIpAddress, lookup.primaryResolver,
			)
		}
	})

	t.Run("ProvidedSecondaryResolverHonored", func(t *testing.T) {
		customSecondaryIpAddress, err := tkValueObject.NewIpAddress("9.9.9.9")
		if err != nil {
			t.Fatalf("CreateCustomSecondaryIpAddressFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			SecondaryResolver: customSecondaryIpAddress,
		})

		if lookup.secondaryResolver != customSecondaryIpAddress {
			t.Errorf(
				"ProvidedSecondaryResolverIgnored: expected %s, got %s",
				customSecondaryIpAddress, lookup.secondaryResolver,
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

// 192.0.2.0/24 (TEST-NET-1, RFC 5737) is reserved for documentation;
// traffic to it is never routed, so a resolver there can never reply.
var nonRoutableTestNetOneIpAddress = tkValueObject.IpAddress("192.0.2.1")

func TestDnsLookupExecuteWithUnreachableResolver(t *testing.T) {
	uniqueHostname, err := tkValueObject.NewUnixHostname(fmt.Sprintf(
		"resolver-isolation-%d.invalid", time.Now().UnixNano(),
	))
	if err != nil {
		t.Fatalf("CreateUniqueHostnameFailed: %v", err)
	}

	t.Run("NonRoutableResolversFailLookup", func(t *testing.T) {
		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver:   nonRoutableTestNetOneIpAddress,
			SecondaryResolver: nonRoutableTestNetOneIpAddress,
			QueryTimeoutSecs:  2,
			DialTimeoutMs:     500,
		})

		results, err := lookup.Execute(uniqueHostname, nil)
		if err == nil {
			t.Errorf(
				"NonRoutableResolversShouldHaveFailed: got %v",
				results,
			)
		}
	})
}

func TestNewDnsLookupHonorsShouldBypassLocalResolver(t *testing.T) {
	lookup := NewDnsLookup(DnsLookupSettings{
		ShouldBypassLocalResolver: true,
	})
	if !lookup.shouldBypassLocalResolver {
		t.Error("ShouldBypassLocalResolverIgnored: expected true")
	}
}

func TestDirectIpAddressResolver(t *testing.T) {
	publicResolver, err := tkValueObject.NewIpAddress("8.8.8.8")
	if err != nil {
		t.Fatalf("CreatePublicResolverIpAddressFailed: %v", err)
	}

	dnsGoogleHostname, err := tkValueObject.NewUnixHostname("dns.google")
	if err != nil {
		t.Fatalf("CreateDnsGoogleHostnameFailed: %v", err)
	}

	lookup := NewDnsLookup(DnsLookupSettings{
		PrimaryResolver:  publicResolver,
		QueryTimeoutSecs: 5,
		DialTimeoutMs:    1000,
	})

	t.Run("ReturnsARecords", func(t *testing.T) {
		results, lookupError := lookup.directIpAddressResolver(
			context.Background(), publicResolver,
			dnsGoogleHostname, tkValueObject.DnsRecordTypeA,
		)
		if lookupError != nil {
			t.Fatalf("DirectARecordLookupFailed: %v", lookupError)
		}
		if len(results) == 0 {
			t.Fatalf("DirectARecordLookupReturnedEmpty")
		}
		foundKnownAddress := false
		for _, result := range results {
			if result == "8.8.8.8" {
				foundKnownAddress = true
				break
			}
		}
		if !foundKnownAddress {
			t.Errorf("ExpectedARecordMissing: got %v", results)
		}
	})

	t.Run("ReturnsAAAARecords", func(t *testing.T) {
		results, lookupError := lookup.directIpAddressResolver(
			context.Background(), publicResolver,
			dnsGoogleHostname, tkValueObject.DnsRecordTypeAAAA,
		)
		if lookupError != nil {
			t.Fatalf("DirectAAAARecordLookupFailed: %v", lookupError)
		}
		if len(results) == 0 {
			t.Fatalf("DirectAAAARecordLookupReturnedEmpty")
		}
	})
}

// The bypass test requires the operator to have set up /etc/hosts so that
// goinfinite.dev points at 127.0.0.1; queried by 8.8.8.8 it resolves to the
// real public IP (216.238.102.32). The contrast proves /etc/hosts is bypassed.
const bypassTestHostname = "goinfinite.dev"
const bypassTestLoopbackAddress = "127.0.0.1"

func TestDnsLookupExecuteBypassesLocalResolver(t *testing.T) {
	bypassHostname, err := tkValueObject.NewUnixHostname(bypassTestHostname)
	if err != nil {
		t.Fatalf("CreateBypassHostnameFailed: %v", err)
	}

	lookup := NewDnsLookup(DnsLookupSettings{
		ShouldBypassLocalResolver: true,
	})

	results, lookupError := lookup.Execute(
		bypassHostname, &tkValueObject.DnsRecordTypeA,
	)
	if lookupError != nil {
		t.Fatalf("BypassLookupFailed: %v", lookupError)
	}
	if len(results) == 0 {
		t.Fatalf("BypassLookupReturnedEmpty")
	}

	for _, result := range results {
		if result == bypassTestLoopbackAddress {
			t.Errorf(
				"LocalResolverNotBypassed: got %v (contains %s, expected public DNS answer)",
				results, bypassTestLoopbackAddress,
			)
		}
	}
}

func TestDnsLookupExecuteIgnoresBypassForNonIpRecordType(t *testing.T) {
	bypassHostname, err := tkValueObject.NewUnixHostname(bypassTestHostname)
	if err != nil {
		t.Fatalf("CreateBypassHostnameFailed: %v", err)
	}

	lookup := NewDnsLookup(DnsLookupSettings{
		ShouldBypassLocalResolver: true,
	})

	results, lookupError := lookup.Execute(
		bypassHostname, &tkValueObject.DnsRecordTypeTXT,
	)
	if lookupError != nil {
		t.Fatalf("TxtLookupFailed: %v", lookupError)
	}
	if len(results) == 0 {
		t.Fatalf("TxtLookupReturnedEmpty")
	}
}
