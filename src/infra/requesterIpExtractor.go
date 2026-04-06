package tkInfra

import (
	"log/slog"
	"net"
	"net/http"
	"os"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
	"github.com/labstack/echo/v4"
)

type RequesterIpExtractor struct {
	disableTrust bool
	trustOptions []echo.TrustOption
}

func TrustOptionsReader() []echo.TrustOption {
	trustOptions := []echo.TrustOption{
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(true),
		echo.TrustPrivateNet(true),
	}

	cidrBlocks, cidrsReadingErr := TrustedCidrsReader()
	if cidrsReadingErr != nil {
		slog.Debug(
			"TrustedCidrsReaderError",
			slog.String("err", cidrsReadingErr.Error()),
		)
		return trustOptions
	}

	for _, cidrBlock := range cidrBlocks {
		_, ipNet, parseErr := net.ParseCIDR(cidrBlock.String())
		if parseErr != nil {
			slog.Debug(
				"InvalidCidrBlockForTrustOption",
				slog.String("cidrBlock", cidrBlock.String()),
			)
			continue
		}
		trustOptions = append(trustOptions, echo.TrustIPRange(ipNet))
	}

	return trustOptions
}

func DisableTrustReader() bool {
	const ipExtractDisableTrustEnvVarName = "IP_EXTRACT_DISABLE_TRUST"

	disableTrustEnvVal := os.Getenv(ipExtractDisableTrustEnvVarName)
	if disableTrustEnvVal == "" {
		return false
	}

	parsedDisableTrust, parseErr := tkVoUtil.InterfaceToBool(disableTrustEnvVal)
	if parseErr != nil {
		slog.Debug(
			"IpExtractDisableTrustEnvVarInvalid",
			slog.String("err", parseErr.Error()),
		)
		return false
	}

	return parsedDisableTrust
}

func NewRequesterIpExtractor() RequesterIpExtractor {
	return RequesterIpExtractor{
		disableTrust: DisableTrustReader(),
		trustOptions: TrustOptionsReader(),
	}
}

func (extractor RequesterIpExtractor) Execute(request *http.Request) string {
	if extractor.disableTrust {
		return echo.ExtractIPDirect()(request)
	}
	return echo.ExtractIPFromXFFHeader(extractor.trustOptions...)(request)
}
