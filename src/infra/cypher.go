package tkInfra

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

const CypherNewSecretKeyLength = 32

type Cypher struct {
	decodedSecretKeyBytes []byte
}

// NewCypherSecretKey generates a cryptographically secure random 32-byte secret key,
// encodes it in base64 for safe storage and transmission, and returns it as a string.
// This key is suitable for AES-GCM encryption and should be kept confidential.
func NewCypherSecretKey() (string, error) {
	secretKeyBytes := make([]byte, CypherNewSecretKeyLength)
	_, readErr := rand.Read(secretKeyBytes)
	if readErr != nil {
		return "", fmt.Errorf("SecretKeyGenerationError: %w", readErr)
	}
	return base64.RawURLEncoding.EncodeToString(secretKeyBytes), nil
}

// NewCypher creates a new Cypher instance with the provided base64-encoded secret key.
// The key must be a valid base64 string that decodes to exactly 16, 24, or 32 bytes
// for AES encryption. Use NewCypherSecretKey to generate a suitable key if needed.
// Returns an error if the key is invalid, providing fail-fast validation.
func NewCypher(encodedSecretKey string) (*Cypher, error) {
	decodedKeyBytes, err := base64.RawURLEncoding.DecodeString(encodedSecretKey)
	if err != nil {
		return nil, fmt.Errorf("SecretKeyDecodeError: %w", err)
	}
	keyLen := len(decodedKeyBytes)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, errors.New("InvalidSecretKeyLength")
	}
	return &Cypher{decodedSecretKeyBytes: decodedKeyBytes}, nil
}

func (cypher *Cypher) Encrypt(plainText string) (encryptedText string, err error) {
	aesBlock, err := aes.NewCipher(cypher.decodedSecretKeyBytes)
	if err != nil {
		return encryptedText, fmt.Errorf("AesCipherCreationError: %w", err)
	}

	gcmCipher, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return encryptedText, fmt.Errorf("GcmCipherCreationError: %w", err)
	}

	nonceBytes := make([]byte, gcmCipher.NonceSize())
	_, readErr := rand.Read(nonceBytes)
	if readErr != nil {
		return encryptedText, fmt.Errorf("NonceGenerationError: %w", readErr)
	}

	inputBytes := []byte(plainText)
	cipherTextWithAuthTag := gcmCipher.Seal(nil, nonceBytes, inputBytes, nil)
	authCipherTextWithNonce := append(nonceBytes, cipherTextWithAuthTag...)

	return base64.StdEncoding.EncodeToString(authCipherTextWithNonce), nil
}

func (cypher *Cypher) Decrypt(encryptedText string) (plainText string, err error) {
	inputBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return plainText, fmt.Errorf("EncryptedTextDecodeError: %w", err)
	}

	aesBlock, err := aes.NewCipher(cypher.decodedSecretKeyBytes)
	if err != nil {
		return plainText, fmt.Errorf("AesCipherCreationError: %w", err)
	}

	gcmCipher, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return plainText, fmt.Errorf("GcmCipherCreationError: %w", err)
	}

	nonceSize := gcmCipher.NonceSize()
	minInputSize := nonceSize + gcmCipher.Overhead()
	if len(inputBytes) < minInputSize {
		return plainText, errors.New("EncryptedTextTooShort")
	}

	noncePart := inputBytes[:nonceSize]
	authCipherTextWithoutNonce := inputBytes[nonceSize:]

	plainTextBytes, err := gcmCipher.Open(nil, noncePart, authCipherTextWithoutNonce, nil)
	if err != nil {
		return plainText, fmt.Errorf("DecryptionError: %w", err)
	}

	return string(plainTextBytes), nil
}
