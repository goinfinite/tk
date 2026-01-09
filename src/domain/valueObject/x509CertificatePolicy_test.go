package tkValueObject

import "testing"

func TestNewX509CertificatePolicy(t *testing.T) {
	t.Run("ValidCertificatePolicy", func(t *testing.T) {
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
				t.Errorf("UnexpectedPolicyIdentifier")
			}

			if testCase.policyName != nil &&
				result.PolicyName != testCase.policyName {
				t.Errorf("UnexpectedPolicyName")
			}

			if testCase.policyQualifiers != nil {
				if len(result.PolicyQualifiers) != len(testCase.policyQualifiers) {
					t.Errorf("UnexpectedPolicyQualifiersLength")
				}
			}
		}
	})
}
