package tkInfra

import (
	"log/slog"
	"os"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	TrustedIpsEnvVarName   string = "TRUSTED_IPS"
	TrustedCidrsEnvVarName string = "TRUSTED_CIDRS"
)

func TrustedCidrsReader() (trustedCidrBlocks []tkValueObject.CidrBlock, err error) {
	rawEntries := os.Getenv(TrustedIpsEnvVarName) + "," +
		os.Getenv(TrustedCidrsEnvVarName)

	for rawEntry := range strings.SplitSeq(rawEntries, ",") {
		if rawEntry == "" {
			continue
		}

		cidrBlock, cidrErr := tkValueObject.NewCidrBlock(rawEntry)
		if cidrErr == nil {
			trustedCidrBlocks = append(trustedCidrBlocks, cidrBlock)
			continue
		}

		ipAddress, ipErr := tkValueObject.NewIpAddress(rawEntry)
		if ipErr == nil {
			trustedCidrBlocks = append(trustedCidrBlocks, ipAddress.ToCidrBlock())
			continue
		}

		slog.Debug(
			"InvalidTrustedEntry",
			slog.String("rawEntry", rawEntry),
		)
	}

	return trustedCidrBlocks, nil
}
