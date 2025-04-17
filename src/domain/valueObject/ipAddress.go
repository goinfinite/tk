package tkValueObject

import (
	"errors"
	"net"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type IpAddress string

var IpAddressLocal = IpAddress("127.0.0.1")

func NewIpAddress(value any) (ipAddress IpAddress, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return ipAddress, errors.New("IpAddressValueMustBeString")
	}

	if stringValue == "" {
		return ipAddress, errors.New("IpAddressValueCannotBeEmpty")
	}

	if net.ParseIP(stringValue) == nil {
		return ipAddress, errors.New("InvalidIpAddress")
	}

	return IpAddress(stringValue), nil
}

func (vo IpAddress) String() string {
	return string(vo)
}

func (vo IpAddress) IsLocal() bool {
	return vo == IpAddressLocal
}

func (vo IpAddress) IsIpv4() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.To4() != nil
}

func (vo IpAddress) IsIpv6() bool {
	return !vo.IsIpv4()
}

func (vo IpAddress) IsPrivate() bool {
	parsedIpAddress := net.ParseIP(vo.String())
	if parsedIpAddress == nil {
		return false
	}
	return parsedIpAddress.IsPrivate()
}

func (vo IpAddress) IsPublic() bool {
	return !vo.IsPrivate()
}

func (vo IpAddress) ToCidrBlock() CidrBlock {
	if vo.IsIpv6() {
		return CidrBlock(vo.String() + "/128")
	}
	return CidrBlock(vo.String() + "/32")
}
