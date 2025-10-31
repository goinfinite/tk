package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type PhoneNumber string

func NewPhoneNumber(value any) (phoneNumber PhoneNumber, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return phoneNumber, errors.New("PhoneNumberMustBeString")
	}

	nonNumericRegexp := regexp.MustCompile(`[^\d]`)
	numericStringValue := nonNumericRegexp.ReplaceAllString(stringValue, "")

	if len(numericStringValue) < 5 {
		return phoneNumber, errors.New("PhoneNumberTooSmall")
	}

	if len(numericStringValue) > 16 {
		return phoneNumber, errors.New("PhoneNumberTooBig")
	}

	return PhoneNumber(numericStringValue), nil
}

func (vo PhoneNumber) String() string {
	return string(vo)
}
