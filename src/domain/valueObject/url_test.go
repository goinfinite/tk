package tkValueObject

import "testing"

func TestNewUrl(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Url
			expectError    bool
		}{
			// cSpell:disable
			{"localhost", Url("https://localhost"), false},
			{"localhost:8080", Url("https://localhost:8080"), false},
			{"goinfinite.net", Url("https://goinfinite.net"), false},
			{"http://goinfinite.net/", Url("http://goinfinite.net/"), false},
			{"http://www.goinfinite.net", Url("http://www.goinfinite.net"), false},
			{"https://goinfinite.net/", Url("https://goinfinite.net/"), false},
			{"https://www.goinfinite.net/", Url("https://www.goinfinite.net/"), false},
			{"http://localhost:8080/v1/ticket/253/attachment/b8680d5cc332672c649f4ff8d9e3b77f.svg", Url("http://localhost:8080/v1/ticket/253/attachment/b8680d5cc332672c649f4ff8d9e3b77f.svg"), false},
			{"https://www.cnn.com/politics/live-news/house-speaker-vote-10-20-23/index.html", Url("https://www.cnn.com/politics/live-news/house-speaker-vote-10-20-23/index.html"), false},
			{"https://blog.goinfinite.net/2023/10/vulnerabilidades-top-10-da-owasp-parte-1/", Url("https://blog.goinfinite.net/2023/10/vulnerabilidades-top-10-da-owasp-parte-1/"), false},
			{"https://upload.wikimedia.org/wikipedia/commons/thumb/9/98/WordPress_blue_logo.svg/1200px-WordPress_blue_logo.svg.png", Url("https://upload.wikimedia.org/wikipedia/commons/thumb/9/98/WordPress_blue_logo.svg/1200px-WordPress_blue_logo.svg.png"), false},
			{123, Url("https://123"), false},
			{true, Url("https://true"), false},
			{"", Url(""), true},
			{" ", Url(""), true},
			{"http://", Url(""), true},
			{"https://", Url(""), true},
			{"http://notãvalidurl.com/", Url(""), true},
			{"https://invalidmaçalink.com.br/", Url(""), true},
			{":8080:/", Url(""), true},
			{"www.GoOgle.com/", Url(""), true},
			{"/home/downloads/", Url(""), true},
			{"DROP TABLE users;", Url(""), true},
			{"SELECT * FROM users;", Url(""), true},
			{"<script>alert('XSS')</script>", Url(""), true},
			{"http://<script>alert('XSS')</script>", Url(""), true},
			{"https://<script>alert('XSS')</script>", Url(""), true},
			{"rm -rf /", Url(""), true},
			{"(){|:& };:", Url(""), true},
			{"INSERT INTO users (name, email) VALUES ('admin', 'admin@example.com');", Url(""), true},
			{"sudo rm -r /", Url(""), true},
			{[]string{"http://example.com"}, Url(""), true},
			// cSpell:enable
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUrl(testCase.inputValue)
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
			inputValue     Url
			expectedOutput string
		}{
			{Url("https://localhost"), "https://localhost"},
			{Url("http://goinfinite.net/"), "http://goinfinite.net/"},
			{Url("https://www.cnn.com/path"), "https://www.cnn.com/path"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
