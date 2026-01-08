package tkValueObject

import "testing"

func TestNewX509Organization(t *testing.T) {
	t.Run("ValidOrganization", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509Organization
			expectError    bool
		}{
			{
				"Acme Corporation",
				X509Organization("Acme Corporation"),
				false,
			},
			{
				"Example Inc.",
				X509Organization("Example Inc."),
				false,
			},
			{
				"Tech-Solutions (Europe)",
				X509Organization("Tech-Solutions (Europe)"),
				false,
			},
			{
				"ABC123",
				X509Organization("ABC123"),
				false,
			},
			{
				"A",
				X509Organization("A"),
				false,
			},
			{
				"Société Française",
				X509Organization("Societe Francaise"),
				false,
			},
			{
				"Müller GmbH",
				X509Organization("Muller GmbH"),
				false,
			},
			{
				"Örnek Şirketi",
				X509Organization("Ornek Sirketi"),
				false,
			},
			{"", X509Organization(""), true},
			{
				"ThisOrganizationNameIsWayTooLongAndExceedsTheMaximumAllowedLengthOf255CharactersWhichIsTheStandardLimitForX509CertificateOrganizationFieldsAndThisStringWillContinueUntilItReachesTheRequiredLengthToFailTheValidationTestSoWeNeedToKeepTypingMoreAndMoreCharactersHere",
				X509Organization(""),
				true,
			},
			{"Invalid\nOrg", X509Organization(""), true},
			{"Invalid@Org", X509Organization(""), true},
			{"123", X509Organization("123"), false},
			{nil, X509Organization(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509Organization(testCase.inputValue)

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
			inputValue     X509Organization
			expectedOutput string
		}{
			{X509Organization("Acme Corporation"), "Acme Corporation"},
			{X509Organization("Example Inc."), "Example Inc."},
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
