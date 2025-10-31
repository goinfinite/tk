package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const unixGroupNameRegex string = `^[a-z_]([a-z0-9_-]{0,31}|[a-z0-9_-]{0,30}\$)$`

type UnixGroupName string

func NewUnixGroupName(value any) (unixGroupName UnixGroupName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unixGroupName, errors.New("UnixGroupNameMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	re := regexp.MustCompile(unixGroupNameRegex)
	if !re.MatchString(stringValue) {
		return unixGroupName, errors.New("InvalidUnixGroupName")
	}

	return UnixGroupName(stringValue), nil
}

func (vo UnixGroupName) String() string {
	return string(vo)
}
