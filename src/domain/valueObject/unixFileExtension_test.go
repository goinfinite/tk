package tkValueObject

import (
	"testing"
)

func TestNewUnixFileExtension(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFileExtension
			expectError    bool
		}{
			{".png", UnixFileExtension("png"), false},
			{"png", UnixFileExtension("png"), false},
			{".c", UnixFileExtension("c"), false},
			{"c", UnixFileExtension("c"), false},
			{".ecelp4800", UnixFileExtension("ecelp4800"), false},
			{".n-gage", UnixFileExtension("n-gage"), false},
			{".application", UnixFileExtension("application"), false},
			{".fe_launch", UnixFileExtension("fe_launch"), false},
			{".cdbcmsg", UnixFileExtension("cdbcmsg"), false},
			{".tar.gz", UnixFileExtension("tar.gz"), false},
			{123, UnixFileExtension("123"), false},
			{true, UnixFileExtension("true"), false},
			// Invalid file extensions
			{"", UnixFileExtension(""), true},
			{"file.php?blabla", UnixFileExtension(""), true},
			{"@<php52.sandbox.ntorga.com>.php", UnixFileExtension(""), true},
			{"../file.php", UnixFileExtension(""), true},
			{"hello10/info.php", UnixFileExtension(""), true},
			{[]string{"png"}, UnixFileExtension(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFileExtension(testCase.inputValue)
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
			inputValue     UnixFileExtension
			expectedOutput string
		}{
			{UnixFileExtension("png"), "png"},
			{UnixFileExtension("txt"), "txt"},
			{UnixFileExtension("tar.gz"), "tar.gz"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadMimeTypeMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFileExtension
			expectedOutput MimeType
		}{
			{UnixFileExtension("png"), MimeType("image/png")},
			{UnixFileExtension("txt"), MimeType("text/plain")},
			{UnixFileExtension("pdf"), MimeType("application/pdf")},
			{UnixFileExtension("unknown"), MimeType("generic")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadMimeType()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
