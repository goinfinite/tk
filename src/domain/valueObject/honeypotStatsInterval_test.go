package tkValueObject

import (
	"testing"
	"time"
)

func TestNewHoneypotStatsInterval(t *testing.T) {
	testCaseStructs := []struct {
		name             string
		inputValue       any
		expectedDuration time.Duration
	}{
		{"DefaultValueIs30m", nil, 30 * time.Minute},
		{"DurationStringParsed", "10m", 10 * time.Minute},
		{"AtFloorNoClamp", "5m", 5 * time.Minute},
		{"IntValueParsedAsSeconds", "600", 10 * time.Minute},
		{"BelowFloorClampsToFloor", "30s", 5 * time.Minute},
		{"ZeroClampsToFloor", "0", 5 * time.Minute},
		{"NegativeClampsToFloor", "-1m", 5 * time.Minute},
		{"NonParseableDefaults", "abc", 30 * time.Minute},
		{"EmptyStringDefaults", "", 30 * time.Minute},
		{"ExtremeValueAccepted", "8760h", 8760 * time.Hour},
		{"IntInputAsSeconds", int(600), 10 * time.Minute},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput, conversionErr := NewHoneypotStatsInterval(
				testCase.inputValue,
			)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue)
			}
			if actualOutput.Duration() != testCase.expectedDuration {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput.Duration(),
					testCase.expectedDuration,
					testCase.inputValue)
			}
		})
	}
}
