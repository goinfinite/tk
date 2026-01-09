package tkValueObject

import (
	"os/exec"
	"strings"
	"testing"
)

func selfSignedCertificatePemGenerator(t *testing.T) string {
	t.Helper()

	rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
	rsaKeyOutput, err := rsaKeyCmd.Output()
	if err != nil {
		t.Skipf("OpenSslNotAvailable: %s", err.Error())
	}

	// We're unable to use tkInfra.Synthesizer here due to circular dependency.
	certCmd := exec.Command(
		"openssl", "req", "-new", "-x509", "-key", "/dev/stdin",
		"-days", "1", "-subj", "/CN=test",
	)
	certCmd.Stdin = strings.NewReader(string(rsaKeyOutput))
	certOutput, err := certCmd.Output()
	if err != nil {
		t.Fatalf("CertificateGenerationFailed: %s", err.Error())
	}

	return strings.TrimSpace(string(certOutput))
}

func TestNewX509EnvelopedCertificate(t *testing.T) {
	validCert := selfSignedCertificatePemGenerator(t)

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
			{
				"malicious payload" + validCert,
				X509EnvelopedCertificate(""),
				true,
			},
			{
				validCert + "malicious payload",
				X509EnvelopedCertificate(""),
				true,
			},
			{
				"<script>alert('xss')</script>\n" + validCert,
				X509EnvelopedCertificate(""),
				true,
			},
			{
				validCert + "\n; rm -rf /",
				X509EnvelopedCertificate(""),
				true,
			},
			{
				validCert + "\n" + validCert, // Multiple certificates
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
