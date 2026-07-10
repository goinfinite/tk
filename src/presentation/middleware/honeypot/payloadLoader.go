package tkPresentationMiddlewareHoneypot

import (
	"embed"
	"encoding/base64"
	"errors"
	"io/fs"
)

const payloadsDirectoryName string = "payloads"

type PayloadLoader struct {
	payloadsFs embed.FS
}

func NewPayloadLoader(payloadsFs embed.FS) *PayloadLoader {
	return &PayloadLoader{payloadsFs: payloadsFs}
}

func (loader *PayloadLoader) Load(fileName string) ([]byte, error) {
	filePath := payloadsDirectoryName + "/" + fileName
	rawContent, readErr := fs.ReadFile(loader.payloadsFs, filePath)
	if readErr != nil {
		return nil, errors.New("PayloadFileReadError")
	}

	decodedContent, decodeErr := base64.StdEncoding.DecodeString(
		string(rawContent),
	)
	if decodeErr != nil {
		return nil, errors.New("PayloadBase64DecodeError")
	}

	return decodedContent, nil
}

func (loader *PayloadLoader) LoadAll() (map[string][]byte, error) {
	fileEntries, readDirErr := fs.ReadDir(
		loader.payloadsFs, payloadsDirectoryName,
	)
	if readDirErr != nil {
		return nil, errors.New("PayloadDirectoryReadError")
	}

	payloads := make(map[string][]byte, len(fileEntries))
	for _, entry := range fileEntries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		decodedContent, loadErr := loader.Load(fileName)
		if loadErr != nil {
			return nil, loadErr
		}
		payloads[fileName] = decodedContent
	}

	return payloads, nil
}
