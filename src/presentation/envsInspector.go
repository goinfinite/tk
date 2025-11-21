package tkPresentation

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
	"github.com/joho/godotenv"
)

const (
	EnvsInspectorEnvFilePathEnvVarName string = "ENV_FILE_PATH"
)

type EnvsInspector struct {
	envFilePath         *tkValueObject.UnixAbsoluteFilePath
	requiredEnvVars     []string
	autoFillableEnvVars []string
}

func NewEnvsInspector(
	envFilePath *tkValueObject.UnixAbsoluteFilePath,
	requiredEnvVars, autoFillableEnvVars []string,
) *EnvsInspector {
	return &EnvsInspector{
		envFilePath:         envFilePath,
		requiredEnvVars:     requiredEnvVars,
		autoFillableEnvVars: autoFillableEnvVars,
	}
}

func (envsInspector *EnvsInspector) Inspect() (err error) {
	if envsInspector.envFilePath == nil {
		rawEnvFilePath := os.Getenv(EnvsInspectorEnvFilePathEnvVarName)
		if rawEnvFilePath == "" {
			rawEnvFilePath, err = filepath.Abs(".env")
			if err != nil {
				return errors.New("RetrieveEnvFilePathFailed")
			}
		}
		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawEnvFilePath, false)
		if err != nil {
			return errors.New("InvalidEnvFilePath")
		}
		envsInspector.envFilePath = &envFilePath
	}
	envFilePathStr := envsInspector.envFilePath.String()

	fileClerk := tkInfra.FileClerk{}
	if !fileClerk.FileExists(envFilePathStr) {
		err = fileClerk.CreateFile(envFilePathStr)
		if err != nil {
			return errors.New("EnvsInspectorCreateEnvFileError: " + err.Error())
		}
	}
	envFileWritePermissions := int(0600)
	err = fileClerk.UpdateFilePermissions(envFilePathStr, &envFileWritePermissions)
	if err != nil {
		return errors.New("EnvsInspectorUpdateEnvFileWritePermissionsError: " + err.Error())
	}

	err = godotenv.Load(envFilePathStr)
	if err != nil {
		return errors.New("EnvsInspectorLoadEnvFileError: " + err.Error())
	}

	missingRequiredEnvVars := []string{}
	for _, envVarName := range envsInspector.requiredEnvVars {
		envVarValue := os.Getenv(envVarName)
		if envVarValue != "" {
			continue
		}

		if !slices.Contains(envsInspector.autoFillableEnvVars, envVarName) {
			missingRequiredEnvVars = append(missingRequiredEnvVars, envVarName)
			continue
		}
	}

	for _, envVarName := range envsInspector.autoFillableEnvVars {
		if os.Getenv(envVarName) != "" {
			continue
		}

		cryptographicallySecureSecretKey, err := tkInfra.NewCypherSecretKey()
		if err != nil {
			return errors.New("EnvsInspectorCypherSecretKeyCreationError: " + err.Error())
		}

		envVarStr := envVarName + "=" + cryptographicallySecureSecretKey + "\n"
		err = fileClerk.UpdateFileContent(envFilePathStr, envVarStr, false)
		if err != nil {
			return errors.New("EnvsInspectorWriteEnvFileError: " + err.Error())
		}

		os.Setenv(envVarName, cryptographicallySecureSecretKey)
	}

	envFileReadOnlyPermissions := int(0400)
	err = fileClerk.UpdateFilePermissions(envFilePathStr, &envFileReadOnlyPermissions)
	if err != nil {
		return errors.New("EnvsInspectorUpdateEnvFileReadOnlyPermissionsError: " + err.Error())
	}

	if len(missingRequiredEnvVars) > 0 {
		return errors.New(
			"EnvsInspectorMissingRequiredEnvVars: " + strings.Join(missingRequiredEnvVars, ", "),
		)
	}

	return nil
}
