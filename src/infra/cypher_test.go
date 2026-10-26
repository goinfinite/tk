package tkInfra

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCypher(test *testing.T) {
	encodedSecretKey, err := NewCypherSecretKey()
	if err != nil {
		test.Fatalf("FailedToGenerateValidKey: %v", err)
	}

	test.Run("NewCypherSecretKey", func(subtest *testing.T) {
		encodedKey, err := NewCypherSecretKey()
		if err != nil {
			subtest.Errorf("NewCypherSecretKeyFailed: %v", err)
		}
		if encodedKey == "" {
			subtest.Error("GeneratedKeyIsEmpty")
		}

		decodedBytes, err := base64.RawURLEncoding.DecodeString(encodedKey)
		if err != nil {
			subtest.Errorf("GeneratedKeyInvalidBase64: %v", err)
		}
		if len(decodedBytes) != CypherNewSecretKeyLength {
			subtest.Errorf(
				"GeneratedKeyWrongLength: expected %d, got %d",
				CypherNewSecretKeyLength, len(decodedBytes),
			)
		}
	})

	test.Run("Cypher", func(subtest *testing.T) {
		cypher, err := NewCypher(encodedSecretKey)
		if err != nil {
			subtest.Fatalf("CreateCypherInstanceFailed: %v", err)
		}

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
			{"InvalidSecretKey", "invalid!", "encrypt", true, "SecretKeyDecodeError"},
			{"InvalidEncryptedText", "invalid base64", "decrypt", true, "EncryptedTextDecodeError"},
			{"EncryptedTextTooShort", "AA==", "decrypt", true, "EncryptedTextTooShort"},                           // InputLengthLessThanNonceSize
			{"EncryptedTextTooShortForAuthTag", "YWJjZGVmZ2hpamtsbQ==", "decrypt", true, "EncryptedTextTooShort"}, // InputLengthBetweenNonceSizeAndMinSize
		}

		for _, testCase := range testCases {
			subtest.Run(testCase.name, func(subtest *testing.T) {
				if testCase.expectError {
					var err error
					switch testCase.operationType {
					case "encrypt":
						_, err = cypher.Encrypt(testCase.input)
						if testCase.name == "InvalidSecretKey" {
							_, err = NewCypher(testCase.input)
						}
					case "decrypt":
						_, err = cypher.Decrypt(testCase.input)
					default:
						subtest.Errorf("InvalidOperationType: %s", testCase.operationType)
					}

					if testCase.expectError && err == nil {
						subtest.Errorf("MissingExpectedError: %s", testCase.name)
					}
					if !testCase.expectError && err != nil {
						subtest.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.name)
					}
					if testCase.expectError && err != nil && !strings.Contains(err.Error(), testCase.errorContains) {
						subtest.Errorf(
							"ErrorDoesNotContainExpectedText: '%s' vs '%s' [%s]",
							err.Error(), testCase.errorContains, testCase.name,
						)
					}

					return
				}

				encryptedText, err := cypher.Encrypt(testCase.input)
				if err != nil {
					subtest.Errorf("EncryptFailed: %v", err)
				}
				decryptedText, err := cypher.Decrypt(encryptedText)
				if err != nil {
					subtest.Errorf("DecryptFailed: %v", err)
				}
				if decryptedText != testCase.input {
					subtest.Errorf(
						"UnexpectedDecryptedText: expected %q, got %q",
						testCase.input, decryptedText,
					)
				}
			})
		}
	})
}
