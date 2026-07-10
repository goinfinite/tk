package tkPresentationMiddlewareHoneypot

import (
	"embed"
	"encoding/base64"
	"testing"
)

//go:embed payloads
var testPayloadsFs embed.FS

//go:embed testdata
var testCorruptedFs embed.FS

func TestPayloadLoader(t *testing.T) {
	loader := NewPayloadLoader(testPayloadsFs)

	t.Run("LoadExistingFile", func(t *testing.T) {
		decodedContent, err := loader.Load("wp-config-php.bin")
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}
		if len(decodedContent) == 0 {
			t.Errorf("EmptyContent: ExpectedNonEmpty")
		}
	})

	t.Run("LoadNonExistentFile", func(t *testing.T) {
		_, err := loader.Load("nonexistent-file.bin")
		if err == nil {
			t.Errorf("MissingExpectedError")
		}
		if err.Error() != "PayloadFileReadError" {
			t.Errorf(
				"UnexpectedErrorMessage: Expected='PayloadFileReadError', Actual='%s'",
				err.Error(),
			)
		}
	})

	t.Run("LoadCorruptedBase64", func(t *testing.T) {
		corruptedLoader := NewPayloadLoader(testCorruptedFs)
		_, err := corruptedLoader.Load("corrupted-base64.bin")
		if err == nil {
			t.Errorf("MissingExpectedError")
		}
	})

	t.Run("LoadAllReturnsCorrectCount", func(t *testing.T) {
		allPayloads, err := loader.LoadAll()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}
		if len(allPayloads) != 114 {
			t.Errorf(
				"PayloadCountMismatch: Expected=114, Actual=%d",
				len(allPayloads),
			)
		}
	})

	t.Run("LoadAllDecodesAllFiles", func(t *testing.T) {
		allPayloads, err := loader.LoadAll()
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		for fileName, content := range allPayloads {
			if len(content) == 0 {
				t.Errorf("EmptyContentForFile: '%s'", fileName)
			}
		}
	})
}

func TestPayloadLoaderContentRoundtrip(t *testing.T) {
	loader := NewPayloadLoader(testPayloadsFs)

	originalText := "test payload content for roundtrip verification"
	encoded := base64.StdEncoding.EncodeToString([]byte(originalText))

	_ = encoded
	_ = loader

	decodedContent, err := loader.Load("wp-config-php.bin")
	if err != nil {
		t.Errorf("UnexpectedError: '%s'", err.Error())
		return
	}

	reEncoded := base64.StdEncoding.EncodeToString(decodedContent)
	reDecoded, decodeErr := base64.StdEncoding.DecodeString(reEncoded)
	if decodeErr != nil {
		t.Errorf("RoundtripDecodeError: '%s'", decodeErr.Error())
	}

	if string(reDecoded) != string(decodedContent) {
		t.Errorf("RoundtripContentMismatch")
	}
}
