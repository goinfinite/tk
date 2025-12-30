package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	DnsRecordTypeA       DnsRecordType = "A"
	DnsRecordTypeAAAA    DnsRecordType = "AAAA"
	DnsRecordTypeMX      DnsRecordType = "MX"
	DnsRecordTypeTXT     DnsRecordType = "TXT"
	DnsRecordTypeNS      DnsRecordType = "NS"
	DnsRecordTypeCNAME   DnsRecordType = "CNAME"
	DnsRecordTypePTR     DnsRecordType = "PTR"
	DnsRecordTypeDefault               = DnsRecordTypeA
)

type DnsRecordType string

func NewDnsRecordType(value any) (dnsRecordType DnsRecordType, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return dnsRecordType, errors.New("DnsRecordTypeMustBeString")
	}
	stringValue = strings.ToUpper(stringValue)

	dnsRecordType = DnsRecordType(stringValue)
	switch dnsRecordType {
	case DnsRecordTypeA, DnsRecordTypeAAAA, DnsRecordTypeMX,
		DnsRecordTypeTXT, DnsRecordTypeNS, DnsRecordTypeCNAME, DnsRecordTypePTR:
		return dnsRecordType, nil
	default:
		return dnsRecordType, errors.New("InvalidDnsRecordType")
	}
}

func (vo DnsRecordType) String() string {
	return string(vo)
}
