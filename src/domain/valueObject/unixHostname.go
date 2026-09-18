package tkValueObject

import (
	"errors"
	"net/netip"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	unixHostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	ipv6ZoneRegex     = regexp.MustCompile(`^[a-zA-Z0-9._~-]{1,255}$`)
)

type UnixHostname string

func NewUnixHostname(value any) (hostname UnixHostname, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return hostname, errors.New("UnixHostnameMustBeString")
	}

	if ipAddress, parseErr := netip.ParseAddr(stringValue); parseErr == nil {
		if ipAddress.Zone() != "" && !ipv6ZoneRegex.MatchString(ipAddress.Zone()) {
			return hostname, errors.New("InvalidUnixHostname")
		}
		return UnixHostname(ipAddress.String()), nil
	}

	stringValue = strings.ToLower(stringValue)

	if len(stringValue) > 253 {
		return hostname, errors.New("UnixHostnameTooBig")
	}

	if !unixHostnameRegex.MatchString(stringValue) {
		return hostname, errors.New("InvalidUnixHostname")
	}

	return UnixHostname(stringValue), nil
}

func (vo UnixHostname) String() string {
	return string(vo)
}

func (vo UnixHostname) ToUrlHost() string {
	ipAddress, parseErr := netip.ParseAddr(vo.String())
	if parseErr != nil || ipAddress.Is4() {
		return vo.String()
	}
	return "[" + vo.String() + "]"
}

func (vo UnixHostname) ToUrlEncodedHost() string {
	return strings.ReplaceAll(vo.ToUrlHost(), "%", "%25")
}
