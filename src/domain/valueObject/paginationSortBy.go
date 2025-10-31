package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const paginationSortByRegex string = `^[\p{L}\d\.\_\-\ ]{1,256}$`

type PaginationSortBy string

func NewPaginationSortBy(value any) (sortBy PaginationSortBy, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return sortBy, errors.New("PaginationSortByMustBeString")
	}

	re := regexp.MustCompile(paginationSortByRegex)
	if !re.MatchString(stringValue) {
		return sortBy, errors.New("InvalidPaginationSortBy")
	}

	return PaginationSortBy(stringValue), nil
}

func (vo PaginationSortBy) String() string {
	return string(vo)
}
