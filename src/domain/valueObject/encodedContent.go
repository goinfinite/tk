package tkValueObject

import (
	"encoding/base64"
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var encodedContentRegex = regexp.MustCompile(`^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{4}|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{2}={2})$`)

type EncodedContent string

func NewEncodedContent(value any) (encodedContent EncodedContent, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return encodedContent, errors.New("EncodedContentMustBeString")
	}

	if len(stringValue) == 0 {
		return encodedContent, errors.New("EncodedContentCannotBeEmpty")
	}

	if len(stringValue) > 10485760 {
		return encodedContent, errors.New("EncodedContentTooBig")
	}

	if !encodedContentRegex.MatchString(stringValue) {
		return encodedContent, errors.New("InvalidEncodedContent")
	}

	return EncodedContent(stringValue), nil
}

func (vo EncodedContent) ReadDecodedContent() (decodedContent string, err error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(string(vo))
	if err != nil {
		return decodedContent, err
	}

	return string(decodedBytes), nil
}

func (vo EncodedContent) String() string {
	return string(vo)
}
