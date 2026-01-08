package tkValueObject

type X509CertificatePolicy struct {
	PolicyIdentifier X509PolicyOID         `json:"policyIdentifier"`
	PolicyName       *X509PolicyName       `json:"policyName"`
	PolicyQualifiers []X509PolicyQualifier `json:"policyQualifiers"`
}

func NewX509CertificatePolicy(
	policyIdentifier X509PolicyOID,
	policyName *X509PolicyName,
	policyQualifiers []X509PolicyQualifier,
) X509CertificatePolicy {
	return X509CertificatePolicy{
		PolicyIdentifier: policyIdentifier,
		PolicyName:       policyName,
		PolicyQualifiers: policyQualifiers,
	}
}
