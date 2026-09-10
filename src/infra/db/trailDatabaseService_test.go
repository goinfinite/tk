package tkInfraDb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Note: Setup/teardown are intentionally inline — test independence
// requires each file to own its preconditions, even if it duplicates code.

func TestNewTrailDatabaseService(t *testing.T) {
	unixAbsoluteFilePathMaxLength := 4096
	overlongFilePath := "/tmp/" + strings.Repeat(
		"a", unixAbsoluteFilePathMaxLength+1,
	) + ".db"

	testCases := []struct {
		name              string
		envValue          string
		shouldUseTempPath bool
		expectedError     string
	}{
		{
			name:          "EnvVarEmpty",
			envValue:      "",
			expectedError: errTrailDatabaseFilePathNotSet,
		},
		{
			name:          "OverlongPath",
			envValue:      overlongFilePath,
			expectedError: errTrailDatabaseFilePathNotValid,
		},
		{
			name:              "ValidPath",
			shouldUseTempPath: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dbFilePath := filepath.Join(t.TempDir(), "trail.db")
			envValue := testCase.envValue
			if testCase.shouldUseTempPath {
				envValue = dbFilePath
			}
			t.Setenv(TrailDatabaseFilePathEnvVarName, envValue)

			dbSvc, err := NewTrailDatabaseService([]any{})
			if testCase.expectedError != "" {
				if err == nil {
					t.Fatalf("MissingExpectedError: %s", testCase.expectedError)
				}
				if err.Error() != testCase.expectedError {
					t.Fatalf(
						"UnexpectedErrorMessage: '%s' vs '%s'",
						err.Error(), testCase.expectedError,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if dbSvc == nil {
				t.Fatal("ServiceIsNil")
			}
			if dbSvc.Handler == nil {
				t.Fatal("HandlerIsNil")
			}

			_, statErr := os.Stat(dbFilePath)
			if statErr != nil {
				t.Fatalf("DatabaseFileNotCreated: %s", dbFilePath)
			}
		})
	}
}

func TestNewTrailDatabaseServiceWithExtraModels(t *testing.T) {
	type testValidModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"column:name"`
	}

	testCases := []struct {
		name             string
		extraModels      []any
		expectedError    string
		expectedTablePtr any
	}{
		{
			name:        "NoExtraModels",
			extraModels: []any{},
		},
		{
			name:             "ValidExtraModel",
			extraModels:      []any{&testValidModel{}},
			expectedTablePtr: &testValidModel{},
		},
		{
			name:          "InvalidExtraModel",
			extraModels:   []any{"invalid_string"},
			expectedError: errTrailDatabaseMigrationError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			dbFilePath := filepath.Join(t.TempDir(), "trail.db")
			t.Setenv(TrailDatabaseFilePathEnvVarName, dbFilePath)

			dbSvc, err := NewTrailDatabaseService(testCase.extraModels)
			if testCase.expectedError != "" {
				if err == nil {
					t.Fatalf("MissingExpectedError: %s", testCase.expectedError)
				}
				if !strings.HasPrefix(err.Error(), testCase.expectedError) {
					t.Fatalf(
						"UnexpectedErrorMessage: '%s' vs '%s'",
						err.Error(), testCase.expectedError,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}

			if testCase.expectedTablePtr == nil {
				return
			}
			if !dbSvc.Handler.Migrator().HasTable(testCase.expectedTablePtr) {
				t.Fatalf("ExpectedTableMissing: %T", testCase.expectedTablePtr)
			}
		})
	}
}
