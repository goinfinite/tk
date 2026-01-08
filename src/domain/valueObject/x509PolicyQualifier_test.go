package tkValueObject

import "testing"

func TestNewX509PolicyQualifier(t *testing.T) {
	t.Run("ValidPolicyQualifier", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PolicyQualifier
			expectError    bool
		}{
			{"cps", X509PolicyQualifierCPS, false},
			{"userNotice", X509PolicyQualifierUserNotice, false},
			{"", X509PolicyQualifier(""), true},
			{"invalidQualifier", X509PolicyQualifier(""), true},
			{"CPS", X509PolicyQualifier(""), true},
			{"UserNotice", X509PolicyQualifier(""), true},
			{123, X509PolicyQualifier(""), true},
			{nil, X509PolicyQualifier(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PolicyQualifier(testCase.inputValue)

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
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
			inputValue     X509PolicyQualifier
			expectedOutput string
		}{
			{X509PolicyQualifierCPS, "cps"},
			{X509PolicyQualifierUserNotice, "userNotice"},
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
