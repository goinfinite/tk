package tkValueObject

import (
	"crypto/x509"
	"errors"
	"log/slog"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	X509KeyUsageDigitalSignature  X509KeyUsage = "digitalSignature"
	X509KeyUsageContentCommitment X509KeyUsage = "contentCommitment"
	X509KeyUsageKeyEncipherment   X509KeyUsage = "keyEncipherment"
	X509KeyUsageDataEncipherment  X509KeyUsage = "dataEncipherment"
	X509KeyUsageKeyAgreement      X509KeyUsage = "keyAgreement"
	X509KeyUsageKeyCertSign       X509KeyUsage = "keyCertSign"
	X509KeyUsageCRLSign           X509KeyUsage = "cRLSign"
	X509KeyUsageEncipherOnly      X509KeyUsage = "encipherOnly"
	X509KeyUsageDecipherOnly      X509KeyUsage = "decipherOnly"
)

type X509KeyUsage string

func NewX509KeyUsage(value any) (usage X509KeyUsage, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return usage, errors.New("X509KeyUsageMustBeString")
	}

	usage = X509KeyUsage(stringValue)
	switch usage {
	case X509KeyUsageDigitalSignature, X509KeyUsageContentCommitment,
		X509KeyUsageKeyEncipherment, X509KeyUsageDataEncipherment,
		X509KeyUsageKeyAgreement, X509KeyUsageKeyCertSign,
		X509KeyUsageCRLSign, X509KeyUsageEncipherOnly,
		X509KeyUsageDecipherOnly:
		return usage, nil
	default:
		return usage, errors.New("InvalidX509KeyUsage")
	}
}

func NewX509KeyUsageSliceFromStdlib(
	stdlibKeyUsage x509.KeyUsage,
) ([]X509KeyUsage, error) {
	var keyUsageSlice []X509KeyUsage

	keyUsagePairs := []struct {
		flag x509.KeyUsage
		name string
	}{
		{x509.KeyUsageDigitalSignature, "digitalSignature"},
		{x509.KeyUsageContentCommitment, "contentCommitment"},
		{x509.KeyUsageKeyEncipherment, "keyEncipherment"},
		{x509.KeyUsageDataEncipherment, "dataEncipherment"},
		{x509.KeyUsageKeyAgreement, "keyAgreement"},
		{x509.KeyUsageCertSign, "keyCertSign"},
		{x509.KeyUsageCRLSign, "cRLSign"},
		{x509.KeyUsageEncipherOnly, "encipherOnly"},
		{x509.KeyUsageDecipherOnly, "decipherOnly"},
	}

	for _, rawKeyUsagePair := range keyUsagePairs {
		flagIsSet := stdlibKeyUsage&rawKeyUsagePair.flag != 0
		if flagIsSet {
			keyUsage, err := NewX509KeyUsage(rawKeyUsagePair.name)
			if err != nil {
				slog.Debug(
					"SkipInvalidKeyUsage",
					slog.String("name", rawKeyUsagePair.name),
				)
				continue
			}
			keyUsageSlice = append(keyUsageSlice, keyUsage)
		}
	}

	return keyUsageSlice, nil
}

func (vo X509KeyUsage) String() string {
	return string(vo)
}
