package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type ActivityRecordId uint64

func NewActivityRecordId(value any) (recordId ActivityRecordId, err error) {
	uintValue, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return recordId, errors.New("ActivityRecordIdMustBeUint64")
	}

	return ActivityRecordId(uintValue), nil
}

func (vo ActivityRecordId) Uint64() uint64 {
	return uint64(vo)
}

func (vo ActivityRecordId) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
