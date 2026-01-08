package entity

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewX509CertificateFromEnvelopedCertificate(t *testing.T) {
	validSelfSignedCertPEM := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKL0UG+mRmKjMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMjQwMTAxMTIwMDAwWhcNMjUwMTAxMTIwMDAwWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAu1SU1LfVLPHCozMxH2Mo4lgOEePzNm0tRgeLezV6ffAt0gunVTLw7onL
RnrQ0/IzPqkCDKKSWnrs3DstcC7bmcMiLGNGGbqBJjE+7h3cVRCsC1WV5DfFJJrF
pF8qhVzWYdHPCj9zzVm/eDMnwbpH+eVBiZJTiGiTUCiCCKdI8hB5vT1xHQS3MPRX
7TYwLPQhWPtUbSPJpgFYJD8B9sxqFxWJ6S9sE4eRd4k1KDCNq8J6MJQyKqNXLuYf
f0xN0J2rSBLVrD5cqQXYxUjPDM5QGDUjLgCqPAC5LPGXQ7ZqHIVKCCDaJPYL1qFb
vTvVMWDJhBXBdKLJZn6h9OEcQC+LxQIDAQABo1AwTjAdBgNVHQ4EFgQU5Z1ZMIJd
jzbfpCL5aw2C0dYZPXowHwYDVR0jBBgwFoAU5Z1ZMIJdjzbfpCL5aw2C0dYZPXow
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAk2fc1E73/kmOlm/wrL7V
gVlt9vL6qxJlJiYn2Fh+mCNvk8P88PYoQCQYvT0zqhL6pGmLfLCQCVMGJCGUqKjQ
0wHOCCEVHh3xXlLQfFH/sBBKJPqE3w9qPpjqNFvvXBDXvEjYaLqMXlI4yqJUxvQI
gghvhCKZELG9KvAQGGBqpL1lqZvKXNuDnXGVVKGYqbFQXTxW7qYvwQXD5pqvZxqJ
UHzrJkL2kJ6N3YVxCVXZjcMZ9rG6x3kGFCfnXB/KQmNOLCJE5r7AHMHZqVKCWCJZ
vQpqBvM3cJJNQNqQSvCDUJJN0qP0V9vXFXBBhGSBVJ0wQZRjVNGJLMOHVHKJmEOE
Og==
-----END CERTIFICATE-----`

	malformedCertDataPEM := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKL0UG+mRmKjMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMjQwMTAxMTIwMDAwWhcNMjUwMTAxMTIwMDAwWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
InvalidDataHere!!!
-----END CERTIFICATE-----`

	testCaseStructs := []struct {
		inputCertificate string
		expectError      bool
		expectedError    string
	}{
		{validSelfSignedCertPEM, false, ""},
		{malformedCertDataPEM, true, "DecodePEMFailed"},
	}

	for _, testCase := range testCaseStructs {
		envelopedCert, err := tkValueObject.NewX509EnvelopedCertificate(
			testCase.inputCertificate,
		)
		if err != nil {
			t.Fatalf("FailedToCreateEnvelopedCertificate: %s", err.Error())
		}

		x509CertEntity, err := NewX509CertificateFromEnvelopedCertificate(
			envelopedCert,
		)

		if testCase.expectError && err == nil {
			t.Fatalf("MissingExpectedError: %s", testCase.expectedError)
		}

		if testCase.expectError && err != nil {
			if err.Error() != testCase.expectedError {
				t.Fatalf(
					"UnexpectedError: got '%s', expected '%s'",
					err.Error(), testCase.expectedError,
				)
			}
			continue
		}

		if !testCase.expectError && err != nil {
			t.Fatalf("UnexpectedError: %s", err.Error())
		}

		if x509CertEntity.SerialNumber.String() == "" {
			t.Errorf("SerialNumberEmpty")
		}

		expectedVersion := uint8(3)
		if x509CertEntity.VersionNumber.Uint8() != expectedVersion {
			t.Errorf(
				"UnexpectedVersionNumber: got %d, expected %d",
				x509CertEntity.VersionNumber.Uint8(), expectedVersion,
			)
		}

		if x509CertEntity.SubjectDistinguishedName == nil {
			t.Errorf("SubjectDistinguishedNameIsNil")
		}

		if x509CertEntity.SubjectDistinguishedName != nil {
			subjectDNString := x509CertEntity.SubjectDistinguishedName.String()
			expectedSubjectDN := "O=Internet Widgits Pty Ltd, ST=Some-State, C=AU"
			if subjectDNString != expectedSubjectDN {
				t.Errorf(
					"UnexpectedSubjectDN: got '%s', expected '%s'",
					subjectDNString, expectedSubjectDN,
				)
			}
		}

		if x509CertEntity.IssuerDistinguishedName == nil {
			t.Errorf("IssuerDistinguishedNameIsNil")
		}

		if x509CertEntity.ValidityNotBefore.Int64() <= 0 {
			t.Errorf("InvalidValidityNotBefore")
		}

		if x509CertEntity.ValidityNotAfter.Int64() <= 0 {
			t.Errorf("InvalidValidityNotAfter")
		}

		expectedPublicKeyAlgorithm := "RSA"
		if x509CertEntity.PublicKeyAlgorithm.String() != expectedPublicKeyAlgorithm {
			t.Errorf(
				"UnexpectedPublicKeyAlgorithm: got %s, expected %s",
				x509CertEntity.PublicKeyAlgorithm.String(), expectedPublicKeyAlgorithm,
			)
		}

		expectedPublicKeySize := uint16(2048)
		if x509CertEntity.PublicKeySize.Uint16() != expectedPublicKeySize {
			t.Errorf(
				"UnexpectedPublicKeySize: got %d, expected %d",
				x509CertEntity.PublicKeySize.Uint16(), expectedPublicKeySize,
			)
		}

		if x509CertEntity.PublicKeyValue.String() == "" {
			t.Errorf("PublicKeyValueEmpty")
		}

		expectedSignatureAlgorithm := "SHA256WithRSA"
		actualSignatureAlgorithm := x509CertEntity.SignatureAlgorithm.String()
		if actualSignatureAlgorithm != expectedSignatureAlgorithm {
			t.Errorf(
				"UnexpectedSignatureAlgorithm: got %s, expected %s",
				actualSignatureAlgorithm, expectedSignatureAlgorithm,
			)
		}

		if x509CertEntity.SignatureValue == nil {
			t.Errorf("SignatureValueIsNil")
		}

		if x509CertEntity.FingerprintSHA256.String() == "" {
			t.Errorf("FingerprintSHA256Empty")
		}

		if x509CertEntity.FingerprintSHA1.String() == "" {
			t.Errorf("FingerprintSHA1Empty")
		}

		if x509CertEntity.BasicConstraints == nil {
			t.Errorf("BasicConstraintsIsNil")
		}

		if x509CertEntity.BasicConstraints != nil {
			certIsAuthority := x509CertEntity.BasicConstraints.IsAuthority
			if !certIsAuthority {
				t.Errorf("CertificateShouldBeCA")
			}
		}

		envelopedCertString := x509CertEntity.EnvelopedCertificate.String()
		if envelopedCertString != testCase.inputCertificate {
			t.Errorf("EnvelopedCertificateMismatch")
		}
	}
}
