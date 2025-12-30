package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var userAgentRegex = regexp.MustCompile(
	`^[a-zA-Z0-9][a-zA-Z0-9\(\)\[\]\{\}\/\?\:;@&=+$,\-_\.!~*'\% ]{0,499}$`,
)

type UserAgent string

func NewUserAgent(value any) (userAgent UserAgent, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return userAgent, errors.New("UserAgentMustBeString")
	}

	if !userAgentRegex.MatchString(stringValue) {
		return userAgent, errors.New("InvalidUserAgent")
	}

	return UserAgent(stringValue), nil
}

func (vo UserAgent) String() string {
	return string(vo)
}
