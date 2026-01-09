package tkValueObject

import "testing"

func TestNewX509PolicyName(t *testing.T) {
	t.Run("ValidPolicyName", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509PolicyName
			expectError    bool
		}{
			{
				"Extended Validation",
				X509PolicyName("Extended Validation"),
				false,
			},
			{"Domain Validation", X509PolicyName("Domain Validation"), false},
			{"Organization Validation", X509PolicyName("Organization Validation"), false},
			{"EV SSL", X509PolicyName("EV SSL"), false},
			{"A", X509PolicyName("A"), false},
			{"", X509PolicyName(""), true},
			{
				"ThisPolicyNameIsWayTooLongAndExceedsTheMaximumAllowedLengthOf128CharactersWhichIsTheStandardLimitForX509CertificatePolicyNamesAndContinues",
				X509PolicyName(""),
				true,
			},
			{"123", X509PolicyName("123"), false},
			{nil, X509PolicyName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509PolicyName(testCase.inputValue)

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
			inputValue     X509PolicyName
			expectedOutput string
		}{
			{
				X509PolicyName("Extended Validation"),
				"Extended Validation",
			},
			{X509PolicyName("EV SSL"), "EV SSL"},
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
