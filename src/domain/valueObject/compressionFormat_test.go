package tkValueObject

import (
	"testing"
)

func TestNewCompressionFormat(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CompressionFormat
			expectError    bool
		}{
			// ValidFormats
			{"tar", CompressionFormat("tar"), false},
			{"gzip", CompressionFormat("gzip"), false},
			{"zip", CompressionFormat("zip"), false},
			{"xz", CompressionFormat("xz"), false},
			{"br", CompressionFormat("br"), false},
			{"TAR", CompressionFormat("tar"), false},
			{"GZIP", CompressionFormat("gzip"), false},
			{"ZIP", CompressionFormat("zip"), false},
			{"XZ", CompressionFormat("xz"), false},
			{"BR", CompressionFormat("br"), false},
			{".tar", CompressionFormat("tar"), false},
			{".gzip", CompressionFormat("gzip"), false},
			{".zip", CompressionFormat("zip"), false},
			{".xz", CompressionFormat("xz"), false},
			{".br", CompressionFormat("br"), false},
			{"gz", CompressionFormat("gzip"), false},
			{"tarball", CompressionFormat("tar"), false},
			{"brotli", CompressionFormat("br"), false},
			{".gz", CompressionFormat("gzip"), false},
			{".tarball", CompressionFormat("tar"), false},
			{".brotli", CompressionFormat("br"), false},
			{"tar.gz", CompressionFormat("tar.gz"), false},
			{"tgz", CompressionFormat("tar.gz"), false},
			{"gzipped-tarball", CompressionFormat("tar.gz"), false},
			// InvalidFormats
			{"invalid", CompressionFormat(""), true},
			{"rar", CompressionFormat(""), true},
			{"7z", CompressionFormat(""), true},
			{"", CompressionFormat(""), true},
			{"   ", CompressionFormat(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCompressionFormat(testCase.inputValue)
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

	t.Run("NonStringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CompressionFormat
			expectError    bool
		}{
			{123, CompressionFormat(""), true},
			{true, CompressionFormat(""), true},
			{false, CompressionFormat(""), true},
			{[]string{"tar"}, CompressionFormat(""), true},
			{nil, CompressionFormat(""), true},
			{map[string]string{"format": "tar"}, CompressionFormat(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCompressionFormat(testCase.inputValue)
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
			inputValue     CompressionFormat
			expectedOutput string
		}{
			{CompressionFormat("tar"), "tar"},
			{CompressionFormat("gzip"), "gzip"},
			{CompressionFormat("zip"), "zip"},
			{CompressionFormat("xz"), "xz"},
			{CompressionFormat("br"), "br"},
			{CompressionFormat("tar.gz"), "tar.gz"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
