package tkInfra

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestTrustedCidrsReader(t *testing.T) {
	testCaseStructs := []struct {
		description    string
		cidrsEnvValue  string
		ipsEnvValue    string
		expectedOutput []tkValueObject.CidrBlock
	}{
		{
			description:    "BothEnvVarsEmpty",
			cidrsEnvValue:  "",
			ipsEnvValue:    "",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			description:   "SingleCidrViaTrustedCidrs",
			cidrsEnvValue: "192.168.1.0/24",
			ipsEnvValue:   "",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
			},
		},
		{
			description:   "MultipleCidrsViaTrustedCidrs",
			cidrsEnvValue: "192.168.1.0/24,10.0.0.0/8",
			ipsEnvValue:   "",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			description:   "InvalidCidrSkippedViaTrustedCidrs",
			cidrsEnvValue: "192.168.1.0/24,notacidr,10.0.0.0/8",
			ipsEnvValue:   "",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			description:    "AllInvalidCidrsReturnEmpty",
			cidrsEnvValue:  "notacidr,alsoinvalid",
			ipsEnvValue:    "",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			description:   "EmptyCommaEntriesSkippedViaTrustedCidrs",
			cidrsEnvValue: "192.168.1.0/24,,10.0.0.0/8",
			ipsEnvValue:   "",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			description:   "WhitespaceAroundCidrsTrimmed",
			cidrsEnvValue: " 192.168.1.0/24 , 10.0.0.0/8 ",
			ipsEnvValue:   "",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.0/24"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			description:    "InvalidIpOctetInCidrSkipped",
			cidrsEnvValue:  "999.999.999.999/24",
			ipsEnvValue:    "",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			description:    "InvalidCidrPrefixLengthSkipped",
			cidrsEnvValue:  "192.168.1.0/999",
			ipsEnvValue:    "",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			description:    "OverlongCidrStringSkipped",
			cidrsEnvValue:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/24",
			ipsEnvValue:    "",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			description:   "SingleIpv4ViaTrustedIpsBecomesSlash32",
			cidrsEnvValue: "",
			ipsEnvValue:   "192.168.1.1",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.1/32"),
			},
		},
		{
			description:   "MultipleIpv4ViaTrustedIpsBecomesSlash32",
			cidrsEnvValue: "",
			ipsEnvValue:   "192.168.1.1,10.0.0.1",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.1/32"),
				tkValueObject.CidrBlock("10.0.0.1/32"),
			},
		},
		{
			description:   "InvalidIpViaTrustedIpsSkipped",
			cidrsEnvValue: "",
			ipsEnvValue:   "192.168.1.1,invalid,10.0.0.1",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.1/32"),
				tkValueObject.CidrBlock("10.0.0.1/32"),
			},
		},
		{
			description:    "AllInvalidIpsReturnEmpty",
			cidrsEnvValue:  "",
			ipsEnvValue:    "invalid,alsoinvalid",
			expectedOutput: []tkValueObject.CidrBlock{},
		},
		{
			description:   "Ipv6ViaTrustedIpsBecomesSlash128",
			cidrsEnvValue: "",
			ipsEnvValue:   "::1,2001:db8::1",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("::1/128"),
				tkValueObject.CidrBlock("2001:db8::1/128"),
			},
		},
		{
			description:   "MixedIpAndCidrFromBothEnvVars",
			cidrsEnvValue: "10.0.0.0/8",
			ipsEnvValue:   "192.168.1.1",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("192.168.1.1/32"),
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
		{
			description:   "CidrViaTrustedIpsIsAccepted",
			cidrsEnvValue: "",
			ipsEnvValue:   "10.0.0.0/8",
			expectedOutput: []tkValueObject.CidrBlock{
				tkValueObject.CidrBlock("10.0.0.0/8"),
			},
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			t.Setenv(TrustedCidrsEnvVarName, testCase.cidrsEnvValue)
			t.Setenv(TrustedIpsEnvVarName, testCase.ipsEnvValue)

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
