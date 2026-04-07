package tkInfra

import (
	"net/http"
	"testing"
)

func TestRequesterIpExtractor(t *testing.T) {
	testCaseStructs := []struct {
		description        string
		headerChainEnvVal  string
		trustedCidrsEnvVal string
		remoteAddr         string
		requestHeaders     map[string]string
		expectedIp         string
		expectError        bool
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
			description:        "XffFromUntrustedRemoteExtractsHeaderValue",
			headerChainEnvVal:  "X-Forwarded-For",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.20:5555",
			requestHeaders:     map[string]string{"X-Forwarded-For": "1.2.3.4"},
			expectedIp:         "1.2.3.4",
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
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			t.Setenv(ipExtractHeaderEnvVarName, testCase.headerChainEnvVal)
			t.Setenv(TrustedCidrsEnvVarName, testCase.trustedCidrsEnvVal)

			extractor := NewRequesterIpExtractor()

			httpRequest, _ := http.NewRequest(http.MethodGet, "/", nil)
			httpRequest.RemoteAddr = testCase.remoteAddr
			for headerName, headerVal := range testCase.requestHeaders {
				httpRequest.Header.Set(headerName, headerVal)
			}

			actualIpAddress, extractionErr := extractor.Execute(httpRequest)
			if testCase.expectError {
				if extractionErr == nil {
					t.Errorf("MissingExpectedError")
				}
				return
			}
			if extractionErr != nil {
				t.Errorf("UnexpectedError: '%s'", extractionErr.Error())
				return
			}
			if actualIpAddress.String() != testCase.expectedIp {
				t.Errorf(
					"IpAddressMismatch: got='%s', want='%s'",
					actualIpAddress.String(), testCase.expectedIp,
				)
			}
		})
	}
}
