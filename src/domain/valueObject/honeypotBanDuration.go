package tkValueObject

import (
	"errors"
	"time"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotBanDuration string

func NewHoneypotBanDuration(value any) (duration HoneypotBanDuration, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return duration, errors.New("HoneypotBanDurationMustBeString")
	}

	parsedDuration, parseErr := time.ParseDuration(stringValue)
	if parseErr != nil {
		return duration, errors.New("InvalidHoneypotBanDuration")
	}

	if parsedDuration <= 0 {
		return duration, errors.New("HoneypotBanDurationMustBePositive")
	}

	return HoneypotBanDuration(stringValue), nil
}

func (vo HoneypotBanDuration) Duration() time.Duration {
	duration, _ := time.ParseDuration(string(vo))
	return duration
}
