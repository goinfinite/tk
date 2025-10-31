package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const unixUsernameRegex string = `^[a-z_]([a-z0-9_-]{0,31}|[a-z0-9_-]{0,30}\$)$`

type UnixUsername string

func NewUnixUsername(value any) (unixUsername UnixUsername, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unixUsername, errors.New("UnixUsernameMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	re := regexp.MustCompile(unixUsernameRegex)
	if !re.MatchString(stringValue) {
		return unixUsername, errors.New("InvalidUnixUsername")
	}

	return UnixUsername(stringValue), nil
}

func (vo UnixUsername) String() string {
	return string(vo)
}
