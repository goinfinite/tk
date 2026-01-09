package tkValueObject

import (
	"crypto/x509"
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	X509SignatureAlgorithmSHA256WithRSA   X509SignatureAlgorithm = "SHA256WithRSA"
	X509SignatureAlgorithmSHA384WithRSA   X509SignatureAlgorithm = "SHA384WithRSA"
	X509SignatureAlgorithmSHA512WithRSA   X509SignatureAlgorithm = "SHA512WithRSA"
	X509SignatureAlgorithmECDSAWithSHA256 X509SignatureAlgorithm = "ECDSAWithSHA256"
	X509SignatureAlgorithmECDSAWithSHA384 X509SignatureAlgorithm = "ECDSAWithSHA384"
	X509SignatureAlgorithmECDSAWithSHA512 X509SignatureAlgorithm = "ECDSAWithSHA512"
	X509SignatureAlgorithmEd25519         X509SignatureAlgorithm = "Ed25519"
)

type X509SignatureAlgorithm string

func NewX509SignatureAlgorithm(
	value any,
) (alg X509SignatureAlgorithm, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return alg, errors.New("X509SignatureAlgorithmMustBeString")
	}

	alg = X509SignatureAlgorithm(stringValue)
	switch alg {
	case X509SignatureAlgorithmSHA256WithRSA,
		X509SignatureAlgorithmSHA384WithRSA,
		X509SignatureAlgorithmSHA512WithRSA,
		X509SignatureAlgorithmECDSAWithSHA256,
		X509SignatureAlgorithmECDSAWithSHA384,
		X509SignatureAlgorithmECDSAWithSHA512,
		X509SignatureAlgorithmEd25519:
		return alg, nil
	default:
		return alg, errors.New("InvalidX509SignatureAlgorithm")
	}
}

func NewX509SignatureAlgorithmFromStdlib(
	stdlibAlgorithm x509.SignatureAlgorithm,
) (X509SignatureAlgorithm, error) {
	algorithmMap := map[x509.SignatureAlgorithm]string{
		x509.SHA256WithRSA:   "SHA256WithRSA",
		x509.SHA384WithRSA:   "SHA384WithRSA",
		x509.SHA512WithRSA:   "SHA512WithRSA",
		x509.ECDSAWithSHA256: "ECDSAWithSHA256",
		x509.ECDSAWithSHA384: "ECDSAWithSHA384",
		x509.ECDSAWithSHA512: "ECDSAWithSHA512",
		x509.PureEd25519:     "Ed25519",
	}

	algorithmStr, algorithmExists := algorithmMap[stdlibAlgorithm]
	if !algorithmExists {
		return "", errors.New("UnsupportedSignatureAlgorithm")
	}

	return NewX509SignatureAlgorithm(algorithmStr)
}

func (vo X509SignatureAlgorithm) String() string {
	return string(vo)
}
