package tkInfra

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	SerializationFormatJson = "json"
	SerializationFormatYaml = "yaml"
)

func dataDeserializer(
	serializedReader io.Reader,
	serializationFormat string,
) (outputMap map[string]any, err error) {
	if serializationFormat == SerializationFormatYaml {
		itemYamlDecoder := yaml.NewDecoder(serializedReader)
		err = itemYamlDecoder.Decode(&outputMap)
		if err != nil {
			return outputMap, err
		}

		return outputMap, nil
	}

	itemJsonDecoder := json.NewDecoder(serializedReader)
	err = itemJsonDecoder.Decode(&outputMap)
	if err != nil {
		return outputMap, err
	}

	return outputMap, nil
}

func FileDeserializer(
	filePath string,
) (outputMap map[string]any, err error) {
	fileHandler, err := os.Open(filePath)
	if err != nil {
		return outputMap, err
	}
	defer fileHandler.Close()

	serializationFormat := SerializationFormatJson
	itemFileExt := filepath.Ext(filePath)
	if itemFileExt == ".yml" || itemFileExt == ".yaml" {
		serializationFormat = SerializationFormatYaml
	}

	return dataDeserializer(fileHandler, serializationFormat)
}

func StringDeserializer(
	serializedString string,
	serializationFormat string,
) (outputMap map[string]any, err error) {
	serializedReader := strings.NewReader(serializedString)

	return dataDeserializer(serializedReader, serializationFormat)
}
