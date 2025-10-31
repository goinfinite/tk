package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewUnixFilePath(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFilePath
			expectError    bool
		}{
			{"/", UnixFilePath("/"), false},
			{"/root", UnixFilePath("/root"), false},
			{"/root/", UnixFilePath("/root/"), false},
			{"/home/sandbox/file.php", UnixFilePath("/home/sandbox/file.php"), false},
			{"/home/sandbox/file", UnixFilePath("/home/sandbox/file"), false},
			{"/home/sandbox/file with space.php", UnixFilePath("/home/sandbox/file with space.php"), false},
			{"/home/100sandbox/file.php", UnixFilePath("/home/100sandbox/file.php"), false},
			{"/home/100sandbox/Imagem - Sem Título.jpg", UnixFilePath("/home/100sandbox/Imagem - Sem Título.jpg"), false},
			{"/home/@directory/file.gif", UnixFilePath("/home/@directory/file.gif"), false},
			{"/file.php", UnixFilePath("/file.php"), false},
			{"/file.tar.br", UnixFilePath("/file.tar.br"), false},
			{"/file with space.php", UnixFilePath("/file with space.php"), false},
			// Invalid file paths
			{"", UnixFilePath(""), true},
			{"/home/user/file.php?blabla", UnixFilePath(""), true},
			{"/home/sandbox/domains/@<php52.sandbox.ntorga.com>", UnixFilePath(""), true},
			{"../file.php", UnixFilePath(""), true},
			{"./file.php", UnixFilePath(""), true},
			{"file.php", UnixFilePath(""), true},
			{"/home/../file.php", UnixFilePath(""), true},
			{"/home/../../file.php", UnixFilePath(""), true},
			{"/home/file" + strings.Repeat("e", 5000) + ".php", UnixFilePath(""), true},
			{123, UnixFilePath(""), true},
			{true, UnixFilePath(""), true},
			{[]string{"/file.php"}, UnixFilePath(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFilePath(testCase.inputValue)
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
			inputValue     UnixFilePath
			expectedOutput string
		}{
			{UnixFilePath("/home/file.php"), "/home/file.php"},
			{UnixFilePath("/root/"), "/root/"},
			{UnixFilePath("/file.txt"), "/file.txt"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadWithoutExtensionMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePath
			expectedOutput UnixFilePath
		}{
			{UnixFilePath("/home/file.php"), UnixFilePath("/home/file")},
			{UnixFilePath("/home/file.txt"), UnixFilePath("/home/file")},
			{UnixFilePath("/home/file"), UnixFilePath("/home/file")},
			{UnixFilePath("/home/file.tar.gz"), UnixFilePath("/home/file")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadWithoutExtension()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadFileNameMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePath
			expectedOutput UnixFileName
		}{
			{UnixFilePath("/home/file.php"), UnixFileName("file.php")},
			{UnixFilePath("/root/dir/"), UnixFileName("dir")},
			{UnixFilePath("/file.txt"), UnixFileName("file.txt")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadFileName()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadFileExtensionMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePath
			expectedOutput UnixFileExtension
			expectError    bool
		}{
			{UnixFilePath("/home/file.php"), UnixFileExtension("php"), false},
			{UnixFilePath("/home/file.txt"), UnixFileExtension("txt"), false},
			{UnixFilePath("/home/file"), UnixFileExtension(""), true},
			{UnixFilePath("/home/file.tar.gz"), UnixFileExtension("gz"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadFileExtension()
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
	})

	t.Run("ReadCompoundFileExtensionMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePath
			expectedOutput UnixFileExtension
			expectError    bool
		}{
			{UnixFilePath("/home/file.php"), UnixFileExtension("php"), false},
			{UnixFilePath("/home/file.txt"), UnixFileExtension("txt"), false},
			{UnixFilePath("/home/file"), UnixFileExtension(""), true},
			{UnixFilePath("/home/file.tar.gz"), UnixFileExtension("tar.gz"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadCompoundFileExtension()
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
	})

	t.Run("ReadFileNameWithoutExtensionMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePath
			expectedOutput UnixFileName
		}{
			{UnixFilePath("/home/file.php"), UnixFileName("file")},
			{UnixFilePath("/home/file.txt"), UnixFileName("file")},
			{UnixFilePath("/home/file"), UnixFileName("file")},
			{UnixFilePath("/home/file.tar.gz"), UnixFileName("file")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadFileNameWithoutExtension()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadFileDirMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixFilePath
			expectedOutput UnixFilePath
		}{
			{UnixFilePath("/home/file.php"), UnixFilePath("/home")},
			{UnixFilePath("/root/dir/file.txt"), UnixFilePath("/root/dir")},
			{UnixFilePath("/file.txt"), UnixFilePath("/")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadFileDir()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
