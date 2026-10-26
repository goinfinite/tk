package tkPresentation

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
)

const (
	ipExtractHeaderEnvVarName              string = "IP_EXTRACT_HEADER"
	ipExtractConfirmationHeadersEnvVarName string = "IP_EXTRACT_CONFIRMATION_HEADERS"
	ipExtractHeaderDefaults                string = "X-Forwarded-For,X-Real-IP"
	ipExtractDirectKeyword                 string = "Direct"
	ipExtractRemoteAddrKeyword             string = "RemoteAddr"
)

type RequesterIpExtractor struct {
	extractionHeaders   []tkValueObject.HttpHeader
	confirmationHeaders map[tkValueObject.HttpHeader]tkValueObject.HttpHeaderValue
	trustedCidrBlocks   []tkValueObject.CidrBlock
}

func NewRequesterIpExtractor() RequesterIpExtractor {
	cidrBlocks, cidrsReadingErr := tkInfra.TrustedCidrsReader()
	if cidrsReadingErr != nil {
		slog.Debug(
			"TrustedCidrsReaderError",
			slog.String("err", cidrsReadingErr.Error()),
		)
	}

	extractor := RequesterIpExtractor{
		trustedCidrBlocks: cidrBlocks,
	}
	extractor.extractionHeaders = extractor.readExtractionHeaders()
	extractor.confirmationHeaders = extractor.readConfirmationHeaders()

	return extractor
}

func (RequesterIpExtractor) readExtractionHeaders() []tkValueObject.HttpHeader {
	rawEnvVal, envVarWasSet := os.LookupEnv(ipExtractHeaderEnvVarName)
	if !envVarWasSet || rawEnvVal == "" {
		rawEnvVal = ipExtractHeaderDefaults
	}

	var extractionHeaders []tkValueObject.HttpHeader
	for rawHeader := range strings.SplitSeq(rawEnvVal, ",") {
		httpHeader, err := tkValueObject.NewHttpHeader(rawHeader)
		if err != nil {
			slog.Debug(
				"InvalidIpExtractHeaderName",
				slog.String("headerName", rawHeader),
			)
			continue
		}
		extractionHeaders = append(
			extractionHeaders, httpHeader,
		)
	}

	if !envVarWasSet && len(extractionHeaders) == 0 {
		for rawHeader := range strings.SplitSeq(
			ipExtractHeaderDefaults, ",",
		) {
			httpHeader, _ := tkValueObject.NewHttpHeader(rawHeader)
			extractionHeaders = append(
				extractionHeaders, httpHeader,
			)
		}
	}

	return extractionHeaders
}

func (RequesterIpExtractor) readConfirmationHeaders() map[tkValueObject.HttpHeader]tkValueObject.HttpHeaderValue {
	rawHeadersList := strings.TrimSpace(
		os.Getenv(ipExtractConfirmationHeadersEnvVarName),
	)
	if rawHeadersList == "" {
		return nil
	}

	hasPresenceOnlyHeader := false
	hasValueCheckedHeader := false
	confirmationHeaders := map[tkValueObject.HttpHeader]tkValueObject.HttpHeaderValue{}
	for headerIndex, rawHeader := range strings.Split(rawHeadersList, ",") {
		rawHeaderName, rawExpectedValue, _ := strings.Cut(rawHeader, "=")

		header, headerErr := tkValueObject.NewHttpHeader(rawHeaderName)
		if headerErr != nil {
			slog.Error(
				"InvalidIpExtractConfirmationHeaderName",
				slog.Int("headerIndex", headerIndex),
			)
			return map[tkValueObject.HttpHeader]tkValueObject.HttpHeaderValue{}
		}

		expectedValue, expectedValueErr := tkValueObject.NewHttpHeaderValue(
			rawExpectedValue,
		)
		if expectedValueErr != nil {
			slog.Error(
				"InvalidIpExtractConfirmationHeaderValue",
				slog.Int("headerIndex", headerIndex),
			)
			return map[tkValueObject.HttpHeader]tkValueObject.HttpHeaderValue{}
		}

		isExpectedValueRequired := expectedValue.String() != ""
		hasPresenceOnlyHeader = hasPresenceOnlyHeader ||
			!isExpectedValueRequired
		hasValueCheckedHeader = hasValueCheckedHeader ||
			isExpectedValueRequired
		confirmationHeaders[header.ReadAsLowerCase()] = expectedValue
	}

	if hasPresenceOnlyHeader && hasValueCheckedHeader {
		slog.Error("MixedIpExtractConfirmationHeaders")
		return map[tkValueObject.HttpHeader]tkValueObject.HttpHeaderValue{}
	}

	return confirmationHeaders
}

func (extractor RequesterIpExtractor) remoteAddrParser(
	remoteAddr string,
) (ipAddress tkValueObject.IpAddress, err error) {
	directIp, _, splitErr := net.SplitHostPort(remoteAddr)
	if splitErr != nil {
		directIp = remoteAddr
	}

	ipAddress, err = tkValueObject.NewIpAddress(directIp)
	if err != nil {
		return ipAddress, errors.New("UnparsableRemoteAddr")
	}

	return ipAddress, nil
}

func (extractor RequesterIpExtractor) isIpTrusted(
	ipAddress tkValueObject.IpAddress,
) bool {
	if ipAddress.IsLocal() || ipAddress.IsPrivate() || ipAddress.IsLinkLocal() {
		return true
	}

	for _, cidrBlock := range extractor.trustedCidrBlocks {
		if cidrBlock.Contains(ipAddress) {
			return true
		}
	}

	return false
}

func (extractor RequesterIpExtractor) headerEntriesParser(
	httpRequest *http.Request,
	extractionHeader tkValueObject.HttpHeader,
) (parsedEntries []tkValueObject.IpAddress) {
	headerValues := httpRequest.Header.Values(
		extractionHeader.String(),
	)

	var rawEntries []string
	for _, headerValue := range headerValues {
		rawEntries = append(rawEntries, strings.Split(headerValue, ",")...)
	}

	for _, rawEntry := range rawEntries {
		entryIpAddress, parseErr := tkValueObject.NewIpAddress(rawEntry)
		if parseErr != nil {
			continue
		}
		parsedEntries = append(parsedEntries, entryIpAddress)
	}

	return parsedEntries
}

func (extractor RequesterIpExtractor) firstUntrustedIpExtractor(
	httpRequest *http.Request,
	remoteAddrIpAddress tkValueObject.IpAddress,
) (ipAddress tkValueObject.IpAddress, found bool) {
	for _, extractionHeader := range extractor.extractionHeaders {
		headerStr := extractionHeader.String()
		if headerStr == ipExtractDirectKeyword ||
			headerStr == ipExtractRemoteAddrKeyword {
			return remoteAddrIpAddress, true
		}

		headerEntries := extractor.headerEntriesParser(
			httpRequest, extractionHeader,
		)
		for _, candidateIp := range slices.Backward(headerEntries) {
			if !extractor.isIpTrusted(candidateIp) {
				return candidateIp, true
			}
		}
	}

	return ipAddress, false
}

func (extractor RequesterIpExtractor) firstChainEntryIpExtractor(
	httpRequest *http.Request,
) (ipAddress tkValueObject.IpAddress, found bool) {
	for _, extractionHeader := range extractor.extractionHeaders {
		parsedEntries := extractor.headerEntriesParser(
			httpRequest, extractionHeader,
		)
		if len(parsedEntries) == 0 {
			continue
		}
		return parsedEntries[0], true
	}

	return ipAddress, false
}

func (extractor RequesterIpExtractor) hasPassingConfirmationHeader(
	httpRequest *http.Request,
) bool {
	for headerName, expectedValue := range extractor.confirmationHeaders {
		headerValues := httpRequest.Header.Values(headerName.String())
		for _, headerValue := range headerValues {
			if headerValue == "" {
				continue
			}

			isExpectedValueRequired := expectedValue.String() != ""
			if !isExpectedValueRequired {
				return true
			}

			doesValueMatch := subtle.ConstantTimeCompare(
				[]byte(headerValue), []byte(expectedValue.String()),
			) == 1
			if doesValueMatch {
				return true
			}
		}
	}

	return false
}

func (extractor RequesterIpExtractor) Execute(
	httpRequest *http.Request,
) (tkValueObject.IpAddress, error) {
	remoteAddrIpAddress, remoteAddrIpErr := extractor.remoteAddrParser(
		httpRequest.RemoteAddr,
	)
	if remoteAddrIpErr != nil {
		return remoteAddrIpAddress, remoteAddrIpErr
	}

	if !extractor.isIpTrusted(remoteAddrIpAddress) {
		return remoteAddrIpAddress, nil
	}

	isConfirmationGateFailed := extractor.confirmationHeaders != nil &&
		!extractor.hasPassingConfirmationHeader(httpRequest)
	if isConfirmationGateFailed {
		return remoteAddrIpAddress, nil
	}

	firstUntrustedIp, untrustedIpFound := extractor.firstUntrustedIpExtractor(
		httpRequest, remoteAddrIpAddress,
	)
	if untrustedIpFound {
		return firstUntrustedIp, nil
	}

	firstChainEntryIp, chainEntryFound := extractor.firstChainEntryIpExtractor(
		httpRequest,
	)
	if chainEntryFound {
		return firstChainEntryIp, nil
	}

	return remoteAddrIpAddress, nil
}
