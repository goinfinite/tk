package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var catalogItemSlugRegex = regexp.MustCompile(`^[a-z0-9_-]{2,64}$`)

type CatalogItemSlug string

func NewCatalogItemSlug(value any) (catalogItemSlug CatalogItemSlug, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return catalogItemSlug, errors.New("CatalogItemSlugMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	if !catalogItemSlugRegex.MatchString(stringValue) {
		return catalogItemSlug, errors.New("InvalidCatalogItemSlug")
	}

	return CatalogItemSlug(stringValue), nil
}

func (vo CatalogItemSlug) String() string {
	return string(vo)
}
