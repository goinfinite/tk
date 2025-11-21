package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	resourceTypeRegex                            = regexp.MustCompile(`^[a-zA-Z][\w-]{0,255}$`)
	SystemResourceTypeAccount SystemResourceType = "account"
)

type SystemResourceType string

func NewSystemResourceType(value any) (resourceType SystemResourceType, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return resourceType, errors.New("SystemResourceTypeMustBeString")
	}

	if !resourceTypeRegex.MatchString(stringValue) {
		return resourceType, errors.New("InvalidSystemResourceType")
	}

	return SystemResourceType(stringValue), nil
}

func (vo SystemResourceType) String() string {
	return string(vo)
}
