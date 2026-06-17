package tkValueObject

import (
	"log/slog"
	"time"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotStatsInterval time.Duration

func NewHoneypotStatsInterval(
	rawValue any,
) (statsInterval HoneypotStatsInterval, err error) {
	defaultValue := 30 * time.Minute
	floorValue := 5 * time.Minute

	stringValue, stringConvErr := tkVoUtil.InterfaceToString(rawValue)
	if stringConvErr != nil {
		return HoneypotStatsInterval(defaultValue), nil
	}

	if stringValue == "" {
		return HoneypotStatsInterval(defaultValue), nil
	}

	parsedDuration, parseErr := time.ParseDuration(stringValue)
	if parseErr != nil {
		intValue, intConvErr := tkVoUtil.InterfaceToInt(rawValue)
		if intConvErr != nil {
			slog.Debug("StatsIntervalNonNumericFallback",
				slog.String("raw", stringValue),
				slog.String("resolved", defaultValue.String()))
			return HoneypotStatsInterval(defaultValue), nil
		}
		parsedDuration = time.Duration(intValue) * time.Second
	}

	if parsedDuration < floorValue {
		slog.Debug("StatsIntervalBelowFloorClamped",
			slog.String("raw", parsedDuration.String()),
			slog.String("floor", floorValue.String()),
			slog.String("resolved", floorValue.String()))
		return HoneypotStatsInterval(floorValue), nil
	}

	return HoneypotStatsInterval(parsedDuration), nil
}

func (vo HoneypotStatsInterval) Duration() time.Duration {
	return time.Duration(vo)
}
