package tkValueObject

import (
	"os/exec"
	"strings"
	"testing"
)

func privateKeyPemGenerator(
	t *testing.T, algorithm PrivateKeyAlgorithm,
) string {
	t.Helper()

	switch algorithm {
	case PrivateKeyAlgorithmRSA:
		rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
		rsaKeyOutput, err := rsaKeyCmd.Output()
		if err != nil {
			t.Skipf("OpenSslNotAvailable: %s", err.Error())
		}
		return strings.TrimSpace(string(rsaKeyOutput))

	case "PKCS8":
		rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
		rsaKeyOutput, err := rsaKeyCmd.Output()
		if err != nil {
			t.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		pkcs8Cmd := exec.Command(
			"openssl", "pkcs8", "-topk8", "-inform", "PEM", "-outform", "PEM",
			"-nocrypt",
		)
		pkcs8Cmd.Stdin = strings.NewReader(string(rsaKeyOutput))
		pkcs8Output, err := pkcs8Cmd.Output()
		if err != nil {
			t.Skipf("PKCS8ConversionFailed: %s", err.Error())
		}
		return strings.TrimSpace(string(pkcs8Output))

	case PrivateKeyAlgorithmECDSA:
		ecKeyCmd := exec.Command("openssl", "ecparam", "-genkey", "-name", "prime256v1")
		ecKeyOutput, err := ecKeyCmd.Output()
		if err != nil {
			t.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		ecKeyOutputStr := string(ecKeyOutput)
		keyStart := strings.Index(ecKeyOutputStr, "-----BEGIN EC PRIVATE KEY-----")
		if keyStart == -1 {
			t.Fatalf("ECPrivateKeyNotFoundInOutput")
		}
		return strings.TrimSpace(ecKeyOutputStr[keyStart:])

	case "Encrypted":
		rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
		rsaKeyOutput, err := rsaKeyCmd.Output()
		if err != nil {
			t.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		encryptCmd := exec.Command(
			"openssl", "pkcs8", "-topk8", "-inform", "PEM", "-outform", "PEM",
			"-v2", "aes-256-cbc", "-passout", "pass:testpassword",
		)
		encryptCmd.Stdin = strings.NewReader(string(rsaKeyOutput))
		encryptOutput, err := encryptCmd.Output()
		if err != nil {
			t.Skipf("PrivateKeyEncryptionFailed: %s", err.Error())
		}
		return strings.TrimSpace(string(encryptOutput))

	case PrivateKeyAlgorithmDSA:
		dsaCmd := exec.Command("openssl", "dsaparam", "-genkey", "2048")
		dsaOutput, err := dsaCmd.Output()
		if err != nil {
			t.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		dsaOutputStr := string(dsaOutput)
		keyStart := strings.Index(dsaOutputStr, "-----BEGIN PRIVATE KEY-----")
		if keyStart == -1 {
			keyStart = strings.Index(dsaOutputStr, "-----BEGIN DSA PRIVATE KEY-----")
		}
		if keyStart == -1 {
			t.Fatalf("DSAPrivateKeyNotFoundInOutput")
		}
		return strings.TrimSpace(dsaOutputStr[keyStart:])

	default:
		t.Fatalf("UnsupportedPrivateKeyAlgorithm: %s", algorithm)
		return ""
	}
}

func TestNewEnvelopedPrivateKey(t *testing.T) {
	validPrivateKey := privateKeyPemGenerator(t, "PKCS8")
	validRsaPrivateKey := privateKeyPemGenerator(t, PrivateKeyAlgorithmRSA)
	validEcPrivateKey := privateKeyPemGenerator(t, PrivateKeyAlgorithmECDSA)
	validEncryptedPrivateKey := privateKeyPemGenerator(t, "Encrypted")
	validDsaPrivateKey := privateKeyPemGenerator(t, PrivateKeyAlgorithmDSA)

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
			{
				validDsaPrivateKey,
				EnvelopedPrivateKey(validDsaPrivateKey),
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
			{EnvelopedPrivateKey(validDsaPrivateKey), validDsaPrivateKey},
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

	t.Run("ShortInputErrorMessage", func(t *testing.T) {
		shortInput := "-----BEGIN PRIVATE KEY-----\nshort\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(shortInput)
		if err == nil {
			t.Fatalf("MissingExpectedError: short input should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyTooShort"
		if err.Error() != expectedError {
			t.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	t.Run("MultipleBeginTagsErrorMessage", func(t *testing.T) {
		multipleBeginInput := "-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 50)) +
			"\n-----END PRIVATE KEY-----\n-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 50)) + "\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(multipleBeginInput)
		if err == nil {
			t.Fatalf("MissingExpectedError: multiple begin tags should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMultipleBeginTags"
		if err.Error() != expectedError {
			t.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	t.Run("MultipleEndTagsErrorMessage", func(t *testing.T) {
		multipleEndInput := "-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 100)) +
			"\n-----END PRIVATE KEY-----\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(multipleEndInput)
		if err == nil {
			t.Fatalf("MissingExpectedError: multiple end tags should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMultipleEndTags"
		if err.Error() != expectedError {
			t.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	t.Run("MissingBeginTagErrorMessage", func(t *testing.T) {
		missingBeginInput := string(make([]byte, 100)) +
			"\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(missingBeginInput)
		if err == nil {
			t.Fatalf("MissingExpectedError: missing begin tag should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMissingBeginTag"
		if err.Error() != expectedError {
			t.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	t.Run("MissingEndTagErrorMessage", func(t *testing.T) {
		missingEndInput := "-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 100))
		_, err := NewEnvelopedPrivateKey(missingEndInput)
		if err == nil {
			t.Fatalf("MissingExpectedError: missing end tag should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMissingEndTag"
		if err.Error() != expectedError {
			t.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})
}
