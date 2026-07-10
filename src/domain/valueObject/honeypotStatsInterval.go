package tkValueObject

import (
	"errors"
	"time"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotStatsInterval string

func NewHoneypotStatsInterval(value any) (interval HoneypotStatsInterval, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return interval, errors.New("HoneypotStatsIntervalMustBeString")
	}

	_, parseErr := time.ParseDuration(stringValue)
	if parseErr != nil {
		return interval, errors.New("InvalidHoneypotStatsInterval")
	}

	return HoneypotStatsInterval(stringValue), nil
}

func (vo HoneypotStatsInterval) Duration() time.Duration {
	duration, _ := time.ParseDuration(string(vo))
	return duration
}
