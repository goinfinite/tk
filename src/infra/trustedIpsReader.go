package tkInfra

import (
	"log/slog"
	"os"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	TrustedIpsEnvVarName string = "TRUSTED_IPS"
)

func TrustedIpsReader() (trustedIpAddresses []tkValueObject.IpAddress, err error) {
	rawTrustedIpsEnvValue := os.Getenv(TrustedIpsEnvVarName)
	if rawTrustedIpsEnvValue == "" {
		return trustedIpAddresses, nil
	}

	for rawTrustedIp := range strings.SplitSeq(rawTrustedIpsEnvValue, ",") {
		trustedIpAddress, err := tkValueObject.NewIpAddress(rawTrustedIp)
		if err != nil {
			slog.Debug("InvalidTrustedIpAddress", slog.String("rawTrustedIp", rawTrustedIp))
			continue
		}
		trustedIpAddresses = append(trustedIpAddresses, trustedIpAddress)
	}

	return trustedIpAddresses, nil
}
