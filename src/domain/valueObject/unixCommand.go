package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type UnixCommand string

func NewUnixCommand(value any) (unixCommand UnixCommand, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unixCommand, errors.New("UnixCommandValueMustBeString")
	}

	if len(stringValue) < 2 {
		return unixCommand, errors.New("UnixCommandTooShort")
	}

	if len(stringValue) > 4096 {
		return unixCommand, errors.New("UnixCommandTooLong")
	}

	return UnixCommand(stringValue), nil
}

func (vo UnixCommand) String() string {
	return string(vo)
}
