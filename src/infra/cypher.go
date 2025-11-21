package tkInfra

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

const CypherNewSecretKeyLength = 32

type Cypher struct {
	encodedSecretKey string
}

// NewCypherSecretKey generates a cryptographically secure random 32-byte secret key,
// encodes it in base64 for safe storage and transmission, and returns it as a string.
// This key is suitable for AES-GCM encryption and should be kept confidential.
func NewCypherSecretKey() (string, error) {
	secretKeyBytes := make([]byte, CypherNewSecretKeyLength)
	if _, err := rand.Read(secretKeyBytes); err != nil {
		return "", errors.New("SecretKeyGenerationError: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(secretKeyBytes), nil
}

// NewCypher creates a new Cypher instance with the provided base64-encoded secret key.
// The key must be a valid base64 string that decodes to 16, 24, or 32 bytes for AES encryption.
// Use NewCypherSecretKey to generate a suitable key if needed.
func NewCypher(encodedSecretKey string) *Cypher {
	return &Cypher{encodedSecretKey: encodedSecretKey}
}

func (cypher *Cypher) Encrypt(plainText string) (encryptedText string, err error) {
	decodedKey, err := base64.RawURLEncoding.DecodeString(cypher.encodedSecretKey)
	if err != nil {
		return encryptedText, errors.New("SecretKeyDecodeError: " + err.Error())
	}

	aesBlock, err := aes.NewCipher(decodedKey)
	if err != nil {
		return encryptedText, errors.New("AesCipherCreationError: " + err.Error())
	}

	gcmCipher, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return encryptedText, errors.New("GcmCipherCreationError: " + err.Error())
	}

	nonceBytes := make([]byte, gcmCipher.NonceSize())
	if _, err := rand.Read(nonceBytes); err != nil {
		return encryptedText, errors.New("NonceGenerationError: " + err.Error())
	}

	inputBytes := []byte(plainText)
	cipherTextWithAuthTag := gcmCipher.Seal(nil, nonceBytes, inputBytes, nil)
	authCipherTextWithNonce := append(nonceBytes, cipherTextWithAuthTag...)

	return base64.StdEncoding.EncodeToString(authCipherTextWithNonce), nil
}

func (cypher *Cypher) Decrypt(encryptedText string) (plainText string, err error) {
	inputBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return plainText, errors.New("EncryptedTextDecodeError: " + err.Error())
	}

	decodedKey, err := base64.RawURLEncoding.DecodeString(cypher.encodedSecretKey)
	if err != nil {
		return plainText, errors.New("SecretKeyDecodeError: " + err.Error())
	}

	aesBlock, err := aes.NewCipher(decodedKey)
	if err != nil {
		return plainText, errors.New("AesCipherCreationError: " + err.Error())
	}

	gcmCipher, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return plainText, errors.New("GcmCipherCreationError: " + err.Error())
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
		return plainText, errors.New("DecryptionError: " + err.Error())
	}

	return string(plainTextBytes), nil
}
