package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509EnvelopedCertificateRegex = regexp.MustCompile(
	`^-----BEGIN CERTIFICATE-----[\s\S]+-----END CERTIFICATE-----$`,
)

type X509EnvelopedCertificate string

func NewX509EnvelopedCertificate(
	value any,
) (envelopedCert X509EnvelopedCertificate, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return envelopedCert, errors.New("X509EnvelopedCertificateMustBeString")
	}

	if !x509EnvelopedCertificateRegex.MatchString(stringValue) {
		return envelopedCert, errors.New("InvalidX509EnvelopedCertificateFormat")
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
