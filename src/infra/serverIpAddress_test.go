package tkInfra

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestServerIpAddress(t *testing.T) {
	t.Run("FunctionalTest", func(t *testing.T) {
		t.Setenv(ServerPublicIpAddressEnvVarName, "")

		_, err := ReadServerPrivateIpAddress()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		_, err = ReadServerPublicIpAddress()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}
	})

	t.Run("PublicIpAddressEnvVarTakesPriority", func(t *testing.T) {
		rawEnvIpAddress := "203.0.113.42"
		t.Setenv(ServerPublicIpAddressEnvVarName, rawEnvIpAddress)

		ipAddress, err := ReadServerPublicIpAddress()
		if err != nil {
			t.Fatalf("UnexpectedError: '%s'", err.Error())
		}

		if ipAddress.String() != rawEnvIpAddress {
			t.Errorf(
				"ExpectedIpFromEnvVar '%s', got '%s'",
				rawEnvIpAddress, ipAddress.String(),
			)
		}
	})
}

type publicIpResolverResponseStub struct {
	statusCode int
	body       string
}

func TestPublicIpAddressResolver(t *testing.T) {
	testCaseStructs := []struct {
		description                   string
		publicIpResolverResponseStubs []publicIpResolverResponseStub
		expectedIpAddress             string
		expectError                   bool
		expectedErrorIn               string
	}{
		{
			description: "SuccessFromSingleEndpoint",
			publicIpResolverResponseStubs: []publicIpResolverResponseStub{
				{statusCode: 200, body: "203.0.113.10\n"},
			},
			expectedIpAddress: "203.0.113.10",
		},
		{
			description: "FallsThroughOnHttpFailure",
			publicIpResolverResponseStubs: []publicIpResolverResponseStub{
				{statusCode: 500, body: "upstream-down"},
				{statusCode: 200, body: "203.0.113.99\n"},
			},
			expectedIpAddress: "203.0.113.99",
		},
		{
			description: "FallsThroughOnMalformedResponse",
			publicIpResolverResponseStubs: []publicIpResolverResponseStub{
				{statusCode: 200, body: "definitely-not-an-ip\n"},
				{statusCode: 200, body: "203.0.113.50\n"},
			},
			expectedIpAddress: "203.0.113.50",
		},
		{
			description: "TrimsWhitespaceFromResponse",
			publicIpResolverResponseStubs: []publicIpResolverResponseStub{
				{statusCode: 200, body: "  203.0.113.77\n\n"},
			},
			expectedIpAddress: "203.0.113.77",
		},
		{
			description: "ErrorsWhenResponseBodyIsEmpty",
			publicIpResolverResponseStubs: []publicIpResolverResponseStub{
				{statusCode: 200, body: ""},
			},
			expectError:     true,
			expectedErrorIn: "PublicIpAddressNotFound",
		},
		{
			description: "ReturnsErrorWhenAllEndpointsFail",
			publicIpResolverResponseStubs: []publicIpResolverResponseStub{
				{statusCode: 500, body: ""},
				{statusCode: 500, body: ""},
			},
			expectError:     true,
			expectedErrorIn: "PublicIpAddressNotFound",
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			var resolverUrls []tkValueObject.Url
			for _, responseStub := range testCase.publicIpResolverResponseStubs {
				testServer := httptest.NewServer(http.HandlerFunc(
					func(writeResponse http.ResponseWriter, readRequest *http.Request) {
						writeResponse.Header().Set("Content-Type", "text/plain")
						writeResponse.WriteHeader(responseStub.statusCode)
						if responseStub.body != "" {
							_, _ = writeResponse.Write([]byte(responseStub.body))
						}
					},
				))
				defer testServer.Close()
				resolverUrls = append(
					resolverUrls, tkValueObject.Url(testServer.URL),
				)
			}

			resolver := &PublicIpAddressResolver{
				httpClient:   &http.Client{Timeout: 5 * time.Second},
				resolverUrls: resolverUrls,
			}

			ipAddress, err := resolver.Resolve()

			if testCase.expectError {
				if err == nil {
					t.Fatalf("ExpectedErrorButGotNone")
				}
				if !strings.Contains(err.Error(), testCase.expectedErrorIn) {
					t.Errorf(
						"ExpectedErrorToContain '%s', got '%s'",
						testCase.expectedErrorIn, err.Error(),
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if ipAddress.String() != testCase.expectedIpAddress {
				t.Errorf(
					"ExpectedIp '%s', got '%s'",
					testCase.expectedIpAddress, ipAddress.String(),
				)
			}
		})
	}

	t.Run("SpreadsLoadAcrossEndpoints", func(t *testing.T) {
		ipAddressesByServer := []string{
			"203.0.113.1",
			"203.0.113.2",
			"203.0.113.3",
		}

		var testServers []*httptest.Server
		for _, ipAddress := range ipAddressesByServer {
			testServer := httptest.NewServer(http.HandlerFunc(
				func(writeResponse http.ResponseWriter, readRequest *http.Request) {
					writeResponse.Header().Set("Content-Type", "text/plain")
					writeResponse.WriteHeader(http.StatusOK)
					_, _ = writeResponse.Write([]byte(ipAddress + "\n"))
				},
			))
			defer testServer.Close()
			testServers = append(testServers, testServer)
		}

		var resolverUrls []tkValueObject.Url
		for _, testServer := range testServers {
			resolverUrls = append(resolverUrls, tkValueObject.Url(testServer.URL))
		}
		resolver := &PublicIpAddressResolver{
			httpClient:   &http.Client{Timeout: 5 * time.Second},
			resolverUrls: resolverUrls,
		}

		seenIpAddresses := make(map[string]bool)
		for callIndex := 0; callIndex < 60; callIndex++ {
			ipAddress, err := resolver.Resolve()
			if err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			seenIpAddresses[ipAddress.String()] = true
		}

		for _, expectedIpAddress := range ipAddressesByServer {
			if !seenIpAddresses[expectedIpAddress] {
				t.Errorf(
					"ExpectedIpAddressSeenAtLeastOnce '%s'",
					expectedIpAddress,
				)
			}
		}
	})

	t.Run("AvailableResolverUrlsAreUsedByPublicConstructor", func(t *testing.T) {
		resolver := NewPublicIpAddressResolver()
		if len(resolver.resolverUrls) != len(publicIpResolverAvailableUrls) {
			t.Errorf(
				"ExpectedAvailableResolverUrlsCount %d, got %d",
				len(publicIpResolverAvailableUrls), len(resolver.resolverUrls),
			)
		}
		for index, expectedUrl := range publicIpResolverAvailableUrls {
			if resolver.resolverUrls[index] != expectedUrl {
				t.Errorf(
					"ExpectedResolverUrlAtIndex %d to be '%s', got '%s'",
					index, expectedUrl.String(),
					resolver.resolverUrls[index].String(),
				)
			}
		}
	})
}
