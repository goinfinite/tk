package tkVoUtil

import (
	"regexp"
	"testing"
)

func TestNamedGroupsExtractor(t *testing.T) {
	t.Run("ValidRegexWithNamedGroupsMatching", func(t *testing.T) {
		testCases := []struct {
			regexPattern   string
			inputString    string
			expectedGroups map[string]string
		}{
			{
				regexPattern:   `^(?P<name>[A-Za-z]+) (?P<age>\d+)$`,
				inputString:    "John 25",
				expectedGroups: map[string]string{"name": "John", "age": "25"},
			},
			{
				regexPattern:   `^(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})$`,
				inputString:    "2023-10-15",
				expectedGroups: map[string]string{"year": "2023", "month": "10", "day": "15"},
			},
		}

		for _, testCase := range testCases {
			compiledRegex := regexp.MustCompile(testCase.regexPattern)
			namedGroupsValuesMap := NamedGroupsExtractor(compiledRegex, testCase.inputString)
			if len(namedGroupsValuesMap) != len(testCase.expectedGroups) {
				t.Errorf(
					"UnexpectedOutputLength: '%d' vs '%d' [pattern: %s, input: %s, namedGroupsFound: %v]",
					len(namedGroupsValuesMap), len(testCase.expectedGroups),
					testCase.regexPattern, testCase.inputString, namedGroupsValuesMap,
				)
			}
			for groupName, expectedValue := range testCase.expectedGroups {
				actualValue, groupExists := namedGroupsValuesMap[groupName]
				if !groupExists {
					t.Errorf(
						"MissingExpectedKey: '%s' [pattern: %s, input: %s]",
						groupName, testCase.regexPattern, testCase.inputString,
					)
				}
				if groupExists && actualValue != expectedValue {
					t.Errorf(
						"UnexpectedOutputValue: '%s' vs '%s' for key '%s' [pattern: %s, input: %s]",
						actualValue, expectedValue, groupName, testCase.regexPattern,
						testCase.inputString,
					)
				}
			}
		}
	})

	t.Run("ValidRegexWithNamedGroupsPartialMatch", func(t *testing.T) {
		regexPattern := `(?P<name>[A-Za-z]+) ?(?P<age>\d+)?`
		inputValue := "John"
		expectedGroups := map[string]string{"name": "John"}

		compiledRegex := regexp.MustCompile(regexPattern)
		namedGroupsValuesMap := NamedGroupsExtractor(compiledRegex, inputValue)
		if len(namedGroupsValuesMap) != len(expectedGroups) {
			t.Errorf(
				"UnexpectedOutputLength: '%d' vs '%d' [pattern: %s, input: %s, namedGroupsFound: %v]",
				len(namedGroupsValuesMap), len(expectedGroups), regexPattern, inputValue, namedGroupsValuesMap,
			)
		}
		for groupName, expectedValue := range expectedGroups {
			actualValue, groupExists := namedGroupsValuesMap[groupName]
			if !groupExists {
				t.Errorf(
					"MissingExpectedKey: '%s' [pattern: %s, input: %s]",
					groupName, regexPattern, inputValue,
				)
			}
			if groupExists && actualValue != expectedValue {
				t.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' for key '%s' [pattern: %s, input: %s]",
					actualValue, expectedValue, groupName, regexPattern, inputValue,
				)
			}
		}
	})

	t.Run("ValidRegexWithNamedGroupsNoMatch", func(t *testing.T) {
		regexPattern := `^(?P<name>[A-Za-z]+) ?(?P<age>\d+)?`
		inputValue := "@John"
		expectedGroups := map[string]string{}

		compiledRegex := regexp.MustCompile(regexPattern)
		namedGroupsValuesMap := NamedGroupsExtractor(compiledRegex, inputValue)
		if len(namedGroupsValuesMap) != len(expectedGroups) {
			t.Errorf(
				"UnexpectedOutputLength: '%d' vs '%d' [pattern: %s, input: %s, namedGroupsFound: %v]",
				len(namedGroupsValuesMap), len(expectedGroups), regexPattern, inputValue, namedGroupsValuesMap,
			)
		}
	})

	t.Run("RegexWithNoNamedGroups", func(t *testing.T) {
		regexPattern := `[A-Za-z]+ \d+`
		inputValue := "John 25"
		expectedGroups := map[string]string{}

		compiledRegex := regexp.MustCompile(regexPattern)
		namedGroupsValuesMap := NamedGroupsExtractor(compiledRegex, inputValue)
		if len(namedGroupsValuesMap) != len(expectedGroups) {
			t.Errorf(
				"UnexpectedOutputLength: '%d' vs '%d' [pattern: %s, input: %s, namedGroupsFound: %v]",
				len(namedGroupsValuesMap), len(expectedGroups), regexPattern, inputValue, namedGroupsValuesMap,
			)
		}
	})

	t.Run("RegexWithMixedNamedAndUnnamedGroups", func(t *testing.T) {
		regexPattern := `^([A-Za-z]+) ?(?P<age>\d+)?`
		inputValue := "John 25"
		expectedGroups := map[string]string{"age": "25"}

		compiledRegex := regexp.MustCompile(regexPattern)
		namedGroupsValuesMap := NamedGroupsExtractor(compiledRegex, inputValue)
		if len(namedGroupsValuesMap) != len(expectedGroups) {
			t.Errorf(
				"UnexpectedOutputLength: '%d' vs '%d' [pattern: %s, input: %s, namedGroupsFound: %v]",
				len(namedGroupsValuesMap), len(expectedGroups), regexPattern, inputValue, namedGroupsValuesMap,
			)
		}
		for groupName, expectedValue := range expectedGroups {
			actualValue, groupExists := namedGroupsValuesMap[groupName]
			if !groupExists {
				t.Errorf(
					"MissingExpectedKey: '%s' [pattern: %s, input: %s]",
					groupName, regexPattern, inputValue,
				)
			}
			if groupExists && actualValue != expectedValue {
				t.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' for key '%s' [pattern: %s, input: %s]",
					actualValue, expectedValue, groupName, regexPattern, inputValue,
				)
			}
		}
	})
}
