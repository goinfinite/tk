package tkValueObject

import "testing"

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
