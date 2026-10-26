package tkPresentation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestEnvsInspectorInspect(test *testing.T) {
	fileClerk := tkInfra.FileClerk{}

	test.Run("NewEnvsInspector", func(subtest *testing.T) {
		envFilePath, _ := tkValueObject.NewUnixAbsoluteFilePath("/tmp/.env", false)
		requiredEnvVars := []string{"DB_HOST", "DB_PORT"}
		autoFillableEnvVars := []string{"DB_PASSWORD"}

		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		if envsInspector.envFilePath.String() != "/tmp/.env" {
			subtest.Errorf("EnvFilePathNotSetCorrectly: %s", envsInspector.envFilePath.String())
		}

		if len(envsInspector.requiredEnvVars) != 2 || envsInspector.requiredEnvVars[0] != "DB_HOST" {
			subtest.Errorf("RequiredEnvVarsNotSetCorrectly: %v", envsInspector.requiredEnvVars)
		}

		if len(envsInspector.autoFillableEnvVars) != 1 || envsInspector.autoFillableEnvVars[0] != "DB_PASSWORD" {
			subtest.Errorf("AutoFillableEnvVarsNotSetCorrectly: %v", envsInspector.autoFillableEnvVars)
		}
	})

	test.Run("SuccessWithProvidedEnvFilePath", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost\nDB_PORT=5432\n"
		err := os.WriteFile(rawEnvFilePath, []byte(envFileContent), 0644)
		if err != nil {
			subtest.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			subtest.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_PORT"}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		err = envsInspector.Inspect()
		if err != nil {
			subtest.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		if os.Getenv("DB_HOST") != "localhost" {
			subtest.Errorf("EnvVarNotLoaded: DB_HOST != localhost")
		}
		if os.Getenv("DB_PORT") != "5432" {
			subtest.Errorf("EnvVarNotLoaded: DB_PORT != 5432")
		}
	})

	test.Run("AutoFillMissingEnvVar", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost\n"
		err := os.WriteFile(rawEnvFilePath, []byte(envFileContent), 0644)
		if err != nil {
			subtest.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			subtest.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_PASSWORD"}
		autoFillableEnvVars := []string{"DB_PASSWORD"}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		subtest.Setenv("DB_PASSWORD", "")

		err = envsInspector.Inspect()
		if err != nil {
			subtest.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		dbPassword := os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			subtest.Errorf("EnvVarNotAutoFilled: DB_PASSWORD is empty")
		}
		fileContent, err := fileClerk.ReadFileContent(rawEnvFilePath, nil)
		if err != nil {
			subtest.Fatalf("ReadEnvFileFailed: %v", err)
		}

		if !strings.Contains(fileContent, "DB_PASSWORD="+dbPassword) {
			subtest.Errorf("EnvVarNotAppendedToFile: DB_PASSWORD not found in file")
		}

		expectedPasswordLength := 43 // base64 raw URL encoding of 32 bytes
		if len(dbPassword) != expectedPasswordLength {
			subtest.Errorf("AutoFilledPasswordWrongLength: expected %d, got %d", expectedPasswordLength, len(dbPassword))
		}
	})

	test.Run("AutoFillAfterFileWithoutTrailingNewline", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost"
		err := os.WriteFile(rawEnvFilePath, []byte(envFileContent), 0644)
		if err != nil {
			subtest.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			subtest.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_PASSWORD"}
		autoFillableEnvVars := []string{"DB_PASSWORD"}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		subtest.Setenv("DB_PASSWORD", "")

		err = envsInspector.Inspect()
		if err != nil {
			subtest.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		loadedVars, loadErr := godotenv.Read(rawEnvFilePath)
		if loadErr != nil {
			subtest.Fatalf("ReloadEnvFileFailed: %v", loadErr)
		}
		if loadedVars["DB_HOST"] != "localhost" {
			subtest.Errorf("ExistingLineCorrupted: DB_HOST=%q", loadedVars["DB_HOST"])
		}
		if loadedVars["DB_PASSWORD"] == "" {
			subtest.Errorf("AutoFilledVarNotSeparatelyLoadable")
		}
	})

	test.Run("MissingRequiredEnvVar", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "DB_HOST=localhost\n"
		err := os.WriteFile(rawEnvFilePath, []byte(envFileContent), 0644)
		if err != nil {
			subtest.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			subtest.Fatalf("CreateFilePathVoFailed: %v", err)
		}

		requiredEnvVars := []string{"DB_HOST", "DB_USER"}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(&envFilePath, requiredEnvVars, autoFillableEnvVars)

		subtest.Setenv("DB_USER", "")

		err = envsInspector.Inspect()
		if err == nil {
			subtest.Errorf("InspectSucceededWhenItShouldFail: MissingRequiredEnvVar")
		}

		expectedError := "EnvsInspectorMissingRequiredEnvVars: DB_USER"
		if err.Error() != expectedError {
			subtest.Errorf("UnexpectedErrorMessage: expected '%s', got '%s'", expectedError, err.Error())
		}
	})

	test.Run("DefaultEnvFilePathFromEnvVar", func(subtest *testing.T) {
		tempDir := subtest.TempDir()
		rawEnvFilePath := filepath.Join(tempDir, ".env")
		envFileContent := "API_KEY=secret\n"
		err := os.WriteFile(rawEnvFilePath, []byte(envFileContent), 0644)
		if err != nil {
			subtest.Fatalf("CreateTestEnvFileFailed: %v", err)
		}

		subtest.Setenv(EnvsInspectorEnvFilePathEnvVarName, rawEnvFilePath)

		requiredEnvVars := []string{"API_KEY"}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(nil, requiredEnvVars, autoFillableEnvVars)

		err = envsInspector.Inspect()
		if err != nil {
			subtest.Errorf("InspectFailedWhenItShouldSucceed: %v", err)
		}

		if os.Getenv("API_KEY") != "secret" {
			subtest.Errorf("EnvVarNotLoaded: API_KEY != secret")
		}
	})

	test.Run("InvalidEnvFilePath", func(subtest *testing.T) {
		invalidPath := "invalid/path/.env"
		subtest.Setenv(EnvsInspectorEnvFilePathEnvVarName, invalidPath)

		requiredEnvVars := []string{}
		autoFillableEnvVars := []string{}
		envsInspector := NewEnvsInspector(nil, requiredEnvVars, autoFillableEnvVars)

		err := envsInspector.Inspect()
		if err == nil {
			subtest.Errorf("InspectSucceededWhenItShouldFail: InvalidEnvFilePath")
		}

		if !strings.Contains(err.Error(), "EnvsInspectorEnvTouchFileError") {
			subtest.Errorf("UnexpectedError: expected EnvsInspectorEnvTouchFileError, got %v", err)
		}
	})
}
