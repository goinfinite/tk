package tkValueObject

import (
	"errors"
	"regexp"
	"slices"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const unixFileNameRegexExpression = `^[^\n\r\t\f\0\?\[\]\<\>\/]{1,512}$`

var reservedUnixFileNames = []string{".", "..", "*", "/", "\\"}

type UnixFileName string

func NewUnixFileName(value any) (fileName UnixFileName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return fileName, errors.New("UnixFileNameValueMustBeString")
	}

	re := regexp.MustCompile(unixFileNameRegexExpression)
	if !re.MatchString(stringValue) {
		return fileName, errors.New("InvalidUnixFileName")
	}

	if slices.Contains(reservedUnixFileNames, stringValue) {
		return fileName, errors.New("ReservedUnixFileName")
	}

	return UnixFileName(stringValue), nil
}

func (vo UnixFileName) String() string {
	return string(vo)
}
