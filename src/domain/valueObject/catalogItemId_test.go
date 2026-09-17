package tkValueObject

import "testing"

func TestNewCatalogItemId(t *testing.T) {
	t.Run("ValidCatalogItemId", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemId
			expectError    bool
		}{
			{"0", CatalogItemId(0), false},
			{int(0), CatalogItemId(0), false},
			{uint16(11), CatalogItemId(11), false},
			{CatalogItemId(11), CatalogItemId(11), false},
			{float64(11), CatalogItemId(11), false},
			// Invalid ids
			{"-1", CatalogItemId(0), true},
			{int(-1), CatalogItemId(0), true},
			{uint32(65536), CatalogItemId(0), true},
			{"invalid", CatalogItemId(0), true},
			{[]string{"1"}, CatalogItemId(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemId(testCase.inputValue)
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

	t.Run("Uint16Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     CatalogItemId
			expectedOutput uint16
		}{
			{CatalogItemId(0), 0},
			{CatalogItemId(11), 11},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint16()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     CatalogItemId
			expectedOutput string
		}{
			{CatalogItemId(0), "0"},
			{CatalogItemId(11), "11"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
