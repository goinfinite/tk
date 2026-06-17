package tkValueObject

import "testing"

func TestNewHoneypotMaxEntries(t *testing.T) {
	testCaseStructs := []struct {
		name          string
		inputValue    any
		expectedValue int
	}{
		{"DefaultValueIs5000", nil, 5000},
		{"StringValueParsed", "1234", 1234},
		{"AtFloorNoClamp", "100", 100},
		{"AtCeilingNoClamp", "50000", 50000},
		{"BelowFloorClampsToDefault", "50", 5000},
		{"AboveCeilingClamps", "75000", 50000},
		{"NegativeDefaults", "-100", 5000},
		{"NonNumericDefaults", "abc", 5000},
		{"EmptyStringDefaults", "", 5000},
		{"ExtremeValueClampsToCeiling", "999999999", 50000},
		{"IntInputParsed", int(3000), 3000},
		{"Int64InputParsed", int64(2500), 2500},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput, conversionErr := NewHoneypotMaxEntries(
				testCase.inputValue,
			)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue)
			}
			if actualOutput.Int() != testCase.expectedValue {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput.Int(), testCase.expectedValue, testCase.inputValue)
			}
		})
	}
}
