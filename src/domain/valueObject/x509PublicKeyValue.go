package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509PublicKeyValueRegex = regexp.MustCompile(`^[A-Za-z0-9+/=\n\r ]{100,}$`)

type X509PublicKeyValue string

func NewX509PublicKeyValue(value any) (key X509PublicKeyValue, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return key, errors.New("X509PublicKeyValueMustBeString")
	}

	if !x509PublicKeyValueRegex.MatchString(stringValue) {
		return key, errors.New("InvalidX509PublicKeyValue")
	}

	return X509PublicKeyValue(stringValue), nil
}

func (vo X509PublicKeyValue) String() string {
	return string(vo)
}
