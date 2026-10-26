package tkInfra

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

// 192.0.2.0/24 (TEST-NET-1, RFC 5737) is reserved for documentation;
// traffic to it is never routed, so a resolver there can never reply.
var nonRoutableTestNetOneIpAddress = tkValueObject.IpAddress("192.0.2.1")

const localhostLoopbackIpAddress = "127.0.0.1"

func TestNewDnsLookup(test *testing.T) {
	customPrimaryIpAddress, primaryIpErr := tkValueObject.NewIpAddress("1.1.1.1")
	if primaryIpErr != nil {
		test.Fatalf("CreateCustomPrimaryIpAddressFailed: %v", primaryIpErr)
	}

	customSecondaryIpAddress, secondaryIpErr := tkValueObject.NewIpAddress("9.9.9.9")
	if secondaryIpErr != nil {
		test.Fatalf("CreateCustomSecondaryIpAddressFailed: %v", secondaryIpErr)
	}

	testCaseStructs := []struct {
		name                 string
		settings             DnsLookupSettings
		expectedPrimary      tkValueObject.IpAddress
		expectedSecondary    tkValueObject.IpAddress
		expectedQueryTimeout uint
		expectedDialTimeout  uint
		expectedBypass       bool
	}{
		{
			name:                 "AllDefaultsWhenAllFieldsZero",
			settings:             DnsLookupSettings{},
			expectedPrimary:      dnsLookupPrimaryResolverDefault,
			expectedSecondary:    dnsLookupSecondaryResolverDefault,
			expectedQueryTimeout: dnsLookupQueryTimeoutSecsDefault,
			expectedDialTimeout:  dnsLookupDialTimeoutMsDefault,
			expectedBypass:       false,
		},
		{
			name:                 "CustomPrimaryResolverHonored",
			settings:             DnsLookupSettings{PrimaryResolver: customPrimaryIpAddress},
			expectedPrimary:      customPrimaryIpAddress,
			expectedSecondary:    dnsLookupSecondaryResolverDefault,
			expectedQueryTimeout: dnsLookupQueryTimeoutSecsDefault,
			expectedDialTimeout:  dnsLookupDialTimeoutMsDefault,
			expectedBypass:       false,
		},
		{
			name: "CustomSecondaryResolverHonored",
			settings: DnsLookupSettings{
				SecondaryResolver: customSecondaryIpAddress,
			},
			expectedPrimary:      dnsLookupPrimaryResolverDefault,
			expectedSecondary:    customSecondaryIpAddress,
			expectedQueryTimeout: dnsLookupQueryTimeoutSecsDefault,
			expectedDialTimeout:  dnsLookupDialTimeoutMsDefault,
			expectedBypass:       false,
		},
		{
			name:                 "ShouldBypassLocalResolverHonored",
			settings:             DnsLookupSettings{ShouldBypassLocalResolver: true},
			expectedPrimary:      dnsLookupPrimaryResolverDefault,
			expectedSecondary:    dnsLookupSecondaryResolverDefault,
			expectedQueryTimeout: dnsLookupQueryTimeoutSecsDefault,
			expectedDialTimeout:  dnsLookupDialTimeoutMsDefault,
			expectedBypass:       true,
		},
	}

	for _, testCase := range testCaseStructs {
		test.Run(testCase.name, func(subtest *testing.T) {
			lookup := NewDnsLookup(testCase.settings)

			if lookup.primaryResolver != testCase.expectedPrimary {
				subtest.Errorf(
					"WrongPrimaryResolver: expected '%s', got '%s'",
					testCase.expectedPrimary, lookup.primaryResolver,
				)
			}
			if lookup.secondaryResolver != testCase.expectedSecondary {
				subtest.Errorf(
					"WrongSecondaryResolver: expected '%s', got '%s'",
					testCase.expectedSecondary, lookup.secondaryResolver,
				)
			}
			if lookup.queryTimeoutSecs != testCase.expectedQueryTimeout {
				subtest.Errorf(
					"WrongQueryTimeoutSecs: expected %d, got %d",
					testCase.expectedQueryTimeout, lookup.queryTimeoutSecs,
				)
			}
			if lookup.dialTimeoutMs != testCase.expectedDialTimeout {
				subtest.Errorf(
					"WrongDialTimeoutMs: expected %d, got %d",
					testCase.expectedDialTimeout, lookup.dialTimeoutMs,
				)
			}
			if lookup.shouldBypassLocalResolver != testCase.expectedBypass {
				subtest.Errorf(
					"WrongShouldBypassLocalResolver: expected %t, got %t",
					testCase.expectedBypass, lookup.shouldBypassLocalResolver,
				)
			}
		})
	}
}

func TestDnsLookupExecute(test *testing.T) {
	validHostname, err := tkValueObject.NewUnixHostname("example.com")
	if err != nil {
		test.Fatalf("CreateValidHostnameFailed: %v", err)
	}

	ptrIpAddress, err := tkValueObject.NewIpAddress("8.8.8.8")
	if err != nil {
		test.Fatalf("CreatePtrIpAddressFailed: %v", err)
	}
	ptrHostname := tkValueObject.UnixHostname(ptrIpAddress.String())

	dnsLookup := NewDnsLookup(DnsLookupSettings{})

	recordTypeTestCases := []struct {
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

	for _, testCase := range recordTypeTestCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			results, lookupErr := dnsLookup.Execute(
				testCase.hostname, testCase.recordType,
			)
			if lookupErr != nil {
				subtest.Errorf(
					"UnexpectedError: '%s' [%s]",
					lookupErr.Error(), testCase.name,
				)
			}
			if len(results) == 0 {
				subtest.Errorf("NoResultsReturned: %s", testCase.name)
			}
		})
	}

	test.Run("UnreachableResolversFailLookup", func(subtest *testing.T) {
		uniqueHostname, err := tkValueObject.NewUnixHostname(fmt.Sprintf(
			"resolver-isolation-%d.invalid", time.Now().UnixNano(),
		))
		if err != nil {
			subtest.Fatalf("CreateUniqueHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver:   nonRoutableTestNetOneIpAddress,
			SecondaryResolver: nonRoutableTestNetOneIpAddress,
			QueryTimeoutSecs:  2,
			DialTimeoutMs:     500,
		})

		results, lookupErr := lookup.Execute(uniqueHostname, nil)
		if lookupErr == nil {
			subtest.Errorf(
				"NonRoutableResolversShouldHaveFailed: got %v",
				results,
			)
		}
	})

	test.Run("LocalhostReturnsLoopbackWithoutDnsServer", func(subtest *testing.T) {
		localhostHostname, err := tkValueObject.NewUnixHostname("localhost")
		if err != nil {
			subtest.Fatalf("CreateLocalhostHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver:   nonRoutableTestNetOneIpAddress,
			SecondaryResolver: nonRoutableTestNetOneIpAddress,
			QueryTimeoutSecs:  1,
			DialTimeoutMs:     500,
		})

		results, lookupErr := lookup.Execute(
			localhostHostname, &tkValueObject.DnsRecordTypeA,
		)
		if lookupErr != nil {
			subtest.Fatalf("LocalhostLookupFailed: %v", lookupErr)
		}

		if !slices.Contains(results, localhostLoopbackIpAddress) {
			subtest.Errorf(
				"LocalResolverShouldReturnLoopback: got %v, expected to contain '%s'",
				results, localhostLoopbackIpAddress,
			)
		}
	})

	test.Run("ZoneScopedIpv6LiteralReturnsItself", func(subtest *testing.T) {
		zoneScopedHostname, err := tkValueObject.NewUnixHostname("fe80::1%eth0")
		if err != nil {
			subtest.Fatalf("CreateZoneScopedHostnameFailed: %v", err)
		}

		results, lookupErr := dnsLookup.Execute(
			zoneScopedHostname, &tkValueObject.DnsRecordTypeAAAA,
		)
		if lookupErr != nil {
			subtest.Fatalf("ZoneScopedLookupFailed: %v", lookupErr)
		}

		if !slices.Contains(results, "fe80::1%eth0") {
			subtest.Errorf(
				"ZoneScopedLiteralMissing: got %v, expected to contain 'fe80::1%%eth0'",
				results,
			)
		}
	})

	test.Run("ZoneScopedIpv6LiteralReturnsItselfWithBypass", func(subtest *testing.T) {
		zoneScopedHostname, err := tkValueObject.NewUnixHostname("fe80::1%eth0")
		if err != nil {
			subtest.Fatalf("CreateZoneScopedHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver:           nonRoutableTestNetOneIpAddress,
			SecondaryResolver:         nonRoutableTestNetOneIpAddress,
			QueryTimeoutSecs:          1,
			DialTimeoutMs:             500,
			ShouldBypassLocalResolver: true,
		})

		results, lookupErr := lookup.Execute(
			zoneScopedHostname, &tkValueObject.DnsRecordTypeAAAA,
		)
		if lookupErr != nil {
			subtest.Fatalf("ZoneScopedBypassLookupFailed: %v", lookupErr)
		}

		if !slices.Contains(results, "fe80::1%eth0") {
			subtest.Errorf(
				"ZoneScopedLiteralMissing: got %v, expected to contain 'fe80::1%%eth0'",
				results,
			)
		}
	})

	test.Run("LocalhostBypassedSkipsHostsFile", func(subtest *testing.T) {
		localhostHostname, err := tkValueObject.NewUnixHostname("localhost")
		if err != nil {
			subtest.Fatalf("CreateLocalhostHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver:           nonRoutableTestNetOneIpAddress,
			SecondaryResolver:         nonRoutableTestNetOneIpAddress,
			QueryTimeoutSecs:          1,
			DialTimeoutMs:             500,
			ShouldBypassLocalResolver: true,
		})

		results, lookupErr := lookup.Execute(
			localhostHostname, &tkValueObject.DnsRecordTypeA,
		)
		if lookupErr == nil {
			subtest.Errorf(
				"BypassedLocalhostShouldNotResolveWithoutDnsServer: got %v",
				results,
			)
		}

		if slices.Contains(results, localhostLoopbackIpAddress) {
			subtest.Errorf(
				"BypassedLocalhostNotContainLoopback: %v contains '%s'",
				results, localhostLoopbackIpAddress,
			)
		}
	})

	test.Run("BypassIgnoredForNonIpRecordType", func(subtest *testing.T) {
		nonIpHostname, err := tkValueObject.NewUnixHostname("example.com")
		if err != nil {
			subtest.Fatalf("CreateNonIpHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			ShouldBypassLocalResolver: true,
		})

		results, lookupErr := lookup.Execute(
			nonIpHostname, &tkValueObject.DnsRecordTypeTXT,
		)
		if lookupErr != nil {
			subtest.Fatalf("TxtLookupFailed: %v", lookupErr)
		}
		if len(results) == 0 {
			subtest.Fatalf("TxtLookupReturnedEmpty")
		}
	})
}

func TestDnsLookupDirectResolver(test *testing.T) {
	publicResolver, err := tkValueObject.NewIpAddress("8.8.8.8")
	if err != nil {
		test.Fatalf("CreatePublicResolverIpAddressFailed: %v", err)
	}

	dnsGoogleHostname, err := tkValueObject.NewUnixHostname("dns.google")
	if err != nil {
		test.Fatalf("CreateDnsGoogleHostnameFailed: %v", err)
	}

	lookup := NewDnsLookup(DnsLookupSettings{
		PrimaryResolver:  publicResolver,
		QueryTimeoutSecs: 5,
		DialTimeoutMs:    1000,
	})

	testCaseStructs := []struct {
		name          string
		recordType    tkValueObject.DnsRecordType
		expectError   bool
		expectKnownIp string
	}{
		{"ReturnsARecords", tkValueObject.DnsRecordTypeA, false, "8.8.8.8"},
		{"ReturnsAAAARecords", tkValueObject.DnsRecordTypeAAAA, false, ""},
		{"RejectsUnsupportedRecordType", tkValueObject.DnsRecordTypeMX, true, ""},
	}

	for _, testCase := range testCaseStructs {
		test.Run(testCase.name, func(subtest *testing.T) {
			results, lookupErr := lookup.directIpAddressResolver(
				context.Background(), publicResolver,
				dnsGoogleHostname, testCase.recordType,
			)

			if testCase.expectError {
				if lookupErr == nil {
					subtest.Errorf("ExpectedErrorButGotNone")
				}
				return
			}

			if lookupErr != nil {
				subtest.Fatalf("DirectLookupFailed: %v", lookupErr)
			}
			if len(results) == 0 {
				subtest.Fatalf("DirectLookupReturnedEmpty")
			}

			if testCase.expectKnownIp == "" {
				return
			}

			if !slices.Contains(results, testCase.expectKnownIp) {
				subtest.Errorf(
					"ExpectedIpMissing: got %v, expected to contain '%s'",
					results, testCase.expectKnownIp,
				)
			}
		})
	}
}

func TestDnsMessageValidatorRejectsNonResponseOrMismatchedQuestionWithMatchingId(
	test *testing.T,
) {
	lookup := NewDnsLookup(DnsLookupSettings{})

	transactionId := uint16(0x1234)

	expectedDnsName, nameParseError := dnsmessage.NewName("example.com.")
	if nameParseError != nil {
		test.Fatalf("CreateExpectedNameFailed: %v", nameParseError)
	}
	expectedQuestion := dnsmessage.Question{
		Name:  expectedDnsName,
		Type:  dnsmessage.TypeA,
		Class: dnsmessage.ClassINET,
	}

	test.Run("RejectsNonResponsePacket", func(subtest *testing.T) {
		queryShapedPacket := dnsmessage.Message{
			Header: dnsmessage.Header{
				ID:               transactionId,
				Response:         false,
				RecursionDesired: true,
			},
			Questions: []dnsmessage.Question{expectedQuestion},
		}
		packedBytes, packError := queryShapedPacket.Pack()
		if packError != nil {
			subtest.Fatalf("PackQueryShapedPacketFailed: %v", packError)
		}

		_, validateError := lookup.dnsMessageValidator(
			packedBytes, transactionId, expectedQuestion,
		)
		if validateError != ErrDnsLookupResponseNotResponse {
			subtest.Errorf(
				"ExpectedErrDnsLookupResponseNotResponse: got '%v'",
				validateError,
			)
		}
	})

	test.Run("RejectsMismatchedQuestion", func(subtest *testing.T) {
		attackerDnsName, nameParseError := dnsmessage.NewName("attacker.com.")
		if nameParseError != nil {
			subtest.Fatalf("CreateAttackerNameFailed: %v", nameParseError)
		}
		mismatchedPacket := dnsmessage.Message{
			Header: dnsmessage.Header{
				ID:               transactionId,
				Response:         true,
				RecursionDesired: true,
			},
			Questions: []dnsmessage.Question{{
				Name:  attackerDnsName,
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
			}},
		}
		packedBytes, packError := mismatchedPacket.Pack()
		if packError != nil {
			subtest.Fatalf("PackMismatchedPacketFailed: %v", packError)
		}

		_, validateError := lookup.dnsMessageValidator(
			packedBytes, transactionId, expectedQuestion,
		)
		if validateError != ErrDnsLookupResponseQuestionMismatch {
			subtest.Errorf(
				"ExpectedErrDnsLookupResponseQuestionMismatch: got '%v'",
				validateError,
			)
		}
	})
}
