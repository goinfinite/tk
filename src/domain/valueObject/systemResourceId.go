package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var resourceIdRegex = regexp.MustCompile(`^([a-zA-Z0-9][\w\.\-]{0,512}|\*)$`)

type SystemResourceId string

func NewSystemResourceId(value any) (resourceId SystemResourceId, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return resourceId, errors.New("SystemResourceIdMustBeString")
	}

	if !resourceIdRegex.MatchString(stringValue) {
		return resourceId, errors.New("InvalidSystemResourceId")
	}

	return SystemResourceId(stringValue), nil
}

func (vo SystemResourceId) String() string {
	return string(vo)
}
