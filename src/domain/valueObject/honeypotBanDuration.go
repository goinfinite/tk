package tkValueObject

import (
	"log/slog"
	"time"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotBanDuration time.Duration

func NewHoneypotBanDuration(
	rawValue any,
) (banDuration HoneypotBanDuration, err error) {
	defaultValue := 24 * time.Hour

	switch typedValue := rawValue.(type) {
	case time.Duration:
		if typedValue <= 0 {
			return HoneypotBanDuration(defaultValue), nil
		}
		return HoneypotBanDuration(typedValue), nil
	}

	stringValue, stringConvErr := tkVoUtil.InterfaceToString(rawValue)
	if stringConvErr != nil {
		return HoneypotBanDuration(defaultValue), nil
	}

	if stringValue == "" {
		return HoneypotBanDuration(defaultValue), nil
	}

	parsedDuration, parseErr := time.ParseDuration(stringValue)
	if parseErr != nil {
		intValue, intConvErr := tkVoUtil.InterfaceToInt(rawValue)
		if intConvErr != nil {
			slog.Debug("BanDurationNonNumericFallback",
				slog.String("raw", stringValue),
				slog.String("resolved",
					defaultValue.String()))
			return HoneypotBanDuration(defaultValue), nil
		}
		parsedDuration = time.Duration(intValue) * time.Second
	}

	if parsedDuration <= 0 {
		slog.Debug("BanDurationNonPositiveClamped",
			slog.String("raw", parsedDuration.String()),
			slog.String("resolved",
				defaultValue.String()))
		return HoneypotBanDuration(defaultValue), nil
	}

	return HoneypotBanDuration(parsedDuration), nil
}

func (vo HoneypotBanDuration) Duration() time.Duration {
	return time.Duration(vo)
}
