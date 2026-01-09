package tkValueObject

import (
	"errors"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var x509SubjectNameRegex = regexp.MustCompile(
	`^[a-zA-Z0-9 .\-_]+$`,
)

var x509WildcardSubjectNameRegex = regexp.MustCompile(
	`^\*\.[a-zA-Z0-9.\-_]+$`,
)

type X509SubjectName string

func NewX509SubjectName(value any) (name X509SubjectName, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return name, errors.New("X509SubjectNameMustBeString")
	}

	if len(stringValue) < 1 || len(stringValue) > 253 {
		return name, errors.New("InvalidX509SubjectNameLength")
	}

	hasWildcard := strings.Contains(stringValue, "*")
	if hasWildcard {
		wildcardCount := strings.Count(stringValue, "*")
		if wildcardCount > 1 {
			return name, errors.New("InvalidX509SubjectNameMultipleWildcards")
		}

		if !x509WildcardSubjectNameRegex.MatchString(stringValue) {
			return name, errors.New("InvalidX509SubjectNameWildcardFormat")
		}

		return X509SubjectName(stringValue), nil
	}

	if !x509SubjectNameRegex.MatchString(stringValue) {
		return name, errors.New("InvalidX509SubjectName")
	}

	return X509SubjectName(stringValue), nil
}

func (vo X509SubjectName) String() string {
	return string(vo)
}
