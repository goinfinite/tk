package tkValueObject

import (
	"testing"
)

func TestNewMimeType(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput MimeType
			expectError    bool
		}{
			{"directory", MimeType("directory"), false},
			{"generic", MimeType("generic"), false},
			{"application/cdmi-object", MimeType("application/cdmi-object"), false},
			{"application/cdmi-queue", MimeType("application/cdmi-queue"), false},
			{"application/cu-seeme", MimeType("application/cu-seeme"), false},
			{"application/davmount+xml", MimeType("application/davmount+xml"), false},
			{"application/dssc+der", MimeType("application/dssc+der"), false},
			{"application/dssc+xml", MimeType("application/dssc+xml"), false},
			{"application/vnd.ms-excel.sheet.macroenabled.12", MimeType("application/vnd.ms-excel.sheet.macroenabled.12"), false},
			{"application/vnd.ms-excel.template.macroenabled.12", MimeType("application/vnd.ms-excel.template.macroenabled.12"), false},
			{"video/vnd.ms-playready.media.pyv", MimeType("video/vnd.ms-playready.media.pyv"), false},
			{"application/vnd.openxmlformats-officedocument.presentationml.presentation", MimeType("application/vnd.openxmlformats-officedocument.presentationml.presentation"), false},
			// Invalid MIME types
			{"", MimeType(""), true},
			{".", MimeType(""), true},
			{"..", MimeType(""), true},
			{"blabla", MimeType(""), true},
			{"application+blabla/vnd.ms~excel", MimeType(""), true},
			{"csv", MimeType(""), true},
			{"text/plain;charset=utf-8", MimeType(""), true}, // with parameters
			{123, MimeType(""), true},
			{true, MimeType(""), true},
			{[]string{"directory"}, MimeType(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewMimeType(testCase.inputValue)
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
			inputValue     MimeType
			expectedOutput string
		}{
			{MimeType("directory"), "directory"},
			{MimeType("generic"), "generic"},
			{MimeType("application/json"), "application/json"},
			{MimeType("text/plain"), "text/plain"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsDirMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     MimeType
			expectedOutput bool
		}{
			{MimeType("directory"), true},
			{MimeType("generic"), false},
			{MimeType("application/json"), false},
			{MimeType("text/plain"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.IsDir()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
