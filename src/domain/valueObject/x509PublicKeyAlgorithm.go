package tkValueObject

import (
	"crypto/x509"
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	X509PublicKeyAlgorithmRSA     X509PublicKeyAlgorithm = "RSA"
	X509PublicKeyAlgorithmECDSA   X509PublicKeyAlgorithm = "ECDSA"
	X509PublicKeyAlgorithmEd25519 X509PublicKeyAlgorithm = "Ed25519"
	X509PublicKeyAlgorithmDSA     X509PublicKeyAlgorithm = "DSA"
)

type X509PublicKeyAlgorithm string

func NewX509PublicKeyAlgorithm(
	value any,
) (alg X509PublicKeyAlgorithm, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return alg, errors.New("X509PublicKeyAlgorithmMustBeString")
	}

	stringValue = strings.ToUpper(stringValue)
	if stringValue == "ED25519" {
		stringValue = "Ed25519"
	}

	alg = X509PublicKeyAlgorithm(stringValue)
	switch alg {
	case X509PublicKeyAlgorithmRSA, X509PublicKeyAlgorithmECDSA,
		X509PublicKeyAlgorithmEd25519, X509PublicKeyAlgorithmDSA:
		return alg, nil
	default:
		return alg, errors.New("InvalidX509PublicKeyAlgorithm")
	}
}

func NewX509PublicKeyAlgorithmFromStdlib(
	stdlibAlgorithm x509.PublicKeyAlgorithm,
) (X509PublicKeyAlgorithm, error) {
	algorithmMap := map[x509.PublicKeyAlgorithm]string{
		x509.RSA:     "RSA",
		x509.ECDSA:   "ECDSA",
		x509.Ed25519: "Ed25519",
		x509.DSA:     "DSA",
	}

	algorithmStr, algorithmExists := algorithmMap[stdlibAlgorithm]
	if !algorithmExists {
		return "", errors.New("UnsupportedPublicKeyAlgorithm")
	}

	return NewX509PublicKeyAlgorithm(algorithmStr)
}

func (vo X509PublicKeyAlgorithm) String() string {
	return string(vo)
}
