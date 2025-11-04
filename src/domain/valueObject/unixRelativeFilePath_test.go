package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewUnixRelativeFilePath(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixRelativeFilePath
			expectError    bool
		}{
			{".", UnixRelativeFilePath("./"), false},
			{"..", UnixRelativeFilePath("../"), false},
			{"/", UnixRelativeFilePath("./"), false},
			{"file.php", UnixRelativeFilePath("./file.php"), false},
			{"./file.php", UnixRelativeFilePath("./file.php"), false},
			{"../file.php", UnixRelativeFilePath("../file.php"), false},
			{"dir/file.php", UnixRelativeFilePath("./dir/file.php"), false},
			{"file", UnixRelativeFilePath("./file"), false},
			{"file with space.php", UnixRelativeFilePath("./file with space.php"), false},
			{"file.tar.br", UnixRelativeFilePath("./file.tar.br"), false},
			{"subdir/file.php", UnixRelativeFilePath("./subdir/file.php"), false},
			{"./subdir/file.php", UnixRelativeFilePath("./subdir/file.php"), false},
			{"../subdir/file.php", UnixRelativeFilePath("../subdir/file.php"), false},
			{123, UnixRelativeFilePath("./123"), false},
			{true, UnixRelativeFilePath("./true"), false},
			{"file<php", UnixRelativeFilePath("./file<php"), false},
			{"file.php?blabla", UnixRelativeFilePath("./file.php?blabla"), false},
			{"/file.php", UnixRelativeFilePath("./file.php"), false},
			{"/home/file.php", UnixRelativeFilePath("./home/file.php"), false},
			{[]string{"file.php"}, UnixRelativeFilePath(""), true},
			{"файл.txt", UnixRelativeFilePath("./файл.txt"), false},
			{"file_ñame.go", UnixRelativeFilePath("./file_ñame.go"), false},
			{"dir/file+test.md", UnixRelativeFilePath("./dir/file+test.md"), false},
			{"path/to/file(1).txt", UnixRelativeFilePath("./path/to/file(1).txt"), false},
			{"file[bracket].js", UnixRelativeFilePath("./file[bracket].js"), false},
			{"~/config", UnixRelativeFilePath("~/config"), false},
			{"/~/home", UnixRelativeFilePath("~/home"), false},
			{"../../file", UnixRelativeFilePath("../../file"), false},
			{"./../file", UnixRelativeFilePath("./../file"), false},
			{"file@domain.com", UnixRelativeFilePath("./file@domain.com"), false},
			{"dir/sub-file_name.ext", UnixRelativeFilePath("./dir/sub-file_name.ext"), false},
			// Invalid file paths
			{"", UnixRelativeFilePath(""), true},
			{"file" + strings.Repeat("e", 5000) + ".php", UnixRelativeFilePath(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixRelativeFilePath(testCase.inputValue)
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
			inputValue     UnixRelativeFilePath
			expectedOutput string
		}{
			{UnixRelativeFilePath("./file.php"), "./file.php"},
			{UnixRelativeFilePath("./file.php"), "./file.php"},
			{UnixRelativeFilePath("../dir/file.txt"), "../dir/file.txt"},
			{UnixRelativeFilePath("~/config"), "~/config"},
			{UnixRelativeFilePath("./файл.txt"), "./файл.txt"},
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
			inputValue     UnixRelativeFilePath
			expectedOutput UnixRelativeFilePath
		}{
			{UnixRelativeFilePath("./file.php"), UnixRelativeFilePath("./file")},
			{UnixRelativeFilePath("./dir/file.txt"), UnixRelativeFilePath("./dir/file")},
			{UnixRelativeFilePath("./file.tar.gz"), UnixRelativeFilePath("./file")},
			{UnixRelativeFilePath("./file"), UnixRelativeFilePath("./file")},
			{UnixRelativeFilePath("./файл.txt"), UnixRelativeFilePath("./файл")},
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
			inputValue     UnixRelativeFilePath
			expectedOutput UnixFileName
		}{
			{UnixRelativeFilePath("./file.php"), UnixFileName("file.php")},
			{UnixRelativeFilePath("./dir/file.txt"), UnixFileName("file.txt")},
			{UnixRelativeFilePath("./subdir/file.txt"), UnixFileName("file.txt")},
			{UnixRelativeFilePath("./файл.txt"), UnixFileName("файл.txt")},
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
			inputValue     UnixRelativeFilePath
			expectedOutput UnixFileExtension
			expectError    bool
		}{
			{UnixRelativeFilePath("./file.php"), UnixFileExtension("php"), false},
			{UnixRelativeFilePath("./dir/file.txt"), UnixFileExtension("txt"), false},
			{UnixRelativeFilePath("./file.tar.gz"), UnixFileExtension("gz"), false},
			{UnixRelativeFilePath("./file"), UnixFileExtension(""), true},
			{UnixRelativeFilePath("./file.файл"), UnixFileExtension("файл"), true},
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
			inputValue     UnixRelativeFilePath
			expectedOutput UnixFileExtension
			expectError    bool
		}{
			{UnixRelativeFilePath("./file.php"), UnixFileExtension("php"), false},
			{UnixRelativeFilePath("./file.txt"), UnixFileExtension("txt"), false},
			{UnixRelativeFilePath("./file.tar.gz"), UnixFileExtension("tar.gz"), false},
			{UnixRelativeFilePath("./file"), UnixFileExtension(""), true},
			{UnixRelativeFilePath("./file.tar.gz.bz2"), UnixFileExtension("tar.gz.bz2"), true},
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
			inputValue     UnixRelativeFilePath
			expectedOutput UnixFileName
		}{
			{UnixRelativeFilePath("./file.php"), UnixFileName("file")},
			{UnixRelativeFilePath("./dir/file.txt"), UnixFileName("file")},
			{UnixRelativeFilePath("./file.tar.gz"), UnixFileName("file")},
			{UnixRelativeFilePath("./файл.txt"), UnixFileName("файл")},
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
			inputValue     UnixRelativeFilePath
			expectedOutput UnixRelativeFilePath
		}{
			{UnixRelativeFilePath("./file.php"), UnixRelativeFilePath("./")},
			{UnixRelativeFilePath("./dir/file.txt"), UnixRelativeFilePath("./dir")},
			{UnixRelativeFilePath("./subdir/file.txt"), UnixRelativeFilePath("./subdir")},
			{UnixRelativeFilePath("../file.txt"), UnixRelativeFilePath("../")},
			{UnixRelativeFilePath("./a/b/c/file.txt"), UnixRelativeFilePath("./a/b/c")},
			{UnixRelativeFilePath("~/dir/file"), UnixRelativeFilePath("~/dir")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadFileDir()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
