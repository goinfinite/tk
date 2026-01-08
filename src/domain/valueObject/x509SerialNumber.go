package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509SerialNumberRegex = regexp.MustCompile(`^[0-9A-Fa-f]{1,40}$`)

type X509SerialNumber string

func NewX509SerialNumber(value any) (serial X509SerialNumber, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return serial, errors.New("X509SerialNumberMustBeString")
	}

	stringValue = strings.ReplaceAll(stringValue, ":", "")
	stringValue = strings.ReplaceAll(stringValue, " ", "")

	if !x509SerialNumberRegex.MatchString(stringValue) {
		return serial, errors.New("InvalidX509SerialNumber")
	}

	return X509SerialNumber(stringValue), nil
}

func (vo X509SerialNumber) String() string {
	return string(vo)
}
