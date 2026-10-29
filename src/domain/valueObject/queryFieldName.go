package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var queryFieldNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]{0,255}$`)

type QueryFieldName string

func NewQueryFieldName(value any) (queryFieldName QueryFieldName, err error) {
	existentQueryFieldName, isQueryFieldName := value.(QueryFieldName)
	if isQueryFieldName {
		return existentQueryFieldName, nil
	}

	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return queryFieldName, errors.New("QueryFieldNameMustBeString")
	}

	if stringValue == "" {
		return queryFieldName, errors.New("QueryFieldNameCannotBeEmpty")
	}

	if len(stringValue) > 256 {
		return queryFieldName, errors.New("QueryFieldNameTooBig")
	}

	if !queryFieldNameRegex.MatchString(stringValue) {
		return queryFieldName, errors.New("InvalidQueryFieldName")
	}

	return QueryFieldName(stringValue), nil
}

func (vo QueryFieldName) String() string {
	return string(vo)
}
