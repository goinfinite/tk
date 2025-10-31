package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	AccessTokenTypeSessionToken AccessTokenType = "sessionToken"
	AccessTokenTypeSecretKey    AccessTokenType = "secretKey"
)

type AccessTokenType string

func NewAccessTokenType(value any) (tokenType AccessTokenType, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return tokenType, errors.New("AccessTokenTypeMustBeString")
	}

	tokenType = AccessTokenType(stringValue)
	switch tokenType {
	case AccessTokenTypeSessionToken, AccessTokenTypeSecretKey:
		return tokenType, nil
	default:
		return tokenType, errors.New("InvalidAccessTokenType")
	}
}

func (vo AccessTokenType) String() string {
	return string(vo)
}
