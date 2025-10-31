package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type NetworkPort uint16

func NewNetworkPort(value interface{}) (networkPort NetworkPort, err error) {
	uintValue, err := tkVoUtil.InterfaceToUint16(value)
	if err != nil {
		return networkPort, errors.New("NetworkPortMustBeUint16")
	}

	return NetworkPort(uintValue), nil
}

func (vo NetworkPort) Uint16() uint16 {
	return uint16(vo)
}

func (vo NetworkPort) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
