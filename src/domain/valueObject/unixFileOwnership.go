package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	UnixFileOwnershipNobodyNogroup UnixFileOwnership = "nobody:nogroup"
	UnixFileOwnershipRootRoot      UnixFileOwnership = "root:root"
)

type UnixFileOwnership string

func NewUnixFileOwnership(value any) (fileOwnership UnixFileOwnership, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return fileOwnership, errors.New("UnixFileOwnershipValueMustBeString")
	}

	stringValueParts := strings.Split(stringValue, ":")
	if len(stringValueParts) != 2 {
		return fileOwnership, errors.New("InvalidUnixFileOwnership")
	}

	ownerUsername, err := NewUnixUsername(stringValueParts[0])
	if err != nil {
		return fileOwnership, err
	}

	ownerGroupName, err := NewUnixGroupName(stringValueParts[1])
	if err != nil {
		return fileOwnership, err
	}

	return UnixFileOwnership(ownerUsername.String() + ":" + ownerGroupName.String()), nil
}

func (vo UnixFileOwnership) String() string {
	return string(vo)
}

func (vo UnixFileOwnership) ReadUsername() (UnixUsername, error) {
	return NewUnixUsername(strings.Split(string(vo), ":")[0])
}

func (vo UnixFileOwnership) ReadGroupName() (UnixGroupName, error) {
	return NewUnixGroupName(strings.Split(string(vo), ":")[1])
}
