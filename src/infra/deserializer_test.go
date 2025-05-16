package tkInfra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStringDeserializer(t *testing.T) {
	t.Run("JsonFormat", func(t *testing.T) {
		testCaseStructs := []struct {
			inputJson     string
			expectedMap   map[string]any
			expectError   bool
			errorContains string
		}{
			{
				`{"name": "test", "value": 123}`,
				map[string]any{"name": "test", "value": float64(123)},
				false,
				"",
			},
			{
				`{"nested": {"key": "value"}, "array": [1, 2, 3]}`,
				map[string]any{
					"nested": map[string]any{"key": "value"},
					"array":  []any{float64(1), float64(2), float64(3)},
				},
				false,
				"",
			},
			{
				`invalid json`,
				nil,
				true,
				"invalid",
			},
		}

		for _, testCase := range testCaseStructs {
			actualMap, err := StringDeserializer(testCase.inputJson, SerializationFormatJson)

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputJson)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.inputJson)
			}
			if testCase.expectError && err != nil && !strings.Contains(err.Error(), testCase.errorContains) {
				t.Errorf(
					"ErrorDoesNotContainExpectedText: '%s' vs '%s' [%s]",
					err.Error(), testCase.errorContains, testCase.inputJson,
				)
			}
			if !testCase.expectError {
				assertMapsEqual(t, actualMap, testCase.expectedMap)
			}
		}
	})

	t.Run("YamlFormat", func(t *testing.T) {
		testCaseStructs := []struct {
			inputYaml     string
			expectedMap   map[string]any
			expectError   bool
			errorContains string
		}{
			{
				"name: test\nvalue: 123",
				map[string]any{"name": "test", "value": 123},
				false,
				"",
			},
			{
				"nested:\n  key: value\narray:\n  - 1\n  - 2\n  - 3",
				map[string]any{
					"nested": map[string]any{"key": "value"},
					"array":  []any{1, 2, 3},
				},
				false,
				"",
			},
			{
				"invalid: yaml:\n  - missing colon",
				nil,
				true,
				"yaml",
			},
		}

		for _, testCase := range testCaseStructs {
			actualMap, err := StringDeserializer(testCase.inputYaml, SerializationFormatYaml)

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputYaml)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.inputYaml)
			}
			if testCase.expectError && err != nil && !strings.Contains(err.Error(), testCase.errorContains) {
				t.Errorf(
					"ErrorDoesNotContainExpectedText: '%s' vs '%s' [%s]",
					err.Error(), testCase.errorContains, testCase.inputYaml,
				)
			}
			if !testCase.expectError {
				assertMapsEqual(t, actualMap, testCase.expectedMap)
			}
		}
	})
}

func TestFileDeserializer(t *testing.T) {
	t.Run("JsonFile", func(t *testing.T) {
		jsonContent := `{"name": "test", "value": 123, "nested": {"key": "value"}}`
		jsonFile := createTempFile(t, "test.json", jsonContent)
		defer os.Remove(jsonFile)

		expectedMap := map[string]any{
			"name":   "test",
			"value":  float64(123),
			"nested": map[string]any{"key": "value"},
		}

		actualMap, err := FileDeserializer(jsonFile)
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		assertMapsEqual(t, actualMap, expectedMap)
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := FileDeserializer("non_existent_file.json")
		if err == nil {
			t.Errorf("MissingExpectedError: FileNotFound")
		}
	})

	t.Run("InvalidJsonFile", func(t *testing.T) {
		invalidContent := "invalid json content"
		invalidFile := createTempFile(t, "invalid.json", invalidContent)
		defer os.Remove(invalidFile)

		_, err := FileDeserializer(invalidFile)
		if err == nil {
			t.Errorf("MissingExpectedError: InvalidJsonFormat")
		}
	})

	t.Run("YamlFile", func(t *testing.T) {
		yamlContent := "name: test\nvalue: 123\nnested:\n  key: value"
		yamlFile := createTempFile(t, "test.yaml", yamlContent)
		defer os.Remove(yamlFile)

		expectedMap := map[string]any{
			"name":   "test",
			"value":  123,
			"nested": map[string]any{"key": "value"},
		}

		actualMap, err := FileDeserializer(yamlFile)
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		} else {
			assertMapsEqual(t, actualMap, expectedMap)
		}
	})
}

func createTempFile(
	t *testing.T,
	filename, content string,
) string {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("CreateTempFileFailed: %v", err)
	}

	return filePath
}

func assertMapsEqual(
	t *testing.T,
	actual, expected map[string]any,
) {
	if actual == nil && expected == nil {
		return
	}
	if actual == nil && expected != nil {
		t.Errorf("ActualMapIsNil")
		return
	}
	if actual != nil && expected == nil {
		t.Errorf("ExpectedMapIsNil")
		return
	}

	actualJson, err := json.Marshal(actual)
	if err != nil {
		t.Errorf("MarshalActualMapFailed: %v", err)
		return
	}

	expectedJson, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("MarshalExpectedMapFailed: %v", err)
		return
	}

	if string(actualJson) != string(expectedJson) {
		t.Errorf(
			"MapMismatch\nExpected JSON: %s\nActual JSON: %s",
			string(expectedJson), string(actualJson),
		)
	}
}
