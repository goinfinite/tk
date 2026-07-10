package tkValueObject

import "testing"

func TestNewHoneypotActivePathCount(t *testing.T) {
	testCaseStructs := []struct {
		inputValue     any
		inputCeiling   int
		expectedOutput HoneypotActivePathCount
		expectError    bool
	}{
		{50, 114, HoneypotActivePathCount(50), false},
		{30, 114, HoneypotActivePathCount(30), false},
		{114, 114, HoneypotActivePathCount(114), false},
		{-1, 114, HoneypotActivePathCount(30), false},
		{0, 114, HoneypotActivePathCount(30), false},
		{99999, 114, HoneypotActivePathCount(114), false},
		{"50", 114, HoneypotActivePathCount(50), false},
		{"invalid", 114, HoneypotActivePathCount(0), true},
	}

	for _, testCase := range testCaseStructs {
		actualOutput, err := NewHoneypotActivePathCount(testCase.inputValue, testCase.inputCeiling)
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
