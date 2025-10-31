package tkValueObject

import (
	"testing"
)

func TestNewHttpMethod(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput HttpMethod
			expectError    bool
		}{
			{"GET", HttpMethod("GET"), false},
			{"HEAD", HttpMethod("HEAD"), false},
			{"POST", HttpMethod("POST"), false},
			{"PUT", HttpMethod("PUT"), false},
			{"DELETE", HttpMethod("DELETE"), false},
			{"CONNECT", HttpMethod("CONNECT"), false},
			{"OPTIONS", HttpMethod("OPTIONS"), false},
			{"TRACE", HttpMethod("TRACE"), false},
			{"PATCH", HttpMethod("PATCH"), false},
			// Invalid HTTP methods
			{"get", HttpMethod(""), true},
			{"INVALID", HttpMethod(""), true},
			{"", HttpMethod(""), true},
			{123, HttpMethod(""), true},
			{true, HttpMethod(""), true},
			{[]string{"GET"}, HttpMethod(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewHttpMethod(testCase.inputValue)
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
			inputValue     HttpMethod
			expectedOutput string
		}{
			{HttpMethod("GET"), "GET"},
			{HttpMethod("POST"), "POST"},
			{HttpMethod("PUT"), "PUT"},
			{HttpMethod("DELETE"), "DELETE"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("HasBodySupportMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     HttpMethod
			expectedOutput bool
		}{
			{HttpMethod("GET"), false},
			{HttpMethod("HEAD"), false},
			{HttpMethod("POST"), true},
			{HttpMethod("PUT"), true},
			{HttpMethod("DELETE"), false},
			{HttpMethod("CONNECT"), false},
			{HttpMethod("OPTIONS"), true},
			{HttpMethod("TRACE"), false},
			{HttpMethod("PATCH"), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.HasBodySupport()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
