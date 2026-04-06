package tkInfra

import (
	"log/slog"
	"os"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	TrustedCidrsEnvVarName string = "TRUSTED_CIDRS"
)

func TrustedCidrsReader() (trustedCidrBlocks []tkValueObject.CidrBlock, err error) {
	rawTrustedCidrsEnvValue := os.Getenv(TrustedCidrsEnvVarName)
	if rawTrustedCidrsEnvValue == "" {
		return trustedCidrBlocks, nil
	}

	for rawTrustedCidr := range strings.SplitSeq(rawTrustedCidrsEnvValue, ",") {
		trustedCidrBlock, err := tkValueObject.NewCidrBlock(rawTrustedCidr)
		if err != nil {
			slog.Debug("InvalidTrustedCidrBlock", slog.String("rawTrustedCidr", rawTrustedCidr))
			continue
		}
		trustedCidrBlocks = append(trustedCidrBlocks, trustedCidrBlock)
	}

	return trustedCidrBlocks, nil
}
