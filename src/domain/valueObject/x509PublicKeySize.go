package tkValueObject

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"math"
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
) (size X509PublicKeySize, err error) {
	switch typedPublicKey := stdlibPublicKey.(type) {
	case *rsa.PublicKey:
		if typedPublicKey == nil || typedPublicKey.N == nil {
			return size, errors.New("InvalidX509PublicKeySize")
		}
		rsaKeyBitLength := typedPublicKey.N.BitLen()
		if rsaKeyBitLength > math.MaxUint16 {
			return size, errors.New("InvalidX509PublicKeySize")
		}
		return NewX509PublicKeySize(uint16(rsaKeyBitLength))
	case *ecdsa.PublicKey:
		if typedPublicKey == nil || typedPublicKey.Curve == nil {
			return size, errors.New("InvalidX509PublicKeySize")
		}
		curveParams := typedPublicKey.Params()
		if curveParams == nil {
			return size, errors.New("InvalidX509PublicKeySize")
		}
		ecdsaCurveBitSize := curveParams.BitSize
		if ecdsaCurveBitSize > math.MaxUint16 {
			return size, errors.New("InvalidX509PublicKeySize")
		}
		return NewX509PublicKeySize(uint16(ecdsaCurveBitSize))
	default:
		return size, errors.New("UnsupportedPublicKeyType")
	}
}

func (vo X509PublicKeySize) Uint16() uint16 {
	return uint16(vo)
}

func (vo X509PublicKeySize) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
