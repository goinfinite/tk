package tkValueObject

import "testing"

func TestSystemResourceIdentifier(test *testing.T) {
	test.Run("NewSystemResourceIdentifier/Valid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"sri://1:account/120", false},
			{"sri://10:secureAccessPublicKey/1", false},
			{"sri://100:cron/1", false},
			{"sri://1000:database/myDb", false},
			{"sri://1:databaseUser/myDbUser", false},
			{"sri://10:marketplaceCatalogItem/1", false},
			{"sri://100:marketplaceCatalogItem/php", false},
			{"sri://1000:marketplaceInstalledItem/1", false},
			{"sri://1:phpRuntime/local.os", false},
			{"sri://10:installableService/node", false},
			{"sri://100:customService/node-e87qxc21", false},
			{"sri://1000:installedService/1", false},
			{"sri://1:ssl/1", false},
			{"sri://10:ssl/*", false},
			{"sri://10:virtualHost/local.os", false},
			{"sri://100:mapping/1", false},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewSystemResourceIdentifier(testCase.inputValue)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedErrorForInput: %v", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedErrorForInput: %s %v", err.Error(), testCase.inputValue)
			}
		}
	})

	test.Run("NewSystemResourceIdentifier/Invalid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"", true},
			{"sri://0:/", true},
			{true, true},
			{1000, true},
			{"sri://1000:unixFile//app/.trash", true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewSystemResourceIdentifier(testCase.inputValue)
			if testCase.expectError && err == nil {
				subtest.Errorf("MissingExpectedErrorForInput: %v", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedErrorForInput: %s %v", err.Error(), testCase.inputValue)
			}
		}
	})

	test.Run("NewSystemResourceIdentifierMustCreate/Valid", func(subtest *testing.T) {
		validSri := "sri://1:account/120"
		sri := NewSystemResourceIdentifierMustCreate(validSri)
		if sri.String() != validSri {
			subtest.Errorf("UnexpectedSriValue: expected %s, got %s", validSri, sri.String())
		}
	})

	test.Run("NewSystemResourceIdentifierMustCreate/Panic", func(subtest *testing.T) {
		invalidSri := "invalid"
		defer func() {
			if recovered := recover(); recovered == nil {
				subtest.Errorf("MissingExpectedPanicForInvalidSri")
			}
		}()
		NewSystemResourceIdentifierMustCreate(invalidSri)
	})

	test.Run("String/Valid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputSri    string
			expectedStr string
		}{
			{"sri://1:account/120", "sri://1:account/120"},
			{"sri://10:database/myDb", "sri://10:database/myDb"},
		}

		for _, testCase := range testCaseStructs {
			sri := NewSystemResourceIdentifierMustCreate(testCase.inputSri)
			actualStr := sri.String()
			if actualStr != testCase.expectedStr {
				subtest.Errorf(
					"UnexpectedStringValueForSri: expected %s, got %s",
					testCase.expectedStr, actualStr,
				)
			}
		}
	})

	test.Run("ReadAccountId/Valid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputSri        string
			expectedAccount string
		}{
			{"sri://1:account/120", "1"},
			{"sri://10:database/myDb", "10"},
			{"sri://1000:ssl/*", "1000"},
		}

		for _, testCase := range testCaseStructs {
			sri := NewSystemResourceIdentifierMustCreate(testCase.inputSri)
			accountId, err := sri.ReadAccountId()
			if err != nil {
				subtest.Errorf("UnexpectedErrorReadingAccountIdForSri")
			}
			if accountId.String() != testCase.expectedAccount {
				subtest.Errorf(
					"UnexpectedAccountIdForSri: expected %s, got %s",
					testCase.expectedAccount, accountId.String(),
				)
			}
		}
	})

	test.Run("ReadAccountId/Invalid", func(subtest *testing.T) {
		invalidSri := "invalid"
		sri := SystemResourceIdentifier(invalidSri)
		_, err := sri.ReadAccountId()
		if err == nil {
			subtest.Errorf("MissingExpectedErrorForInvalidSri")
		}
	})

	test.Run("ReadResourceType/Valid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputSri             string
			expectedResourceType string
		}{
			{"sri://1:account/120", "account"},
			{"sri://10:database/myDb", "database"},
			{"sri://1000:ssl/*", "ssl"},
		}

		for _, testCase := range testCaseStructs {
			sri := NewSystemResourceIdentifierMustCreate(testCase.inputSri)
			resourceType, err := sri.ReadResourceType()
			if err != nil {
				subtest.Errorf("UnexpectedErrorReadingResourceTypeForSri")
			}
			if resourceType.String() != testCase.expectedResourceType {
				subtest.Errorf(
					"UnexpectedResourceTypeForSri: expected %s, got %s",
					testCase.expectedResourceType, resourceType.String(),
				)
			}
		}
	})

	test.Run("ReadResourceType/Invalid", func(subtest *testing.T) {
		invalidSri := "invalid"
		sri := SystemResourceIdentifier(invalidSri)
		_, err := sri.ReadResourceType()
		if err == nil {
			subtest.Errorf("MissingExpectedErrorForInvalidSri")
		}
	})

	test.Run("ReadResourceId/Valid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputSri           string
			expectedResourceId string
		}{
			{"sri://1:account/120", "120"},
			{"sri://10:database/myDb", "myDb"},
			{"sri://1000:ssl/*", "*"},
		}

		for _, testCase := range testCaseStructs {
			sri := NewSystemResourceIdentifierMustCreate(testCase.inputSri)
			resourceId, err := sri.ReadResourceId()
			if err != nil {
				subtest.Errorf("UnexpectedErrorReadingResourceIdForSri")
			}
			if resourceId.String() != testCase.expectedResourceId {
				subtest.Errorf(
					"UnexpectedResourceIdForSri: expected %s, got %s",
					testCase.expectedResourceId, resourceId.String(),
				)
			}
		}
	})

	test.Run("ReadResourceId/Invalid", func(subtest *testing.T) {
		invalidSri := "invalid"
		sri := SystemResourceIdentifier(invalidSri)
		_, err := sri.ReadResourceId()
		if err == nil {
			subtest.Errorf("MissingExpectedErrorForInvalidSri")
		}
	})

	test.Run("NewSriAccount/Valid", func(subtest *testing.T) {
		testCaseStructs := []struct {
			inputAccountId string
			expectedSri    string
		}{
			{"1", "sri://0:account/1"},
			{"123", "sri://0:account/123"},
			{"1000", "sri://0:account/1000"},
		}

		for _, testCase := range testCaseStructs {
			accountId, err := NewAccountId(testCase.inputAccountId)
			if err != nil {
				subtest.Errorf("UnexpectedErrorCreatingAccountIdForInput")
			}
			sri := NewSriAccount(accountId)
			if sri.String() != testCase.expectedSri {
				subtest.Errorf(
					"UnexpectedSriAccountForAccountId: expected %s, got %s",
					testCase.expectedSri, sri.String(),
				)
			}
		}
	})
}
