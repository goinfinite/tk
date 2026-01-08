package tkValueObject

import (
	"crypto/x509"
	"errors"
	"log/slog"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	X509ExtendedKeyUsageServerAuth      X509ExtendedKeyUsage = "serverAuth"
	X509ExtendedKeyUsageClientAuth      X509ExtendedKeyUsage = "clientAuth"
	X509ExtendedKeyUsageCodeSigning     X509ExtendedKeyUsage = "codeSigning"
	X509ExtendedKeyUsageEmailProtection X509ExtendedKeyUsage = "emailProtection"
	X509ExtendedKeyUsageTimeStamping    X509ExtendedKeyUsage = "timeStamping"
	X509ExtendedKeyUsageOCSPSigning     X509ExtendedKeyUsage = "ocspSigning"
)

type X509ExtendedKeyUsage string

func NewX509ExtendedKeyUsage(
	value any,
) (usage X509ExtendedKeyUsage, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return usage, errors.New("X509ExtendedKeyUsageMustBeString")
	}

	usage = X509ExtendedKeyUsage(stringValue)
	switch usage {
	case X509ExtendedKeyUsageServerAuth, X509ExtendedKeyUsageClientAuth,
		X509ExtendedKeyUsageCodeSigning, X509ExtendedKeyUsageEmailProtection,
		X509ExtendedKeyUsageTimeStamping, X509ExtendedKeyUsageOCSPSigning:
		return usage, nil
	default:
		return usage, errors.New("InvalidX509ExtendedKeyUsage")
	}
}

func NewX509ExtendedKeyUsageSliceFromStdlib(
	stdlibExtendedKeyUsages []x509.ExtKeyUsage,
) ([]X509ExtendedKeyUsage, error) {
	var extendedKeyUsageSlice []X509ExtendedKeyUsage
	extendedKeyUsageMap := map[x509.ExtKeyUsage]string{
		x509.ExtKeyUsageServerAuth:      "serverAuth",
		x509.ExtKeyUsageClientAuth:      "clientAuth",
		x509.ExtKeyUsageCodeSigning:     "codeSigning",
		x509.ExtKeyUsageEmailProtection: "emailProtection",
		x509.ExtKeyUsageTimeStamping:    "timeStamping",
		x509.ExtKeyUsageOCSPSigning:     "ocspSigning",
	}

	for _, stdlibExtendedKeyUsage := range stdlibExtendedKeyUsages {
		extendedKeyUsageName, extendedKeyUsageExists :=
			extendedKeyUsageMap[stdlibExtendedKeyUsage]
		if !extendedKeyUsageExists {
			slog.Debug(
				"SkipUnsupportedExtendedKeyUsage",
				slog.Int("value", int(stdlibExtendedKeyUsage)),
			)
			continue
		}
		extendedKeyUsage, err := NewX509ExtendedKeyUsage(
			extendedKeyUsageName,
		)
		if err != nil {
			slog.Debug(
				"SkipInvalidExtendedKeyUsage",
				slog.String("name", extendedKeyUsageName),
			)
			continue
		}
		extendedKeyUsageSlice = append(extendedKeyUsageSlice, extendedKeyUsage)
	}

	return extendedKeyUsageSlice, nil
}

func (vo X509ExtendedKeyUsage) String() string {
	return string(vo)
}
