package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type RegexPattern string

func NewRegexPattern(value any) (regexPattern RegexPattern, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return regexPattern, errors.New("RegexPatternMustBeString")
	}

	if stringValue == "" {
		return regexPattern, errors.New("RegexPatternCannotBeEmpty")
	}

	_, err = regexp.Compile(stringValue)
	if err != nil {
		return regexPattern, errors.New("InvalidRegexPattern")
	}

	return RegexPattern(stringValue), nil
}

func (vo RegexPattern) CompiledRegexp() (*regexp.Regexp, error) {
	return regexp.Compile(string(vo))
}

func (vo RegexPattern) String() string {
	return string(vo)
}
