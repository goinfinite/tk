package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	HttpMethodGet     HttpMethod = "GET"
	HttpMethodHead    HttpMethod = "HEAD"
	HttpMethodPost    HttpMethod = "POST"
	HttpMethodPut     HttpMethod = "PUT"
	HttpMethodDelete  HttpMethod = "DELETE"
	HttpMethodConnect HttpMethod = "CONNECT"
	HttpMethodOptions HttpMethod = "OPTIONS"
	HttpMethodTrace   HttpMethod = "TRACE"
	HttpMethodPatch   HttpMethod = "PATCH"
)

type HttpMethod string

func NewHttpMethod(value any) (vo HttpMethod, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return vo, errors.New("HttpMethodMustBeString")
	}

	stringValueVo := HttpMethod(stringValue)
	switch stringValueVo {
	case HttpMethodGet, HttpMethodPost, HttpMethodPut, HttpMethodDelete,
		HttpMethodHead, HttpMethodConnect, HttpMethodOptions, HttpMethodTrace,
		HttpMethodPatch:
		return stringValueVo, nil
	default:
		return vo, errors.New("InvalidHttpMethod")
	}
}

func (vo HttpMethod) String() string {
	return string(vo)
}

func (vo HttpMethod) HasBodySupport() bool {
	switch vo {
	case HttpMethodPost, HttpMethodPut, HttpMethodPatch, HttpMethodOptions:
		return true
	default:
		return false
	}
}
