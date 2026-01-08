package tkValueObject

import "testing"

func TestNewX509SubjectName(t *testing.T) {
	t.Run("ValidSubjectName", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509SubjectName
			expectError    bool
		}{
			{"example.com", X509SubjectName("example.com"), false},
			{"*.example.com", X509SubjectName("*.example.com"), false},
			{"subdomain.example.com", X509SubjectName("subdomain.example.com"), false},
			{"example-site.com", X509SubjectName("example-site.com"), false},
			{"John Doe", X509SubjectName("John Doe"), false},
			{"server123", X509SubjectName("server123"), false},
			{"localhost", X509SubjectName("localhost"), false},
			{"A", X509SubjectName("A"), false},
			{"", X509SubjectName(""), true},
			{
				"this-is-a-very-long-domain-name-that-exceeds-the-maximum-allowed-length-of-253-characters-which-is-the-standard-limit-for-dns-names-and-x509-certificate-common-names-this-string-continues-until-it-reaches-the-required-length-to-fail-the-validation-test-so-we-need-to-keep-typing-more",
				X509SubjectName(""),
				true,
			},
			{"invalid@domain.com", X509SubjectName(""), true},
			{"invalid\ndomain", X509SubjectName(""), true},
			{"invalid&domain", X509SubjectName(""), true},
			{"123", X509SubjectName("123"), false},
			{nil, X509SubjectName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509SubjectName(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf(
					"MissingExpectedError: [%v]",
					testCase.inputValue,
				)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf(
					"UnexpectedError: '%s' [%v]",
					err.Error(), testCase.inputValue,
				)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509SubjectName
			expectedOutput string
		}{
			{X509SubjectName("example.com"), "example.com"},
			{X509SubjectName("*.example.com"), "*.example.com"},
			{X509SubjectName("John Doe"), "John Doe"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
