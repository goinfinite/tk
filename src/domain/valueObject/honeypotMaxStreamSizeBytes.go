package tkValueObject

import (
	"log/slog"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotMaxStreamSizeBytes int64

func NewHoneypotMaxStreamSizeBytes(
	rawValue any,
) (maxStreamSizeBytes HoneypotMaxStreamSizeBytes, err error) {
	defaultValue := int64(20 * 1024 * 1024)
	floorValue := int64(5 * 1024 * 1024)

	parsedValue, parseErr := tkVoUtil.InterfaceToInt64(rawValue)
	if parseErr != nil {
		return HoneypotMaxStreamSizeBytes(defaultValue), nil
	}

	if parsedValue < floorValue {
		slog.Debug("MaxStreamSizeBelowFloorClamped",
			slog.Int64("raw", parsedValue),
			slog.Int64("floor", floorValue),
			slog.Int64("resolved", defaultValue))
		return HoneypotMaxStreamSizeBytes(defaultValue), nil
	}

	return HoneypotMaxStreamSizeBytes(parsedValue), nil
}

func (vo HoneypotMaxStreamSizeBytes) Int64() int64 {
	return int64(vo)
}
