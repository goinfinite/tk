package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewHttpHeader(test *testing.T) {
	test.Run("StringInput", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput HttpHeader
			expectError    bool
		}{
			{"X-Forwarded-For", HttpHeader("X-Forwarded-For"), false},
			{"X-Real-IP", HttpHeader("X-Real-IP"), false},
			{"CF-Connecting-IP", HttpHeader("CF-Connecting-IP"), false},
			{"Authorization", HttpHeader("Authorization"), false},
			{"Content-Type", HttpHeader("Content-Type"), false},
			{"X_Custom_Header", HttpHeader("X_Custom_Header"), false},
			{123, HttpHeader("123"), false},
			{true, HttpHeader("true"), false},
			{"", HttpHeader(""), true},
			{"X Forwarded For", HttpHeader(""), true},
			{"X-Forwarded-For!", HttpHeader(""), true},
			{"header\ninjection", HttpHeader(""), true},
			{strings.Repeat("A", 257), HttpHeader(""), true},
			{[]string{"X-Forwarded-For"}, HttpHeader(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewHttpHeader(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf(
					"UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue,
				)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     HttpHeader
			expectedOutput string
		}{
			{HttpHeader("X-Forwarded-For"), "X-Forwarded-For"},
			{HttpHeader("X-Real-IP"), "X-Real-IP"},
			{HttpHeader("CF-Connecting-IP"), "CF-Connecting-IP"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("ReadAsLowerCase", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     HttpHeader
			expectedOutput HttpHeader
		}{
			{HttpHeader("X-Forwarded-For"), HttpHeader("x-forwarded-for")},
			{HttpHeader("x-real-ip"), HttpHeader("x-real-ip")},
			{HttpHeader("CF-Connecting-IP"), HttpHeader("cf-connecting-ip")},
			{HttpHeader("X_Custom_Header"), HttpHeader("x_custom_header")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadAsLowerCase()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})
}
