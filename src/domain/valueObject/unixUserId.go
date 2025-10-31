package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type UnixUserId uint64

func NewUnixUserId(value any) (unixUserId UnixUserId, err error) {
	uint64Value, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return unixUserId, errors.New("InvalidUnixUserId")
	}

	return UnixUserId(uint64Value), nil
}

func (vo UnixUserId) Uint64() uint64 {
	return uint64(vo)
}

func (vo UnixUserId) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
