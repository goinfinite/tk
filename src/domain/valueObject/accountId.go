package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var AccountIdSystem = AccountId(0)

type AccountId uint64

func NewAccountId(value any) (accountId AccountId, err error) {
	uint64Value, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return accountId, errors.New("AccountIdMustBeUint64")
	}

	return AccountId(uint64Value), nil
}

func (vo AccountId) Uint64() uint64 {
	return uint64(vo)
}

func (vo AccountId) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
