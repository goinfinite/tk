package tkValueObject

import (
	"log/slog"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotMaxEntries int

func NewHoneypotMaxEntries(
	rawValue any,
) (maxEntries HoneypotMaxEntries, err error) {
	defaultValue := 5000
	floorValue := 100
	ceilingValue := 50000

	parsedValue, parseErr := tkVoUtil.InterfaceToInt(rawValue)
	if parseErr != nil {
		return HoneypotMaxEntries(defaultValue), nil
	}

	if parsedValue < floorValue {
		slog.Debug("MaxEntriesBelowFloorClamped",
			slog.Int("raw", parsedValue),
			slog.Int("floor", floorValue),
			slog.Int("resolved", defaultValue))
		return HoneypotMaxEntries(defaultValue), nil
	}

	if parsedValue > ceilingValue {
		return HoneypotMaxEntries(ceilingValue), nil
	}

	return HoneypotMaxEntries(parsedValue), nil
}

func (vo HoneypotMaxEntries) Int() int {
	return int(vo)
}
