package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewCatalogItemName(t *testing.T) {
	t.Run("ValidCatalogItemName", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemName
			expectError    bool
		}{
			{"Drupal", CatalogItemName("Drupal"), false},
			{"OpenSearch", CatalogItemName("OpenSearch"), false},
			{"opensearch", CatalogItemName("opensearch"), false},
			{"php-webserver", CatalogItemName("php-webserver"), false},
			{"node.js", CatalogItemName("node.js"), false},
			{"John's App", CatalogItemName("John's App"), false},
			{"my_item", CatalogItemName("my_item"), false},
			{"  Drupal  ", CatalogItemName("Drupal"), false},
			{strings.Repeat("a", 64), CatalogItemName(strings.Repeat("a", 64)), false},
			// Invalid names
			{"", CatalogItemName(""), true},
			{"ôpencart", CatalogItemName(""), true},
			{"Ωmega", CatalogItemName(""), true},
			{"-invalid", CatalogItemName(""), true},
			{"invalid@name", CatalogItemName(""), true},
			{"invalid/name", CatalogItemName(""), true},
			{strings.Repeat("a", 65), CatalogItemName(""), true},
			{123, CatalogItemName("123"), false},
			{[]string{"Drupal"}, CatalogItemName(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemName(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     CatalogItemName
			expectedOutput string
		}{
			{CatalogItemName("Drupal"), "Drupal"},
			{CatalogItemName("OpenSearch"), "OpenSearch"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
