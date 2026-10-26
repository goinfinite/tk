package tkValueObject

import (
	"os/exec"
	"strings"
	"testing"
)

func privateKeyPemGenerator(
	test *testing.T, algorithm PrivateKeyAlgorithm,
) string {
	test.Helper()

	switch algorithm {
	case PrivateKeyAlgorithmRSA:
		rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
		rsaKeyOutput, err := rsaKeyCmd.Output()
		if err != nil {
			test.Skipf("OpenSslNotAvailable: %s", err.Error())
		}
		return strings.TrimSpace(string(rsaKeyOutput))

	case "PKCS8":
		rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
		rsaKeyOutput, err := rsaKeyCmd.Output()
		if err != nil {
			test.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		pkcs8Cmd := exec.Command(
			"openssl", "pkcs8", "-topk8", "-inform", "PEM", "-outform", "PEM",
			"-nocrypt",
		)
		pkcs8Cmd.Stdin = strings.NewReader(string(rsaKeyOutput))
		pkcs8Output, err := pkcs8Cmd.Output()
		if err != nil {
			test.Skipf("PKCS8ConversionFailed: %s", err.Error())
		}
		return strings.TrimSpace(string(pkcs8Output))

	case PrivateKeyAlgorithmECDSA:
		ecKeyCmd := exec.Command("openssl", "ecparam", "-genkey", "-name", "prime256v1")
		ecKeyOutput, err := ecKeyCmd.Output()
		if err != nil {
			test.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		ecKeyOutputStr := string(ecKeyOutput)
		keyStart := strings.Index(ecKeyOutputStr, "-----BEGIN EC PRIVATE KEY-----")
		if keyStart == -1 {
			test.Fatalf("ECPrivateKeyNotFoundInOutput")
		}
		return strings.TrimSpace(ecKeyOutputStr[keyStart:])

	case "Encrypted":
		rsaKeyCmd := exec.Command("openssl", "genrsa", "2048")
		rsaKeyOutput, err := rsaKeyCmd.Output()
		if err != nil {
			test.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		encryptCmd := exec.Command(
			"openssl", "pkcs8", "-topk8", "-inform", "PEM", "-outform", "PEM",
			"-v2", "aes-256-cbc", "-passout", "pass:testpassword",
		)
		encryptCmd.Stdin = strings.NewReader(string(rsaKeyOutput))
		encryptOutput, err := encryptCmd.Output()
		if err != nil {
			test.Skipf("PrivateKeyEncryptionFailed: %s", err.Error())
		}
		return strings.TrimSpace(string(encryptOutput))

	case PrivateKeyAlgorithmDSA:
		dsaCmd := exec.Command("openssl", "dsaparam", "-genkey", "2048")
		dsaOutput, err := dsaCmd.Output()
		if err != nil {
			test.Skipf("OpenSslNotAvailable: %s", err.Error())
		}

		dsaOutputStr := string(dsaOutput)
		keyStart := strings.Index(dsaOutputStr, "-----BEGIN PRIVATE KEY-----")
		if keyStart == -1 {
			keyStart = strings.Index(dsaOutputStr, "-----BEGIN DSA PRIVATE KEY-----")
		}
		if keyStart == -1 {
			test.Fatalf("DSAPrivateKeyNotFoundInOutput")
		}
		return strings.TrimSpace(dsaOutputStr[keyStart:])

	default:
		test.Fatalf("UnsupportedPrivateKeyAlgorithm: %s", algorithm)
		return ""
	}
}

func TestNewEnvelopedPrivateKey(test *testing.T) {
	validPrivateKey := privateKeyPemGenerator(test, "PKCS8")
	validRsaPrivateKey := privateKeyPemGenerator(test, PrivateKeyAlgorithmRSA)
	validEcPrivateKey := privateKeyPemGenerator(test, PrivateKeyAlgorithmECDSA)
	validEncryptedPrivateKey := privateKeyPemGenerator(test, "Encrypted")
	validDsaPrivateKey := privateKeyPemGenerator(test, PrivateKeyAlgorithmDSA)

	test.Run("ValidEnvelopedPrivateKey", func(subtest *testing.T) {
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
				"-----BEGIN PRIVATE KEY-----\n" + strings.Repeat("A", 1048576) +
					"\n-----END PRIVATE KEY-----",
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
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}

			if !testCase.expectError && err != nil {
				subtest.Fatalf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}

			if !testCase.expectError &&
				actualOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v' [%v]",
					actualOutput, testCase.expectedOutput, testCase.inputValue,
				)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
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
				subtest.Errorf(
					"UnexpectedOutputValue: '%v' vs '%v'",
					actualOutput, testCase.expectedOutput,
				)
			}
		}
	})

	test.Run("ShortInputErrorMessage", func(subtest *testing.T) {
		shortInput := "-----BEGIN PRIVATE KEY-----\nshort\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(shortInput)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: short input should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyTooShort"
		if err.Error() != expectedError {
			subtest.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	test.Run("MultipleBeginTagsErrorMessage", func(subtest *testing.T) {
		multipleBeginInput := "-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 50)) +
			"\n-----END PRIVATE KEY-----\n-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 50)) + "\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(multipleBeginInput)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: multiple begin tags should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMultipleBeginTags"
		if err.Error() != expectedError {
			subtest.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	test.Run("MultipleEndTagsErrorMessage", func(subtest *testing.T) {
		multipleEndInput := "-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 100)) +
			"\n-----END PRIVATE KEY-----\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(multipleEndInput)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: multiple end tags should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMultipleEndTags"
		if err.Error() != expectedError {
			subtest.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	test.Run("MissingBeginTagErrorMessage", func(subtest *testing.T) {
		missingBeginInput := string(make([]byte, 100)) +
			"\n-----END PRIVATE KEY-----"
		_, err := NewEnvelopedPrivateKey(missingBeginInput)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: missing begin tag should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMissingBeginTag"
		if err.Error() != expectedError {
			subtest.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})

	test.Run("MissingEndTagErrorMessage", func(subtest *testing.T) {
		missingEndInput := "-----BEGIN PRIVATE KEY-----\n" +
			string(make([]byte, 100))
		_, err := NewEnvelopedPrivateKey(missingEndInput)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: missing end tag should fail")
		}

		expectedError := "InvalidEnvelopedPrivateKeyMissingEndTag"
		if err.Error() != expectedError {
			subtest.Errorf(
				"UnexpectedErrorMessage: got '%s', expected '%s'",
				err.Error(), expectedError,
			)
		}
	})
}
