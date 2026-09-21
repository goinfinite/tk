package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var shortDescriptionRegex = regexp.MustCompile(`^[^\p{Cc}]+$`)

type ShortDescription string

func NewShortDescription(value any) (
	shortDescription ShortDescription, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return shortDescription, errors.New("ShortDescriptionMustBeString")
	}

	if len(stringValue) < 2 {
		return shortDescription, errors.New("ShortDescriptionTooSmall")
	}

	if len(stringValue) > 2048 {
		return shortDescription, errors.New("ShortDescriptionTooBig")
	}

	if !shortDescriptionRegex.MatchString(stringValue) {
		return shortDescription, errors.New("InvalidShortDescription")
	}

	return ShortDescription(stringValue), nil
}

func (vo ShortDescription) String() string {
	return string(vo)
}
