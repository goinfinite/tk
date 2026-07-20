package tkInfra

import (
	"context"
	"errors"
	"io"
	"log/slog"
	mathRand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	ServerPublicIpAddressEnvVarName string = "SERVER_PUBLIC_IP_ADDR"

	publicIpResolverRequestTimeout          = 5 * time.Second
	publicIpResolverResponseByteLimit int64 = 256
)

var publicIpResolverAvailableUrls = []tkValueObject.Url{
	"https://goinfinite.net/ip",
	"https://checkip.amazonaws.com",
	"https://icanhazip.com",
}

type PublicIpAddressResolver struct {
	httpClient   *http.Client
	resolverUrls []tkValueObject.Url
}

func NewPublicIpAddressResolver() *PublicIpAddressResolver {
	return &PublicIpAddressResolver{
		httpClient:   &http.Client{Timeout: publicIpResolverRequestTimeout},
		resolverUrls: publicIpResolverAvailableUrls,
	}
}

func (resolver *PublicIpAddressResolver) resolverFetcher(
	resolverUrl tkValueObject.Url,
) (rawIpAddress string, err error) {
	resolverContext, contextCancel := context.WithTimeout(
		context.Background(), publicIpResolverRequestTimeout,
	)
	defer contextCancel()

	httpRequest, newRequestError := http.NewRequestWithContext(
		resolverContext, http.MethodGet, resolverUrl.String(), nil,
	)
	if newRequestError != nil {
		return rawIpAddress, newRequestError
	}
	httpRequest.Header.Set("User-Agent", "goinfinite-tk/1.0")
	httpRequest.Header.Set("Accept", "text/plain")

	httpResponse, requestSendError := resolver.httpClient.Do(httpRequest)
	if requestSendError != nil {
		return rawIpAddress, requestSendError
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return rawIpAddress, errors.New(
			"UnexpectedPublicIpResolverHttpStatus: " +
				strconv.Itoa(httpResponse.StatusCode),
		)
	}

	bodyBytes, bodyReadError := io.ReadAll(io.LimitReader(
		httpResponse.Body, publicIpResolverResponseByteLimit,
	))
	if bodyReadError != nil {
		return rawIpAddress, bodyReadError
	}

	if len(bodyBytes) == 0 {
		return rawIpAddress, errors.New("PublicIpResolverEmptyResponse")
	}

	return strings.TrimSpace(string(bodyBytes)), nil
}

func (resolver *PublicIpAddressResolver) Resolve() (
	ipAddress tkValueObject.IpAddress, err error,
) {
	resolverUrlsCount := len(resolver.resolverUrls)
	if resolverUrlsCount == 0 {
		return ipAddress, errors.New("PublicIpResolverHasNoEndpoints")
	}

	randomOrderedResolvers := make([]tkValueObject.Url, resolverUrlsCount)
	for shuffledPosition, originalPosition := range mathRand.Perm(resolverUrlsCount) {
		randomOrderedResolvers[shuffledPosition] =
			resolver.resolverUrls[originalPosition]
	}

	for _, resolverUrl := range randomOrderedResolvers {
		rawIpAddress, fetchError := resolver.resolverFetcher(resolverUrl)
		if fetchError != nil {
			slog.Debug("PublicIpResolverEndpointFailed",
				slog.String("resolverUrl", resolverUrl.String()),
				slog.String("error", fetchError.Error()),
			)
			continue
		}

		parsedIpAddress, parseError := tkValueObject.NewIpAddress(rawIpAddress)
		if parseError != nil {
			slog.Debug("PublicIpResolverEndpointReturnedInvalidIp",
				slog.String("resolverUrl", resolverUrl.String()),
				slog.String("rawResponse", rawIpAddress),
			)
			continue
		}

		return parsedIpAddress, nil
	}

	return ipAddress, errors.New("PublicIpAddressNotFound")
}

func ReadServerPrivateIpAddress() (ipAddress tkValueObject.IpAddress, err error) {
	rawIpAddress, err := NewShell(
		ShellSettings{Command: "hostname", Args: []string{"-I"}},
	).Run()
	if err != nil {
		return ipAddress, err
	}

	rawIpAddresses := strings.Split(rawIpAddress, " ")
	if len(rawIpAddresses) > 0 {
		rawIpAddress = rawIpAddresses[0]
	}

	rawIpAddress = strings.TrimSpace(rawIpAddress)
	if rawIpAddress == "" {
		return ipAddress, errors.New("PrivateIpAddressNotFound")
	}

	return tkValueObject.NewIpAddress(rawIpAddress)
}

func ReadServerPublicIpAddress() (ipAddress tkValueObject.IpAddress, err error) {
	rawEnvIpAddress := os.Getenv(ServerPublicIpAddressEnvVarName)
	if rawEnvIpAddress != "" {
		envIpAddress, envError := tkValueObject.NewIpAddress(rawEnvIpAddress)
		if envError == nil {
			return envIpAddress, nil
		}

		slog.Debug(
			"InvalidServerPublicIpAddressEnvVar",
			slog.String("envVarName", ServerPublicIpAddressEnvVarName),
			slog.String("envValue", rawEnvIpAddress),
		)
	}

	return NewPublicIpAddressResolver().Resolve()
}
