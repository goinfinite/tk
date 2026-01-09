package tkValueObject

import (
	"strings"
	"testing"
)

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
			// Allowed punctuation characters
			{"Company.Inc", X509Organization("Company.Inc"), false},
			{"Company,LLC", X509Organization("Company,LLC"), false},
			{"Tech-Solutions", X509Organization("Tech-Solutions"), false},
			{"My_Company", X509Organization("My_Company"), false},
			{"Company(US)", X509Organization("Company(US)"), false},
			{"Smith&Sons", X509Organization("Smith&Sons"), false},
			{"Parent/Subsidiary", X509Organization("Parent/Subsidiary"), false},
			{"O'Reilly", X509Organization("O'Reilly"), false},
			{"McDonald's Inc.", X509Organization("McDonald's Inc."), false},
			{"L'Oreal", X509Organization("L'Oreal"), false},
			// Rejected characters
			{"Company@Domain", X509Organization(""), true},
			{"Company#1", X509Organization(""), true},
			{"Company$Corp", X509Organization(""), true},
			{"Company%Inc", X509Organization(""), true},
			{"Company*Star", X509Organization(""), true},
			{"Company+Plus", X509Organization(""), true},
			{"Company=Equal", X509Organization(""), true},
			{"Company[Bracket]", X509Organization(""), true},
			{"Company{Brace}", X509Organization(""), true},
			{"Company|Pipe", X509Organization(""), true},
			{"Company\\Backslash", X509Organization(""), true},
			{"Company:Colon", X509Organization(""), true},
			{"Company;Semicolon", X509Organization(""), true},
			{"Company\"Quote", X509Organization(""), true},
			{"Company<Less>", X509Organization(""), true},
			{"Company?Question", X509Organization(""), true},
			{"Company`Backtick", X509Organization(""), true},
			{"Company~Tilde", X509Organization(""), true},
			{"Company!Exclaim", X509Organization(""), true},
			{
				strings.Repeat("A", 255),
				X509Organization(strings.Repeat("A", 255)),
				false,
			},
			{
				strings.Repeat("A", 256),
				X509Organization(""),
				true,
			},
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

	t.Run("StringMethodReturnsNormalizedValue", func(t *testing.T) {
		normalizedOrg, err := NewX509Organization("Société Française")
		if err != nil {
			t.Fatalf("UnexpectedError: '%s'", err.Error())
		}

		expectedNormalizedString := "Societe Francaise"
		actualOutput := normalizedOrg.String()
		if actualOutput != expectedNormalizedString {
			t.Errorf(
				"StringMethodShouldReturnNormalizedValue: '%v' vs '%v'",
				actualOutput, expectedNormalizedString,
			)
		}
	})
}
