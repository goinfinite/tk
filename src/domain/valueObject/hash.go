package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var hashRegex = regexp.MustCompile(`^[\w\-\=]{6,512}$`)

type Hash string

func NewHash(value any) (hash Hash, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return hash, errors.New("HashMustBeString")
	}

	if !hashRegex.MatchString(stringValue) {
		return hash, errors.New("InvalidHash")
	}

	return Hash(stringValue), nil
}

func (vo Hash) String() string {
	return string(vo)
}
