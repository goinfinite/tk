package tkValueObject

import "testing"

func TestNewCatalogItemType(t *testing.T) {
	t.Run("ValidCatalogItemType", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemType
			expectError    bool
		}{
			{"app", CatalogItemTypeApp, false},
			{"framework", CatalogItemTypeFramework, false},
			{"stack", CatalogItemTypeStack, false},
			{"system", CatalogItemTypeSystem, false},
			{"database", CatalogItemTypeDatabase, false},
			{"runtime", CatalogItemTypeRuntime, false},
			{"webserver", CatalogItemTypeWebServer, false},
			{"other", CatalogItemTypeOther, false},
			{"APP", CatalogItemTypeApp, false},
			// Unknown types coerce to other
			{"", CatalogItemTypeOther, false},
			{"invalid", CatalogItemTypeOther, false},
			{"tool", CatalogItemTypeOther, false},
			{123, CatalogItemTypeOther, false},
			{[]string{"app"}, CatalogItemType(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemType(testCase.inputValue)
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
			inputValue     CatalogItemType
			expectedOutput string
		}{
			{CatalogItemTypeApp, "app"},
			{CatalogItemTypeDatabase, "database"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
