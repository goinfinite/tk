package tkValueObject_test

import (
	"errors"
	"testing"
	"time"

	tkUseCase "github.com/goinfinite/tk/src/domain/useCase"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewHoneypotBanDuration(t *testing.T) {
	testCaseStructs := []struct {
		name             string
		inputValue       any
		expectedDuration time.Duration
	}{
		{"ZeroDurationDefaultsTo24h", time.Duration(0), 24 * time.Hour},
		{"PositiveDurationAccepted", 48 * time.Hour, 48 * time.Hour},
		{"StringDurationParsed", "12h", 12 * time.Hour},
		{"Int64SecondsParsed", int64(3600), 1 * time.Hour},
		{"NegativeDurationDefaults", -1 * time.Hour, 24 * time.Hour},
		{"InvalidStringDefaults", "not-a-duration", 24 * time.Hour},
		{"EmptyStringDefaults", "", 24 * time.Hour},
		{"GarbageMapDefaults", map[string]int{"x": 1}, 24 * time.Hour},
		{"NilDefaults", nil, 24 * time.Hour},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			rawInput := testCase.inputValue
			actualOutput, conversionErr :=
				tkValueObject.NewHoneypotBanDuration(rawInput)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: [%v] [%v]",
					conversionErr, rawInput)
			}
			actualDuration := actualOutput.Duration()
			if actualDuration != testCase.expectedDuration {
				t.Errorf(
					"DurationMismatch: got=%v, want=%v [%v]",
					actualDuration,
					testCase.expectedDuration,
					rawInput,
				)
			}
		})
	}
}

func TestSentinelErrNilHoneypotQueryRepoIsComparable(
	t *testing.T,
) {
	isComparable := errors.Is(
		tkUseCase.ErrNilHoneypotQueryRepo,
		tkUseCase.ErrNilHoneypotQueryRepo,
	)
	if !isComparable {
		t.Errorf("SentinelNotComparable")
	}
}
