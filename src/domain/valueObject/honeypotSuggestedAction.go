package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	HoneypotSuggestedActionBan          HoneypotSuggestedAction = "ban"
	HoneypotSuggestedActionServePayload HoneypotSuggestedAction = "servePayload"
	HoneypotSuggestedActionServeStream  HoneypotSuggestedAction = "serveStream"
	HoneypotSuggestedActionServeAiTrap  HoneypotSuggestedAction = "serveAiTrap"
	HoneypotSuggestedActionServeMixed   HoneypotSuggestedAction = "serveMixed"
	HoneypotSuggestedActionPassthrough  HoneypotSuggestedAction = "passthrough"
)

type HoneypotSuggestedAction string

func NewHoneypotSuggestedAction(value any) (action HoneypotSuggestedAction, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return action, errors.New("HoneypotSuggestedActionMustBeString")
	}
	lowerValue := strings.ToLower(stringValue)

	switch lowerValue {
	case "ban":
		return HoneypotSuggestedActionBan, nil
	case "servepayload":
		return HoneypotSuggestedActionServePayload, nil
	case "servestream":
		return HoneypotSuggestedActionServeStream, nil
	case "serveaitrap":
		return HoneypotSuggestedActionServeAiTrap, nil
	case "servemixed":
		return HoneypotSuggestedActionServeMixed, nil
	case "passthrough":
		return HoneypotSuggestedActionPassthrough, nil
	default:
		return action, errors.New("InvalidHoneypotSuggestedAction")
	}
}

func (vo HoneypotSuggestedAction) String() string {
	return string(vo)
}
