package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type CatalogItemType string

const (
	CatalogItemTypeApp       CatalogItemType = "app"
	CatalogItemTypeFramework CatalogItemType = "framework"
	CatalogItemTypeStack     CatalogItemType = "stack"
	CatalogItemTypeSystem    CatalogItemType = "system"
	CatalogItemTypeDatabase  CatalogItemType = "database"
	CatalogItemTypeRuntime   CatalogItemType = "runtime"
	CatalogItemTypeWebServer CatalogItemType = "webserver"
	CatalogItemTypeOther     CatalogItemType = "other"
)

var knownCatalogItemTypes = []string{
	CatalogItemTypeApp.String(), CatalogItemTypeFramework.String(),
	CatalogItemTypeStack.String(), CatalogItemTypeSystem.String(),
	CatalogItemTypeDatabase.String(), CatalogItemTypeRuntime.String(),
	CatalogItemTypeWebServer.String(), CatalogItemTypeOther.String(),
}

var catalogItemTypeRegex = regexp.MustCompile(
	"^(" + strings.Join(knownCatalogItemTypes, "|") + ")$",
)

func NewCatalogItemType(value any) (catalogItemType CatalogItemType, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return catalogItemType, errors.New("CatalogItemTypeMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	if !catalogItemTypeRegex.MatchString(stringValue) {
		return CatalogItemTypeOther, nil
	}

	return CatalogItemType(stringValue), nil
}

func (vo CatalogItemType) String() string {
	return string(vo)
}
