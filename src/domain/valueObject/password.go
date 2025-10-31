package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type Password string

var (
	hasLetterRegex       = regexp.MustCompile(`[a-zA-Z]`)
	hasNumberRegex       = regexp.MustCompile(`[0-9]`)
	hasSpecialCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

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

	if !hasLetterRegex.MatchString(stringValue) {
		return password, errors.New("PasswordMustHaveLetter")
	}

	if !hasNumberRegex.MatchString(stringValue) {
		return password, errors.New("PasswordMustHaveNumber")
	}

	if !hasSpecialCharsRegex.MatchString(stringValue) {
		return password, errors.New("PasswordMustHaveSpecialCharacter")
	}

	return Password(stringValue), nil
}

func (vo Password) String() string {
	return string(vo)
}
