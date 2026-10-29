package tkPresentation

import (
	"errors"
	"fmt"
	"log/slog"
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

var ErrEnvsInspectorEnvLoad = errors.New("InvalidEnvFileContent")

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
				slog.Error(
					"EnvsInspectorEnvFilePathResolutionError",
					slog.String("error", err.Error()),
				)
				return errors.New("RetrieveEnvFilePathFailed")
			}
		}
		envFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(
			rawEnvFilePath, false,
		)
		if err != nil {
			slog.Error(
				"EnvsInspectorInvalidEnvFilePath",
				slog.String("rawEnvFilePath", rawEnvFilePath),
				slog.String("error", err.Error()),
			)
			return errors.New("InvalidEnvFilePath")
		}
		envsInspector.envFilePath = &envFilePath
	}
	envFilePathStr := envsInspector.envFilePath.String()

	fileClerk := tkInfra.FileClerk{}
	err = fileClerk.TouchFile(envFilePathStr)
	if err != nil {
		slog.Error(
			"EnvsInspectorEnvTouchFileError",
			slog.String("envFilePath", envFilePathStr),
			slog.String("error", err.Error()),
		)
		return errors.New("EnvsInspectorEnvTouchFileError")
	}

	envFileWritePermissions := os.FileMode(0600)
	err = fileClerk.UpdateFilePermissions(
		envFilePathStr, &envFileWritePermissions,
	)
	if err != nil {
		slog.Error(
			"EnvsInspectorEnvUpdateFileWritePermissionsError",
			slog.String("envFilePath", envFilePathStr),
			slog.String("permissions", envFileWritePermissions.String()),
			slog.String("error", err.Error()),
		)
		return errors.New(
			"EnvsInspectorEnvUpdateFileWritePermissionsError",
		)
	}

	err = godotenv.Load(envFilePathStr)
	if err != nil {
		slog.Error(
			"EnvsInspectorEnvLoadError",
			slog.String("envFilePath", envFilePathStr),
			slog.String("error", ErrEnvsInspectorEnvLoad.Error()),
		)
		return fmt.Errorf(
			"EnvsInspectorEnvLoadError: %w", ErrEnvsInspectorEnvLoad,
		)
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

		envVarValue, err = tkInfra.NewCypherSecretKey()
		if err != nil {
			slog.Error(
				"EnvsInspectorSecretKeyGenerationError",
				slog.String("envVarName", envVarName),
				slog.String("error", err.Error()),
			)
			return errors.New("EnvsInspectorSecretKeyGenerationError")
		}
		envVarStr := envVarName + "=" + envVarValue + "\n"

		envFileContent, readErr := fileClerk.ReadFileContent(
			envFilePathStr, nil,
		)
		if readErr != nil {
			slog.Error(
				"EnvsInspectorEnvReadFileError",
				slog.String("envFilePath", envFilePathStr),
				slog.String("error", readErr.Error()),
			)
			return errors.New("EnvsInspectorEnvReadFileError")
		}
		separator := ""
		if len(envFileContent) > 0 && !strings.HasSuffix(envFileContent, "\n") {
			separator = "\n"
		}

		err = fileClerk.AppendFileContent(
			tkInfra.FileAppendSettings{
				FilePath: *envsInspector.envFilePath,
			},
			separator+envVarStr,
		)
		if err != nil {
			slog.Error(
				"EnvsInspectorEnvAppendFileError",
				slog.String("envFilePath", envFilePathStr),
				slog.String("envVarName", envVarName),
				slog.String("error", err.Error()),
			)
			return errors.New("EnvsInspectorEnvAppendFileError")
		}

		setenvErr := os.Setenv(envVarName, envVarValue)
		if setenvErr != nil {
			slog.Error(
				"EnvsInspectorEnvSetError",
				slog.String("envVarName", envVarName),
				slog.String("error", setenvErr.Error()),
			)
			return errors.New("EnvsInspectorEnvSetError")
		}
	}

	envFileReadOnlyPermissions := os.FileMode(0400)
	err = fileClerk.UpdateFilePermissions(
		envFilePathStr, &envFileReadOnlyPermissions,
	)
	if err != nil {
		slog.Error(
			"EnvsInspectorEnvUpdateFileReadOnlyPermissionsError",
			slog.String("envFilePath", envFilePathStr),
			slog.String("permissions", envFileReadOnlyPermissions.String()),
			slog.String("error", err.Error()),
		)
		return errors.New(
			"EnvsInspectorEnvUpdateFileReadOnlyPermissionsError",
		)
	}

	if len(missingRequiredEnvVars) > 0 {
		slog.Error(
			"EnvsInspectorMissingRequiredEnvVars",
			slog.String(
				"missingEnvVars", strings.Join(missingRequiredEnvVars, ","),
			),
		)
		return errors.New("EnvsInspectorMissingRequiredEnvVars")
	}

	return nil
}
