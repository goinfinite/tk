package tkValueObject

import "testing"

func TestNewX509EnvelopedCertificate(t *testing.T) {
	validCert := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKL0UG+mRkSvMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcwODE0MDk1NzU3WhcNMTgwODE0MDk1NzU3WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAyWm
-----END CERTIFICATE-----`

	t.Run("ValidEnvelopedCertificate", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput X509EnvelopedCertificate
			expectError    bool
		}{
			{
				validCert,
				X509EnvelopedCertificate(validCert),
				false,
			},
			{"", X509EnvelopedCertificate(""), true},
			{"Invalid certificate", X509EnvelopedCertificate(""), true},
			{
				"-----BEGIN CERTIFICATE-----\nshort\n-----END CERTIFICATE-----",
				X509EnvelopedCertificate(""),
				true,
			},
			{
				"-----BEGIN CERTIFICATE-----\n" + string(make([]byte, 100)),
				X509EnvelopedCertificate(""),
				true,
			},
			{123, X509EnvelopedCertificate(""), true},
			{nil, X509EnvelopedCertificate(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewX509EnvelopedCertificate(
				testCase.inputValue,
			)

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     X509EnvelopedCertificate
			expectedOutput string
		}{
			{X509EnvelopedCertificate(validCert), validCert},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()

			if actualOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})
}
