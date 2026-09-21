package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var catalogItemNameRegex = regexp.MustCompile(`^[A-Za-z0-9][\p{L}0-9._' -]{0,63}$`)

type CatalogItemName string

func NewCatalogItemName(value any) (catalogItemName CatalogItemName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return catalogItemName, errors.New("CatalogItemNameMustBeString")
	}

	stringValue = strings.TrimSpace(stringValue)

	if !catalogItemNameRegex.MatchString(stringValue) {
		return catalogItemName, errors.New("InvalidCatalogItemName")
	}

	return CatalogItemName(stringValue), nil
}

func (vo CatalogItemName) String() string {
	return string(vo)
}
