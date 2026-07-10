package tkValueObject

import "testing"

func TestNewHoneypotMaxEntries(t *testing.T) {
	testCaseStructs := []struct {
		inputValue     any
		expectedOutput HoneypotMaxEntries
		expectError    bool
	}{
		{5000, HoneypotMaxEntries(5000), false},
		{"5000", HoneypotMaxEntries(5000), false},
		{1, HoneypotMaxEntries(1), false},
		{0, HoneypotMaxEntries(0), true},
		{-100, HoneypotMaxEntries(0), true},
		{-1, HoneypotMaxEntries(0), true},
		{"invalid", HoneypotMaxEntries(0), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewHoneypotMaxEntries(testCase.inputValue)
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
