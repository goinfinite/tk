package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509StateOrProvinceRegex = regexp.MustCompile(
	`^[a-zA-Z \-']{1,128}$`,
)

type X509StateOrProvince string

func NewX509StateOrProvince(
	value any,
) (state X509StateOrProvince, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return state, errors.New("X509StateOrProvinceMustBeString")
	}

	normalizedValue, err := tkVoUtil.StripAccents(stringValue)
	if err != nil {
		return state, errors.New("InvalidX509StateOrProvinceNormalizationFailed")
	}

	if !x509StateOrProvinceRegex.MatchString(normalizedValue) {
		return state, errors.New("InvalidX509StateOrProvince")
	}

	return X509StateOrProvince(normalizedValue), nil
}

func (vo X509StateOrProvince) String() string {
	return string(vo)
}
