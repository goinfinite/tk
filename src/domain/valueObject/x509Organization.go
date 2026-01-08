package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509OrganizationRegex = regexp.MustCompile(
	`^[a-zA-Z0-9 .,\-_()&/]{1,255}$`,
)

type X509Organization string

func NewX509Organization(value any) (org X509Organization, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return org, errors.New("X509OrganizationMustBeString")
	}

	normalizedValue, err := tkVoUtil.StripAccents(stringValue)
	if err != nil {
		return org, errors.New("InvalidX509OrganizationNormalizationFailed")
	}

	if !x509OrganizationRegex.MatchString(normalizedValue) {
		return org, errors.New("InvalidX509Organization")
	}

	return X509Organization(normalizedValue), nil
}

func (vo X509Organization) String() string {
	return string(vo)
}
