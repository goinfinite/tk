package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	HoneypotAggressivenessModeImmediate HoneypotAggressivenessMode = "immediate"
	HoneypotAggressivenessModeBalanced  HoneypotAggressivenessMode = "balanced"
	HoneypotAggressivenessModeTolerant  HoneypotAggressivenessMode = "tolerant"
	HoneypotAggressivenessModeObserve   HoneypotAggressivenessMode = "observe"
)

type HoneypotAggressivenessMode string

func NewHoneypotAggressivenessMode(value any) (mode HoneypotAggressivenessMode, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return mode, errors.New("HoneypotAggressivenessModeMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	stringValueVo := HoneypotAggressivenessMode(stringValue)
	switch stringValueVo {
	case HoneypotAggressivenessModeImmediate, HoneypotAggressivenessModeBalanced,
		HoneypotAggressivenessModeTolerant, HoneypotAggressivenessModeObserve:
		return stringValueVo, nil
	default:
		return mode, errors.New("InvalidHoneypotAggressivenessMode")
	}
}

func (vo HoneypotAggressivenessMode) String() string {
	return string(vo)
}
