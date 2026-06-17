package tkValueObject

import (
	"log/slog"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotActivePathCount int

func NewHoneypotActivePathCount(
	rawValue any,
	poolCeiling int,
) (activePathCount HoneypotActivePathCount, err error) {
	defaultValue := 30
	floorValue := 30

	parsedValue, parseErr := tkVoUtil.InterfaceToInt(rawValue)
	if parseErr != nil {
		return HoneypotActivePathCount(defaultValue), nil
	}

	if parsedValue < floorValue {
		slog.Debug("ActivePathCountBelowFloorClamped",
			slog.Int("raw", parsedValue),
			slog.Int("floor", floorValue),
			slog.Int("resolved", defaultValue))
		return HoneypotActivePathCount(defaultValue), nil
	}

	if poolCeiling > 0 {
		if poolCeiling < floorValue {
			slog.Debug("ActivePathCountPoolCeilingBelowFloor",
				slog.Int("ceiling", poolCeiling),
				slog.Int("floor", floorValue),
				slog.Int("resolved", defaultValue))
			return HoneypotActivePathCount(defaultValue), nil
		}

		if parsedValue > poolCeiling {
			return HoneypotActivePathCount(poolCeiling), nil
		}
	}

	return HoneypotActivePathCount(parsedValue), nil
}

func (vo HoneypotActivePathCount) Int() int {
	return int(vo)
}
