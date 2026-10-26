package tkPresentation

import (
	"net/http"
	"testing"

	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestRequesterIpExtractor(test *testing.T) {
	testCaseStructs := []struct {
		description               string
		headerChainEnvVal         string
		confirmationHeadersEnvVal string
		trustedCidrsEnvVal        string
		remoteAddr                string
		requestHeaders            map[string]string
		expectedIp                string
		expectError               bool
	}{
		{
			description:        "XffInChainExtractsRightmostUntrusted",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "203.0.113.5"},
			expectedIp:         "203.0.113.5",
		},
		{
			description:        "DefaultChainBehaviorUsesXffFirst",
			headerChainEnvVal:  "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "203.0.113.10"},
			expectedIp:         "203.0.113.10",
		},
		{
			description:               "ConfirmationGateDisabledByEmptyConfig",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders:            map[string]string{"X-Forwarded-For": "203.0.113.11"},
			expectedIp:                "203.0.113.11",
		},
		{
			description:               "ConfirmationGateAbsentHeaderFallsBackToRemoteAddr",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders:            map[string]string{"X-Forwarded-For": "203.0.113.12"},
			expectedIp:                "127.0.0.1",
		},
		{
			description:               "ConfirmationGatePresentHeaderUnlocksForwardedHeader",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.13",
				"X-Edge-Confirmation": "present",
			},
			expectedIp: "203.0.113.13",
		},
		{
			description:               "ConfirmationGateMatchingValueUnlocksForwardedHeader",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=shared-secret",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.14",
				"X-Edge-Confirmation": "shared-secret",
			},
			expectedIp: "203.0.113.14",
		},
		{
			description:               "ConfirmationGateHeaderNameIsCaseInsensitive",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "x-edge-confirmation=shared-secret",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.27",
				"X-Edge-Confirmation": "shared-secret",
			},
			expectedIp: "203.0.113.27",
		},
		{
			description:               "ConfirmationGateWrongValueFallsBackToRemoteAddr",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=shared-secret",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.15",
				"X-Edge-Confirmation": "wrong-secret",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:               "ConfirmationGateMissingExpectedValueFallsBackToRemoteAddr",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=shared-secret",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders:            map[string]string{"X-Forwarded-For": "203.0.113.16"},
			expectedIp:                "127.0.0.1",
		},
		{
			description:               "ConfirmationGateEmptyHeaderValueFailsClosed",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.23",
				"X-Edge-Confirmation": "",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:               "ConfirmationGateAnyConfiguredHeaderSatisfies",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation,X-Edge-Signature",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":  "203.0.113.17",
				"X-Edge-Signature": "present",
			},
			expectedIp: "203.0.113.17",
		},
		{
			description:               "ConfirmationGateInvalidConfigFailsClosed",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X.Edge.Confirmation,=broken",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.18",
				"X.Edge.Confirmation": "present",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:               "ConfirmationGateInvalidEntryInvalidatesList",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=shared-secret,X.Edge.Confirmation",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.24",
				"X-Edge-Confirmation": "shared-secret",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:               "ConfirmationGateInvalidValueFailsClosed",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=invalid\nvalue",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.28",
				"X-Edge-Confirmation": "present",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:               "ConfirmationGateEmptyExpectedValueRequiresPresence",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.19",
				"X-Edge-Confirmation": "present",
			},
			expectedIp: "203.0.113.19",
		},
		{
			description:               "ConfirmationGateMixedConfigFailsClosed",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Presence,X-Edge-Secret=shared-secret",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For": "203.0.113.25",
				"X-Edge-Presence": "present",
				"X-Edge-Secret":   "shared-secret",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:               "ConfirmationGateAnyValueCheckedHeaderSatisfies",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Secret-A=secret-a,X-Edge-Secret-B=secret-b",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For": "203.0.113.26",
				"X-Edge-Secret-B": "secret-b",
			},
			expectedIp: "203.0.113.26",
		},
		{
			description:               "ConfirmationGateTrimsWhitespace",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: " X-Edge-Confirmation = shared-secret , X-Edge-Other = other-secret ",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.21",
				"X-Edge-Confirmation": "shared-secret",
			},
			expectedIp: "203.0.113.21",
		},
		{
			description:               "ConfirmationGateValueMayContainEquals",
			headerChainEnvVal:         "X-Forwarded-For",
			confirmationHeadersEnvVal: "X-Edge-Confirmation=abc=def",
			trustedCidrsEnvVal:        "",
			remoteAddr:                "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For":     "203.0.113.22",
				"X-Edge-Confirmation": "abc=def",
			},
			expectedIp: "203.0.113.22",
		},
		{
			description:        "HeaderChainFallbackToXffWhenXRealIpEmpty",
			headerChainEnvVal:  "X-Real-IP,X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "203.0.113.20"},
			expectedIp:         "203.0.113.20",
		},
		{
			description:        "HeaderChainPrefersXRealIpOverXff",
			headerChainEnvVal:  "X-Real-IP,X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Real-IP":       "203.0.113.30",
				"X-Forwarded-For": "203.0.113.31",
			},
			expectedIp: "203.0.113.30",
		},
		{
			description:        "XffChainWithMultipleProxiesReturnsNearestUntrusted",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "10.0.0.1:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "1.1.1.1, 203.0.113.1"},
			expectedIp:         "203.0.113.1",
		},
		{
			description:        "CustomHeaderInChainReadWhenDirectConnectionTrusted",
			headerChainEnvVal:  "CF-Connecting-IP,X-Real-IP",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:8080",
			requestHeaders:     map[string]string{"CF-Connecting-IP": "203.0.113.40"},
			expectedIp:         "203.0.113.40",
		},
		{
			description:        "TrustedCidrAppliesToCustomHeader",
			headerChainEnvVal:  "X-Real-IP",
			trustedCidrsEnvVal: "198.51.100.0/24",
			remoteAddr:         "198.51.100.5:9000",
			requestHeaders:     map[string]string{"X-Real-IP": "203.0.113.50"},
			expectedIp:         "203.0.113.50",
		},
		{
			description:        "AllHeadersEmptyInChainReturnsDirect",
			headerChainEnvVal:  "X-Real-IP,CF-Connecting-IP",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:3333",
			requestHeaders:     map[string]string{},
			expectedIp:         "127.0.0.1",
		},
		{
			description:        "MalformedIpInCustomHeaderFallsBackToDirect",
			headerChainEnvVal:  "X-Real-IP",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:4444",
			requestHeaders:     map[string]string{"X-Real-IP": "not-an-ip"},
			expectedIp:         "127.0.0.1",
		},
		{
			description:        "UntrustedRemoteIgnoresXffEntries",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.20:5555",
			requestHeaders:     map[string]string{"X-Forwarded-For": "1.2.3.4"},
			expectedIp:         "203.0.113.20",
		},
		{
			description:        "Ipv6RemoteAddrTrustedLoopback",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "[::1]:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "2001:db8::1"},
			expectedIp:         "2001:db8::1",
		},
		{
			description:        "InvalidCidrInEnvVarDoesNotCrash",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "notacidr,10.0.0.0/8",
			remoteAddr:         "10.0.0.1:1111",
			requestHeaders:     map[string]string{"X-Forwarded-For": "203.0.113.1"},
			expectedIp:         "203.0.113.1",
		},
		{
			description:        "PrivateNetworkRemoteAddrTrustedForCustomHeader",
			headerChainEnvVal:  "X-Real-IP",
			trustedCidrsEnvVal: "",
			remoteAddr:         "192.168.1.1:7777",
			requestHeaders:     map[string]string{"X-Real-IP": "203.0.113.60"},
			expectedIp:         "203.0.113.60",
		},
		{
			description:        "XffChainSkipsTrustedProxyInMiddle",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "10.0.0.0/8",
			remoteAddr:         "10.0.0.2:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "1.2.3.4, 10.0.0.1"},
			expectedIp:         "1.2.3.4",
		},
		{
			description:        "FullyTrustedChainWithTrustedDirectPeerReturnsOriginalRequester",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "172.17.0.1:54321",
			requestHeaders:     map[string]string{"X-Forwarded-For": "127.0.0.1, 172.17.0.1"},
			expectedIp:         "127.0.0.1",
		},
		{
			description:        "FullyTrustedChainWithUntrustedDirectPeerReturnsDirect",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.99:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "127.0.0.1, 10.0.0.1"},
			expectedIp:         "203.0.113.99",
		},
		{
			description:        "FullyTrustedChainSkipsMalformedEntryForOriginalRequester",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "172.17.0.1:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "not-an-ip, 127.0.0.1"},
			expectedIp:         "127.0.0.1",
		},
		{
			description:        "FullyTrustedChainsReturnOriginalRequesterOfFirstHeader",
			headerChainEnvVal:  "X-Forwarded-For,X-Real-IP",
			trustedCidrsEnvVal: "",
			remoteAddr:         "172.17.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For": "127.0.0.1",
				"X-Real-IP":       "10.0.0.9",
			},
			expectedIp: "127.0.0.1",
		},
		{
			description:        "SingleTrustedXffEntryWithTrustedDirectPeerReturnsIt",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "172.17.0.1:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "192.168.1.50"},
			expectedIp:         "192.168.1.50",
		},
		{
			description:        "FullyTrustedIpv6ChainReturnsOriginalRequester",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "[fd00::1]:1234",
			requestHeaders:     map[string]string{"X-Forwarded-For": "::1, fd00::2"},
			expectedIp:         "::1",
		},
		{
			description:        "XffChainMalformedEntrySkipped",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			requestHeaders: map[string]string{
				"X-Forwarded-For": "1.2.3.4, not-an-ip, 203.0.113.1",
			},
			expectedIp: "203.0.113.1",
		},
		{
			description:        "DirectKeywordInChainParsesRemoteAddr",
			headerChainEnvVal:  "Direct",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.99:5555",
			requestHeaders:     map[string]string{},
			expectedIp:         "203.0.113.99",
		},
		{
			description:        "RemoteAddrKeywordInChainAfterEmptyHeaders",
			headerChainEnvVal:  "X-Real-IP,RemoteAddr",
			trustedCidrsEnvVal: "",
			remoteAddr:         "198.51.100.1:8080",
			requestHeaders:     map[string]string{},
			expectedIp:         "198.51.100.1",
		},
		{
			description:        "RemoteAddrKeywordWithTrustedPeerParsesRemoteAddr",
			headerChainEnvVal:  "X-Real-IP,RemoteAddr",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:8080",
			requestHeaders:     map[string]string{"X-Real-IP": "10.0.0.1"},
			expectedIp:         "127.0.0.1",
		},
		{
			description:        "UnparsableRemoteAddrFails",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "not-an-ip",
			requestHeaders:     map[string]string{"X-Forwarded-For": "1.2.3.4"},
			expectError:        true,
		},
	}

	for _, testCase := range testCaseStructs {
		test.Run(testCase.description, func(subtest *testing.T) {
			subtest.Setenv(ipExtractHeaderEnvVarName, testCase.headerChainEnvVal)
			subtest.Setenv(
				ipExtractConfirmationHeadersEnvVarName,
				testCase.confirmationHeadersEnvVal,
			)
			subtest.Setenv(tkInfra.TrustedCidrsEnvVarName, testCase.trustedCidrsEnvVal)

			extractor := NewRequesterIpExtractor()

			httpRequest, _ := http.NewRequest(http.MethodGet, "/", nil)
			httpRequest.RemoteAddr = testCase.remoteAddr
			for headerName, headerVal := range testCase.requestHeaders {
				httpRequest.Header.Set(headerName, headerVal)
			}

			actualIpAddress, extractionErr := extractor.Execute(httpRequest)
			if testCase.expectError {
				if extractionErr == nil {
					subtest.Errorf("MissingExpectedError")
				}
				return
			}
			if extractionErr != nil {
				subtest.Errorf("UnexpectedError: '%s'", extractionErr.Error())
				return
			}
			if actualIpAddress.String() != testCase.expectedIp {
				subtest.Errorf(
					"IpAddressMismatch: got='%s', want='%s'",
					actualIpAddress.String(), testCase.expectedIp,
				)
			}
		})
	}
}

func TestRequesterIpExtractor_InvalidHeaderConfig_FailsClosed(
	test *testing.T,
) {
	test.Setenv(ipExtractHeaderEnvVarName, "!!!invalid!!!")
	test.Setenv(ipExtractConfirmationHeadersEnvVarName, "")
	test.Setenv(tkInfra.TrustedCidrsEnvVarName, "")

	extractor := NewRequesterIpExtractor()
	if len(extractor.extractionHeaders) != 0 {
		test.Errorf(
			"ExpectedEmptyHeaders: got=%d",
			len(extractor.extractionHeaders),
		)
	}
}

func TestRequesterIpExtractor_HeaderChainUsesAllValues(
	test *testing.T,
) {
	test.Setenv(ipExtractHeaderEnvVarName, "X-Forwarded-For")
	test.Setenv(ipExtractConfirmationHeadersEnvVarName, "")
	test.Setenv(tkInfra.TrustedCidrsEnvVarName, "")

	extractor := NewRequesterIpExtractor()

	httpRequest, _ := http.NewRequest(http.MethodGet, "/", nil)
	httpRequest.RemoteAddr = "127.0.0.1:1234"
	httpRequest.Header.Add("X-Forwarded-For", "1.1.1.1")
	httpRequest.Header.Add("X-Forwarded-For", "203.0.113.5")

	actualIpAddress, extractionErr := extractor.Execute(httpRequest)
	if extractionErr != nil {
		test.Errorf("UnexpectedError: '%s'", extractionErr.Error())
		return
	}
	if actualIpAddress.String() != "203.0.113.5" {
		test.Errorf(
			"IpAddressMismatch: got='%s', want='203.0.113.5'",
			actualIpAddress.String(),
		)
	}
}
