package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509OrganizationalUnitRegex = x509OrgFieldRegex

type X509OrganizationalUnit string

func NewX509OrganizationalUnit(
	value any,
) (unit X509OrganizationalUnit, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unit, errors.New("X509OrganizationalUnitMustBeString")
	}

	normalizedValue, err := tkVoUtil.StripAccents(stringValue)
	if err != nil {
		return unit, errors.New("InvalidX509OrganizationalUnitNormalizationFailed")
	}

	if !x509OrganizationalUnitRegex.MatchString(normalizedValue) {
		return unit, errors.New("InvalidX509OrganizationalUnit")
	}

	return X509OrganizationalUnit(normalizedValue), nil
}

func (vo X509OrganizationalUnit) String() string {
	return string(vo)
}
