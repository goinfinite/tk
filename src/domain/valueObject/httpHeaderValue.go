package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const (
	httpHeaderValueVisibleAsciiMin byte = 0x21
	httpHeaderValueVisibleAsciiMax byte = 0x7E
	httpHeaderValueObsTextMin      byte = 0x80
)

type HttpHeaderValue string

func NewHttpHeaderValue(value any) (httpHeaderValue HttpHeaderValue, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return httpHeaderValue, errors.New("HttpHeaderValueMustBeString")
	}

	for byteIndex := 0; byteIndex < len(stringValue); byteIndex++ {
		character := stringValue[byteIndex]
		isVisibleAscii := character >= httpHeaderValueVisibleAsciiMin &&
			character <= httpHeaderValueVisibleAsciiMax
		isObsText := character >= httpHeaderValueObsTextMin
		isSpaceOrTab := character == ' ' || character == '\t'
		if isVisibleAscii || isObsText || isSpaceOrTab {
			continue
		}
		return httpHeaderValue, errors.New("InvalidHttpHeaderValue")
	}

	return HttpHeaderValue(stringValue), nil
}

func (vo HttpHeaderValue) String() string {
	return string(vo)
}
