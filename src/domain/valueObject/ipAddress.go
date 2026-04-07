package tkValueObject

import (
	"errors"
	"net"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type IpAddress string

var IpAddressLocal = IpAddress("127.0.0.1")

func NewIpAddress(value any) (ipAddress IpAddress, err error) {
	if existentIpAddress, assertOk := value.(IpAddress); assertOk {
		return existentIpAddress, nil
	}

	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return ipAddress, errors.New("IpAddressValueMustBeString")
	}

	if stringValue == "" {
		return ipAddress, errors.New("IpAddressValueCannotBeEmpty")
	}

	stringValue, _, _ = strings.Cut(stringValue, "%")
	if net.ParseIP(stringValue) == nil {
		return ipAddress, errors.New("InvalidIpAddress")
	}

	return IpAddress(stringValue), nil
}

func (vo IpAddress) String() string {
	return string(vo)
}

func (vo IpAddress) IsLocal() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.IsLoopback()
}

func (vo IpAddress) IsIpv4() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.To4() != nil
}

func (vo IpAddress) IsIpv6() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.To4() == nil
}

func (vo IpAddress) IsLinkLocal() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.IsLinkLocalUnicast()
}

func (vo IpAddress) IsPrivate() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.IsPrivate()
}

func (vo IpAddress) IsPublic() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return !parsedIpAddress.IsPrivate() && !parsedIpAddress.IsLoopback()
}

func (vo IpAddress) ToCidrBlock() CidrBlock {
	if vo.IsIpv6() {
		return CidrBlock(vo.String() + "/128")
	}
	return CidrBlock(vo.String() + "/32")
}
