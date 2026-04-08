package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var httpHeaderRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

type HttpHeader string

func NewHttpHeader(value any) (httpHeader HttpHeader, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return httpHeader, errors.New("HttpHeaderMustBeString")
	}

	if stringValue == "" {
		return httpHeader, errors.New("HttpHeaderCannotBeEmpty")
	}

	if !httpHeaderRegex.MatchString(stringValue) {
		return httpHeader, errors.New("InvalidHttpHeader")
	}

	return HttpHeader(stringValue), nil
}

func (vo HttpHeader) String() string {
	return string(vo)
}
