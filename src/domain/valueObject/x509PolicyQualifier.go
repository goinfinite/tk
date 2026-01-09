package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	X509PolicyQualifierCPS        X509PolicyQualifier = "cps"
	X509PolicyQualifierUserNotice X509PolicyQualifier = "userNotice"
)

type X509PolicyQualifier string

func NewX509PolicyQualifier(
	value any,
) (qualifier X509PolicyQualifier, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return qualifier, errors.New("X509PolicyQualifierMustBeString")
	}

	qualifier = X509PolicyQualifier(stringValue)
	switch qualifier {
	case X509PolicyQualifierCPS, X509PolicyQualifierUserNotice:
		return qualifier, nil
	default:
		return qualifier, errors.New("InvalidX509PolicyQualifier")
	}
}

func (vo X509PolicyQualifier) String() string {
	return string(vo)
}
