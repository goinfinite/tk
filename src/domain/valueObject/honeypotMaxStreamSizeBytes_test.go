package tkValueObject

import "testing"

func TestNewHoneypotMaxStreamSizeBytes(t *testing.T) {
	testCaseStructs := []struct {
		name          string
		inputValue    any
		expectedValue int64
	}{
		{"DefaultValueIs20MB", nil, 20971520},
		{"StringValueParsed", "31457280", 31457280},
		{"AtFloorNoClamp", "5242880", 5242880},
		{"BelowFloorClampsToDefault", "1048576", 20971520},
		{"NegativeDefaults", "-100", 20971520},
		{"NonNumericDefaults", "abc", 20971520},
		{"EmptyStringDefaults", "", 20971520},
		{"ExtremeValueAccepted", "10737418240", 10737418240},
		{"IntInputParsed", int(10485760), 10485760},
		{"Int64InputParsed", int64(15728640), 15728640},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput, conversionErr := NewHoneypotMaxStreamSizeBytes(
				testCase.inputValue,
			)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]",
					conversionErr.Error(), testCase.inputValue)
			}
			if actualOutput.Int64() != testCase.expectedValue {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput.Int64(), testCase.expectedValue, testCase.inputValue)
			}
		})
	}
}
