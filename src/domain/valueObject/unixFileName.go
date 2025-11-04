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
	// \p{Zs}	a whitespace character that is invisible, but does take up space.
	// Unsafe allows for additional characters:
	// \p{S}	math symbols, currency signs, dingbats, box-drawing characters, etc.
	// \p{P}	any kind of punctuation character.
	unixFileNameStrictRegex    = regexp.MustCompile(`^[\p{L}\p{N}\p{Pc}\p{Pd}\.]?[\p{L}\p{N}\p{Pc}\p{Pd}\p{Zs}\(\)\[\]\+\.]*[\p{L}\p{N}\p{Pc}\p{Pd}]$`)
	unixFileNameUnsafeRegex    = regexp.MustCompile(`^[\p{L}\p{N}\p{Pc}\p{Pd}\.]?[\p{L}\p{N}\p{Pc}\p{Pd}\p{Zs}\p{S}\p{P}\(\)\[\]\+\.]*[\p{L}\p{N}\p{Pc}\p{Pd}]$`)
	forbiddenUnixFileNameRegex = regexp.MustCompile(`^(\.|\.\.|\~|\^|\*|\/|\\)$|[\|\/\\]`)
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
