package tkValueObject

import "testing"

func TestNewX509CertificatePolicy(test *testing.T) {
	test.Run("ValidCertificatePolicy", func(subtest *testing.T) {
		oid, _ := NewX509PolicyOID("2.5.29.32.0")
		policyName, _ := NewX509PolicyName("Extended Validation")
		qualifier1, _ := NewX509PolicyQualifier("cps")
		qualifier2, _ := NewX509PolicyQualifier("userNotice")

		testCaseStructs := []struct {
			policyIdentifier X509PolicyOID
			policyName       *X509PolicyName
			policyQualifiers []X509PolicyQualifier
		}{
			{oid, &policyName, []X509PolicyQualifier{qualifier1}},
			{oid, &policyName, []X509PolicyQualifier{qualifier1, qualifier2}},
			{oid, nil, []X509PolicyQualifier{qualifier1}},
			{oid, nil, nil},
		}

		for _, testCase := range testCaseStructs {
			result := NewX509CertificatePolicy(
				testCase.policyIdentifier,
				testCase.policyName,
				testCase.policyQualifiers,
			)

			if result.PolicyIdentifier != testCase.policyIdentifier {
				subtest.Errorf("UnexpectedPolicyIdentifier")
			}

			if testCase.policyName != nil &&
				result.PolicyName != testCase.policyName {
				subtest.Errorf("UnexpectedPolicyName")
			}

			if testCase.policyQualifiers != nil {
				if len(result.PolicyQualifiers) != len(testCase.policyQualifiers) {
					subtest.Errorf("UnexpectedPolicyQualifiersLength")
				}
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		oid, _ := NewX509PolicyOID("2.5.29.32.0")
		policyName, _ := NewX509PolicyName("Extended Validation")

		testCaseStructs := []struct {
			inputValue     X509CertificatePolicy
			expectedOutput string
		}{
			{
				X509CertificatePolicy{PolicyIdentifier: oid},
				"2.5.29.32.0",
			},
			{
				X509CertificatePolicy{PolicyIdentifier: oid, PolicyName: &policyName},
				"2.5.29.32.0 (Extended Validation)",
			},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
