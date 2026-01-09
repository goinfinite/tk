package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509SignatureValueRegex = regexp.MustCompile(
	`^[0-9A-Za-z+/=\n\r ]{64,}$`,
)

type X509SignatureValue string

func NewX509SignatureValue(value any) (sig X509SignatureValue, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return sig, errors.New("X509SignatureValueMustBeString")
	}

	if len(stringValue) < 64 || len(stringValue) > 1024 {
		return sig, errors.New("InvalidX509SignatureValueLength")
	}

	if !x509SignatureValueRegex.MatchString(stringValue) {
		return sig, errors.New("InvalidX509SignatureValue")
	}

	return X509SignatureValue(stringValue), nil
}

func (vo X509SignatureValue) String() string {
	return string(vo)
}
