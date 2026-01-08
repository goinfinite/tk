package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509FingerprintRegex = regexp.MustCompile(
	`^([0-9A-Fa-f]{40}|[0-9A-Fa-f]{64})$`,
)

type X509Fingerprint string

func NewX509Fingerprint(value any) (fp X509Fingerprint, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return fp, errors.New("X509FingerprintMustBeString")
	}

	stringValue = strings.ReplaceAll(stringValue, ":", "")
	stringValue = strings.ReplaceAll(stringValue, " ", "")

	if !x509FingerprintRegex.MatchString(stringValue) {
		return fp, errors.New("InvalidX509Fingerprint")
	}

	return X509Fingerprint(stringValue), nil
}

func (vo X509Fingerprint) String() string {
	return string(vo)
}
