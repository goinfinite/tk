package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var relativeTimeRegex = regexp.MustCompile(`^(?i)(\d+(?:\.\d+)?)\s*(second|minute|hour|day|week|month|year|s|m|h|d|w|M|y)(?:s?)\s*(ago|from now)?$`)

type RelativeTime string

func NewRelativeTime(value any) (relativeTime RelativeTime, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return relativeTime, errors.New("RelativeTimeMustBeString")
	}

	if !relativeTimeRegex.MatchString(stringValue) {
		return relativeTime, errors.New("InvalidRelativeTime")
	}

	return RelativeTime(stringValue), nil
}

func (vo RelativeTime) String() string {
	return string(vo)
}
