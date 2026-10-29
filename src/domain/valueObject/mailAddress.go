package tkValueObject

import (
	"errors"
	"net/mail"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type MailAddress string

func NewMailAddress(value any) (mailAddress MailAddress, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return mailAddress, errors.New("MailAddressMustBeString")
	}

	_, parseErr := mail.ParseAddress(stringValue)
	if parseErr != nil {
		return mailAddress, errors.New("InvalidMailAddress")
	}

	return MailAddress(stringValue), nil
}

func (vo MailAddress) String() string {
	return string(vo)
}
