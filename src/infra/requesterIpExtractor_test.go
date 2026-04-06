package tkInfra

import (
	"net/http"
	"testing"
)

func TestExtractRequesterIpAddress(t *testing.T) {
	testCaseStructs := []struct {
		description         string
		disableTrustEnvVal  string
		trustedCidrsEnvVal  string
		remoteAddr          string
		xffHeader           string
		expectedIp          string
		skipExpectedIpCheck bool
	}{
		{
			description:        "DefaultXffTrustExtractsClientIpBehindTrustedLoopbackProxy",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			xffHeader:          "203.0.113.5",
			expectedIp:         "203.0.113.5",
		},
		{
			description:        "CustomCidrIsTrusted",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "198.51.100.0/24",
			remoteAddr:         "198.51.100.1:5678",
			xffHeader:          "203.0.113.7",
			expectedIp:         "203.0.113.7",
		},
		{
			description:        "DirectExtractionWhenTrustDisabled",
			disableTrustEnvVal: "true",
			trustedCidrsEnvVal: "",
			remoteAddr:         "10.0.0.5:9000",
			xffHeader:          "203.0.113.9",
			expectedIp:         "10.0.0.5",
		},
		{
			description:        "EndToEndXffWithCustomCidr",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "192.0.2.0/24",
			remoteAddr:         "192.0.2.10:4321",
			xffHeader:          "198.51.100.42",
			expectedIp:         "198.51.100.42",
		},
		{
			description:        "InvalidCidrInEnvVarDoesNotCrash",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "notacidr,10.0.0.0/8",
			remoteAddr:         "10.0.0.1:1111",
			xffHeader:          "203.0.113.1",
			expectedIp:         "203.0.113.1",
		},
		{
			description:        "MissingXffFallsBackToRemoteAddr",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.55:2222",
			xffHeader:          "",
			expectedIp:         "203.0.113.55",
		},
		{
			description:        "XffInjectionFromUntrustedRemoteRejected",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.20:5555",
			xffHeader:          "1.2.3.4",
			expectedIp:         "203.0.113.20",
		},
		{
			description:        "Ipv6InXff",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "127.0.0.1:1234",
			xffHeader:          "2001:db8::1",
			expectedIp:         "2001:db8::1",
		},
		{
			description:        "XffChainWithTrustedProxyReturnsNearestUntrusted",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "10.0.0.1:1234",
			xffHeader:          "1.1.1.1, 203.0.113.1",
			expectedIp:         "203.0.113.1",
		},
		{
			description:        "UntrustedProxyInChainReturnsRemoteAddr",
			disableTrustEnvVal: "",
			trustedCidrsEnvVal: "",
			remoteAddr:         "203.0.113.99:9999",
			xffHeader:          "10.0.0.1, 203.0.113.50",
			expectedIp:         "203.0.113.99",
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			t.Setenv("IP_EXTRACT_DISABLE_TRUST", testCase.disableTrustEnvVal)
			t.Setenv(TrustedCidrsEnvVarName, testCase.trustedCidrsEnvVal)

			extractor := NewRequesterIpExtractor()

			httpRequest, _ := http.NewRequest(http.MethodGet, "/", nil)
			httpRequest.RemoteAddr = testCase.remoteAddr
			if testCase.xffHeader != "" {
				httpRequest.Header.Set("X-Forwarded-For", testCase.xffHeader)
			}

			actualIpAddress := extractor.Execute(httpRequest)
			if actualIpAddress != testCase.expectedIp {
				t.Errorf(
					"IpAddressMismatch: got='%s', want='%s'",
					actualIpAddress, testCase.expectedIp,
				)
			}
		})
	}
}
