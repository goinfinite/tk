package tkValueObject

import (
	"testing"
)

func TestNewUnixFileName(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFileName
			expectError    bool
		}{
			{"17795713_1253219528108045_4440713319482755723_n\\ \\(1\\).png", UnixFileName("17795713_1253219528108045_4440713319482755723_n\\ \\(1\\).png"), false},
			{"hello.php", UnixFileName("hello.php"), false},
			{"hello_file.php", UnixFileName("hello_file.php"), false},
			{"hello\\$_file.php", UnixFileName("hello\\$_file.php"), false},
			{"hello (file).php", UnixFileName("hello (file).php"), false},
			{"Imagem - Sem Título.jpg", UnixFileName("Imagem - Sem Título.jpg"), false},
			{"Imagem - Sem Título & BW.jpg", UnixFileName("Imagem - Sem Título & BW.jpg"), false},
			{"Imagem - Sem Título # BW.jpg", UnixFileName("Imagem - Sem Título # BW.jpg"), false},
			{"Imagem - Sem Título @ BW.jpg", UnixFileName("Imagem - Sem Título @ BW.jpg"), false},
			{"Clean Architecture A Craftsman's Guide to Software Structure and Design.pdf", UnixFileName("Clean Architecture A Craftsman's Guide to Software Structure and Design.pdf"), false},
			{".sudo_as_admin_successful", UnixFileName(".sudo_as_admin_successful"), false},
			{"WhatsApp Image 2018-06-22 at 18.05.08.jpeg", UnixFileName("WhatsApp Image 2018-06-22 at 18.05.08.jpeg"), false},
			{123, UnixFileName("123"), false},
			{true, UnixFileName("true"), false},
			// Invalid file names
			{"", UnixFileName(""), true},
			{".", UnixFileName(""), true},
			{"..", UnixFileName(""), true},
			{"/", UnixFileName(""), true},
			{"\\", UnixFileName(""), true},
			{"file.php?blabla", UnixFileName(""), true},
			{"@<php52.sandbox.ntorga.com>.php", UnixFileName(""), true},
			{"../file.php", UnixFileName(""), true},
			{"hello10/info.php", UnixFileName(""), true},
			{[]string{"hello.php"}, UnixFileName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFileName(testCase.inputValue)
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
			inputValue     UnixFileName
			expectedOutput string
		}{
			{UnixFileName("hello.php"), "hello.php"},
			{UnixFileName("file.txt"), "file.txt"},
			{UnixFileName(".hidden"), ".hidden"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
