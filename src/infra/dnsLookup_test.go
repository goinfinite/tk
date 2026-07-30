package tkInfra

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

// 192.0.2.0/24 (TEST-NET-1, RFC 5737) is reserved for documentation;
// traffic to it is never routed, so a resolver there can never reply.
var nonRoutableTestNetOneIpAddress = tkValueObject.IpAddress("192.0.2.1")

const localhostLoopbackIpAddress = "127.0.0.1"

func TestNewDnsLookup(t *testing.T) {
	customPrimaryIpAddress, primaryIpErr := tkValueObject.NewIpAddress("1.1.1.1")
	if primaryIpErr != nil {
		t.Fatalf("CreateCustomPrimaryIpAddressFailed: %v", primaryIpErr)
	}

	customSecondaryIpAddress, secondaryIpErr := tkValueObject.NewIpAddress("9.9.9.9")
	if secondaryIpErr != nil {
		t.Fatalf("CreateCustomSecondaryIpAddressFailed: %v", secondaryIpErr)
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
		t.Run(testCase.name, func(t *testing.T) {
			lookup := NewDnsLookup(testCase.settings)

			if lookup.primaryResolver != testCase.expectedPrimary {
				t.Errorf(
					"WrongPrimaryResolver: expected '%s', got '%s'",
					testCase.expectedPrimary, lookup.primaryResolver,
				)
			}
			if lookup.secondaryResolver != testCase.expectedSecondary {
				t.Errorf(
					"WrongSecondaryResolver: expected '%s', got '%s'",
					testCase.expectedSecondary, lookup.secondaryResolver,
				)
			}
			if lookup.queryTimeoutSecs != testCase.expectedQueryTimeout {
				t.Errorf(
					"WrongQueryTimeoutSecs: expected %d, got %d",
					testCase.expectedQueryTimeout, lookup.queryTimeoutSecs,
				)
			}
			if lookup.dialTimeoutMs != testCase.expectedDialTimeout {
				t.Errorf(
					"WrongDialTimeoutMs: expected %d, got %d",
					testCase.expectedDialTimeout, lookup.dialTimeoutMs,
				)
			}
			if lookup.shouldBypassLocalResolver != testCase.expectedBypass {
				t.Errorf(
					"WrongShouldBypassLocalResolver: expected %t, got %t",
					testCase.expectedBypass, lookup.shouldBypassLocalResolver,
				)
			}
		})
	}
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
		t.Run(testCase.name, func(t *testing.T) {
			results, lookupErr := dnsLookup.Execute(
				testCase.hostname, testCase.recordType,
			)
			if lookupErr != nil {
				t.Errorf(
					"UnexpectedError: '%s' [%s]",
					lookupErr.Error(), testCase.name,
				)
			}
			if len(results) == 0 {
				t.Errorf("NoResultsReturned: %s", testCase.name)
			}
		})
	}

	t.Run("UnreachableResolversFailLookup", func(t *testing.T) {
		uniqueHostname, err := tkValueObject.NewUnixHostname(fmt.Sprintf(
			"resolver-isolation-%d.invalid", time.Now().UnixNano(),
		))
		if err != nil {
			t.Fatalf("CreateUniqueHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			PrimaryResolver:   nonRoutableTestNetOneIpAddress,
			SecondaryResolver: nonRoutableTestNetOneIpAddress,
			QueryTimeoutSecs:  2,
			DialTimeoutMs:     500,
		})

		results, lookupErr := lookup.Execute(uniqueHostname, nil)
		if lookupErr == nil {
			t.Errorf(
				"NonRoutableResolversShouldHaveFailed: got %v",
				results,
			)
		}
	})

	t.Run("LocalhostReturnsLoopbackViaLocalResolver", func(t *testing.T) {
		localhostHostname, err := tkValueObject.NewUnixHostname("localhost")
		if err != nil {
			t.Fatalf("CreateLocalhostHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{})

		results, lookupErr := lookup.Execute(
			localhostHostname, &tkValueObject.DnsRecordTypeA,
		)
		if lookupErr != nil {
			t.Fatalf("LocalhostLookupFailed: %v", lookupErr)
		}

		foundLoopback := false
		for _, result := range results {
			if result == localhostLoopbackIpAddress {
				foundLoopback = true
				break
			}
		}
		if !foundLoopback {
			t.Errorf(
				"LocalResolverShouldReturnLoopback: got %v, expected to contain '%s'",
				results, localhostLoopbackIpAddress,
			)
		}
	})

	t.Run("LocalhostBypassedSkipsLocalLookup", func(t *testing.T) {
		localhostHostname, err := tkValueObject.NewUnixHostname("localhost")
		if err != nil {
			t.Fatalf("CreateLocalhostHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			ShouldBypassLocalResolver: true,
		})

		results, lookupErr := lookup.Execute(
			localhostHostname, &tkValueObject.DnsRecordTypeA,
		)
		if lookupErr != nil &&
			lookupErr.Error() != ErrDnsLookupResponseNameError.Error() {
			t.Fatalf("BypassedLocalhostLookupFailed: %v", lookupErr)
		}

		for _, result := range results {
			if result == localhostLoopbackIpAddress {
				t.Errorf(
					"BypassedLocalhostNotContainLoopback: %v contains '%s'",
					results, localhostLoopbackIpAddress,
				)
				break
			}
		}
	})

	t.Run("BypassIgnoredForNonIpRecordType", func(t *testing.T) {
		nonIpHostname, err := tkValueObject.NewUnixHostname("example.com")
		if err != nil {
			t.Fatalf("CreateNonIpHostnameFailed: %v", err)
		}

		lookup := NewDnsLookup(DnsLookupSettings{
			ShouldBypassLocalResolver: true,
		})

		results, lookupErr := lookup.Execute(
			nonIpHostname, &tkValueObject.DnsRecordTypeTXT,
		)
		if lookupErr != nil {
			t.Fatalf("TxtLookupFailed: %v", lookupErr)
		}
		if len(results) == 0 {
			t.Fatalf("TxtLookupReturnedEmpty")
		}
	})
}

func TestDnsLookupDirectResolver(t *testing.T) {
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
		t.Run(testCase.name, func(t *testing.T) {
			results, lookupErr := lookup.directIpAddressResolver(
				context.Background(), publicResolver,
				dnsGoogleHostname, testCase.recordType,
			)

			if testCase.expectError {
				if lookupErr == nil {
					t.Errorf("ExpectedErrorButGotNone")
				}
				return
			}

			if lookupErr != nil {
				t.Fatalf("DirectLookupFailed: %v", lookupErr)
			}
			if len(results) == 0 {
				t.Fatalf("DirectLookupReturnedEmpty")
			}

			if testCase.expectKnownIp == "" {
				return
			}

			foundKnownAddress := false
			for _, result := range results {
				if result == testCase.expectKnownIp {
					foundKnownAddress = true
					break
				}
			}
			if !foundKnownAddress {
				t.Errorf(
					"ExpectedIpMissing: got %v, expected to contain '%s'",
					results, testCase.expectKnownIp,
				)
			}
		})
	}
}

func TestDnsMessageValidatorRejectsNonResponseOrMismatchedQuestionWithMatchingId(
	t *testing.T,
) {
	lookup := NewDnsLookup(DnsLookupSettings{})

	transactionId := uint16(0x1234)

	expectedDnsName, nameParseError := dnsmessage.NewName("example.com.")
	if nameParseError != nil {
		t.Fatalf("CreateExpectedNameFailed: %v", nameParseError)
	}
	expectedQuestion := dnsmessage.Question{
		Name:  expectedDnsName,
		Type:  dnsmessage.TypeA,
		Class: dnsmessage.ClassINET,
	}

	t.Run("RejectsNonResponsePacket", func(t *testing.T) {
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
			t.Fatalf("PackQueryShapedPacketFailed: %v", packError)
		}

		_, validateError := lookup.dnsMessageValidator(
			packedBytes, transactionId, expectedQuestion,
		)
		if validateError != ErrDnsLookupResponseNotResponse {
			t.Errorf(
				"ExpectedErrDnsLookupResponseNotResponse: got '%v'",
				validateError,
			)
		}
	})

	t.Run("RejectsMismatchedQuestion", func(t *testing.T) {
		attackerDnsName, nameParseError := dnsmessage.NewName("attacker.com.")
		if nameParseError != nil {
			t.Fatalf("CreateAttackerNameFailed: %v", nameParseError)
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
			t.Fatalf("PackMismatchedPacketFailed: %v", packError)
		}

		_, validateError := lookup.dnsMessageValidator(
			packedBytes, transactionId, expectedQuestion,
		)
		if validateError != ErrDnsLookupResponseQuestionMismatch {
			t.Errorf(
				"ExpectedErrDnsLookupResponseQuestionMismatch: got '%v'",
				validateError,
			)
		}
	})
}
