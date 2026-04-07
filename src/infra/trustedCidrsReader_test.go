package tkInfra

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestTrustedCidrsReader(t *testing.T) {
	testCaseStructs := []struct {
		envValue       string
		expectedOutput []tkValueObject.CidrBlock
	}{
		{
			envValue:       "",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			envValue: "192.168.1.0/24",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
			},
		},
		{
			envValue: "192.168.1.0/24,10.0.0.0/8",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			envValue: "192.168.1.0/24,notacidr,10.0.0.0/8",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			envValue:       "notacidr,alsoinvalid",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			envValue: "192.168.1.0/24,,10.0.0.0/8",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			envValue: " 192.168.1.0/24 , 10.0.0.0/8 ",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			envValue:       "999.999.999.999/24",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			envValue:       "192.168.1.0/999",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			envValue:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/24",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run("EnvValue_"+testCase.envValue, func(t *testing.T) {
			t.Setenv(TrustedCidrsEnvVarName, testCase.envValue)

			actualCidrBlocks, err := TrustedCidrsReader()
			if err != nil {
				t.Errorf("UnexpectedError: '%s'", err.Error())
				return
			}

			if len(actualCidrBlocks) != len(testCase.expectedOutput) {
				t.Errorf(
					"OutputLengthMismatch: expected=%d, actual=%d",
					len(testCase.expectedOutput), len(actualCidrBlocks),
				)
				return
			}

			for expectedCidrIndex, expectedCidr := range testCase.expectedOutput {
				if actualCidrBlocks[expectedCidrIndex] != expectedCidr {
					t.Errorf(
						"CidrMismatchAtIndex%d: expected='%s', actual='%s'",
						expectedCidrIndex,
						expectedCidr.String(),
						actualCidrBlocks[expectedCidrIndex].String(),
					)
				}
			}
		})
	}
}
