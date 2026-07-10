package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotMaxEntries uint64

func NewHoneypotMaxEntries(value any) (entries HoneypotMaxEntries, err error) {
	uintValue, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return entries, errors.New("HoneypotMaxEntriesMustBeUint64")
	}

	if uintValue == 0 {
		return entries, errors.New("HoneypotMaxEntriesMustBePositive")
	}

	return HoneypotMaxEntries(uintValue), nil
}

func (vo HoneypotMaxEntries) Uint64() uint64 {
	return uint64(vo)
}
