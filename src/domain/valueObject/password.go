package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type Password string

func NewPassword(value any) (password Password, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return password, errors.New("PasswordMustBeString")
	}

	if len(stringValue) < 5 {
		return password, errors.New("PasswordTooShort")
	}

	if len(stringValue) > 128 {
		return password, errors.New("PasswordTooLong")
	}

	hasLetterRegexp := regexp.MustCompile(`[a-zA-Z]`)
	if !hasLetterRegexp.MatchString(stringValue) {
		return password, errors.New("PasswordMustHaveLetter")
	}

	hasNumberRegexp := regexp.MustCompile(`[0-9]`)
	if !hasNumberRegexp.MatchString(stringValue) {
		return password, errors.New("PasswordMustHaveNumber")
	}

	hasSpecialRegexp := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !hasSpecialRegexp.MatchString(stringValue) {
		return password, errors.New("PasswordMustHaveSpecialCharacter")
	}

	return Password(stringValue), nil
}

func (vo Password) String() string {
	return string(vo)
}
