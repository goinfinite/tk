package tkValueObject

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUnixAbsoluteFilePath(test *testing.T) {
	test.Run("StringInputStrict", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixAbsoluteFilePath
			expectError    bool
		}{
			{"/", UnixAbsoluteFilePath("/"), false},
			{"a", UnixAbsoluteFilePath("/a"), false},
			{"/root", UnixAbsoluteFilePath("/root"), false},
			{"/root/", UnixAbsoluteFilePath("/root/"), false},
			{"/home/sandbox/file.php", UnixAbsoluteFilePath("/home/sandbox/file.php"), false},
			{"/home/sandbox/file", UnixAbsoluteFilePath("/home/sandbox/file"), false},
			{"/home/sandbox/file with space.php", UnixAbsoluteFilePath("/home/sandbox/file with space.php"), false},
			{"/home/100sandbox/file.php", UnixAbsoluteFilePath("/home/100sandbox/file.php"), false},
			{"/home/100sandbox/Imagem - Sem Título.jpg", UnixAbsoluteFilePath("/home/100sandbox/Imagem - Sem Título.jpg"), false},
			{"/file.php", UnixAbsoluteFilePath("/file.php"), false},
			{"/file.tar.br", UnixAbsoluteFilePath("/file.tar.br"), false},
			{"/file with space.php", UnixAbsoluteFilePath("/file with space.php"), false},
			{"file.php", UnixAbsoluteFilePath("/file.php"), false},
			{123, UnixAbsoluteFilePath("/123"), false},
			{true, UnixAbsoluteFilePath("/true"), false},
			{"/var/log/syslog", UnixAbsoluteFilePath("/var/log/syslog"), false},
			{"/usr/local/bin/script.sh", UnixAbsoluteFilePath("/usr/local/bin/script.sh"), false},
			{"/etc/nginx/nginx.conf", UnixAbsoluteFilePath("/etc/nginx/nginx.conf"), false},
			{"/opt/app/config.yaml", UnixAbsoluteFilePath("/opt/app/config.yaml"), false},
			{"/tmp/.hidden", UnixAbsoluteFilePath("/tmp/.hidden"), false},
			{"/var//log///syslog", UnixAbsoluteFilePath("/var//log///syslog"), false},
			{"/etc/passwd\r\n", UnixAbsoluteFilePath("/etc/passwd"), false}, // InterfaceToString trims whitespace
			// Invalid file paths
			{"", UnixAbsoluteFilePath(""), true},
			{"/home/@directory/file.gif", UnixAbsoluteFilePath(""), true},
			{"/home/user/file.php?blabla", UnixAbsoluteFilePath(""), true},
			{"/home/sandbox/domains/@<php52.sandbox.ntorga.com>", UnixAbsoluteFilePath(""), true},
			{"../file.php", UnixAbsoluteFilePath(""), true},
			{"./file.php", UnixAbsoluteFilePath(""), true},
			{"~/", UnixAbsoluteFilePath(""), true},
			{"~/file.php", UnixAbsoluteFilePath(""), true},
			{"/~/file.php", UnixAbsoluteFilePath(""), true},
			{"/home/../file.php", UnixAbsoluteFilePath(""), true},
			{"/home/../../file.php", UnixAbsoluteFilePath(""), true},
			{"/home/file" + strings.Repeat("e", 5000) + ".php", UnixAbsoluteFilePath(""), true},
			{[]string{"/file.php"}, UnixAbsoluteFilePath(""), true},
			{"/etc/passwd%00", UnixAbsoluteFilePath(""), true},
			{"/etc/passwd\x00", UnixAbsoluteFilePath(""), true},
			{"/var/www/<script>alert(1)</script>", UnixAbsoluteFilePath(""), true},
			{"/home/user/file\nanother", UnixAbsoluteFilePath(""), true},
			{"//../etc/passwd", UnixAbsoluteFilePath(""), true},
			// Legitimate Unix paths with special chars
			{"~file.php", UnixAbsoluteFilePath("/~file.php"), false},
			{"/home/user/photo (1).jpg", UnixAbsoluteFilePath("/home/user/photo (1).jpg"), false},
			{"/home/user/[Backup] notes.txt", UnixAbsoluteFilePath("/home/user/[Backup] notes.txt"), false},
			{"/home/user/file~", UnixAbsoluteFilePath("/home/user/file~"), false},
			{"/mime.properties0,v", UnixAbsoluteFilePath("/mime.properties0,v"), false},
			{"/crond.reboot", UnixAbsoluteFilePath("/crond.reboot"), false},
			{"/sys/block/nvme0n1/bdi/subsystem/0:84", UnixAbsoluteFilePath("/sys/block/nvme0n1/bdi/subsystem/0:84"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixAbsoluteFilePath(testCase.inputValue, false)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("StringInputUnsafe", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixAbsoluteFilePath
			expectError    bool
		}{
			{"/", UnixAbsoluteFilePath("/"), false},
			{"a", UnixAbsoluteFilePath("/a"), false},
			{"/root", UnixAbsoluteFilePath("/root"), false},
			{"/root/", UnixAbsoluteFilePath("/root/"), false},
			{"/home/sandbox/file.php", UnixAbsoluteFilePath("/home/sandbox/file.php"), false},
			{"/home/sandbox/file", UnixAbsoluteFilePath("/home/sandbox/file"), false},
			{"/home/sandbox/file with space.php", UnixAbsoluteFilePath("/home/sandbox/file with space.php"), false},
			{"/home/100sandbox/file.php", UnixAbsoluteFilePath("/home/100sandbox/file.php"), false},
			{"/home/100sandbox/Imagem - Sem Título.jpg", UnixAbsoluteFilePath("/home/100sandbox/Imagem - Sem Título.jpg"), false},
			{"/file.php", UnixAbsoluteFilePath("/file.php"), false},
			{"/file.tar.br", UnixAbsoluteFilePath("/file.tar.br"), false},
			{"/file with space.php", UnixAbsoluteFilePath("/file with space.php"), false},
			{"file.php", UnixAbsoluteFilePath("/file.php"), false},
			{123, UnixAbsoluteFilePath("/123"), false},
			{true, UnixAbsoluteFilePath("/true"), false},
			{"/home/@directory/file.gif", UnixAbsoluteFilePath("/home/@directory/file.gif"), false},
			{"/home/user/file.php?blabla", UnixAbsoluteFilePath("/home/user/file.php?blabla"), false},
			{"/home/sandbox/domains/@<php52.sandbox.ntorga.com>", UnixAbsoluteFilePath("/home/sandbox/domains/@<php52.sandbox.ntorga.com>"), false},
			{"~file.php", UnixAbsoluteFilePath("/~file.php"), false},
			{"/var/log/syslog", UnixAbsoluteFilePath("/var/log/syslog"), false},
			{"/usr/local/bin/script.sh", UnixAbsoluteFilePath("/usr/local/bin/script.sh"), false},
			{"/etc/nginx/nginx.conf", UnixAbsoluteFilePath("/etc/nginx/nginx.conf"), false},
			{"/opt/app/config.yaml", UnixAbsoluteFilePath("/opt/app/config.yaml"), false},
			{"/tmp/.hidden", UnixAbsoluteFilePath("/tmp/.hidden"), false},
			{"/var//log///syslog", UnixAbsoluteFilePath("/var//log///syslog"), false},
			{"/var/www/index.html#fragment", UnixAbsoluteFilePath("/var/www/index.html#fragment"), false},
			{"/var/www/<script>alert(1)</script>", UnixAbsoluteFilePath("/var/www/<script>alert(1)</script>"), false},
			{"/etc/passwd%00", UnixAbsoluteFilePath("/etc/passwd%00"), false},
			{"/etc/passwd\r\n", UnixAbsoluteFilePath("/etc/passwd"), false},
			// Invalid file paths
			{"", UnixAbsoluteFilePath(""), true},
			{"../file.php", UnixAbsoluteFilePath(""), true},
			{"./file.php", UnixAbsoluteFilePath(""), true},
			{"~/", UnixAbsoluteFilePath(""), true},
			{"~/file.php", UnixAbsoluteFilePath(""), true},
			{"/~/file.php", UnixAbsoluteFilePath(""), true},
			{"/home/../file.php", UnixAbsoluteFilePath(""), true},
			{"/home/../../file.php", UnixAbsoluteFilePath(""), true},
			{"/home/file" + strings.Repeat("e", 5000) + ".php", UnixAbsoluteFilePath(""), true},
			{[]string{"/file.php"}, UnixAbsoluteFilePath(""), true},
			{"/etc/passwd\x00", UnixAbsoluteFilePath(""), true},
			{"/home/user/file\nanother", UnixAbsoluteFilePath(""), true},
			{"//../etc/passwd", UnixAbsoluteFilePath(""), true},
			// Legitimate Unix paths with special chars (unsafe allows more)
			{"/mime.properties0,v", UnixAbsoluteFilePath("/mime.properties0,v"), false},
			{"/crond.reboot", UnixAbsoluteFilePath("/crond.reboot"), false},
			{"/sys/block/nvme0n1/bdi/subsystem/0:84", UnixAbsoluteFilePath("/sys/block/nvme0n1/bdi/subsystem/0:84"), false},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixAbsoluteFilePath(testCase.inputValue, true)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput string
		}{
			{UnixAbsoluteFilePath("/home/file.php"), "/home/file.php"},
			{UnixAbsoluteFilePath("/root/"), "/root/"},
			{UnixAbsoluteFilePath("/file.txt"), "/file.txt"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadWithoutExtensionMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput UnixAbsoluteFilePath
		}{
			{UnixAbsoluteFilePath("/home/file.php"), UnixAbsoluteFilePath("/home/file")},
			{UnixAbsoluteFilePath("/home/file.txt"), UnixAbsoluteFilePath("/home/file")},
			{UnixAbsoluteFilePath("/home/file"), UnixAbsoluteFilePath("/home/file")},
			{UnixAbsoluteFilePath("/home/file.tar.gz"), UnixAbsoluteFilePath("/home/file")},
			{UnixAbsoluteFilePath("/home/.hidden"), UnixAbsoluteFilePath("/home/.hidden")},
			{UnixAbsoluteFilePath("/home/.config.json"), UnixAbsoluteFilePath("/home/.config")},
			{UnixAbsoluteFilePath("/home/README"), UnixAbsoluteFilePath("/home/README")},
			{
				UnixAbsoluteFilePath("/home/.htaccess"),
				UnixAbsoluteFilePath("/home/.htaccess"),
			},
			{UnixAbsoluteFilePath("/home/file."), UnixAbsoluteFilePath("/home/file.")},
			{UnixAbsoluteFilePath("/home/site.tar.gz"), UnixAbsoluteFilePath("/home/site")},
			{
				UnixAbsoluteFilePath("/home/file.someverylongextension"),
				UnixAbsoluteFilePath("/home/file.someverylongextension"),
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadWithoutExtension(false)
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadFileNameMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput UnixFileName
			expectError    bool
		}{
			{UnixAbsoluteFilePath("/home/file.php"), UnixFileName("file.php"), false},
			{UnixAbsoluteFilePath("/file.txt"), UnixFileName("file.txt"), false},
			{UnixAbsoluteFilePath("/root/dir"), UnixFileName("dir"), false},
			{UnixAbsoluteFilePath("/root/dir/"), UnixFileName("dir"), false},
			{UnixAbsoluteFilePath("/"), UnixFileName(""), true},
			{UnixAbsoluteFilePath("/home/user/."), UnixFileName(""), true},
			{UnixAbsoluteFilePath("/home/user/.."), UnixFileName(""), true},
			{UnixAbsoluteFilePath("/home/user/~"), UnixFileName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadFileName(false)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadFileNameMethodRootPathError", func(subtest *testing.T) {
		_, err := UnixAbsoluteFilePath("/").ReadFileName(false)
		if !errors.Is(err, ErrRootPathHasNoFileName) {
			subtest.Errorf("ExpectedErrRootPathHasNoFileName, got: %v", err)
		}
	})

	test.Run("ReadFileExtensionMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput UnixFileExtension
			expectedError  error
		}{
			{UnixAbsoluteFilePath("/home/file.php"), UnixFileExtension("php"), nil},
			{UnixAbsoluteFilePath("/home/file.txt"), UnixFileExtension("txt"), nil},
			{UnixAbsoluteFilePath("/home/README"), UnixFileExtension(""), nil},
			{UnixAbsoluteFilePath("/home/.htaccess"), UnixFileExtension(""), nil},
			{UnixAbsoluteFilePath("/home/file."), UnixFileExtension(""), nil},
			{UnixAbsoluteFilePath("/home/.config.json"), UnixFileExtension("json"), nil},
			{UnixAbsoluteFilePath("/home/file.tar.gz"), UnixFileExtension("gz"), nil},
			{UnixAbsoluteFilePath("/home/site.tar.gz"), UnixFileExtension("gz"), nil},
			{
				UnixAbsoluteFilePath("/home/file.someverylongextension"),
				UnixFileExtension(""),
				ErrInvalidFileExtension,
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadFileExtension()
			if testCase.expectedError == nil && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if testCase.expectedError != nil && !errors.Is(err, testCase.expectedError) {
				subtest.Errorf(
					"ExpectedError: '%v' vs '%v' [%v]",
					testCase.expectedError, err, testCase.inputValue,
				)
			}
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadFileExtensionMethodPathError", func(subtest *testing.T) {
		_, err := UnixAbsoluteFilePath("/home/user/.").ReadFileExtension()
		if err == nil {
			subtest.Fatal("MissingExpectedError")
		}
		if errors.Is(err, ErrInvalidFileExtension) {
			subtest.Errorf("PathErrorWrappedAsInvalidFileExtension: '%v'", err)
		}
	})

	test.Run("ReadCompoundFileExtensionMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput UnixFileExtension
			expectError    bool
		}{
			{UnixAbsoluteFilePath("/home/file.php"), UnixFileExtension("php"), false},
			{UnixAbsoluteFilePath("/home/file.txt"), UnixFileExtension("txt"), false},
			{UnixAbsoluteFilePath("/home/file"), UnixFileExtension(""), false},
			{UnixAbsoluteFilePath("/home/file.tar.gz"), UnixFileExtension("tar.gz"), false},
			{UnixAbsoluteFilePath("/home/README"), UnixFileExtension(""), false},
			{UnixAbsoluteFilePath("/home/.htaccess"), UnixFileExtension(""), false},
			{UnixAbsoluteFilePath("/home/file."), UnixFileExtension(""), false},
			{UnixAbsoluteFilePath("/home/.config.json"), UnixFileExtension("json"), false},
			{UnixAbsoluteFilePath("/home/site.tar.gz"), UnixFileExtension("tar.gz"), false},
			{
				UnixAbsoluteFilePath("/home/file.someverylongextension"),
				UnixFileExtension(""),
				true,
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadCompoundFileExtension()
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadFileNameWithoutExtensionMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput UnixFileName
			expectError    bool
		}{
			{UnixAbsoluteFilePath("/home/file.php"), UnixFileName("file"), false},
			{UnixAbsoluteFilePath("/home/file.txt"), UnixFileName("file"), false},
			{UnixAbsoluteFilePath("/home/file"), UnixFileName("file"), false},
			{UnixAbsoluteFilePath("/home/file.tar.gz"), UnixFileName("file"), false},
			{UnixAbsoluteFilePath("/home/.hidden"), UnixFileName(".hidden"), false},
			{UnixAbsoluteFilePath("/home/.config.json"), UnixFileName(".config"), false},
			{UnixAbsoluteFilePath("/home/README"), UnixFileName("README"), false},
			{UnixAbsoluteFilePath("/home/.htaccess"), UnixFileName(".htaccess"), false},
			{UnixAbsoluteFilePath("/home/file."), UnixFileName("file."), false},
			{UnixAbsoluteFilePath("/home/site.tar.gz"), UnixFileName("site"), false},
			{
				UnixAbsoluteFilePath("/home/file.someverylongextension"),
				UnixFileName("file.someverylongextension"),
				false,
			},
			{UnixAbsoluteFilePath("/root/dir/"), UnixFileName("dir"), false},
			{UnixAbsoluteFilePath("/home/user/."), UnixFileName(""), true},
			{UnixAbsoluteFilePath("/home/user/.."), UnixFileName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := testCase.inputValue.ReadFileNameWithoutExtension(false)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("ReadFileDirMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixAbsoluteFilePath
			expectedOutput UnixAbsoluteFilePath
		}{
			{UnixAbsoluteFilePath("/home/file.php"), UnixAbsoluteFilePath("/home")},
			{UnixAbsoluteFilePath("/root/dir/file.txt"), UnixAbsoluteFilePath("/root/dir")},
			{UnixAbsoluteFilePath("/file.txt"), UnixAbsoluteFilePath("/")},
			{UnixAbsoluteFilePath("/root/dir/"), UnixAbsoluteFilePath("/root")},
			{UnixAbsoluteFilePath("/root/"), UnixAbsoluteFilePath("/")},
			{UnixAbsoluteFilePath("/"), UnixAbsoluteFilePath("/")},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadFileDir()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("SecurityTestSuite", func(subtest *testing.T) {
		subtest.Run("StrictModeBlocksShellInjection", func(subtest *testing.T) {
			shellInjectionInputValues := []string{
				"/tmp/file;rm -rf /",
				"/tmp/file|cat /etc/passwd",
				"/tmp/file&whoami",
				"/tmp/file$PATH",
				"/tmp/file`id`",
				"/tmp/file>output",
				"/tmp/file<input",
				"/tmp/file{malicious}",
				"/tmp/file!history",
				"/tmp/file#comment",
				"/tmp/file*glob",
				"/tmp/file?glob",
			}

			for _, inputValue := range shellInjectionInputValues {
				_, err := NewUnixAbsoluteFilePath(inputValue, false)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%s]", inputValue)
				}
			}
		})

		subtest.Run("BothModesBlockControlChars", func(subtest *testing.T) {
			controlCharInputValues := []string{
				"/tmp/file" + string([]byte{0x00}) + "injection",
				"/tmp/file" + string([]byte{0x1f}) + "injection",
				"/tmp/file" + string([]byte{0x7f}) + "injection",
				"/tmp/file\ninjection",
				"/tmp/file\rinjection",
				"/tmp/file\tinjection",
			}

			for _, inputValue := range controlCharInputValues {
				_, err := NewUnixAbsoluteFilePath(inputValue, false)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%q]", inputValue)
				}

				_, err = NewUnixAbsoluteFilePath(inputValue, true)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%q]", inputValue)
				}
			}
		})

		subtest.Run("BothModesBlockPathTraversal", func(subtest *testing.T) {
			traversalInputValues := []string{
				"../etc/passwd",
				"./etc/passwd",
				"~/",
				"/home/../etc/passwd",
				"/home/../../etc/passwd",
				"//../etc/passwd",
			}

			for _, inputValue := range traversalInputValues {
				_, err := NewUnixAbsoluteFilePath(inputValue, false)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%s]", inputValue)
				}

				_, err = NewUnixAbsoluteFilePath(inputValue, true)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%s]", inputValue)
				}
			}
		})

		subtest.Run("UnsafeModeAllowsShellChars", func(subtest *testing.T) {
			shellCharInputValues := []string{
				"/tmp/file;malicious",
				"/tmp/file|pipe",
				"/tmp/file&background",
				"/tmp/file$dollar",
				"/tmp/file>redirect",
				"/tmp/file<in",
				"/tmp/file(paren)",
				"/tmp/file{brace}",
				"/tmp/file[bracket]",
				"/tmp/file!bang",
				"/tmp/file#hash",
				"/tmp/file*star",
				"/tmp/file?question",
			}

			for _, inputValue := range shellCharInputValues {
				_, err := NewUnixAbsoluteFilePath(inputValue, true)
				if err != nil {
					subtest.Errorf("UnexpectedError: [%s]", inputValue)
				}
			}
		})

		subtest.Run("StrictModeAllowsLegitimateSpecialChars", func(subtest *testing.T) {
			legitInputValues := []string{
				"/home/user/photo (1).jpg",
				"/home/user/[Backup] notes.txt",
				"/home/user/file~",
				"/~localfile.txt",
			}

			for _, inputValue := range legitInputValues {
				_, err := NewUnixAbsoluteFilePath(inputValue, false)
				if err != nil {
					subtest.Errorf("UnexpectedError: [%s]", inputValue)
				}
			}
		})

		subtest.Run("BothModesBlockTildeExpansionPatterns", func(subtest *testing.T) {
			tildeExpansionInputValues := []string{
				"~/",
				"~/file.php",
				"/~/file.php",
				"~user/file.php",
				"/~user/file.php",
				"/~admin/config",
			}

			for _, inputValue := range tildeExpansionInputValues {
				_, err := NewUnixAbsoluteFilePath(inputValue, false)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%s]", inputValue)
				}

				_, err = NewUnixAbsoluteFilePath(inputValue, true)
				if err == nil {
					subtest.Errorf("MissingExpectedError: [%s]", inputValue)
				}
			}
		})
	})
}
