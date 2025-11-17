package tkValueObject

import (
	"testing"
)

func TestNewSystemResourceId(t *testing.T) {
	t.Run("ValidSystemResourceId", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"*", false},
			{"resource", false},
			{"resource123", false},
			{"a", false},
			{"1", false},
			{"resource.id", false},
			{"resource-id", false},
			{"resource_id", false},
			{"123resource", false},
			{"resource.id-123_test", false},
			{int(123), false},
			{float64(123.45), false},
			{true, false},
			{[]string{"test"}, true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewSystemResourceId(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("InvalidSystemResourceId", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"", true},
			{"-resource", true},
			{"_resource", true},
			{".resource", true},
			{"resource id", true},
			{"resource@id", true},
			{"resource/id", true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewSystemResourceId(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     SystemResourceId
			expectedOutput string
		}{
			{SystemResourceId("*"), "*"},
			{SystemResourceId("resource"), "resource"},
			{SystemResourceId("resource.id"), "resource.id"},
			{SystemResourceId("123test-resource_id"), "123test-resource_id"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})
}
