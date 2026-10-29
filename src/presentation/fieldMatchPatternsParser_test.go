package tkPresentation

import (
	"maps"
	"strings"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestParseFieldMatchPatterns(test *testing.T) {
	fieldNameFactory := func(value string) tkValueObject.QueryFieldName {
		return tkValueObject.QueryFieldName(value)
	}

	testCases := []struct {
		name           string
		input          any
		expectedOutput map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern
		expectedErr    string
	}{
		{
			name:           "NilInput",
			input:          nil,
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
		},
		{
			name:           "UnconvertibleInput",
			input:          map[string]any{"name": "contains"},
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "FieldMatchPatternsMustBeString",
		},
		{
			name:           "NumericInputBecomesInvalidFieldMatchPattern",
			input:          123,
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "InvalidFieldMatchPattern",
		},
		{
			name:           "EmptyInput",
			input:          "",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
		},
		{
			name:  "SingleField",
			input: "name:contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("name"): tkValueObject.MatchPatternContains,
			},
		},
		{
			name:  "MultipleFields",
			input: "name:contains;status:equals",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("name"):   tkValueObject.MatchPatternContains,
				fieldNameFactory("status"): tkValueObject.MatchPatternEquals,
			},
		},
		{
			name:  "TrimsWhitespaceAndSpacedPatterns",
			input: " name : begins with ; status : ends-with ",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("name"):   tkValueObject.MatchPatternBeginsWith,
				fieldNameFactory("status"): tkValueObject.MatchPatternEndsWith,
			},
		},
		{
			name:  "EmptyPatternDefaultsToEquals",
			input: "name:",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("name"): tkValueObject.MatchPatternEquals,
			},
		},
		{
			name:  "SkipsEmptySegments",
			input: ";; name:contains ;;",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("name"): tkValueObject.MatchPatternContains,
			},
		},
		{
			name:  "DuplicateFieldLastValueWins",
			input: "name:contains;name:equals",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("name"): tkValueObject.MatchPatternEquals,
			},
		},
		{
			name:  "PreservesQueryFieldNameCase",
			input: "Name:contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("Name"): tkValueObject.MatchPatternContains,
			},
		},
		{
			name:  "DottedQueryFieldName",
			input: "account.name:contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{
				fieldNameFactory("account.name"): tkValueObject.MatchPatternContains,
			},
		},
		{
			name:           "MissingSeparator",
			input:          "name",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "InvalidFieldMatchPattern",
		},
		{
			name:           "EmptyQueryFieldName",
			input:          ":contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "QueryFieldNameCannotBeEmpty",
		},
		{
			name:           "InvalidQueryFieldNameWithSpace",
			input:          "na me:contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "InvalidQueryFieldName",
		},
		{
			name:           "InvalidQueryFieldNameWithWildcard",
			input:          "na*me:contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "InvalidQueryFieldName",
		},
		{
			name:           "QueryFieldNameTooBig",
			input:          strings.Repeat("a", 257) + ":contains",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "QueryFieldNameTooBig",
		},
		{
			name:           "InvalidMatchPattern",
			input:          "name:starts-with",
			expectedOutput: map[tkValueObject.QueryFieldName]tkValueObject.MatchPattern{},
			expectedErr:    "InvalidMatchPattern",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			actualOutput, err := ParseFieldMatchPatterns(testCase.input)

			if testCase.expectedErr != "" {
				if err == nil {
					subtest.Fatalf("MissingExpectedError: %s", testCase.expectedErr)
				}
				if err.Error() != testCase.expectedErr {
					subtest.Fatalf(
						"UnexpectedError: '%s' vs '%s'", err.Error(), testCase.expectedErr,
					)
				}
				return
			}

			if err != nil {
				subtest.Fatalf("UnexpectedError: %v", err)
			}

			if !maps.Equal(actualOutput, testCase.expectedOutput) {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		})
	}
}
