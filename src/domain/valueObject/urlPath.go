package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var urlPathRegex = regexp.MustCompile(`^(?P<path>\/[A-Za-z0-9\/\_\.\-]*)?(?P<query>\?[\w\/#=&%\-]*)?$`)

type UrlPath string

func NewUrlPath(value any) (urlPath UrlPath, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return urlPath, errors.New("UrlPathValueMustBeString")
	}

	hasLeadingSlash := strings.HasPrefix(stringValue, "/")
	if !hasLeadingSlash {
		stringValue = "/" + stringValue
	}

	if !urlPathRegex.MatchString(stringValue) {
		return urlPath, errors.New("InvalidUrlPath")
	}

	return UrlPath(stringValue), nil
}

func (vo UrlPath) String() string {
	return string(vo)
}

func (vo UrlPath) ReadWithoutQuery() string {
	return strings.Split(vo.String(), "?")[0]
}

func (vo UrlPath) ReadQuery() string {
	pathParts := strings.Split(vo.String(), "?")
	if len(pathParts) < 2 {
		return ""
	}
	return pathParts[1]
}

func (vo UrlPath) ReadWithoutTrailingSlash() string {
	return strings.TrimSuffix(vo.String(), "/")
}
