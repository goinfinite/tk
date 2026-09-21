package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewCpuModelName(t *testing.T) {
	t.Run("ValidCpuModelName", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CpuModelName
			expectError    bool
		}{
			{"Intel Xeon E5-2670 v3", CpuModelName("Intel Xeon E5-2670 v3"), false},
			{"AMD Ryzen 9 5950X", CpuModelName("AMD Ryzen 9 5950X"), false},
			{"Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz", CpuModelName("Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz"), false},
			{"ab", CpuModelName("ab"), false},
			{" Intel Xeon ", CpuModelName("Intel Xeon"), false},
			{strings.Repeat("a", 100), CpuModelName(strings.Repeat("a", 100)), false},
			// Invalid model names
			{"", CpuModelName(""), true},
			{"a", CpuModelName(""), true},
			{"Intel\nXeon", CpuModelName(""), true},
			{"Intel;Xeon", CpuModelName(""), true},
			{strings.Repeat("a", 101), CpuModelName(""), true},
			{[]string{"cpu"}, CpuModelName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCpuModelName(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     CpuModelName
			expectedOutput string
		}{
			{CpuModelName("Intel Xeon E5-2670 v3"), "Intel Xeon E5-2670 v3"},
			{CpuModelName("AMD Ryzen 9 5950X"), "AMD Ryzen 9 5950X"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
