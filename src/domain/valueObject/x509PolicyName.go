package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509PolicyNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 .\-_()]{1,128}$`)

type X509PolicyName string

func NewX509PolicyName(value any) (name X509PolicyName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return name, errors.New("X509PolicyNameMustBeString")
	}

	if !x509PolicyNameRegex.MatchString(stringValue) {
		return name, errors.New("InvalidX509PolicyName")
	}

	return X509PolicyName(stringValue), nil
}

func (vo X509PolicyName) String() string {
	return string(vo)
}
