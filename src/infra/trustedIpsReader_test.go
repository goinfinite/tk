package tkInfra

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestTrustedIpsReader(t *testing.T) {
	testCaseStructs := []struct {
		envValue       string
		expectedOutput []tkValueObject.IpAddress
	}{
		{
			envValue:       "",
			expectedOutput: []tkValueObject.IpAddress{},
		},
		{
			envValue: "192.168.1.1",
			expectedOutput: []tkValueObject.IpAddress{
				tkValueObject.IpAddress("192.168.1.1"),
			},
		},
		{
			envValue: "192.168.1.1,10.0.0.1",
			expectedOutput: []tkValueObject.IpAddress{
				tkValueObject.IpAddress("192.168.1.1"),
				tkValueObject.IpAddress("10.0.0.1"),
			},
		},
		{
			envValue: "192.168.1.1,invalid,10.0.0.1",
			expectedOutput: []tkValueObject.IpAddress{
				tkValueObject.IpAddress("192.168.1.1"),
				tkValueObject.IpAddress("10.0.0.1"),
			},
		},
		{
			envValue:       "invalid,alsoinvalid",
			expectedOutput: []tkValueObject.IpAddress{},
		},
		{
			envValue: "::1,2001:db8::1",
			expectedOutput: []tkValueObject.IpAddress{
				tkValueObject.IpAddress("::1"),
				tkValueObject.IpAddress("2001:db8::1"),
			},
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run("EnvValue_"+testCase.envValue, func(t *testing.T) {
			t.Setenv(TrustedIpsEnvVarName, testCase.envValue)

			actualOutput, err := TrustedIpsReader()
			if err != nil {
				t.Errorf("UnexpectedError: '%s'", err.Error())
				return
			}

			if len(actualOutput) != len(testCase.expectedOutput) {
				t.Errorf(
					"OutputLengthMismatch: expected=%d, actual=%d",
					len(testCase.expectedOutput), len(actualOutput),
				)
				return
			}

			for expectedIpIndex, expectedIp := range testCase.expectedOutput {
				if actualOutput[expectedIpIndex] != expectedIp {
					t.Errorf(
						"IpMismatchAtIndex%d: expected='%s', actual='%s'",
						expectedIpIndex, expectedIp.String(), actualOutput[expectedIpIndex].String(),
					)
				}
			}
		})
	}
}
