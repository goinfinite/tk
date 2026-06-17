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
