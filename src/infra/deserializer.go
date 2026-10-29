package tkInfra

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	SerializationFormatJson = "json"
	SerializationFormatYaml = "yaml"
)

var ErrInvalidYamlContent = errors.New("InvalidYamlContent")

func yamlDecodeErrorPosition(decodeErr error) (line, column int) {
	yamlErr, isYamlError := decodeErr.(yaml.Error)
	if !isYamlError {
		return 0, 0
	}

	yamlToken := yamlErr.GetToken()
	if yamlToken == nil || yamlToken.Position == nil {
		return 0, 0
	}

	return yamlToken.Position.Line, yamlToken.Position.Column
}

func dataDeserializer(
	serializedReader io.Reader,
	serializationFormat string,
) (outputMap map[string]any, err error) {
	if serializationFormat == SerializationFormatYaml {
		itemYamlDecoder := yaml.NewDecoder(serializedReader)
		err = itemYamlDecoder.Decode(&outputMap)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return outputMap, err
			}
			line, column := yamlDecodeErrorPosition(err)
			slog.Error(
				"YamlDeserializerDecodeError",
				slog.Int("line", line),
				slog.Int("column", column),
				slog.String("error", ErrInvalidYamlContent.Error()),
			)
			return outputMap, ErrInvalidYamlContent
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
	defer func() {
		closeErr := fileHandler.Close()
		if closeErr == nil {
			return
		}
		if err == nil {
			err = closeErr
			return
		}
		slog.Warn(
			"FileDeserializerCloseError",
			slog.String("filePath", filePath),
			slog.String("error", closeErr.Error()),
		)
	}()

	serializationFormat := SerializationFormatJson
	itemFileExt := filepath.Ext(filePath)
	if itemFileExt == ".yml" || itemFileExt == ".yaml" {
		serializationFormat = SerializationFormatYaml
	}

	outputMap, deserializeErr := dataDeserializer(
		fileHandler, serializationFormat,
	)
	if deserializeErr != nil {
		slog.Error(
			"FileDeserializerError",
			slog.String("filePath", filePath),
			slog.String("serializationFormat", serializationFormat),
			slog.String("error", deserializeErr.Error()),
		)
		return outputMap, errors.New("FileDeserializerError")
	}

	return outputMap, nil
}

func StringDeserializer(
	serializedString string,
	serializationFormat string,
) (outputMap map[string]any, err error) {
	serializedReader := strings.NewReader(serializedString)

	return dataDeserializer(serializedReader, serializationFormat)
}
