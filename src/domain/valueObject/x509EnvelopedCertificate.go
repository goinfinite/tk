package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type X509EnvelopedCertificate string

func NewX509EnvelopedCertificate(
	value any,
) (envelopedCert X509EnvelopedCertificate, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return envelopedCert, errors.New("X509EnvelopedCertificateMustBeString")
	}

	if !strings.Contains(stringValue, "-----BEGIN CERTIFICATE-----") {
		return envelopedCert, errors.New("InvalidX509EnvelopedCertificateNoBeginMarker")
	}

	if !strings.Contains(stringValue, "-----END CERTIFICATE-----") {
		return envelopedCert, errors.New("InvalidX509EnvelopedCertificateNoEndMarker")
	}

	if len(stringValue) < 100 {
		return envelopedCert, errors.New("InvalidX509EnvelopedCertificateTooShort")
	}

	return X509EnvelopedCertificate(stringValue), nil
}

func (vo X509EnvelopedCertificate) String() string {
	return string(vo)
}

func (vo X509EnvelopedCertificate) Bytes() []byte {
	return []byte(vo)
}
