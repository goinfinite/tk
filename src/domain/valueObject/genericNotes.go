package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const HumanlyUsedCharsRegex string = `^[\p{L}\p{N}\p{Pd}\p{Pi}\p{Pf}\p{Pc}\p{Po}\p{Z}\p{Sc}\(\)\[\]\+\=]+$`

type GenericNotes string

func NewGenericNotes(value any) (genericNotes GenericNotes, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return genericNotes, errors.New("GenericNotesMustBeString")
	}

	if len(stringValue) < 1 {
		return genericNotes, errors.New("GenericNotesTooSmall")
	}

	if len(stringValue) > 5000 {
		return genericNotes, errors.New("GenericNotesTooBig")
	}

	re := regexp.MustCompile(HumanlyUsedCharsRegex)
	if !re.MatchString(stringValue) {
		return genericNotes, errors.New("GenericNotesMustOnlyContainHumanlyUsedChars")
	}

	return GenericNotes(stringValue), nil
}

func (vo GenericNotes) String() string {
	return string(vo)
}
