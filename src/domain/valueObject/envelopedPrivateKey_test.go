package tkValueObject

import "testing"

func TestNewEnvelopedPrivateKey(t *testing.T) {
	validPrivateKey := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC5Wm1Z9ZJvLQ3
LQ5Rk5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
-----END PRIVATE KEY-----`

	validRsaPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAyWmAk8QJ9XJvLQ3LQ5Rk5J2k5J2k5J2k5J2k5J2k5J2k5J2k
5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
-----END RSA PRIVATE KEY-----`

	validEcPrivateKey := `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIJKL0UG+mRkSvMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNVBAYTAkFVMRMw
EQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0e
SBMdGQwHhcNMTcwODE0MDk1NzU3WhcNMTgwODE0MDk1NzU3WjBFMQswCQYDVQQGEwJB
VTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0cyBQ
dHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyWmAk8QJ9XJvLQ3
LQ5Rk5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
-----END EC PRIVATE KEY-----`

	validEncryptedPrivateKey := `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIFHDBOBgkqhkiG9w0BBQ0wQTApBgkqhkiG9w0BBQwwHAQIxJmAk8QJ9XJCAggA
MAwGCCqGSIb3DQIJBQAwFAYIKoZIhvcNAwcECC5Wm1Z9ZJvLQ3LQ5Rk5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J2k5J
-----END ENCRYPTED PRIVATE KEY-----`

	t.Run("ValidEnvelopedPrivateKey", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput EnvelopedPrivateKey
			expectError    bool
		}{
			{
				validPrivateKey,
				EnvelopedPrivateKey(validPrivateKey),
				false,
			},
			{
				validRsaPrivateKey,
				EnvelopedPrivateKey(validRsaPrivateKey),
				false,
			},
			{
				validEcPrivateKey,
				EnvelopedPrivateKey(validEcPrivateKey),
				false,
			},
			{
				validEncryptedPrivateKey,
				EnvelopedPrivateKey(validEncryptedPrivateKey),
				false,
			},
			{"", EnvelopedPrivateKey(""), true},
			{"Invalid private key", EnvelopedPrivateKey(""), true},
			{
				"-----BEGIN PRIVATE KEY-----\nshort\n-----END PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY-----\n" + string(make([]byte, 100)),
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN CERTIFICATE-----\n" + validPrivateKey[28:],
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"Some random text -----BEGIN PRIVATE KEY----- content -----END PRIVATE KEY----- more text",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY----- content -----END PRIVATE KEY-----\n-----BEGIN RSA PRIVATE KEY----- content -----END RSA PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----END PRIVATE KEY----- content -----BEGIN PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY-----\ncontent\n-----END CERTIFICATE-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN RSA PRIVATE KEY----- content -----END PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY----- content -----END RSA PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"  -----BEGIN PRIVATE KEY----- content -----END PRIVATE KEY-----  ",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY----- \n content \n -----END PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"prefix\n-----BEGIN PRIVATE KEY-----\n" + string(make([]byte, 50)) + "\n-----END PRIVATE KEY-----\nsuffix",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY-----\n" + string(make([]byte, 50)) + "\n-----END PRIVATE KEY-----\n-----BEGIN PRIVATE KEY-----\n" + string(make([]byte, 50)) + "\n-----END PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"malformed -----BEGIN PRIVATE KEY----- content -----END PRIVATE KEY----- injection",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"-----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY-----\nEXTRA_DATA",
				EnvelopedPrivateKey(""),
				true,
			},
			{
				"EXTRA_DATA\n-----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY-----",
				EnvelopedPrivateKey(""),
				true,
			},
			{123, EnvelopedPrivateKey(""), true},
			{nil, EnvelopedPrivateKey(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, err := NewEnvelopedPrivateKey(
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
			inputValue     EnvelopedPrivateKey
			expectedOutput string
		}{
			{EnvelopedPrivateKey(validPrivateKey), validPrivateKey},
			{EnvelopedPrivateKey(validRsaPrivateKey), validRsaPrivateKey},
			{EnvelopedPrivateKey(validEcPrivateKey), validEcPrivateKey},
			{EnvelopedPrivateKey(validEncryptedPrivateKey), validEncryptedPrivateKey},
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
