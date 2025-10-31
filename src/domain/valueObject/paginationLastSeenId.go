package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const paginationLastSeenIdRegex string = `^[\w\-]{1,256}$`

type PaginationLastSeenId string

func NewPaginationLastSeenId(value any) (lastSeenId PaginationLastSeenId, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return lastSeenId, errors.New("PaginationLastSeenIdMustBeString")
	}

	re := regexp.MustCompile(paginationLastSeenIdRegex)
	if !re.MatchString(stringValue) {
		return lastSeenId, errors.New("InvalidPaginationLastSeenId")
	}

	return PaginationLastSeenId(stringValue), nil
}

func (vo PaginationLastSeenId) String() string {
	return string(vo)
}
