package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	PrivateKeyAlgorithmRSA     PrivateKeyAlgorithm = "RSA"
	PrivateKeyAlgorithmECDSA   PrivateKeyAlgorithm = "ECDSA"
	PrivateKeyAlgorithmDSA     PrivateKeyAlgorithm = "DSA"
	PrivateKeyAlgorithmEd25519 PrivateKeyAlgorithm = "Ed25519"
)

type PrivateKeyAlgorithm string

var ValidPrivateKeyAlgorithms = []string{
	PrivateKeyAlgorithmRSA.String(),
	PrivateKeyAlgorithmECDSA.String(),
	PrivateKeyAlgorithmDSA.String(),
	PrivateKeyAlgorithmEd25519.String(),
}

func NewPrivateKeyAlgorithm(value any) (PrivateKeyAlgorithm, error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return "", errors.New("PrivateKeyAlgorithmMustBeString")
	}

	stringValue = strings.ToUpper(stringValue)
	if stringValue == "ED25519" {
		stringValue = "Ed25519"
	}

	if stringValue == "" {
		return "", errors.New("PrivateKeyAlgorithmCannotBeEmpty")
	}

	valueVo := PrivateKeyAlgorithm(stringValue)
	switch valueVo {
	case PrivateKeyAlgorithmRSA, PrivateKeyAlgorithmECDSA,
		PrivateKeyAlgorithmDSA, PrivateKeyAlgorithmEd25519:
		return valueVo, nil
	default:
		return "", errors.New("InvalidPrivateKeyAlgorithm")
	}
}

func (vo PrivateKeyAlgorithm) String() string {
	return string(vo)
}
