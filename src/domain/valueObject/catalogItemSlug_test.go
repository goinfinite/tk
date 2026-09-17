package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewCatalogItemSlug(t *testing.T) {
	t.Run("ValidCatalogItemSlug", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemSlug
			expectError    bool
		}{
			{"drupal", CatalogItemSlug("drupal"), false},
			{"opensearch", CatalogItemSlug("opensearch"), false},
			{"php-webserver", CatalogItemSlug("php-webserver"), false},
			{"my_item", CatalogItemSlug("my_item"), false},
			{"DRUPAL", CatalogItemSlug("drupal"), false},
			{strings.Repeat("a", 64), CatalogItemSlug(strings.Repeat("a", 64)), false},
			// Invalid slugs
			{"", CatalogItemSlug(""), true},
			{"a", CatalogItemSlug(""), true},
			{"invalid slug", CatalogItemSlug(""), true},
			{"invalid.slug", CatalogItemSlug(""), true},
			{"invalid@slug", CatalogItemSlug(""), true},
			{strings.Repeat("a", 65), CatalogItemSlug(""), true},
			{123, CatalogItemSlug("123"), false},
			{[]string{"drupal"}, CatalogItemSlug(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemSlug(testCase.inputValue)
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
			inputValue     CatalogItemSlug
			expectedOutput string
		}{
			{CatalogItemSlug("drupal"), "drupal"},
			{CatalogItemSlug("php-webserver"), "php-webserver"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
