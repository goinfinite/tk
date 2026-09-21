package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var zipCodeNonNumericRegex = regexp.MustCompile(`[^\d]`)

type ZipCode string

func NewZipCode(value any) (ZipCode, error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return "", errors.New("ZipCodeMustBeString")
	}

	numericStringValue := zipCodeNonNumericRegex.ReplaceAllString(stringValue, "")

	if len(numericStringValue) < 3 {
		return "", errors.New("ZipCodeTooSmall")
	}

	if len(numericStringValue) > 10 {
		return "", errors.New("ZipCodeTooBig")
	}

	return ZipCode(numericStringValue), nil
}

func (vo ZipCode) String() string {
	return string(vo)
}
