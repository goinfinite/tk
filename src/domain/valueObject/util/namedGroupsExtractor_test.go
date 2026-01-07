package tkVoUtil

import (
	"regexp"
	"testing"
)

func TestNamedGroupsExtractor(t *testing.T) {
	t.Run("AllScenarios", func(t *testing.T) {
		testCases := []struct {
			caseName       string
			regexPattern   string
			inputString    string
			expectedGroups map[string]string
		}{
			{
				caseName:       "Matching",
				regexPattern:   `^(?P<name>[A-Za-z]+) (?P<age>\d+)$`,
				inputString:    "John 25",
				expectedGroups: map[string]string{"name": "John", "age": "25"},
			},
			{
				caseName:       "PartialMatch",
				regexPattern:   `^(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})$`,
				inputString:    "2023-10-15",
				expectedGroups: map[string]string{"year": "2023", "month": "10", "day": "15"},
			},
			{
				caseName:       "PartialMatch",
				regexPattern:   `(?P<name>[A-Za-z]+) ?(?P<age>\d+)?`,
				inputString:    "John",
				expectedGroups: map[string]string{"name": "John"},
			},
			{
				caseName:       "NoMatch",
				regexPattern:   `^(?P<name>[A-Za-z]+) ?(?P<age>\d+)?`,
				inputString:    "@John",
				expectedGroups: map[string]string{},
			},
			{
				caseName:       "NoNamedGroups",
				regexPattern:   `[A-Za-z]+ \d+`,
				inputString:    "John 25",
				expectedGroups: map[string]string{},
			},
			{
				caseName:       "MixedNamedAndUnnamedGroups",
				regexPattern:   `^([A-Za-z]+) ?(?P<age>\d+)?`,
				inputString:    "John 25",
				expectedGroups: map[string]string{"age": "25"},
			},
		}

		for _, testCase := range testCases {
			compiledRegex := regexp.MustCompile(testCase.regexPattern)
			namedGroupsValuesMap := NamedGroupsExtractor(compiledRegex, testCase.inputString)
			if len(namedGroupsValuesMap) != len(testCase.expectedGroups) {
				t.Errorf(
					"[%s] UnexpectedOutputLength: '%d' vs '%d' [pattern: %s, input: %s, namedGroupsFound: %v]",
					testCase.caseName, len(namedGroupsValuesMap), len(testCase.expectedGroups),
					testCase.regexPattern, testCase.inputString, namedGroupsValuesMap,
				)
			}
			for groupName, expectedValue := range testCase.expectedGroups {
				actualValue, groupExists := namedGroupsValuesMap[groupName]
				if !groupExists {
					t.Errorf(
						"[%s] MissingExpectedKey: '%s' [pattern: %s, input: %s]",
						testCase.caseName, groupName, testCase.regexPattern, testCase.inputString,
					)
				}
				if groupExists && actualValue != expectedValue {
					t.Errorf(
						"[%s] UnexpectedOutputValue: '%s' vs '%s' for key '%s' [pattern: %s, input: %s]",
						testCase.caseName, actualValue, expectedValue, groupName, testCase.regexPattern,
						testCase.inputString,
					)
				}
			}
		}
	})
}
