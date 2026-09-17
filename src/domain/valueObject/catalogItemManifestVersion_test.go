package tkValueObject

import "testing"

func TestNewCatalogItemManifestVersion(t *testing.T) {
	t.Run("ValidCatalogItemManifestVersion", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemManifestVersion
			expectError    bool
		}{
			{"v1", CatalogItemManifestVersion("v1"), false},
			{"V1", CatalogItemManifestVersion("v1"), false},
			{"v2", CatalogItemManifestVersion("v2"), false},
			{"v1.2", CatalogItemManifestVersion("v1.2"), false},
			{"v1.2.3", CatalogItemManifestVersion("v1.2.3"), false},
			// Invalid versions
			{"", CatalogItemManifestVersion(""), true},
			{"v", CatalogItemManifestVersion(""), true},
			{"1", CatalogItemManifestVersion(""), true},
			{"v1.", CatalogItemManifestVersion(""), true},
			{"v1.2.", CatalogItemManifestVersion(""), true},
			{"vv1", CatalogItemManifestVersion(""), true},
			{"invalid", CatalogItemManifestVersion(""), true},
			{123, CatalogItemManifestVersion(""), true},
			{[]string{"v1"}, CatalogItemManifestVersion(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemManifestVersion(testCase.inputValue)
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
			inputValue     CatalogItemManifestVersion
			expectedOutput string
		}{
			{CatalogItemManifestVersion("v1"), "v1"},
			{CatalogItemManifestVersion("v1.2.3"), "v1.2.3"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
