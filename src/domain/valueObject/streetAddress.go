package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var streetAddressRegex = regexp.MustCompile(`^[\p{L}\d][\p{L}\d\'\.\,\ \-]{2,512}[\p{L}\d\.]$`)

type StreetAddress string

func NewStreetAddress(value any) (streetAddress StreetAddress, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return streetAddress, errors.New("StreetAddressMustBeString")
	}

	if !streetAddressRegex.MatchString(stringValue) {
		return streetAddress, errors.New("InvalidStreetAddress")
	}

	return StreetAddress(stringValue), nil
}

func (vo StreetAddress) String() string {
	return string(vo)
}
