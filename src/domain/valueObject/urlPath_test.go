package tkValueObject

import "testing"

func TestNewUrlPath(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UrlPath
			expectError    bool
		}{
			{"", UrlPath("/"), false},
			{"/", UrlPath("/"), false},
			{"blog", UrlPath("/blog"), false},
			{"news/new-product-from-Infinite-revolutionizes-the-market", UrlPath("/news/new-product-from-Infinite-revolutionizes-the-market"), false},
			{"/app/html", UrlPath("/app/html"), false},
			{"/info.php", UrlPath("/info.php"), false},
			{"/app/html/goinfinite.net", UrlPath("/app/html/goinfinite.net"), false},
			{"/v1/ticket/253/attachment/b8680d5cc332672c649f4ff8d9e3b77f.svg", UrlPath("/v1/ticket/253/attachment/b8680d5cc332672c649f4ff8d9e3b77f.svg"), false},
			{"/politics/live-news/house-speaker-vote-10-20-23/index.html", UrlPath("/politics/live-news/house-speaker-vote-10-20-23/index.html"), false},
			{"/2023/10/vulnerabilidades-top-10-da-owasp-parte-1/", UrlPath("/2023/10/vulnerabilidades-top-10-da-owasp-parte-1/"), false},
			{"/wikipedia/commons/thumb/9/98/WordPress_blue_logo.svg/1200px-WordPress_blue_logo.svg.png", UrlPath("/wikipedia/commons/thumb/9/98/WordPress_blue_logo.svg/1200px-WordPress_blue_logo.svg.png"), false},
			{"/path?query=value", UrlPath("/path?query=value"), false},
			// Invalid URL paths
			{"/app/html@", UrlPath(""), true},
			{"/path to download", UrlPath(""), true},
			{"index.js=", UrlPath(""), true},
			{"spaces are invalid", UrlPath(""), true},
			{123, UrlPath("/123"), false},
			{true, UrlPath("/true"), false},
			{[]string{"/path"}, UrlPath(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUrlPath(testCase.inputValue)
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
			inputValue     UrlPath
			expectedOutput string
		}{
			{UrlPath("/"), "/"},
			{UrlPath("/blog"), "/blog"},
			{UrlPath("/path?query"), "/path?query"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadWithoutQueryMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UrlPath
			expectedOutput string
		}{
			{UrlPath("/path"), "/path"},
			{UrlPath("/path?query=value"), "/path"},
			{UrlPath("/blog?param=1&other=2"), "/blog"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadWithoutQuery()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadQueryMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UrlPath
			expectedOutput string
		}{
			{UrlPath("/path?query=value"), "query=value"},
			{UrlPath("/blog?param=1&other=2"), "param=1&other=2"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadQuery()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadWithoutTrailingSlashMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UrlPath
			expectedOutput string
		}{
			{UrlPath("/path/"), "/path"},
			{UrlPath("/blog"), "/blog"},
			{UrlPath("/"), ""},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadWithoutTrailingSlash()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
