package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var paginationSortByRegex = regexp.MustCompile(`^[\p{L}\d\.\_\-\ ]{1,256}$`)

type PaginationSortBy string

func NewPaginationSortBy(value any) (sortBy PaginationSortBy, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return sortBy, errors.New("PaginationSortByMustBeString")
	}

	if !paginationSortByRegex.MatchString(stringValue) {
		return sortBy, errors.New("InvalidPaginationSortBy")
	}

	return PaginationSortBy(stringValue), nil
}

func (vo PaginationSortBy) String() string {
	return string(vo)
}
