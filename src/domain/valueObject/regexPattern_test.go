package tkValueObject

import (
	"testing"
)

func TestNewRegexPattern(t *testing.T) {
	testCaseStructs := []struct {
		description    string
		inputValue     any
		expectedOutput RegexPattern
		expectError    bool
		expectedError  string
	}{
		{"ValidSimplePattern", `^alpha=(\d+)$`, RegexPattern(`^alpha=(\d+)$`), false, ""},
		{"ValidAnchoredPattern", `^foo$`, RegexPattern(`^foo$`), false, ""},
		{"ValidGroupPattern", `(foo|bar)`, RegexPattern(`(foo|bar)`), false, ""},
		{"RedosShapedButSafeUnaryNested", `^(a+)+$`, RegexPattern(`^(a+)+$`), false, ""},
		{"RedosShapedButSafeAlternationStar", `(a|a)*`, RegexPattern(`(a|a)*`), false, ""},
		{
			"InvalidUnclosedBracket", `[unclosed`, RegexPattern(""), true,
			"InvalidRegexPattern",
		},
		{
			"InvalidLeadingQuantifier", `*invalid`, RegexPattern(""), true,
			"InvalidRegexPattern",
		},
		{"InvalidUnbalancedParen", `(foo`, RegexPattern(""), true, "InvalidRegexPattern"},
		{"EmptyPattern", "", RegexPattern(""), true, "RegexPatternCannotBeEmpty"},
		{
			"NonStringSlice", []string{"foo"}, RegexPattern(""), true,
			"RegexPatternMustBeString",
		},
		{"NonStringNil", nil, RegexPattern(""), true, "RegexPatternMustBeString"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			actualOutput, conversionErr := NewRegexPattern(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf(
					"UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue,
				)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
			if testCase.expectError && conversionErr != nil &&
				conversionErr.Error() != testCase.expectedError {
				t.Errorf(
					"WrongErrorMessage: '%s' vs '%s' [%v]",
					conversionErr.Error(), testCase.expectedError, testCase.inputValue,
				)
			}
		})
	}
}

func TestRegexPatternCompiledRegexp(t *testing.T) {
	testCaseStructs := []struct {
		description    string
		inputPattern   string
		shouldMatch    []string
		shouldNotMatch []string
	}{
		{
			"SimpleAnchoredPattern",
			`^foo$`,
			[]string{"foo"},
			[]string{"foobar", "barfoo"},
		},
		{
			"DigitCaptureGroup",
			`^(\d+)$`,
			[]string{"123", "0"},
			[]string{"abc", "12a"},
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			pattern, constructionErr := NewRegexPattern(testCase.inputPattern)
			if constructionErr != nil {
				t.Fatalf("NewRegexPatternFailed: %v", constructionErr)
			}

			compiledRegexp, compileErr := pattern.CompiledRegexp()
			if compileErr != nil {
				t.Fatalf("CompiledRegexpFailed: %v", compileErr)
			}

			if compiledRegexp == nil {
				t.Errorf("CompiledRegexpShouldNotBeNil")
				return
			}

			for _, matchCandidate := range testCase.shouldMatch {
				if !compiledRegexp.MatchString(matchCandidate) {
					t.Errorf(
						"CompiledRegexpShouldMatch: '%s' against '%s'",
						matchCandidate, testCase.inputPattern,
					)
				}
			}

			for _, nonMatchCandidate := range testCase.shouldNotMatch {
				if compiledRegexp.MatchString(nonMatchCandidate) {
					t.Errorf(
						"CompiledRegexpShouldNotMatch: '%s' against '%s'",
						nonMatchCandidate, testCase.inputPattern,
					)
				}
			}
		})
	}
}

func TestRegexPatternString(t *testing.T) {
	testCaseStructs := []struct {
		description    string
		inputPattern   RegexPattern
		expectedOutput string
	}{
		{"SimplePattern", RegexPattern(`^test$`), `^test$`},
		{"EmptyPattern", RegexPattern(""), ""},
		{"ComplexPattern", RegexPattern(`^[a-z]+(\d{2})$`), `^[a-z]+(\d{2})$`},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			actualOutput := testCase.inputPattern.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		})
	}
}
