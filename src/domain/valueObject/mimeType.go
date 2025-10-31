package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	MimeTypeDirectory MimeType = "directory"
	MimeTypeGeneric   MimeType = "generic"
)

const mimeTypeRegexExpression = `^[A-Za-z0-9\-]{1,64}\/[A-Za-z0-9\-\_\+\.\,]{2,128}$|^(directory|generic)$`

type MimeType string

func NewMimeType(value any) (mimeType MimeType, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return mimeType, errors.New("MimeTypeValueMustBeString")
	}

	re := regexp.MustCompile(mimeTypeRegexExpression)
	if !re.MatchString(stringValue) {
		return mimeType, errors.New("InvalidMimeTypeValue")
	}

	return MimeType(stringValue), nil
}

func (vo MimeType) String() string {
	return string(vo)
}

func (vo MimeType) IsDir() bool {
	return vo == MimeTypeDirectory
}
