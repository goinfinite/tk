package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509LocalityRegex = regexp.MustCompile(`^[a-zA-Z \-']{1,128}$`)

type X509Locality string

func NewX509Locality(value any) (locality X509Locality, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return locality, errors.New("X509LocalityMustBeString")
	}

	normalizedValue, err := tkVoUtil.StripAccents(stringValue)
	if err != nil {
		return locality, errors.New("InvalidX509LocalityNormalizationFailed")
	}

	if !x509LocalityRegex.MatchString(normalizedValue) {
		return locality, errors.New("InvalidX509Locality")
	}

	return X509Locality(normalizedValue), nil
}

func (vo X509Locality) String() string {
	return string(vo)
}
