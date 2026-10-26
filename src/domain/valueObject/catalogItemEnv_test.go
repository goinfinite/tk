package tkValueObject

import "testing"

func TestNewCatalogItemEnv(test *testing.T) {
	test.Run("ValidCatalogItemEnv", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CatalogItemEnv
			expectError    bool
		}{
			{"KEY=value", CatalogItemEnv("KEY=value"), false},
			{CatalogItemEnv("KEY=value"), CatalogItemEnv("KEY=value"), false},
			{"DATABASE_URL=postgres://localhost:5432", CatalogItemEnv("DATABASE_URL=postgres://localhost:5432"), false},
			{"TOKEN=secret ", CatalogItemEnv("TOKEN=secret "), false},
			{"KEY= ", CatalogItemEnv("KEY= "), false},
			{"EMPTY=", CatalogItemEnv(""), true},
			// Invalid envs
			{"", CatalogItemEnv(""), true},
			{"NO_EQUALS_SIGN", CatalogItemEnv(""), true},
			{"=value", CatalogItemEnv(""), true},
			{"KEY WITH SPACE=value", CatalogItemEnv(""), true},
			{123, CatalogItemEnv(""), true},
			{[]string{"KEY=value"}, CatalogItemEnv(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCatalogItemEnv(testCase.inputValue)
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

	test.Run("ReadKeyAndReadValueMethods", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue    CatalogItemEnv
			expectedKey   string
			expectedValue string
		}{
			{CatalogItemEnv("KEY=value"), "KEY", "value"},
			{CatalogItemEnv("DATABASE_URL=postgres://localhost:5432"), "DATABASE_URL", "postgres://localhost:5432"},
			{CatalogItemEnv("KEY=value=with=equals"), "KEY", "value=with=equals"},
		}

		for _, testCase := range testCaseStructs {
			actualKey := testCase.inputValue.ReadKey()
			if actualKey != testCase.expectedKey {
				subtest.Errorf("UnexpectedKey: '%v' vs '%v' [%v]", actualKey, testCase.expectedKey, testCase.inputValue)
			}
			actualValue := testCase.inputValue.ReadValue()
			if actualValue != testCase.expectedValue {
				subtest.Errorf("UnexpectedValue: '%v' vs '%v' [%v]", actualValue, testCase.expectedValue, testCase.inputValue)
			}
		}
	})

	test.Run("StringMethod", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue     CatalogItemEnv
			expectedOutput string
		}{
			{CatalogItemEnv("KEY=value"), "KEY=value"},
			{CatalogItemEnv("DATABASE_URL=postgres://localhost:5432"), "DATABASE_URL=postgres://localhost:5432"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				subtest.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
