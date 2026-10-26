package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewCatalogItemDescription(test *testing.T) {
	test.Run("ValidCatalogItemDescription", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemDescription
			expectError    bool
		}{
			{"Drupal is an open source platform.", CatalogItemDescription("Drupal is an open source platform."), false},
			{"ab", CatalogItemDescription("ab"), false},
			{strings.Repeat("a", 2048), CatalogItemDescription(strings.Repeat("a", 2048)), false},
			// Invalid descriptions
			{"", CatalogItemDescription(""), true},
			{"a", CatalogItemDescription(""), true},
			{strings.Repeat("a", 2049), CatalogItemDescription(""), true},
			{123, CatalogItemDescription("123"), false},
			{[]string{"description"}, CatalogItemDescription(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemDescription(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				subtest.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				subtest.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     CatalogItemDescription
			expectedOutput string
		}{
			{CatalogItemDescription("Drupal"), "Drupal"},
			{CatalogItemDescription("OpenSearch"), "OpenSearch"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
