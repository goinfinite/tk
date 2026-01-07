package tkValueObject

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var urlRegex = regexp.MustCompile(`^(?:(?:mailto:)?(?P<mailUsername>[a-z0-9._%+-]+)@(?i:(?P<mailHostname>[a-z0-9.-]+\.[a-z]{2,}))|(?:tel:)?(?P<phone>\+?\d{6,15}(?:-\d{1,12})?)|(?:(?P<scheme>(?:https?|wss?|grpcs?|tcp|udp|ftp|ftps|file|data|irc|imap|nntp|pop3|smtp|telnet):\/\/)?(?:(?P<userAuthUsername>[a-z0-9._~%!$&'()*+,;=-]+)(?::(?P<userAuthPassword>[a-z0-9._~%!$&'()*+,;=:-]+))?@)?(?i:(?P<hostname>[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9][a-z0-9-]{0,61}[a-z0-9])*))(?::(?P<networkPort>\d{1,6}))?(?P<path>\/[A-Za-z0-9\/\_\.\-]*)?(?P<query>\?[\w\/#=&%\-]*)?))$`)

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
		if namedGroupsValuesMap["hostname"] != "" {
			stringValue = "https://" + stringValue
		}

		if namedGroupsValuesMap["phone"] != "" && !strings.HasPrefix(stringValue, "tel:") {
			stringValue = "tel:" + stringValue
		}

		if namedGroupsValuesMap["mailUsername"] != "" && !strings.HasPrefix(stringValue, "mailto:") {
			stringValue = "mailto:" + stringValue
		}
	}

	if namedGroupsValuesMap["hostname"] != "" {
		lowercaseHostname := strings.ToLower(namedGroupsValuesMap["hostname"])
		stringValue = strings.ReplaceAll(
			stringValue, namedGroupsValuesMap["hostname"], lowercaseHostname,
		)
	}

	if namedGroupsValuesMap["mailHostname"] != "" {
		lowercaseMailHostname := strings.ToLower(namedGroupsValuesMap["mailHostname"])
		stringValue = strings.ReplaceAll(
			stringValue, namedGroupsValuesMap["mailHostname"], lowercaseMailHostname,
		)
	}

	if namedGroupsValuesMap["networkPort"] != "" {
		networkPort, err := strconv.Atoi(namedGroupsValuesMap["networkPort"])
		if err != nil {
			return url, errors.New("InvalidNetworkPort")
		}

		if networkPort < 0 || networkPort > 65535 {
			return url, errors.New("InvalidNetworkPort")
		}
	}

	return Url(stringValue), nil
}

func (vo Url) String() string {
	return string(vo)
}
