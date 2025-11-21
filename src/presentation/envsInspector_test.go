package tkPresentation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestEnvsInspectorInspect(t *testing.T) {
	fileClerk := tkInfra.FileClerk{}

	t.Run("NewEnvsInspector", func(t *testing.T) {
		envFilePath, _ := tkValueObject.NewUnixAbsoluteFilePath("/tmp/.env", false)
		requiredEnvVars := []string{"DB_HOST", "DB_PORT"}
		autoFillableEnvVars := []string{"DB_PASSWORD"}

		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		if envsInspector.envFilePath.String() != "/tmp/.env" {
			t.Errorf("EnvFilePathNotSetCorrectly: %s", envsInspector.envFilePath.String())
		}

		if len(envsInspector.requiredEnvVars) != 2 || envsInspector.requiredEnvVars[0] != "DB_HOST" {
			t.Errorf("RequiredEnvVarsNotSetCorrectly: %v", envsInspector.requiredEnvVars)
		}

		if len(envsInspector.autoFillableEnvVars) != 1 || envsInspector.autoFillableEnvVars[0] != "DB_PASSWORD" {
			t.Errorf("AutoFillableEnvVarsNotSetCorrectly: %v", envsInspector.autoFillableEnvVars)
		}
	})

	t.Run("SuccessWithProvidedEnvFilePath", func(t *testing.T) {
		tempDir := t.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost\nDB_PORT=5432\n"
		err := fileClerk.UpdateFileContent(rawEnvFilePath, envFileContent, true)
		if err != nil {
			t.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			t.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_PORT"}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		err = envsInspector.Inspect()
		if err != nil {
			t.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		if os.Getenv("DB_HOST") != "localhost" {
			t.Errorf("EnvVarNotLoaded: DB_HOST != localhost")
		}
		if os.Getenv("DB_PORT") != "5432" {
			t.Errorf("EnvVarNotLoaded: DB_PORT != 5432")
		}
	})

	t.Run("AutoFillMissingEnvVar", func(t *testing.T) {
		tempDir := t.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost\n"
		err := fileClerk.UpdateFileContent(rawEnvFilePath, envFileContent, true)
		if err != nil {
			t.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			t.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_PASSWORD"}
		autoFillableEnvVars := []string{"DB_PASSWORD"}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		os.Unsetenv("DB_PASSWORD")

		err = envsInspector.Inspect()
		if err != nil {
			t.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		dbPassword := os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			t.Errorf("EnvVarNotAutoFilled: DB_PASSWORD is empty")
		}
		fileContent, err := fileClerk.ReadFileContent(rawEnvFilePath, nil)
		if err != nil {
			t.Fatalf("ReadEnvFileFailed: %v", err)
		}

		if !strings.Contains(fileContent, "DB_PASSWORD="+dbPassword) {
			t.Errorf("EnvVarNotAppendedToFile: DB_PASSWORD not found in file")
		}

		if len(dbPassword) < 32 {
			t.Errorf("AutoFilledPasswordWrongLength: expected >= 32, got %d", len(dbPassword))
		}
	})

	t.Run("MissingRequiredEnvVar", func(t *testing.T) {
		tempDir := t.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost\n"
		err := fileClerk.UpdateFileContent(rawEnvFilePath, envFileContent, true)
		if err != nil {
			t.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			t.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_USER"}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		os.Unsetenv("DB_USER")

		err = envsInspector.Inspect()
		if err == nil {
			t.Errorf("InspectSucceededWhenItShouldFail: MissingRequiredEnvVar")
		}

		expectedError := "EnvsInspectorMissingRequiredEnvVars: DB_USER"
		if err.Error() != expectedError {
			t.Errorf("UnexpectedErrorMessage: expected '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("DefaultEnvFilePathFromEnvVar", func(t *testing.T) {
		tempDir := t.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "API_KEY=secret\n"
		err := fileClerk.UpdateFileContent(rawEnvFilePath, envFileContent, true)
		if err != nil {
			t.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		t.Setenv(EnvsInspectorEnvFilePathEnvVarName, rawEnvFilePath)

		requiredEnvVars := []string{"API_KEY"}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(nil, requiredEnvVars, autoFillableEnvVars)

		err = envsInspector.Inspect()
		if err != nil {
			t.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		if os.Getenv("API_KEY") != "secret" {
			t.Errorf("EnvVarNotLoaded: API_KEY != secret")
		}
	})

	t.Run("InvalidEnvFilePath", func(t *testing.T) {
		invalidPath := "invalid/path/.env"
		t.Setenv(EnvsInspectorEnvFilePathEnvVarName, invalidPath)

		requiredEnvVars := []string{}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(nil, requiredEnvVars, autoFillableEnvVars)

		err := envsInspector.Inspect()
		if err == nil {
			t.Errorf("InspectSucceededWhenItShouldFail: InvalidEnvFilePath")
		}

		if !strings.Contains(err.Error(), "EnvsInspectorEnvCreateFileError") {
			t.Errorf("UnexpectedError: expected EnvsInspectorEnvCreateFileError, got %v", err)
		}
	})
}
