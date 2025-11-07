package tkInfraDb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTrailDatabaseService(t *testing.T) {
	t.Run("EnvVarNotSet", func(t *testing.T) {
		originalValue := os.Getenv(TrailDatabaseFilePathEnvVarName)
		defer func() { os.Setenv(TrailDatabaseFilePathEnvVarName, originalValue) }()
		os.Unsetenv(TrailDatabaseFilePathEnvVarName)

		_, err := NewTrailDatabaseService([]any{})
		if err == nil {
			t.Errorf("MissingExpectedError: %s", errTrailDatabaseFilePathNotSet)
		}
		if err != nil && err.Error() != errTrailDatabaseFilePathNotSet {
			t.Errorf("UnexpectedErrorMessage: '%s' vs '%s'", err.Error(), errTrailDatabaseFilePathNotSet)
		}
	})

	t.Run("InvalidPath", func(t *testing.T) {
		longPath := "/tmp/" + strings.Repeat("a", 4100) + ".db"

		originalValue := os.Getenv(TrailDatabaseFilePathEnvVarName)
		defer func() { os.Setenv(TrailDatabaseFilePathEnvVarName, originalValue) }()
		os.Setenv(TrailDatabaseFilePathEnvVarName, longPath)

		_, err := NewTrailDatabaseService([]any{})
		if err == nil {
			t.Errorf("MissingExpectedError: %s", errTrailDatabaseFilePathNotValid)
		}
		if err != nil && err.Error() != errTrailDatabaseFilePathNotValid {
			t.Errorf("UnexpectedErrorMessage: '%s' vs '%s'", err.Error(), errTrailDatabaseFilePathNotValid)
		}
	})

	t.Run("ValidPath", func(t *testing.T) {
		tempDir := t.TempDir()
		dbFilePath := filepath.Join(tempDir, "trail.db")

		originalValue := os.Getenv(TrailDatabaseFilePathEnvVarName)
		defer func() { os.Setenv(TrailDatabaseFilePathEnvVarName, originalValue) }()
		os.Setenv(TrailDatabaseFilePathEnvVarName, dbFilePath)

		dbSvc, err := NewTrailDatabaseService([]any{})
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}
		if dbSvc == nil {
			t.Errorf("ServiceIsNil")
		}
		if dbSvc != nil && dbSvc.Handler == nil {
			t.Errorf("HandlerIsNil")
		}

		if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
			t.Errorf("DatabaseFileNotCreated: %s", dbFilePath)
		}
	})
}

func TestDbMigrate(t *testing.T) {
	type TestValidModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"column:name"`
	}

	testCases := []struct {
		name        string
		extraModels []any
		expectError bool
	}{
		{
			name:        "NoExtraModels",
			extraModels: []any{},
			expectError: false,
		},
		{
			name:        "WithValidExtraModel",
			extraModels: []any{&TestValidModel{}},
			expectError: false,
		},
		{
			name:        "WithInvalidExtraModel",
			extraModels: []any{"invalid_string"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			dbFilePath := filepath.Join(tempDir, "trail.db")

			originalValue := os.Getenv(TrailDatabaseFilePathEnvVarName)
			defer func() { os.Setenv(TrailDatabaseFilePathEnvVarName, originalValue) }()
			os.Setenv(TrailDatabaseFilePathEnvVarName, dbFilePath)

			dbSvc, err := NewTrailDatabaseService([]any{})
			if err != nil {
				t.Errorf("SetupFailed: '%s'", err.Error())
				return
			}

			err = dbSvc.dbMigrate(tc.extraModels)
			if tc.expectError && err == nil {
				t.Errorf("MissingExpectedError")
			}
			if !tc.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s'", err.Error())
			}
		})
	}
}
