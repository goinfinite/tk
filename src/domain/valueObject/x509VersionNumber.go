package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type X509VersionNumber uint8

func NewX509VersionNumber(value any) (version X509VersionNumber, err error) {
	uintValue, err := tkVoUtil.InterfaceToUint8(value)
	if err != nil {
		return version, errors.New("X509VersionNumberMustBeUint8")
	}

	if uintValue < 1 || uintValue > 3 {
		return version, errors.New("InvalidX509VersionNumber")
	}

	return X509VersionNumber(uintValue), nil
}

func (vo X509VersionNumber) Uint8() uint8 {
	return uint8(vo)
}

func (vo X509VersionNumber) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
