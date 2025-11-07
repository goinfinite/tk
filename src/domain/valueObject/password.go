package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

// Password must be at least 5 characters long and at most 128 characters long. It must
// contain at least one letter, one number, and one special character. If you cannot
// control the password being validated (e.g. a password from an external service),
// use WeakPassword instead.
type Password string

var (
	hasLetterRegex       = regexp.MustCompile(`[a-zA-Z]`)
	hasNumberRegex       = regexp.MustCompile(`[0-9]`)
	hasSpecialCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

const (
	errPasswordMustBeString             string = "PasswordMustBeString"
	errPasswordTooShort                 string = "PasswordTooShort"
	errPasswordTooLong                  string = "PasswordTooLong"
	errPasswordMustHaveLetter           string = "PasswordMustHaveLetter"
	errPasswordMustHaveNumber           string = "PasswordMustHaveNumber"
	errPasswordMustHaveSpecialCharacter string = "PasswordMustHaveSpecialCharacter"
)

func NewPassword(value any) (password Password, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return password, errors.New(errPasswordMustBeString)
	}

	if len(stringValue) < 5 {
		return password, errors.New(errPasswordTooShort)
	}

	if len(stringValue) > 128 {
		return password, errors.New(errPasswordTooLong)
	}

	if !hasLetterRegex.MatchString(stringValue) {
		return password, errors.New(errPasswordMustHaveLetter)
	}

	if !hasNumberRegex.MatchString(stringValue) {
		return password, errors.New(errPasswordMustHaveNumber)
	}

	if !hasSpecialCharsRegex.MatchString(stringValue) {
		return password, errors.New(errPasswordMustHaveSpecialCharacter)
	}

	return Password(stringValue), nil
}

func (vo Password) String() string {
	return string(vo)
}

// WeakPassword is only for backward compatibility with old code or when the password
// being validated is outside the control of your application (e.g. a password from
// an external service). If possible, never use this type and instead use Password
// which provides stronger validation policies.
type WeakPassword string

func NewWeakPassword(value any) (weakPassword WeakPassword, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return weakPassword, errors.New(errPasswordMustBeString)
	}

	if len(stringValue) < 5 {
		return weakPassword, errors.New(errPasswordTooShort)
	}

	if len(stringValue) > 128 {
		return weakPassword, errors.New(errPasswordTooLong)
	}

	return WeakPassword(stringValue), nil
}

func (vo WeakPassword) String() string {
	return string(vo)
}
