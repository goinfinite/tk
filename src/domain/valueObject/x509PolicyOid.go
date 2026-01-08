package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509PolicyOidRegex = regexp.MustCompile(`^[0-9]+(\.[0-9]+)+$`)

type X509PolicyOID string

func NewX509PolicyOID(value any) (oid X509PolicyOID, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return oid, errors.New("X509PolicyOIDMustBeString")
	}

	if !x509PolicyOidRegex.MatchString(stringValue) {
		return oid, errors.New("InvalidX509PolicyOID")
	}

	components := strings.Split(stringValue, ".")
	if len(components) < 2 {
		return oid, errors.New("InvalidX509PolicyOIDTooFewComponents")
	}

	return X509PolicyOID(stringValue), nil
}

func (vo X509PolicyOID) String() string {
	return string(vo)
}
