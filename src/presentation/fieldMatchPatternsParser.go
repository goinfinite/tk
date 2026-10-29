package tkPresentation

import (
	"errors"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

func ParseFieldMatchPatterns(input any) (
	fieldNameMatchPatternsMap map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern,
	err error,
) {
	fieldNameMatchPatternsMap = map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{}
	if input == nil {
		return fieldNameMatchPatternsMap, nil
	}

	rawFieldMatchPatternList, err := tkVoUtil.InterfaceToString(input)
	if err != nil {
		return fieldNameMatchPatternsMap, errors.New("FieldMatchPatternsMustBeString")
	}

	for rawFieldMatchPattern := range strings.SplitSeq(rawFieldMatchPatternList, ";") {
		rawFieldMatchPattern = strings.TrimSpace(rawFieldMatchPattern)
		if rawFieldMatchPattern == "" {
			continue
		}

		rawFieldName, rawMatchPattern, hasSeparator := strings.Cut(
			rawFieldMatchPattern, ":",
		)
		if !hasSeparator {
			return fieldNameMatchPatternsMap, errors.New("InvalidFieldMatchPattern")
		}

		fieldName, err := tkValueObject.NewQueryFieldName(rawFieldName)
		if err != nil {
			return fieldNameMatchPatternsMap, err
		}

		matchPattern, err := tkValueObject.NewMatchPattern(rawMatchPattern)
		if err != nil {
			return fieldNameMatchPatternsMap, err
		}
		fieldNameMatchPatternsMap[fieldName] = matchPattern
	}

	return fieldNameMatchPatternsMap, nil
}
