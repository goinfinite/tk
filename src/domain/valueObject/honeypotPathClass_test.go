package tkValueObject

import "testing"

func TestNewHoneypotPathClass(t *testing.T) {
	testCaseStructs := []struct {
		inputValue     any
		expectedOutput HoneypotPathClass
		expectError    bool
	}{
		{"staticVulnerability", HoneypotPathClass("staticVulnerability"), false},
		{"STATICVULNERABILITY", HoneypotPathClass("staticVulnerability"), false},
		{"bandwidthExhaust", HoneypotPathClass("bandwidthExhaust"), false},
		{"BANDWIDTHEXHAUST", HoneypotPathClass("bandwidthExhaust"), false},
		{"aiTrap", HoneypotPathClass("aiTrap"), false},
		{"AITRAP", HoneypotPathClass("aiTrap"), false},
		{"unknown", HoneypotPathClass(""), true},
		{"", HoneypotPathClass(""), true},
		{123, HoneypotPathClass(""), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewHoneypotPathClass(testCase.inputValue)
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
