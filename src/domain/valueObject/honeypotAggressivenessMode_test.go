package tkValueObject

import "testing"

func TestNewHoneypotAggressivenessMode(t *testing.T) {
	testCaseStructs := []struct {
		name           string
		inputValue     any
		expectedOutput HoneypotAggressivenessMode
		expectError    bool
	}{
		{
			"ImmediateAccepted",
			"immediate",
			HoneypotAggressivenessModeImmediate,
			false,
		},
		{
			"BalancedAccepted",
			"balanced",
			HoneypotAggressivenessModeBalanced,
			false,
		},
		{
			"TolerantAccepted",
			"tolerant",
			HoneypotAggressivenessModeTolerant,
			false,
		},
		{
			"ObserveAccepted",
			"observe",
			HoneypotAggressivenessModeObserve,
			false,
		},
		{
			"CaseSensitiveRejected",
			"BALANCED",
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"EmptyStringRejected",
			"",
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"DeprecatedStandardRejected",
			"standard",
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"DeprecatedLenientRejected",
			"lenient",
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"DeprecatedPassiveRejected",
			"passive",
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"GarbageRejected",
			"garbage",
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"NonStringRejected",
			123,
			HoneypotAggressivenessMode(""),
			true,
		},
		{
			"SliceRejected",
			[]uint{123},
			HoneypotAggressivenessMode(""),
			true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput, conversionErr := NewHoneypotAggressivenessMode(
				testCase.inputValue,
			)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput,
					testCase.inputValue)
			}
		})
	}
}

func TestHoneypotAggressivenessModeString(t *testing.T) {
	testCaseStructs := []struct {
		name           string
		inputValue     HoneypotAggressivenessMode
		expectedOutput string
	}{
		{"ImmediateString", HoneypotAggressivenessModeImmediate, "immediate"},
		{"BalancedString", HoneypotAggressivenessModeBalanced, "balanced"},
		{"TolerantString", HoneypotAggressivenessModeTolerant, "tolerant"},
		{"ObserveString", HoneypotAggressivenessModeObserve, "observe"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedStringValue: got='%s', want='%s'",
					actualOutput, testCase.expectedOutput)
			}
		})
	}
}

func TestHoneypotAggressivenessModeResolveTier(t *testing.T) {
	testCaseStructs := []struct {
		name         string
		mode         HoneypotAggressivenessMode
		hitCount     int
		expectedTier int
	}{
		{
			"ImmediateZeroHitsTierZero",
			HoneypotAggressivenessModeImmediate, 0, 0,
		},
		{
			"ImmediateOneHitTierThree",
			HoneypotAggressivenessModeImmediate, 1, 3,
		},
		{
			"ImmediateFiveHitsTierThree",
			HoneypotAggressivenessModeImmediate, 5, 3,
		},
		{
			"BalancedZeroHitsTierZero",
			HoneypotAggressivenessModeBalanced, 0, 0,
		},
		{
			"BalancedOneHitTierOne",
			HoneypotAggressivenessModeBalanced, 1, 1,
		},
		{
			"BalancedTwoHitsTierTwo",
			HoneypotAggressivenessModeBalanced, 2, 2,
		},
		{
			"BalancedThreeHitsTierThree",
			HoneypotAggressivenessModeBalanced, 3, 3,
		},
		{
			"BalancedTenHitsTierThree",
			HoneypotAggressivenessModeBalanced, 10, 3,
		},
		{
			"TolerantZeroHitsTierZero",
			HoneypotAggressivenessModeTolerant, 0, 0,
		},
		{
			"TolerantOneHitTierZero",
			HoneypotAggressivenessModeTolerant, 1, 0,
		},
		{
			"TolerantTwoHitsTierOne",
			HoneypotAggressivenessModeTolerant, 2, 1,
		},
		{
			"TolerantFourHitsTierOne",
			HoneypotAggressivenessModeTolerant, 4, 1,
		},
		{
			"TolerantFiveHitsTierTwo",
			HoneypotAggressivenessModeTolerant, 5, 2,
		},
		{
			"TolerantTenHitsTierTwo",
			HoneypotAggressivenessModeTolerant, 10, 2,
		},
		{
			"ObserveZeroHitsTierOne",
			HoneypotAggressivenessModeObserve, 0, 1,
		},
		{
			"ObserveOneHitTierOne",
			HoneypotAggressivenessModeObserve, 1, 1,
		},
		{
			"ObserveFiftyHitsTierOne",
			HoneypotAggressivenessModeObserve, 50, 1,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualTier := testCase.mode.ResolveTier(
				testCase.hitCount,
			)
			if actualTier != testCase.expectedTier {
				t.Errorf("TierMismatch: mode=%s, hits=%d, "+
					"got=%d, want=%d",
					testCase.mode.String(),
					testCase.hitCount,
					actualTier, testCase.expectedTier)
			}
		})
	}
}
