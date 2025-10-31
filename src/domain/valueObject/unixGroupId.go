package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type UnixGroupId uint64

func NewUnixGroupId(value any) (UnixGroupId, error) {
	uint64Value, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return 0, errors.New("InvalidGroupId")
	}

	return UnixGroupId(uint64Value), nil
}

func (vo UnixGroupId) Uint64() uint64 {
	return uint64(vo)
}

func (vo UnixGroupId) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
