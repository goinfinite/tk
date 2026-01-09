package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewX509OrganizationalUnit(t *testing.T) {
	t.Run("ValidOrganizationalUnit", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509OrganizationalUnit
			expectError    bool
		}{
			{
				"Engineering Department",
				X509OrganizationalUnit("Engineering Department"),
				false,
			},
			{
				"IT",
				X509OrganizationalUnit("IT"),
				false,
			},
			{
				"Research & Development",
				X509OrganizationalUnit("Research & Development"),
				false,
			},
			{
				"Sales (North America)",
				X509OrganizationalUnit("Sales (North America)"),
				false,
			},
			{
				"A",
				X509OrganizationalUnit("A"),
				false,
			},
			{
				"Département Informatique",
				X509OrganizationalUnit("Departement Informatique"),
				false,
			},
			{
				"Fürschung",
				X509OrganizationalUnit("Furschung"),
				false,
			},
			{
				"Ñoño Unit",
				X509OrganizationalUnit("Nono Unit"),
				false,
			},
			{"", X509OrganizationalUnit(""), true},
			{
				"ThisOrganizationalUnitNameIsWayTooLongAndExceedsTheMaximumAllowedLengthOf255CharactersWhichIsTheStandardLimitForX509CertificateOrganizationalUnitFieldsAndThisStringWillContinueUntilItReachesTheRequiredLengthToFailTheValidationTestForSureAndEvenMoreTextToMakeItLonger",
				X509OrganizationalUnit(""),
				true,
			},
			{
				"Invalid\nUnit",
				X509OrganizationalUnit(""),
				true,
			},
			{"Invalid@Unit", X509OrganizationalUnit(""), true},
			{"123", X509OrganizationalUnit("123"), false},
			{nil, X509OrganizationalUnit(""), true},
			// Allowed punctuation characters
			{"Dept.Engineering", X509OrganizationalUnit("Dept.Engineering"), false},
			{"Group,Alpha", X509OrganizationalUnit("Group,Alpha"), false},
			{"Sub-Unit", X509OrganizationalUnit("Sub-Unit"), false},
			{"Team_Red", X509OrganizationalUnit("Team_Red"), false},
			{"Division(East)", X509OrganizationalUnit("Division(East)"), false},
			{"Sales&Marketing", X509OrganizationalUnit("Sales&Marketing"), false},
			{"Parent/Child", X509OrganizationalUnit("Parent/Child"), false},
			{"Director's Office", X509OrganizationalUnit("Director's Office"), false},
			{"Men's Department", X509OrganizationalUnit("Men's Department"), false},
			{"L'Equipe", X509OrganizationalUnit("L'Equipe"), false},
			// Rejected characters
			{"Unit@Domain", X509OrganizationalUnit(""), true},
			{"Unit#1", X509OrganizationalUnit(""), true},
			{"Unit$Corp", X509OrganizationalUnit(""), true},
			{"Unit%Inc", X509OrganizationalUnit(""), true},
			{"Unit*Star", X509OrganizationalUnit(""), true},
			{"Unit+Plus", X509OrganizationalUnit(""), true},
			{"Unit=Equal", X509OrganizationalUnit(""), true},
			{"Unit[Bracket]", X509OrganizationalUnit(""), true},
			{"Unit{Brace}", X509OrganizationalUnit(""), true},
			{"Unit|Pipe", X509OrganizationalUnit(""), true},
			{"Unit\\Backslash", X509OrganizationalUnit(""), true},
			{"Unit:Colon", X509OrganizationalUnit(""), true},
			{"Unit;Semicolon", X509OrganizationalUnit(""), true},
			{"Unit\"Quote", X509OrganizationalUnit(""), true},
			{"Unit<Less>", X509OrganizationalUnit(""), true},
			{"Unit?Question", X509OrganizationalUnit(""), true},
			{"Unit`Backtick", X509OrganizationalUnit(""), true},
			{"Unit~Tilde", X509OrganizationalUnit(""), true},
			{"Unit!Exclaim", X509OrganizationalUnit(""), true},
			{
				strings.Repeat("A", 255),
				X509OrganizationalUnit(strings.Repeat("A", 255)),
				false,
			},
			{
				strings.Repeat("A", 256),
				X509OrganizationalUnit(""),
				true,
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509OrganizationalUnit(
				testCase.inputValue,
			)

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
			inputValue     X509OrganizationalUnit
			expectedOutput string
		}{
			{
				X509OrganizationalUnit("Engineering Department"),
				"Engineering Department",
			},
			{X509OrganizationalUnit("IT"), "IT"},
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
