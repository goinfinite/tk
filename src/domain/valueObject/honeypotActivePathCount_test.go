package tkValueObject

import "testing"

func TestNewHoneypotActivePathCount(t *testing.T) {
	testCaseStructs := []struct {
		name          string
		inputValue    any
		poolCeiling   int
		expectedValue int
	}{
		{"NilInputDefaults", nil, 100, 30},
		{"StringValueParsed", "45", 100, 45},
		{"CeilingClamp", "120", 100, 100},
		{"AtFloorNoClamp", "30", 100, 30},
		{"ZeroCeilingProducesFloor", "45", 0, 45},
		{"BelowFloorClampsToDefault", "10", 100, 30},
		{"NegativeDefaults", "-5", 100, 30},
		{"NonNumericDefaults", "abc", 100, 30},
		{"EmptyStringDefaults", "", 100, 30},
		{"NonStringTypeDefaults", 3.14, 100, 30},
		{"ExtremeValueClampsToCeiling", "999999", 100, 100},
		{"IntInputParsed", int(50), 100, 50},
		{"Int64InputParsed", int64(60), 100, 60},
		{"DefaultValueIs30", nil, 100, 30},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput, conversionErr := NewHoneypotActivePathCount(
				testCase.inputValue, testCase.poolCeiling,
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
