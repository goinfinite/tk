package tkInfra

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCypher(t *testing.T) {
	encodedSecretKey, err := NewCypherSecretKey()
	if err != nil {
		t.Fatalf("FailedToGenerateValidKey: %v", err)
	}

	t.Run("NewCypherSecretKey", func(t *testing.T) {
		encodedKey, err := NewCypherSecretKey()
		if err != nil {
			t.Errorf("NewCypherSecretKeyFailed: %v", err)
		}
		if encodedKey == "" {
			t.Error("GeneratedKeyIsEmpty")
		}

		decodedBytes, err := base64.RawURLEncoding.DecodeString(encodedKey)
		if err != nil {
			t.Errorf("GeneratedKeyInvalidBase64: %v", err)
		}
		if len(decodedBytes) != CypherNewSecretKeyLength {
			t.Errorf(
				"GeneratedKeyWrongLength: expected %d, got %d",
				CypherNewSecretKeyLength, len(decodedBytes),
			)
		}
	})

	t.Run("Cypher", func(t *testing.T) {
		testCases := []struct {
			name          string
			input         string
			operationType string
			expectError   bool
			errorContains string
		}{
			{"ValidText", "hello world", "", false, ""},
			{"EmptyText", "", "", false, ""},
			{"SpecialChars", "test@123!#$%", "", false, ""},
			{"LongText", strings.Repeat("a", 1000), "", false, ""},
			{"InvalidSecretKey", "test", "encrypt", true, "SecretKeyDecodeError"},
			{"InvalidEncryptedText", "invalid base64", "decrypt", true, "EncryptedTextDecodeError"},
			{"EncryptedTextTooShort", "AA==", "decrypt", true, "EncryptedTextTooShort"},                           // InputLengthLessThanNonceSize
			{"EncryptedTextTooShortForAuthTag", "YWJjZGVmZ2hpamtsbQ==", "decrypt", true, "EncryptedTextTooShort"}, // InputLengthBetweenNonceSizeAndMinSize
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				invalidEncodedSecretKey := "invalid base64 key"
				cypher := NewCypher(encodedSecretKey)
				if testCase.name == "InvalidSecretKey" {
					cypher = NewCypher(invalidEncodedSecretKey)
				}
				if testCase.expectError {
					var err error
					if testCase.operationType == "encrypt" {
						_, err = cypher.Encrypt(testCase.input)
					}
					if testCase.operationType == "decrypt" {
						_, err = cypher.Decrypt(testCase.input)
					}

					if testCase.expectError && err == nil {
						t.Errorf("MissingExpectedError: %s", testCase.name)
					}
					if !testCase.expectError && err != nil {
						t.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.name)
					}
					if testCase.expectError && err != nil && !strings.Contains(err.Error(), testCase.errorContains) {
						t.Errorf(
							"ErrorDoesNotContainExpectedText: '%s' vs '%s' [%s]",
							err.Error(), testCase.errorContains, testCase.name,
						)
					}

					return
				}

				encryptedText, err := cypher.Encrypt(testCase.input)
				if err != nil {
					t.Errorf("EncryptFailed: %v", err)
				}
				decryptedText, err := cypher.Decrypt(encryptedText)
				if err != nil {
					t.Errorf("DecryptFailed: %v", err)
				}
				if decryptedText != testCase.input {
					t.Errorf(
						"UnexpectedDecryptedText: expected %q, got %q",
						testCase.input, decryptedText,
					)
				}
			})
		}
	})
}
