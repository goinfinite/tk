package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	PaginationSortDirectionAsc  PaginationSortDirection = "asc"
	PaginationSortDirectionDesc PaginationSortDirection = "desc"
)

type PaginationSortDirection string

func NewPaginationSortDirection(value any) (
	sortDirection PaginationSortDirection, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return sortDirection, errors.New("PaginationSortDirectionMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	stringValueVo := PaginationSortDirection(stringValue)
	switch stringValueVo {
	case PaginationSortDirectionAsc, PaginationSortDirectionDesc:
		return stringValueVo, nil
	default:
		return sortDirection, errors.New("InvalidPaginationSortDirection")
	}
}

func (vo PaginationSortDirection) String() string {
	return string(vo)
}
