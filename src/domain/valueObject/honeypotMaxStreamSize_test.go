package tkValueObject

import "testing"

func TestNewHoneypotMaxStreamSize(t *testing.T) {
	testCaseStructs := []struct {
		inputValue     any
		expectedOutput HoneypotMaxStreamSize
		expectError    bool
	}{
		{20971520, HoneypotMaxStreamSize(20971520), false},
		{"20971520", HoneypotMaxStreamSize(20971520), false},
		{1, HoneypotMaxStreamSize(1), false},
		{0, HoneypotMaxStreamSize(0), true},
		{-1, HoneypotMaxStreamSize(0), true},
		{"invalid", HoneypotMaxStreamSize(0), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewHoneypotMaxStreamSize(testCase.inputValue)
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
