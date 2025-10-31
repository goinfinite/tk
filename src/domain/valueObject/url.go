package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const urlRegex string = `^(?P<schema>https?:\/\/)(?P<hostname>[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9][a-z0-9-]{0,61}[a-z0-9])*)(:(?P<port>\d{1,6}))?(?P<path>\/[A-Za-z0-9\/\_\.\-]*)?(?P<query>\?[\w\/#=&%\-]*)?$`

type Url string

func NewUrl(value any) (url Url, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return url, errors.New("UrlValueMustBeString")
	}

	if !strings.HasPrefix(stringValue, "http") {
		stringValue = "https://" + stringValue
	}

	urlRegex := regexp.MustCompile(urlRegex)
	if !urlRegex.MatchString(stringValue) {
		return url, errors.New("InvalidUrl")
	}

	return Url(stringValue), nil
}

func (vo Url) String() string {
	return string(vo)
}
