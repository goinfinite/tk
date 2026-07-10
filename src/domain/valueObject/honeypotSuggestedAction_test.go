package tkValueObject

import "testing"

func TestNewHoneypotSuggestedAction(t *testing.T) {
	testCaseStructs := []struct {
		inputValue     any
		expectedOutput HoneypotSuggestedAction
		expectError    bool
	}{
		{"ban", HoneypotSuggestedAction("ban"), false},
		{"BAN", HoneypotSuggestedAction("ban"), false},
		{"servePayload", HoneypotSuggestedAction("servePayload"), false},
		{"SERVEPAYLOAD", HoneypotSuggestedAction("servePayload"), false},
		{"serveStream", HoneypotSuggestedAction("serveStream"), false},
		{"serveAiTrap", HoneypotSuggestedAction("serveAiTrap"), false},
		{"serveMixed", HoneypotSuggestedAction("serveMixed"), false},
		{"passthrough", HoneypotSuggestedAction("passthrough"), false},
		{"explode", HoneypotSuggestedAction(""), true},
		{"", HoneypotSuggestedAction(""), true},
		{123, HoneypotSuggestedAction(""), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewHoneypotSuggestedAction(testCase.inputValue)
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
