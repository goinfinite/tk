package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	HoneypotPathClassStaticVulnerability HoneypotPathClass = "staticVulnerability"
	HoneypotPathClassBandwidthExhaust    HoneypotPathClass = "bandwidthExhaust"
	HoneypotPathClassAiTrap              HoneypotPathClass = "aiTrap"
)

type HoneypotPathClass string

func NewHoneypotPathClass(value any) (pathClass HoneypotPathClass, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return pathClass, errors.New("HoneypotPathClassMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	stringValueVo := HoneypotPathClass(stringValue)
	switch stringValueVo {
	case "staticvulnerability":
		return HoneypotPathClassStaticVulnerability, nil
	case "bandwidthexhaust":
		return HoneypotPathClassBandwidthExhaust, nil
	case "aitrap":
		return HoneypotPathClassAiTrap, nil
	default:
		return pathClass, errors.New("InvalidHoneypotPathClass")
	}
}

func (vo HoneypotPathClass) String() string {
	return string(vo)
}
