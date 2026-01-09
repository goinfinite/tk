package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var envelopedPrivateKeyRegex = regexp.MustCompile(
	`(?:` +
		`^-----BEGIN PRIVATE KEY-----[\s\S]+-----END PRIVATE KEY-----$` + `|` +
		`^-----BEGIN RSA PRIVATE KEY-----[\s\S]+-----END RSA PRIVATE KEY-----$` + `|` +
		`^-----BEGIN EC PRIVATE KEY-----[\s\S]+-----END EC PRIVATE KEY-----$` + `|` +
		`^-----BEGIN DSA PRIVATE KEY-----[\s\S]+-----END DSA PRIVATE KEY-----$` + `|` +
		`^-----BEGIN ENCRYPTED PRIVATE KEY-----[\s\S]+-----END ENCRYPTED PRIVATE KEY-----$` +
		`)`,
)

type EnvelopedPrivateKey string

func NewEnvelopedPrivateKey(
	value any,
) (envelopedKey EnvelopedPrivateKey, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return envelopedKey, errors.New("EnvelopedPrivateKeyMustBeString")
	}

	if len(stringValue) < 100 {
		return envelopedKey, errors.New("InvalidEnvelopedPrivateKeyTooShort")
	}

	beginTagCount := strings.Count(stringValue, "-----BEGIN")
	if beginTagCount == 0 {
		return envelopedKey, errors.New("InvalidEnvelopedPrivateKeyMissingBeginTag")
	}
	if beginTagCount > 1 {
		return envelopedKey, errors.New("InvalidEnvelopedPrivateKeyMultipleBeginTags")
	}

	endTagCount := strings.Count(stringValue, "-----END")
	if endTagCount == 0 {
		return envelopedKey, errors.New("InvalidEnvelopedPrivateKeyMissingEndTag")
	}
	if endTagCount > 1 {
		return envelopedKey, errors.New("InvalidEnvelopedPrivateKeyMultipleEndTags")
	}

	if !envelopedPrivateKeyRegex.MatchString(stringValue) {
		return envelopedKey, errors.New("InvalidEnvelopedPrivateKeyFormat")
	}

	return EnvelopedPrivateKey(stringValue), nil
}

func (vo EnvelopedPrivateKey) String() string {
	return string(vo)
}

func (vo EnvelopedPrivateKey) Bytes() []byte {
	return []byte(vo)
}
