package tkInfraDb

import (
	"errors"
	"testing"
	"time"
)

// Note: Setup/teardown are intentionally inline — test independence
// requires each file to own its preconditions, even if it duplicates code.

func TestNewTransientDatabaseService(test *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		test.Fatalf("UnexpectedError: '%s'", err.Error())
	}
	if dbSvc == nil {
		test.Fatal("ServiceIsNil")
	}
	if dbSvc.Handler == nil {
		test.Fatal("HandlerIsNil")
	}
}

func TestTransientDatabaseServiceSet(test *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		test.Fatalf("SetupFailed: '%s'", err.Error())
	}

	presetValue := "presetValue"

	testCases := []struct {
		name              string
		key               string
		presetValuePtr    *string
		valueToSet        string
		expectedReadValue string
	}{
		{
			name:              "CreatesMissingKey",
			key:               "setCreatesMissingKey",
			valueToSet:        "createdValue",
			expectedReadValue: "createdValue",
		},
		{
			name:              "OverwritesExistingKey",
			key:               "setOverwritesExistingKey",
			presetValuePtr:    &presetValue,
			valueToSet:        "overwrittenValue",
			expectedReadValue: "overwrittenValue",
		},
		{
			name:              "AcceptsEmptyValue",
			key:               "setAcceptsEmptyValue",
			valueToSet:        "",
			expectedReadValue: "",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			if testCase.presetValuePtr != nil {
				err := dbSvc.Set(testCase.key, *testCase.presetValuePtr, nil)
				if err != nil {
					subtest.Fatalf("PresetFailed: '%s'", err.Error())
				}
			}

			err := dbSvc.Set(testCase.key, testCase.valueToSet, nil)
			if err != nil {
				subtest.Fatalf("SetFailed: '%s'", err.Error())
			}

			readValue, err := dbSvc.Read(testCase.key)
			if err != nil {
				subtest.Fatalf("ReadFailed: '%s'", err.Error())
			}
			if readValue != testCase.expectedReadValue {
				subtest.Fatalf(
					"UnexpectedValue: '%s' vs '%s'",
					readValue, testCase.expectedReadValue,
				)
			}
		})
	}
}

func TestTransientDatabaseServiceRead(test *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		test.Fatalf("SetupFailed: '%s'", err.Error())
	}

	existingKey := "readExistingKey"
	err = dbSvc.Set(existingKey, "existingValue", nil)
	if err != nil {
		test.Fatalf("SetupSetFailed: '%s'", err.Error())
	}

	testCases := []struct {
		name          string
		key           string
		expectedValue string
		expectedError error
	}{
		{
			name:          "ExistingKey",
			key:           existingKey,
			expectedValue: "existingValue",
		},
		{
			name:          "MissingKey",
			key:           "readMissingKey",
			expectedError: ErrKeyNotFound,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			readValue, err := dbSvc.Read(testCase.key)
			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					subtest.Fatalf(
						"ExpectedError: '%s', got: '%v'",
						testCase.expectedError.Error(), err,
					)
				}
				return
			}

			if err != nil {
				subtest.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if readValue != testCase.expectedValue {
				subtest.Fatalf(
					"UnexpectedValue: '%s' vs '%s'",
					readValue, testCase.expectedValue,
				)
			}
		})
	}
}

func TestTransientDatabaseServiceHas(test *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		test.Fatalf("SetupFailed: '%s'", err.Error())
	}

	existingKey := "hasExistingKey"
	err = dbSvc.Set(existingKey, "existingValue", nil)
	if err != nil {
		test.Fatalf("SetupSetFailed: '%s'", err.Error())
	}

	testCases := []struct {
		name          string
		key           string
		expectedValue bool
	}{
		{
			name:          "ExistingKey",
			key:           existingKey,
			expectedValue: true,
		},
		{
			name:          "MissingKey",
			key:           "hasMissingKey",
			expectedValue: false,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			hasKey := dbSvc.Has(testCase.key)
			if hasKey != testCase.expectedValue {
				subtest.Fatalf("UnexpectedResult: '%t' vs '%t'", hasKey, testCase.expectedValue)
			}
		})
	}
}

func TestTransientDatabaseServiceExpiration(test *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		test.Fatalf("SetupFailed: '%s'", err.Error())
	}

	expiredTtl := -time.Minute
	unexpiredTtl := time.Hour

	testCases := []struct {
		name              string
		key               string
		valueToSet        string
		ttlPtr            *time.Duration
		expectedHasKey    bool
		expectedReadValue string
		expectedError     error
	}{
		{
			name:           "ExpiredKeyIsMissing",
			key:            "expirationExpiredKey",
			valueToSet:     "expiredValue",
			ttlPtr:         &expiredTtl,
			expectedHasKey: false,
			expectedError:  ErrKeyNotFound,
		},
		{
			name:              "UnexpiredKeyIsFound",
			key:               "expirationUnexpiredKey",
			valueToSet:        "unexpiredValue",
			ttlPtr:            &unexpiredTtl,
			expectedHasKey:    true,
			expectedReadValue: "unexpiredValue",
		},
		{
			name:              "KeyWithoutTtlNeverExpires",
			key:               "expirationPersistentKey",
			valueToSet:        "persistentValue",
			expectedHasKey:    true,
			expectedReadValue: "persistentValue",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			err := dbSvc.Set(testCase.key, testCase.valueToSet, testCase.ttlPtr)
			if err != nil {
				subtest.Fatalf("SetFailed: '%s'", err.Error())
			}

			hasKey := dbSvc.Has(testCase.key)
			if hasKey != testCase.expectedHasKey {
				subtest.Fatalf(
					"UnexpectedHasResult: '%t' vs '%t'",
					hasKey, testCase.expectedHasKey,
				)
			}

			readValue, err := dbSvc.Read(testCase.key)
			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					subtest.Fatalf(
						"ExpectedError: '%s', got: '%v'",
						testCase.expectedError.Error(), err,
					)
				}
				return
			}

			if err != nil {
				subtest.Fatalf("ReadFailed: '%s'", err.Error())
			}
			if readValue != testCase.expectedReadValue {
				subtest.Fatalf(
					"UnexpectedValue: '%s' vs '%s'",
					readValue, testCase.expectedReadValue,
				)
			}
		})
	}
}

func TestTransientDatabaseServiceSetOverwritesTtl(test *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		test.Fatalf("SetupFailed: '%s'", err.Error())
	}

	expiredTtl := -time.Minute

	test.Run("SetWithoutTtlClearsPreviousTtl", func(subtest *testing.T) {
		key := "overwriteTtlClearedKey"
		err := dbSvc.Set(key, "expiredValue", &expiredTtl)
		if err != nil {
			subtest.Fatalf("PresetFailed: '%s'", err.Error())
		}

		err = dbSvc.Set(key, "persistentValue", nil)
		if err != nil {
			subtest.Fatalf("OverwriteFailed: '%s'", err.Error())
		}

		readValue, err := dbSvc.Read(key)
		if err != nil {
			subtest.Fatalf("ReadFailed: '%s'", err.Error())
		}
		if readValue != "persistentValue" {
			subtest.Fatalf("UnexpectedValue: '%s' vs 'persistentValue'", readValue)
		}
	})

	test.Run("SetWithTtlExpiresPreviousValue", func(subtest *testing.T) {
		key := "overwriteTtlAppliedKey"
		err := dbSvc.Set(key, "persistentValue", nil)
		if err != nil {
			subtest.Fatalf("PresetFailed: '%s'", err.Error())
		}

		err = dbSvc.Set(key, "expiredValue", &expiredTtl)
		if err != nil {
			subtest.Fatalf("OverwriteFailed: '%s'", err.Error())
		}

		_, err = dbSvc.Read(key)
		if !errors.Is(err, ErrKeyNotFound) {
			subtest.Fatalf("ExpectedError: '%s', got: '%v'", ErrKeyNotFound.Error(), err)
		}
	})
}
