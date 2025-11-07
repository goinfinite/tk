package tkValueObject

import "testing"

func TestNewSystemResourceIdentifier(t *testing.T) {
	t.Run("ValidSystemResourceIdentifier", func(t *testing.T) {
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
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("InvalidSystemResourceIdentifier", func(t *testing.T) {
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
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})
}
