package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var fieldNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$`)

type FieldName string

func NewFieldName(value any) (fieldName FieldName, err error) {
	existentFieldName, isFieldName := value.(FieldName)
	if isFieldName {
		return existentFieldName, nil
	}

	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return fieldName, errors.New("FieldNameMustBeString")
	}

	if stringValue == "" {
		return fieldName, errors.New("FieldNameCannotBeEmpty")
	}

	if len(stringValue) > 64 {
		return fieldName, errors.New("FieldNameTooBig")
	}

	if !fieldNameRegex.MatchString(stringValue) {
		return fieldName, errors.New("InvalidFieldName")
	}

	return FieldName(stringValue), nil
}

func (vo FieldName) String() string {
	return string(vo)
}
