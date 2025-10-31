package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var unixHostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

type UnixHostname string

func NewUnixHostname(value any) (hostname UnixHostname, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return hostname, errors.New("UnixHostnameMustBeString")
	}

	stringValue = strings.ToLower(stringValue)

	if !unixHostnameRegex.MatchString(stringValue) {
		return hostname, errors.New("InvalidUnixHostname")
	}

	return UnixHostname(stringValue), nil
}

func (vo UnixHostname) String() string {
	return string(vo)
}
