package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const (
	MatchPatternEquals     MatchPattern = "equals"
	MatchPatternContains   MatchPattern = "contains"
	MatchPatternBeginsWith MatchPattern = "begins-with"
	MatchPatternEndsWith   MatchPattern = "ends-with"
)

type MatchPattern string

func NewMatchPattern(value any) (matchPattern MatchPattern, err error) {
	existentMatchPattern, isMatchPattern := value.(MatchPattern)
	if isMatchPattern {
		return existentMatchPattern, nil
	}

	if value == nil {
		return MatchPatternEquals, nil
	}

	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return matchPattern, errors.New("MatchPatternMustBeString")
	}
	stringValue = strings.ToLower(stringValue)
	stringValue = strings.ReplaceAll(stringValue, " ", "-")

	if stringValue == "" {
		return MatchPatternEquals, nil
	}

	stringValueVo := MatchPattern(stringValue)
	switch stringValueVo {
	case MatchPatternEquals, MatchPatternContains,
		MatchPatternBeginsWith, MatchPatternEndsWith:
		return stringValueVo, nil
	default:
		return matchPattern, errors.New("InvalidMatchPattern")
	}
}

func (vo MatchPattern) String() string {
	return string(vo)
}
