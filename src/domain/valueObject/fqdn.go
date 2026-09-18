package tkValueObject

import (
	"errors"
	"net"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var fqdnRegex = regexp.MustCompile(`^((\*\.)?([a-zA-Z0-9_]+[\w-]*\.)*)?([a-zA-Z0-9_][\w-]*[a-zA-Z0-9])\.([a-zA-Z]{2,})$`)

type Fqdn string

func NewFqdn(value any) (fqdn Fqdn, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return fqdn, errors.New("FqdnMustBeString")
	}
	stringValue = strings.ToLower(stringValue)

	if len(stringValue) > 253 {
		return fqdn, errors.New("FqdnTooBig")
	}

	isIpAddress := net.ParseIP(stringValue) != nil
	if isIpAddress {
		return fqdn, errors.New("FqdnCannotBeIpAddress")
	}

	if !fqdnRegex.MatchString(stringValue) {
		return fqdn, errors.New("InvalidFqdn")
	}

	return Fqdn(stringValue), nil
}

func (vo Fqdn) String() string {
	return string(vo)
}
