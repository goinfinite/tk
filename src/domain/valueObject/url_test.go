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
			{"wss://localhost:8080", Url("wss://localhost:8080"), false},
			{"wss://production-sfo.browserless.io?token=094632bb-e326-4c63-b953-82b55700b14c", Url("wss://production-sfo.browserless.io?token=094632bb-e326-4c63-b953-82b55700b14c"), false},
			{"grpc://localhost:8080", Url("grpc://localhost:8080"), false},
			{"grpcs://localhost:8080", Url("grpcs://localhost:8080"), false},
			{"tcp://localhost:8080", Url("tcp://localhost:8080"), false},
			{"udp://localhost:8080", Url("udp://localhost:8080"), false},
			{"ftp://localhost:8080", Url("ftp://localhost:8080"), false},
			{"ftps://localhost:8080", Url("ftps://localhost:8080"), false},
			{"file://localhost:8080", Url("file://localhost:8080"), false},
			{"data://localhost:8080", Url("data://localhost:8080"), false},
			{"irc://localhost:8080", Url("irc://localhost:8080"), false},
			{"mailto:admin@example.com", Url("mailto:admin@example.com"), false},
			{"imap://localhost:8080", Url("imap://localhost:8080"), false},
			{"nntp://localhost:8080", Url("nntp://localhost:8080"), false},
			{"pop3://localhost:8080", Url("pop3://localhost:8080"), false},
			{"smtp://localhost:8080", Url("smtp://localhost:8080"), false},
			{"telnet://localhost:8080", Url("telnet://localhost:8080"), false},
			{"https://user:password@localhost:8080", Url("https://user:password@localhost:8080"), false},
			{"https://user:password@localhost:8080/path", Url("https://user:password@localhost:8080/path"), false},
			{"https://user:password@localhost:8080/path?query=param", Url("https://user:password@localhost:8080/path?query=param"), false},
			{123456, Url("tel:123456"), false},
			{"tel:5511999999999", Url("tel:5511999999999"), false},
			{"tel:+5511999999999", Url("tel:+5511999999999"), false},
			{true, Url("https://true"), false},
			{"www.GoOgle.com/", Url("https://www.google.com/"), false},
			{"admin@Example.COM", Url("mailto:admin@example.com"), false},
			{"GoInfinite.Net", Url("https://goinfinite.net"), false},
			{"", Url(""), true},
			{" ", Url(""), true},
			{"http://", Url(""), true},
			{"https://", Url(""), true},
			{"http://notãvalidurl.com/", Url(""), true},
			{"https://invalidmaçalink.com.br/", Url(""), true},
			{":8080:/", Url(""), true},
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
			{"http://localhost:65536", Url(""), true},
			{"http://localhost:-1", Url(""), true},
			{"http://localhost:999999", Url(""), true},
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
