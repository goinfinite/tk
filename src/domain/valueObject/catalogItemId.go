package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type CatalogItemId uint16

func NewCatalogItemId(value any) (catalogItemId CatalogItemId, err error) {
	if existentCatalogItemId, assertOk := value.(CatalogItemId); assertOk {
		return existentCatalogItemId, nil
	}

	uint16Value, err := tkVoUtil.InterfaceToUint16(value)
	if err != nil {
		return catalogItemId, errors.New("CatalogItemIdMustBeUint16")
	}

	return CatalogItemId(uint16Value), nil
}

func (vo CatalogItemId) Uint16() uint16 {
	return uint16(vo)
}

func (vo CatalogItemId) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
