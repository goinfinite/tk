package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	// @see https://www.regular-expressions.info/unicodecategory.html
	// \p{L}	any kind of letter from any language.
	// \p{N}	any kind of numeric character in any script.
	// \p{Pc}	a punctuation character such as an underscore that connects words.
	// \p{Pd}	any kind of hyphen or dash.
	// Both strict and unsafe use a blacklist approach for character validation.
	// The first char is restricted to letters, numbers, connectors, hyphens, dots, asterisks, or tildes.
	// Strict blocks control chars and shell-dangerous chars (;|&$`><{}!#?@%\/); allow all else.
	// Unsafe only blocks control chars (\x00-\x1f, \x7f) and directory separators (/, \);
	// shell-dangerous chars are allowed — the caller is responsible for quoting.
	unixFileNameStrictRegex    = regexp.MustCompile(`^[\p{L}\p{N}\p{Pc}\p{Pd}\.\*\~][^\x00-\x1f\x7f;|&$` + "`" + `><{}!#?@%\\/]*$`)
	unixFileNameUnsafeRegex    = regexp.MustCompile(`^[\p{L}\p{N}\p{Pc}\p{Pd}\.\*\~][^\x00-\x1f\x7f\\/]*$`)
	forbiddenUnixFileNameRegex = regexp.MustCompile(`^(\.|\.\.|\~|\^|\*|\/|\\)$|[\|\/\\]|\*{2,}`)
)

type UnixFileName string

func NewUnixFileName(value any, allowUnsafeChars bool) (fileName UnixFileName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return fileName, errors.New("UnixFileNameValueMustBeString")
	}

	if len(stringValue) == 0 {
		return fileName, errors.New("UnixFileNameValueMustNotBeEmpty")
	}

	if len(stringValue) > 255 {
		return fileName, errors.New("UnixFileNameTooBig")
	}

	switch allowUnsafeChars {
	case true:
		if !unixFileNameUnsafeRegex.MatchString(stringValue) {
			return fileName, errors.New("InvalidUnixFileName")
		}
	case false:
		if !unixFileNameStrictRegex.MatchString(stringValue) {
			return fileName, errors.New("InvalidUnixFileName")
		}
	}

	if forbiddenUnixFileNameRegex.MatchString(stringValue) {
		return fileName, errors.New("ForbiddenUnixFileName")
	}

	return UnixFileName(stringValue), nil
}

func (vo UnixFileName) String() string {
	return string(vo)
}
