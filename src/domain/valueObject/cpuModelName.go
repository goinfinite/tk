package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var cpuModelNameRegex = regexp.MustCompile(`^[A-Za-z0-9][\p{L}0-9 ._@()\-]{1,99}$`)

type CpuModelName string

func NewCpuModelName(value any) (cpuModelName CpuModelName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return cpuModelName, errors.New("CpuModelNameMustBeString")
	}

	if !cpuModelNameRegex.MatchString(stringValue) {
		return cpuModelName, errors.New("InvalidCpuModelName")
	}

	return CpuModelName(stringValue), nil
}

func (vo CpuModelName) String() string {
	return string(vo)
}
