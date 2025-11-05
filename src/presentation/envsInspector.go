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

	envFileHandler, err := os.OpenFile(
		envFilePathStr, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0400,
	)
	if err != nil {
		return errors.New("EnvsInspectorEnvOpenFileError")
	}
	defer envFileHandler.Close()

	err = godotenv.Load(envFilePathStr)
	if err != nil {
		return errors.New("EnvsInspectorEnvLoadError: " + err.Error())
	}

	missingRequiredEnvVars := []string{}
	synthesizer := tkInfra.Synthesizer{}
	for _, envVarName := range envsInspector.requiredEnvVars {
		envVarValue := os.Getenv(envVarName)
		if envVarValue != "" {
			continue
		}

		if !slices.Contains(envsInspector.autoFillableEnvVars, envVarName) {
			missingRequiredEnvVars = append(missingRequiredEnvVars, envVarName)
			continue
		}

		envVarValue = synthesizer.PasswordFactory(32, true)

		_, err = envFileHandler.WriteString(envVarName + "=" + envVarValue + "\n")
		if err != nil {
			return errors.New("EnvsInspectorEnvWriteFileError")
		}

		os.Setenv(envVarName, envVarValue)
	}

	if len(missingRequiredEnvVars) > 0 {
		return errors.New(
			"EnvsInspectorMissingRequiredEnvVars: " + strings.Join(missingRequiredEnvVars, ", "),
		)
	}

	return nil
}
