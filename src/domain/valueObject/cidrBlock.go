package tkValueObject

import (
	"errors"
	"net"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type CidrBlock string

func NewCidrBlock(value any) (cidrBlock CidrBlock, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return cidrBlock, errors.New("CidrBlockValueMustBeString")
	}

	if stringValue == "" {
		return cidrBlock, errors.New("CidrBlockValueCannotBeEmpty")
	}

	if _, _, err := net.ParseCIDR(stringValue); err != nil {
		return cidrBlock, errors.New("InvalidCidrBlock")
	}

	return CidrBlock(stringValue), nil
}

func (vo CidrBlock) String() string {
	return string(vo)
}

func (vo CidrBlock) IsIpv4() bool {
	_, cidrBlock, err := net.ParseCIDR(vo.String())
	if err != nil {
		return false
	}
	return cidrBlock.IP.To4() != nil
}

func (vo CidrBlock) IsIpv6() bool {
	return !vo.IsIpv4()
}

func (vo CidrBlock) IsPrivate() bool {
	_, cidrBlock, err := net.ParseCIDR(vo.String())
	if err != nil {
		return false
	}
	return cidrBlock.IP.IsPrivate()
}

func (vo CidrBlock) IsPublic() bool {
	return !vo.IsPrivate()
}
