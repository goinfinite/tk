package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type HoneypotAggressivenessMode string

var (
	HoneypotAggressivenessModeImmediate HoneypotAggressivenessMode = "immediate"
	HoneypotAggressivenessModeBalanced  HoneypotAggressivenessMode = "balanced"
	HoneypotAggressivenessModeTolerant  HoneypotAggressivenessMode = "tolerant"
	HoneypotAggressivenessModeObserve   HoneypotAggressivenessMode = "observe"
)

func NewHoneypotAggressivenessMode(value any) (
	aggressivenessMode HoneypotAggressivenessMode, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return aggressivenessMode, errors.New(
			"HoneypotAggressivenessModeMustBeString",
		)
	}

	aggressivenessMode = HoneypotAggressivenessMode(stringValue)
	switch aggressivenessMode {
	case HoneypotAggressivenessModeImmediate,
		HoneypotAggressivenessModeBalanced,
		HoneypotAggressivenessModeTolerant,
		HoneypotAggressivenessModeObserve:
		return aggressivenessMode, nil
	default:
		return aggressivenessMode, errors.New(
			"InvalidHoneypotAggressivenessMode",
		)
	}
}

func (vo HoneypotAggressivenessMode) String() string {
	return string(vo)
}

func (vo HoneypotAggressivenessMode) ResolveTier(hitCount int) int {
	switch vo {
	case HoneypotAggressivenessModeImmediate:
		if hitCount >= 1 {
			return 3
		}
		return 0
	case HoneypotAggressivenessModeBalanced:
		if hitCount >= 3 {
			return 3
		}
		if hitCount <= 0 {
			return 0
		}
		return hitCount
	case HoneypotAggressivenessModeTolerant:
		if hitCount >= 5 {
			return 2
		}
		if hitCount >= 2 {
			return 1
		}
		return 0
	case HoneypotAggressivenessModeObserve:
		return 1
	default:
		return 0
	}
}
