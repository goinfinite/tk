package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HttpStatusCode uint16

func NewHttpStatusCode(value any) (statusCode HttpStatusCode, err error) {
	uint16Value, err := tkVoUtil.InterfaceToUint16(value)
	if err != nil {
		return statusCode, errors.New("HttpStatusCodeMustBeUint16")
	}

	if uint16Value < 100 || uint16Value > 599 {
		return statusCode, errors.New("InvalidHttpStatusCode")
	}

	return HttpStatusCode(uint16Value), nil
}

func (vo HttpStatusCode) Uint16() uint16 {
	return uint16(vo)
}

func (vo HttpStatusCode) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
