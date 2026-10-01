package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

// @see https://www.regular-expressions.info/unicodecategory.html
// UnixAbsoluteFilePath characters plus the glob wildcards * ? and [].
// Blacklist approach: block control chars and shell-dangerous chars; allow all else.
var unixAbsoluteGlobPathRegex = regexp.MustCompile(
	`^[\/\p{L}\p{N}\p{Pc}\p{Pd}\.][^\x00-\x1f\x7f;|&$` + "`" + `><{}!#@%\\]*$`,
)

type UnixAbsoluteGlobPath string

func NewUnixAbsoluteGlobPath(value any) (
	globPath UnixAbsoluteGlobPath, err error,
) {
	if existentGlobPath, assertOk := value.(UnixAbsoluteGlobPath); assertOk {
		return existentGlobPath, nil
	}

	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return globPath, errors.New("UnixAbsoluteGlobPathValueMustBeString")
	}

	if len(stringValue) == 0 {
		return globPath, errors.New("UnixAbsoluteGlobPathValueMustNotBeEmpty")
	}

	if len(stringValue) > 4096 {
		return globPath, errors.New("UnixAbsoluteGlobPathTooBig")
	}

	if !strings.HasPrefix(stringValue, "/") {
		stringValue = "/" + stringValue
	}

	if !unixAbsoluteGlobPathRegex.MatchString(stringValue) {
		return globPath, errors.New("InvalidUnixAbsoluteGlobPath")
	}

	if unixTypicalRelativeFilePathRegex.MatchString(stringValue) {
		return globPath, errors.New("RelativePathNotAllowed")
	}

	return UnixAbsoluteGlobPath(stringValue), nil
}

func (vo UnixAbsoluteGlobPath) String() string {
	return string(vo)
}

func (vo UnixAbsoluteGlobPath) HasWildcardChars() bool {
	return strings.ContainsAny(vo.String(), "*?[")
}
