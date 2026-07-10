package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const honeypotActivePathCountFloor = 30

type HoneypotActivePathCount int

func NewHoneypotActivePathCount(value any, ceiling int) (count HoneypotActivePathCount, err error) {
	intValue, err := tkVoUtil.InterfaceToInt(value)
	if err != nil {
		return count, errors.New("HoneypotActivePathCountMustBeInteger")
	}

	if intValue < honeypotActivePathCountFloor {
		intValue = honeypotActivePathCountFloor
	}
	if intValue > ceiling {
		intValue = ceiling
	}

	return HoneypotActivePathCount(intValue), nil
}

func (vo HoneypotActivePathCount) Int() int {
	return int(vo)
}
