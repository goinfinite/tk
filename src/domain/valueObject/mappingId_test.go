package tkValueObject

import "testing"

func TestNewMappingId(test *testing.T) {
	test.Run("ValidMappingId", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput MappingId
			expectError    bool
		}{
			{"0", MappingId(0), false},
			{int(0), MappingId(0), false},
			{uint64(42), MappingId(42), false},
			{MappingId(42), MappingId(42), false},
			{float64(42), MappingId(42), false},
			// Invalid ids
			{"-1", MappingId(0), true},
			{int(-1), MappingId(0), true},
			{float64(42.5), MappingId(0), true},
			{"invalid", MappingId(0), true},
			{[]string{"1"}, MappingId(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewMappingId(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("Uint64Method", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     MappingId
			expectedOutput uint64
		}{
			{MappingId(0), 0},
			{MappingId(42), 42},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint64()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     MappingId
			expectedOutput string
		}{
			{MappingId(0), "0"},
			{MappingId(42), "42"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
