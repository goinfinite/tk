package tkPresentation

import (
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
	ipExtractHeaderEnvVarName  string = "IP_EXTRACT_HEADER"
	ipExtractHeaderDefaults    string = "X-Forwarded-For,X-Real-IP"
	ipExtractDirectKeyword     string = "Direct"
	ipExtractRemoteAddrKeyword string = "RemoteAddr"
)

type RequesterIpExtractor struct {
	extractionHeaders []tkValueObject.HttpHeader
	trustedCidrBlocks []tkValueObject.CidrBlock
}

func IpExtractHeaderReader() []tkValueObject.HttpHeader {
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

func NewRequesterIpExtractor() RequesterIpExtractor {
	cidrBlocks, cidrsReadingErr := tkInfra.TrustedCidrsReader()
	if cidrsReadingErr != nil {
		slog.Debug(
			"TrustedCidrsReaderError",
			slog.String("err", cidrsReadingErr.Error()),
		)
	}

	return RequesterIpExtractor{
		extractionHeaders: IpExtractHeaderReader(),
		trustedCidrBlocks: cidrBlocks,
	}
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

func (extractor RequesterIpExtractor) HeaderIpExtractor(
	httpRequest *http.Request,
	extractionHeader tkValueObject.HttpHeader,
) (ipAddress tkValueObject.IpAddress, err error) {
	headerValues := httpRequest.Header.Values(
		extractionHeader.String(),
	)

	var allEntries []string
	for _, headerValue := range headerValues {
		entries := strings.Split(headerValue, ",")
		allEntries = append(allEntries, entries...)
	}

	for _, headerEntry := range slices.Backward(allEntries) {
		candidateIp, parseErr := tkValueObject.NewIpAddress(
			headerEntry,
		)
		if parseErr != nil {
			continue
		}
		if extractor.isIpTrusted(candidateIp) {
			continue
		}
		return candidateIp, nil
	}

	return ipAddress, errors.New("NoUntrustedIpInHeader")
}

func (extractor RequesterIpExtractor) Execute(
	httpRequest *http.Request,
) (tkValueObject.IpAddress, error) {
	for _, extractionHeader := range extractor.extractionHeaders {
		headerStr := extractionHeader.String()
		if headerStr == ipExtractDirectKeyword ||
			headerStr == ipExtractRemoteAddrKeyword {
			return extractor.remoteAddrParser(httpRequest.RemoteAddr)
		}

		extractedIp, extractionErr := extractor.HeaderIpExtractor(
			httpRequest, extractionHeader,
		)
		if extractionErr == nil {
			return extractedIp, nil
		}
	}

	return extractor.remoteAddrParser(httpRequest.RemoteAddr)
}
