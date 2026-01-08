package tkValueObject

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type X509PublicKeySize uint16

func NewX509PublicKeySize(value any) (size X509PublicKeySize, err error) {
	uintValue, err := tkVoUtil.InterfaceToUint16(value)
	if err != nil {
		return size, errors.New("X509PublicKeySizeMustBeUint16")
	}

	switch uintValue {
	case 256, 384, 521, 1024, 2048, 3072, 4096, 8192:
		return X509PublicKeySize(uintValue), nil
	default:
		return size, errors.New("InvalidX509PublicKeySize")
	}
}

func NewX509PublicKeySizeFromStdlib(
	stdlibPublicKey any,
) (X509PublicKeySize, error) {
	switch typedPublicKey := stdlibPublicKey.(type) {
	case *rsa.PublicKey:
		rsaKeyBitLength := uint16(typedPublicKey.N.BitLen())
		return NewX509PublicKeySize(rsaKeyBitLength)
	case *ecdsa.PublicKey:
		ecdsaCurveBitSize := uint16(typedPublicKey.Curve.Params().BitSize)
		return NewX509PublicKeySize(ecdsaCurveBitSize)
	default:
		defaultKeySize := uint16(2048)
		return NewX509PublicKeySize(defaultKeySize)
	}
}

func (vo X509PublicKeySize) Uint16() uint16 {
	return uint16(vo)
}

func (vo X509PublicKeySize) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
