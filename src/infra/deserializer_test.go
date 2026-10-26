package tkInfra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStringDeserializer(test *testing.T) {
	test.Run("JsonFormat", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%s]", testCase.inputJson)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.inputJson)
			}
			if testCase.expectError && err != nil && !strings.Contains(err.Error(), testCase.errorContains) {
				subtest.Errorf(
					"ErrorDoesNotContainExpectedText: '%s' vs '%s' [%s]",
					err.Error(), testCase.errorContains, testCase.inputJson,
				)
			}
			if !testCase.expectError {
				assertMapsEqual(subtest, actualMap, testCase.expectedMap)
			}
		}
	})

	test.Run("YamlFormat", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%s]", testCase.inputYaml)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%s]", err.Error(), testCase.inputYaml)
			}
			if testCase.expectError && err != nil && !strings.Contains(err.Error(), testCase.errorContains) {
				subtest.Errorf(
					"ErrorDoesNotContainExpectedText: '%s' vs '%s' [%s]",
					err.Error(), testCase.errorContains, testCase.inputYaml,
				)
			}
			if !testCase.expectError {
				assertMapsEqual(subtest, actualMap, testCase.expectedMap)
			}
		}
	})
}

func TestFileDeserializer(test *testing.T) {
	test.Run("JsonFile", func(subtest *testing.T) {
		jsonContent := `{"name": "test", "value": 123, "nested": {"key": "value"}}`
		jsonFile := createTempFile(subtest, "test.json", jsonContent)
		defer func() { _ = os.Remove(jsonFile) }()

		expectedMap := map[string]any{
			"name":   "test",
			"value":  float64(123),
			"nested": map[string]any{"key": "value"},
		}

		actualMap, err := FileDeserializer(jsonFile)
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		}

		assertMapsEqual(subtest, actualMap, expectedMap)
	})

	test.Run("NonExistentFile", func(subtest *testing.T) {
		_, err := FileDeserializer("non_existent_file.json")
		if err == nil {
			subtest.Errorf("MissingExpectedError: FileNotFound")
		}
	})

	test.Run("InvalidJsonFile", func(subtest *testing.T) {
		invalidContent := "invalid json content"
		invalidFile := createTempFile(subtest, "invalid.json", invalidContent)
		defer func() { _ = os.Remove(invalidFile) }()

		_, err := FileDeserializer(invalidFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: InvalidJsonFormat")
		}
	})

	test.Run("YamlFile", func(subtest *testing.T) {
		yamlContent := "name: test\nvalue: 123\nnested:\n  key: value"
		yamlFile := createTempFile(subtest, "test.yaml", yamlContent)
		defer func() { _ = os.Remove(yamlFile) }()

		expectedMap := map[string]any{
			"name":   "test",
			"value":  123,
			"nested": map[string]any{"key": "value"},
		}

		actualMap, err := FileDeserializer(yamlFile)
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		} else {
			assertMapsEqual(subtest, actualMap, expectedMap)
		}
	})
}

func createTempFile(
	test *testing.T,
	filename, content string,
) string {
	test.Helper()
	tempDir := test.TempDir()
	filePath := filepath.Join(tempDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		test.Fatalf("CreateTempFileFailed: %v", err)
	}

	return filePath
}

func assertMapsEqual(
	test *testing.T,
	actual, expected map[string]any,
) {
	test.Helper()
	if actual == nil && expected == nil {
		return
	}
	if actual == nil && expected != nil {
		test.Errorf("ActualMapIsNil")
		return
	}
	if actual != nil && expected == nil {
		test.Errorf("ExpectedMapIsNil")
		return
	}

	actualJson, err := json.Marshal(actual)
	if err != nil {
		test.Errorf("MarshalActualMapFailed: %v", err)
		return
	}

	expectedJson, err := json.Marshal(expected)
	if err != nil {
		test.Errorf("MarshalExpectedMapFailed: %v", err)
		return
	}

	if string(actualJson) != string(expectedJson) {
		test.Errorf(
			"MapMismatch\nExpected JSON: %s\nActual JSON: %s",
			string(expectedJson), string(actualJson),
		)
	}
}
