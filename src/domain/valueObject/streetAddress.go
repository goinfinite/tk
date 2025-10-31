package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const streetAddressRegex = `^[\p{L}\d][\p{L}\d\'\.\,\ \-]{2,512}[\p{L}\d\.]$`

type StreetAddress string

func NewStreetAddress(value any) (streetAddress StreetAddress, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return streetAddress, errors.New("StreetAddressMustBeString")
	}

	re := regexp.MustCompile(streetAddressRegex)
	if !re.MatchString(stringValue) {
		return streetAddress, errors.New("InvalidStreetAddress")
	}

	return StreetAddress(stringValue), nil
}

func (vo StreetAddress) String() string {
	return string(vo)
}
