package tkValueObject

import (
	"testing"
)

func TestNewUnixFileName(t *testing.T) {
	t.Run("StringInputStrict", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFileName
			expectError    bool
		}{
			{"a", UnixFileName("a"), false},
			{"17795713_1253219528108045_4440713319482755723_n (1).png", UnixFileName("17795713_1253219528108045_4440713319482755723_n (1).png"), false},
			{"hello.php", UnixFileName("hello.php"), false},
			{"hello_file.php", UnixFileName("hello_file.php"), false},
			{"hello (file).php", UnixFileName("hello (file).php"), false},
			{"Imagem - Sem Título.jpg", UnixFileName("Imagem - Sem Título.jpg"), false},
			{".sudo_as_admin_successful", UnixFileName(".sudo_as_admin_successful"), false},
			{"WhatsApp Image 2018-06-22 at 18.05.08.jpeg", UnixFileName("WhatsApp Image 2018-06-22 at 18.05.08.jpeg"), false},
			{"VID-20181201-WA0001.mp4", UnixFileName("VID-20181201-WA0001.mp4"), false},
			{"AUD-20181201-WA0001.opus", UnixFileName("AUD-20181201-WA0001.opus"), false},
			{"Screenshot_20200101-120000.png", UnixFileName("Screenshot_20200101-120000.png"), false},
			{"unknown.png", UnixFileName("unknown.png"), false},
			{"SPOILER_image.jpg", UnixFileName("SPOILER_image.jpg"), false},
			{"README.md", UnixFileName("README.md"), false},
			{"LICENSE", UnixFileName("LICENSE"), false},
			{"Dockerfile", UnixFileName("Dockerfile"), false},
			{"go.mod", UnixFileName("go.mod"), false},
			{"package.json", UnixFileName("package.json"), false},
			{"index.html", UnixFileName("index.html"), false},
			{"style.css", UnixFileName("style.css"), false},
			{"main.go", UnixFileName("main.go"), false},
			{"budget.xlsx", UnixFileName("budget.xlsx"), false},
			{"report.pdf", UnixFileName("report.pdf"), false},
			{"data.csv", UnixFileName("data.csv"), false},
			{".env", UnixFileName(".env"), false},
			{".gitignore", UnixFileName(".gitignore"), false},
			{"Makefile", UnixFileName("Makefile"), false},
			{"CMakeLists.txt", UnixFileName("CMakeLists.txt"), false},
			{"wp-config.php", UnixFileName("wp-config.php"), false},
			{123, UnixFileName("123"), false},
			{true, UnixFileName("true"), false},
			// Invalid file names
			{"Imagem - Sem Título & BW.jpg", UnixFileName("Imagem - Sem Título & BW.jpg"), true},
			{"Imagem - Sem Título # BW.jpg", UnixFileName("Imagem - Sem Título # BW.jpg"), true},
			{"Imagem - Sem Título @ BW.jpg", UnixFileName("Imagem - Sem Título @ BW.jpg"), true},
			{"hello$_file.php", UnixFileName("hello$_file.php"), true},
			{"Clean Architecture A Craftsman's Guide to Software Structure and Design.pdf", UnixFileName("Clean Architecture A Craftsman's Guide to Software Structure and Design.pdf"), true},
			{"", UnixFileName(""), true},
			{".", UnixFileName(""), true},
			{"..", UnixFileName(""), true},
			{"/", UnixFileName(""), true},
			{"\\", UnixFileName(""), true},
			{"file.php?blabla", UnixFileName(""), true},
			{"@<php52.sandbox.ntorga.com>.php", UnixFileName(""), true},
			{"../../../etc/passwd", UnixFileName(""), true},
			{"..\\..\\windows\\system32\\cmd.exe", UnixFileName(""), true},
			{"file.php;id", UnixFileName(""), true},
			{"script.sh|rm -rf /", UnixFileName(""), true},
			{"../../../../root/.ssh/id_rsa", UnixFileName(""), true},
			{"../../../boot.ini", UnixFileName(""), true},
			{"../../config/database.yml", UnixFileName(""), true},
			{"file.php\r\nContent-Type: text/html", UnixFileName(""), true},
			{"../../../../../../../proc/self/environ", UnixFileName(""), true},
			{"../file.php", UnixFileName(""), true},
			{"hello10/info.php", UnixFileName(""), true},
			{[]string{"hello.php"}, UnixFileName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFileName(testCase.inputValue, false)
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

	t.Run("StringInputUnsafe", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixFileName
			expectError    bool
		}{
			{"17795713_1253219528108045_4440713319482755723_n (1).png", UnixFileName("17795713_1253219528108045_4440713319482755723_n (1).png"), false},
			{"hello.php", UnixFileName("hello.php"), false},
			{"hello_file.php", UnixFileName("hello_file.php"), false},
			{"hello$_file.php", UnixFileName("hello$_file.php"), false},
			{"hello (file).php", UnixFileName("hello (file).php"), false},
			{"Imagem - Sem Título.jpg", UnixFileName("Imagem - Sem Título.jpg"), false},
			{"Imagem - Sem Título & BW.jpg", UnixFileName("Imagem - Sem Título & BW.jpg"), false},
			{"Imagem - Sem Título # BW.jpg", UnixFileName("Imagem - Sem Título # BW.jpg"), false},
			{"Imagem - Sem Título @ BW.jpg", UnixFileName("Imagem - Sem Título @ BW.jpg"), false},
			{"Clean Architecture A Craftsman's Guide to Software Structure and Design.pdf", UnixFileName("Clean Architecture A Craftsman's Guide to Software Structure and Design.pdf"), false},
			{".sudo_as_admin_successful", UnixFileName(".sudo_as_admin_successful"), false},
			{"WhatsApp Image 2018-06-22 at 18.05.08.jpeg", UnixFileName("WhatsApp Image 2018-06-22 at 18.05.08.jpeg"), false},
			{"VID-20181201-WA0001.mp4", UnixFileName("VID-20181201-WA0001.mp4"), false},
			{"AUD-20181201-WA0001.opus", UnixFileName("AUD-20181201-WA0001.opus"), false},
			{"Screenshot_20200101-120000.png", UnixFileName("Screenshot_20200101-120000.png"), false},
			{"unknown.png", UnixFileName("unknown.png"), false},
			{"SPOILER_image.jpg", UnixFileName("SPOILER_image.jpg"), false},
			{"README.md", UnixFileName("README.md"), false},
			{"LICENSE", UnixFileName("LICENSE"), false},
			{"Dockerfile", UnixFileName("Dockerfile"), false},
			{"go.mod", UnixFileName("go.mod"), false},
			{"package.json", UnixFileName("package.json"), false},
			{"index.html", UnixFileName("index.html"), false},
			{"style.css", UnixFileName("style.css"), false},
			{"main.go", UnixFileName("main.go"), false},
			{"budget.xlsx", UnixFileName("budget.xlsx"), false},
			{"report.pdf", UnixFileName("report.pdf"), false},
			{"data.csv", UnixFileName("data.csv"), false},
			{".env", UnixFileName(".env"), false},
			{".gitignore", UnixFileName(".gitignore"), false},
			{"Makefile", UnixFileName("Makefile"), false},
			{"CMakeLists.txt", UnixFileName("CMakeLists.txt"), false},
			{"wp-config.php", UnixFileName("wp-config.php"), false},
			{123, UnixFileName("123"), false},
			{true, UnixFileName("true"), false},
			{"file.php?blabla", UnixFileName("file.php?blabla"), false},
			{"file.php;id", UnixFileName("file.php;id"), false},
			{"@<php52.sandbox.ntorga.com>.php", UnixFileName("@<php52.sandbox.ntorga.com>.php"), false},
			// Invalid file names
			{"", UnixFileName(""), true},
			{".", UnixFileName(""), true},
			{"..", UnixFileName(""), true},
			{"/", UnixFileName(""), true},
			{"\\", UnixFileName(""), true},
			{"../file.php", UnixFileName(""), true},
			{"hello10/info.php", UnixFileName(""), true},
			{"../../../etc/passwd", UnixFileName(""), true},
			{"..\\..\\windows\\system32\\cmd.exe", UnixFileName(""), true},
			{"../../../../root/.ssh/id_rsa", UnixFileName(""), true},
			{"../../../boot.ini", UnixFileName(""), true},
			{"../../config/database.yml", UnixFileName(""), true},
			{"../../../../../../../proc/self/environ", UnixFileName(""), true},
			{[]string{"hello.php"}, UnixFileName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixFileName(testCase.inputValue, true)
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
