package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewUserAgent(t *testing.T) {
	testCases := []struct {
		name        string
		input       any
		expected    string
		expectError bool
	}{
		{
			name:        "SimpleUserAgent",
			input:       "Mozilla/5.0",
			expected:    "Mozilla/5.0",
			expectError: false,
		},
		{
			name:        "ChromeUserAgent",
			input:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectError: false,
		},
		{
			name:        "BotUserAgent",
			input:       "Googlebot/2.1 (+http://www.google.com/bot.html)",
			expected:    "Googlebot/2.1 (+http://www.google.com/bot.html)",
			expectError: false,
		},
		{
			name:        "CurlUserAgent",
			input:       "curl/7.68.0",
			expected:    "curl/7.68.0",
			expectError: false,
		},
		{
			name:        "UserAgentWithSpecialChars",
			input:       "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)",
			expected:    "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)",
			expectError: false,
		},
		{
			name:        "MinimalUserAgent",
			input:       "a",
			expected:    "a",
			expectError: false,
		},
		{
			name:        "UserAgentWithSpaces",
			input:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			expected:    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			expectError: false,
		},
		{
			name: "UserAgentWithAllValidChars",
			input: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
				"KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 " +
				"[compatible] {bot} +https://example.com/bot",
			expected: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
				"KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 " +
				"[compatible] {bot} +https://example.com/bot",
			expectError: false,
		},
		{
			name:        "MaxLengthUserAgent",
			input:       strings.Repeat("a", 500),
			expected:    strings.Repeat("a", 500),
			expectError: false,
		},
		{
			name:        "EmptyString",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "NonString",
			input:       []uint64{123},
			expected:    "",
			expectError: true,
		},
		{
			name:        "InvalidCharacters",
			input:       "Mozilla/5.0\nBlabla",
			expected:    "",
			expectError: true,
		},
		{
			name:        "TooLong",
			input:       strings.Repeat("a", 501),
			expected:    "",
			expectError: true,
		},
		{
			name:        "NullBytes",
			input:       "Mozilla/5.0\x00",
			expected:    "",
			expectError: true,
		},
		{
			name:        "ControlCharacters",
			input:       "Mozilla/5.0\x01",
			expected:    "",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			userAgent, err := NewUserAgent(testCase.input)

			if testCase.expectError {
				if err == nil {
					t.Errorf("MissingExpectedError: [%v]", testCase.input)
				}
				return
			}

			if err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.input)
				return
			}

			if userAgent.String() != testCase.expected {
				t.Errorf("Expected '%s', got '%s'", testCase.expected, userAgent.String())
			}
		})
	}
}
