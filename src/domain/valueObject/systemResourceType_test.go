package tkValueObject

import (
	"testing"
)

func TestNewSystemResourceType(t *testing.T) {
	t.Run("ValidSystemResourceType", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"account", false},
			{"resource", false},
			{"a", false},
			{"A", false},
			{"resourceType", false},
			{"resource-type", false},
			{"resource_type", false},
			{"resourceType123", false},
			{"a1b2c3", false},
			{"Test-Resource_Type123", false},
			{int(123), true},
			{float64(123.45), true},
			{true, false},
			{[]string{"test"}, true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewSystemResourceType(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("InvalidSystemResourceType", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"", true},
			{"123resource", true},
			{"-resource", true},
			{"_resource", true},
			{"resource type", true},
			{"resource@type", true},
			{"resource.type", true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewSystemResourceType(testCase.inputValue)
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
			inputValue     SystemResourceType
			expectedOutput string
		}{
			{SystemResourceType("account"), "account"},
			{SystemResourceType("resource-type"), "resource-type"},
			{SystemResourceType("Test_Resource123"), "Test_Resource123"},
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
