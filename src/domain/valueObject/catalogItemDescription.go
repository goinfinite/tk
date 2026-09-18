package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type CatalogItemDescription string

func NewCatalogItemDescription(value any) (
	catalogItemDescription CatalogItemDescription, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return catalogItemDescription, errors.New(
			"CatalogItemDescriptionMustBeString",
		)
	}

	if len(stringValue) < 2 {
		return catalogItemDescription, errors.New(
			"CatalogItemDescriptionTooSmall",
		)
	}

	if len(stringValue) > 2048 {
		return catalogItemDescription, errors.New(
			"CatalogItemDescriptionTooBig",
		)
	}

	return CatalogItemDescription(stringValue), nil
}

func (vo CatalogItemDescription) String() string {
	return string(vo)
}
