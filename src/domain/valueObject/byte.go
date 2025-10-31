package tkValueObject

import (
	"errors"
	"fmt"
	"math"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type Byte uint64

func NewByte(value any) (byteVo Byte, err error) {
	valueUint64, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return byteVo, errors.New("ByteMustBeUint64")
	}

	return Byte(valueUint64), nil
}

func NewKibibyte(value any) (byteVo Byte, err error) {
	valueUint64, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return byteVo, errors.New("KibibytesMustBeUint64")
	}

	return Byte(valueUint64 * 1024), nil
}

func NewMebibyte(value any) (byteVo Byte, err error) {
	valueUint64, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return byteVo, errors.New("MebibytesMustBeUint64")
	}

	return Byte(valueUint64 * 1048576), nil
}

func NewGibibyte(value any) (byteVo Byte, err error) {
	valueUint64, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return byteVo, errors.New("GibibytesMustBeUint64")
	}

	return Byte(valueUint64 * 1073741824), nil
}

func NewTebibyte(value any) (byteVo Byte, err error) {
	valueUint64, err := tkVoUtil.InterfaceToUint64(value)
	if err != nil {
		return byteVo, errors.New("TebibytesMustBeUint64")
	}

	return Byte(valueUint64 * 1099511627776), nil
}

func (vo Byte) Int64() int64 {
	return int64(vo)
}

func (vo Byte) Uint64() uint64 {
	return uint64(vo)
}

func (vo Byte) Float64() float64 {
	return float64(vo)
}

func (vo Byte) ToKiB() uint64 {
	return uint64(math.Round(vo.Float64() / 1024))
}

func (vo Byte) ToMiB() uint64 {
	return uint64(math.Round(vo.Float64() / 1048576))
}

func (vo Byte) ToGiB() uint64 {
	return uint64(math.Round(vo.Float64() / 1073741824))
}

func (vo Byte) ToTiB() uint64 {
	return uint64(math.Round(vo.Float64() / 1099511627776))
}

func (vo Byte) String() string {
	return fmt.Sprintf("%d", vo.Int64())
}

func (vo Byte) StringWithSuffix() string {
	voUint64 := vo.Uint64()
	switch {
	case voUint64 < 1024:
		return fmt.Sprintf("%d B", voUint64)
	case voUint64 < 1048576:
		return fmt.Sprintf("%d KiB", vo.ToKiB())
	case voUint64 < 1073741824:
		return fmt.Sprintf("%d MiB", vo.ToMiB())
	case voUint64 < 1099511627776:
		return fmt.Sprintf("%d GiB", vo.ToGiB())
	case voUint64 < 1125899906842624:
		return fmt.Sprintf("%d TiB", vo.ToTiB())
	default:
		return fmt.Sprintf("%d B", voUint64)
	}
}
