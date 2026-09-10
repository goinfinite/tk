package tkInfraDb

import (
	"errors"
	"testing"
)

// Note: Setup/teardown are intentionally inline — test independence
// requires each file to own its preconditions, even if it duplicates code.

func TestNewTransientDatabaseService(t *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		t.Fatalf("UnexpectedError: '%s'", err.Error())
	}
	if dbSvc == nil {
		t.Fatal("ServiceIsNil")
	}
	if dbSvc.Handler == nil {
		t.Fatal("HandlerIsNil")
	}
}

func TestTransientDatabaseServiceSet(t *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		t.Fatalf("SetupFailed: '%s'", err.Error())
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
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.presetValuePtr != nil {
				err := dbSvc.Set(testCase.key, *testCase.presetValuePtr)
				if err != nil {
					t.Fatalf("PresetFailed: '%s'", err.Error())
				}
			}

			err := dbSvc.Set(testCase.key, testCase.valueToSet)
			if err != nil {
				t.Fatalf("SetFailed: '%s'", err.Error())
			}

			readValue, err := dbSvc.Read(testCase.key)
			if err != nil {
				t.Fatalf("ReadFailed: '%s'", err.Error())
			}
			if readValue != testCase.expectedReadValue {
				t.Fatalf(
					"UnexpectedValue: '%s' vs '%s'",
					readValue, testCase.expectedReadValue,
				)
			}
		})
	}
}

func TestTransientDatabaseServiceRead(t *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		t.Fatalf("SetupFailed: '%s'", err.Error())
	}

	existingKey := "readExistingKey"
	err = dbSvc.Set(existingKey, "existingValue")
	if err != nil {
		t.Fatalf("SetupSetFailed: '%s'", err.Error())
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
		t.Run(testCase.name, func(t *testing.T) {
			readValue, err := dbSvc.Read(testCase.key)
			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					t.Fatalf(
						"ExpectedError: '%s', got: '%v'",
						testCase.expectedError.Error(), err,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if readValue != testCase.expectedValue {
				t.Fatalf(
					"UnexpectedValue: '%s' vs '%s'",
					readValue, testCase.expectedValue,
				)
			}
		})
	}
}

func TestTransientDatabaseServiceHas(t *testing.T) {
	dbSvc, err := NewTransientDatabaseService()
	if err != nil {
		t.Fatalf("SetupFailed: '%s'", err.Error())
	}

	existingKey := "hasExistingKey"
	err = dbSvc.Set(existingKey, "existingValue")
	if err != nil {
		t.Fatalf("SetupSetFailed: '%s'", err.Error())
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
		t.Run(testCase.name, func(t *testing.T) {
			hasKey := dbSvc.Has(testCase.key)
			if hasKey != testCase.expectedValue {
				t.Fatalf("UnexpectedResult: '%t' vs '%t'", hasKey, testCase.expectedValue)
			}
		})
	}
}
