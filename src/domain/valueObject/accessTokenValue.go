package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const accessTokenValueRegex string = `^[a-zA-Z0-9\-_=+/.]+$`

type AccessTokenValue string

func NewAccessTokenValue(value any) (tokenValue AccessTokenValue, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return tokenValue, errors.New("AccessTokenValueMustBeString")
	}

	if len(stringValue) < 10 {
		return tokenValue, errors.New("InvalidAccessTokenValueTooShort")
	}

	if len(stringValue) > 3072 {
		return tokenValue, errors.New("InvalidAccessTokenValueTooLong")
	}

	re := regexp.MustCompile(accessTokenValueRegex)
	if !re.MatchString(stringValue) {
		return tokenValue, errors.New("InvalidAccessTokenValue")
	}
	return AccessTokenValue(stringValue), nil
}

func (vo AccessTokenValue) String() string {
	return string(vo)
}
