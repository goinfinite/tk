package tkValueObject

import (
	"testing"
)

func TestNewMatchPattern(test *testing.T) {
	test.Run("ValidMatchPattern", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput MatchPattern
		}{
			{nil, MatchPatternEquals},
			{"", MatchPatternEquals},
			{"   ", MatchPatternEquals},
			{"equals", MatchPatternEquals},
			{MatchPatternEquals, MatchPatternEquals},
			{MatchPatternContains, MatchPatternContains},
			{"contains", MatchPatternContains},
			{"begins-with", MatchPatternBeginsWith},
			{"ends-with", MatchPatternEndsWith},
			{"CONTAINS", MatchPatternContains},
			{"Begins With", MatchPatternBeginsWith},
			{" ends with ", MatchPatternEndsWith},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewMatchPattern(testCase.inputValue)
			if err != nil {
				subtest.Errorf(
					"UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue,
				)
				continue
			}
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("InvalidMatchPattern", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectedErr string
		}{
			{"starts-with", "InvalidMatchPattern"},
			{"equal", "InvalidMatchPattern"},
			{"begins  with", "InvalidMatchPattern"},
			{123, "InvalidMatchPattern"},
			{true, "InvalidMatchPattern"},
			{[]string{"contains"}, "MatchPatternMustBeString"},
			{map[string]any{}, "MatchPatternMustBeString"},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewMatchPattern(testCase.inputValue)
			if err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
				continue
			}
			if err.Error() != testCase.expectedErr {
				subtest.Errorf(
					"UnexpectedError: '%s' vs '%s' [%v]",
					err.Error(), testCase.expectedErr, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     MatchPattern
			expectedOutput string
		}{
			{MatchPatternEquals, "equals"},
			{MatchPatternContains, "contains"},
			{MatchPatternBeginsWith, "begins-with"},
			{MatchPatternEndsWith, "ends-with"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})
}
