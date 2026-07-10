package tkValueObject

import "testing"

func TestNewHoneypotAggressivenessMode(t *testing.T) {
	testCaseStructs := []struct {
		inputValue     any
		expectedOutput HoneypotAggressivenessMode
		expectError    bool
	}{
		{"balanced", HoneypotAggressivenessMode("balanced"), false},
		{"BALANCED", HoneypotAggressivenessMode("balanced"), false},
		{"immediate", HoneypotAggressivenessMode("immediate"), false},
		{"tolerant", HoneypotAggressivenessMode("tolerant"), false},
		{"observe", HoneypotAggressivenessMode("observe"), false},
		{"OBSERVE", HoneypotAggressivenessMode("observe"), false},
		{"aggressive", HoneypotAggressivenessMode(""), true},
		{"unknown", HoneypotAggressivenessMode(""), true},
		{"", HoneypotAggressivenessMode(""), true},
		{123, HoneypotAggressivenessMode(""), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewHoneypotAggressivenessMode(testCase.inputValue)
		if testCase.expectError && err == nil {
			t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
		}
		if !testCase.expectError && err != nil {
			t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
		}
		if !testCase.expectError && actualOutput != testCase.expectedOutput {
			t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
		}
	}
}
