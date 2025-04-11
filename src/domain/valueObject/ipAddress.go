package tkValueObject

import (
	"errors"
	"net"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type IpAddress string

var IpAddressSystem = IpAddress("127.0.0.1")

func NewIpAddress(value any) (ipAddress IpAddress, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return ipAddress, errors.New("IpAddressValueMustBeString")
	}

	parsedIpAddress := net.ParseIP(stringValue)
	if parsedIpAddress == nil {
		return ipAddress, errors.New("InvalidIpAddress")
	}

	return IpAddress(stringValue), nil
}

func (vo IpAddress) String() string {
	return string(vo)
}
