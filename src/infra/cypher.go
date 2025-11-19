package tkInfra

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const CypherNewSecretKeyLength = 32

type Cypher struct {
	encodedSecretKey string
}

// NewCypherSecretKey generates a cryptographically secure random 32-byte secret key,
// encodes it in base64 for safe storage and transmission, and returns it as a string.
// This key is suitable for AES-256 encryption and should be kept confidential.
func NewCypherSecretKey() (string, error) {
	secretKeyBytes := make([]byte, CypherNewSecretKeyLength)
	if _, err := rand.Read(secretKeyBytes); err != nil {
		return "", err
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
		return encryptedText, errors.New("CipherCreationError: " + err.Error())
	}

	inputBytes := []byte(plainText)
	outputBytes := make([]byte, aes.BlockSize+len(inputBytes))
	ivBuffer := outputBytes[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, ivBuffer); err != nil {
		return encryptedText, errors.New("IvGenerationError: " + err.Error())
	}

	ctrStream := cipher.NewCTR(aesBlock, ivBuffer)
	ctrStream.XORKeyStream(outputBytes[aes.BlockSize:], inputBytes)

	return base64.StdEncoding.EncodeToString(outputBytes), nil
}

func (cypher *Cypher) Decrypt(encryptedText string) (plainText string, err error) {
	inputBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return plainText, errors.New("EncryptedTextDecodeError: " + err.Error())
	}
	if len(inputBytes) < aes.BlockSize {
		return plainText, errors.New("EncryptedTextTooShort")
	}

	decodedKey, err := base64.RawURLEncoding.DecodeString(cypher.encodedSecretKey)
	if err != nil {
		return plainText, errors.New("SecretKeyDecodeError: " + err.Error())
	}

	aesBlock, err := aes.NewCipher(decodedKey)
	if err != nil {
		return plainText, errors.New("CipherCreationError: " + err.Error())
	}

	outputBytes := make([]byte, len(inputBytes)-aes.BlockSize)
	ivBuffer := inputBytes[:aes.BlockSize]

	ctrStream := cipher.NewCTR(aesBlock, ivBuffer)
	ctrStream.XORKeyStream(outputBytes, inputBytes[aes.BlockSize:])

	return string(outputBytes), nil
}
