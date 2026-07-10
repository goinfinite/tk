package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotMaxStreamSize uint64

func NewHoneypotMaxStreamSize(value any) (size HoneypotMaxStreamSize, err error) {
	uintValue, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return size, errors.New("HoneypotMaxStreamSizeMustBeUint64")
	}

	if uintValue == 0 {
		return size, errors.New("HoneypotMaxStreamSizeMustBePositive")
	}

	return HoneypotMaxStreamSize(uintValue), nil
}

func (vo HoneypotMaxStreamSize) Uint64() uint64 {
	return uint64(vo)
}
