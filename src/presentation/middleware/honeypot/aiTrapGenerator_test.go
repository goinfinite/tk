package tkPresentationMiddlewareHoneypot

import (
	"strings"
	"testing"
)

func TestNewAiTrapGenerator(t *testing.T) {
	generator := NewAiTrapGenerator()
	if generator == nil {
		t.Errorf("ExpectedNonNilGenerator")
	}
}

func TestAiTrapGeneratorGenerate(t *testing.T) {
	testCaseStructs := []struct {
		name        string
		requestSize int
		shouldEmpty bool
	}{
		{
			name:        "ValidSize1000",
			requestSize: 1000,
			shouldEmpty: false,
		},
		{
			name:        "ZeroSize",
			requestSize: 0,
			shouldEmpty: true,
		},
		{
			name:        "NegativeSize",
			requestSize: -100,
			shouldEmpty: true,
		},
		{
			name:        "SmallSize10",
			requestSize: 10,
			shouldEmpty: false,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			generator := NewAiTrapGenerator()
			result := generator.Generate(testCase.requestSize)

			if testCase.shouldEmpty && result != "" {
				t.Errorf(
					"ExpectedEmptyString: GotLength=%d",
					len(result),
				)
			}

			if !testCase.shouldEmpty && len(result) == 0 {
				t.Errorf("ExpectedNonEmptyString")
			}

			if !testCase.shouldEmpty && testCase.requestSize > 0 {
				if len(result) < testCase.requestSize {
					t.Errorf(
						"ResultTooShort: Expected>=%d, Actual=%d",
						testCase.requestSize, len(result),
					)
				}
			}
		})
	}
}

func TestAiTrapGeneratorContainsNoSecrets(t *testing.T) {
	generator := NewAiTrapGenerator()
	result := generator.Generate(5000)

	secretPatterns := []string{
		"password", "secret", "api_key", "token",
		"PRIVATE.KEY", "BEGIN RSA",
	}

	for _, pattern := range secretPatterns {
		if strings.Contains(strings.ToLower(result), strings.ToLower(pattern)) {
			t.Errorf(
				"ResultContainsSecretPattern: Pattern='%s'",
				pattern,
			)
		}
	}
}

func TestAiTrapGeneratorProducesVariedOutput(t *testing.T) {
	generator := NewAiTrapGenerator()
	first := generator.Generate(500)
	second := generator.Generate(500)

	if first == second {
		t.Errorf("ExpectedVariedOutput: TwoGenerationsIdentical")
	}
}

func TestResolveHallucinationPattern(t *testing.T) {
	generator := NewAiTrapGenerator()

	seenPatterns := make(map[string]bool)
	for range 20 {
		pattern := generator.resolveHallucinationPattern()
		if pattern == "" {
			t.Errorf("ExpectedNonEmptyPattern")
		}
		seenPatterns[pattern] = true
	}

	if len(seenPatterns) < 2 {
		t.Errorf(
			"ExpectedVariedPatterns: UniqueCount=%d",
			len(seenPatterns),
		)
	}
}
