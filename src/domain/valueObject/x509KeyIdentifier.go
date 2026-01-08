package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509KeyIdentifierRegex = regexp.MustCompile(`^[0-9A-Fa-f]{40}$`)

type X509KeyIdentifier string

func NewX509KeyIdentifier(value any) (id X509KeyIdentifier, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return id, errors.New("X509KeyIdentifierMustBeString")
	}

	stringValue = strings.ReplaceAll(stringValue, ":", "")
	stringValue = strings.ReplaceAll(stringValue, " ", "")

	if !x509KeyIdentifierRegex.MatchString(stringValue) {
		return id, errors.New("InvalidX509KeyIdentifier")
	}

	return X509KeyIdentifier(stringValue), nil
}

func (vo X509KeyIdentifier) String() string {
	return string(vo)
}
