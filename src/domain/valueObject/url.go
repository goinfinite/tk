package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var urlRegex = regexp.MustCompile(`^(?P<scheme>(https?|wss?|grpcs?|tcp|udp|ftp|mailto|file|data|irc):\/\/)?(?P<hostname>[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9][a-z0-9-]{0,61}[a-z0-9])*)(:(?P<port>\d{1,6}))?(?P<path>\/[A-Za-z0-9\/\_\.\-]*)?(?P<query>\?[\w\/#=&%\-]*)?$`)

type Url string

func NewUrl(value any) (url Url, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return url, errors.New("UrlValueMustBeString")
	}

	if !urlRegex.MatchString(stringValue) {
		return url, errors.New("InvalidUrl")
	}

	namedGroupsValuesMap := tkVoUtil.NamedGroupsExtractor(urlRegex, stringValue)
	if namedGroupsValuesMap["scheme"] == "" {
		stringValue = "https://" + stringValue
	}

	return Url(stringValue), nil
}

func (vo Url) String() string {
	return string(vo)
}
