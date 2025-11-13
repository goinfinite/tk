package tkPresentationMiddleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestIsOperatorTrustworthy(t *testing.T) {
	originalEnv := os.Getenv(tkInfra.TrustedIpsEnvVarName)
	defer os.Setenv(tkInfra.TrustedIpsEnvVarName, originalEnv)

	testCaseStructs := []struct {
		name                string
		trustedIpsEnv       string
		requestIp           string
		shouldBeTrustworthy bool
	}{
		{
			name:                "NoTrustedIpsEnv",
			trustedIpsEnv:       "",
			requestIp:           "192.168.1.1",
			shouldBeTrustworthy: false,
		},
		{
			name:                "TrustedIpMatches",
			trustedIpsEnv:       "192.168.1.1,10.0.0.1",
			requestIp:           "192.168.1.1",
			shouldBeTrustworthy: true,
		},
		{
			name:                "TrustedIpDoesNotMatch",
			trustedIpsEnv:       "192.168.1.2,10.0.0.1",
			requestIp:           "192.168.1.1",
			shouldBeTrustworthy: false,
		},
		{
			name:                "EmptyRequestIp",
			trustedIpsEnv:       "192.168.1.1",
			requestIp:           "",
			shouldBeTrustworthy: false,
		},
		{
			name:                "InvalidTrustedIp",
			trustedIpsEnv:       "invalid,192.168.1.1",
			requestIp:           "192.168.1.1",
			shouldBeTrustworthy: true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			os.Setenv(tkInfra.TrustedIpsEnvVarName, testCase.trustedIpsEnv)

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
			httpRequest.RemoteAddr = testCase.requestIp + ":12345"
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

			isOperatorTrustworthy := isOperatorTrustworthy(echoContext)
			if isOperatorTrustworthy != testCase.shouldBeTrustworthy {
				t.Errorf(
					"OperatorTrustworthinessMismatch: Expected=%v, Actual=%v",
					testCase.shouldBeTrustworthy, isOperatorTrustworthy,
				)
			}
		})
	}
}

func TestLogPanic(t *testing.T) {
	tempDir := t.TempDir()
	originalWorkingDir, _ := os.Getwd()
	defer os.Chdir(originalWorkingDir)
	os.Chdir(tempDir)

	fileClerk := tkInfra.FileClerk{}
	panicReport := &PanicReport{
		StackTrace:        "test stack trace",
		RequestUri:        "/test",
		OperatorIpAddress: "192.168.1.1",
	}

	logPanic(panicReport)

	logFilePath := filepath.Join("logs", "panic.log")
	if !fileClerk.FileExists(logFilePath) {
		t.Errorf("LogFileCreationFailed: %s", logFilePath)
		return
	}

	logFileContent, err := fileClerk.ReadFileContent(logFilePath, nil)
	if err != nil {
		t.Errorf("LogFileReadFailed: %v", err)
		return
	}

	if !strings.Contains(logFileContent, "test stack trace") {
		t.Errorf("StackTraceNotLogged: ExpectedContainTestStackTrace")
	}
	if !strings.Contains(logFileContent, "/test") {
		t.Errorf("RequestUriNotLogged: ExpectedContainTestUri")
	}
	if !strings.Contains(logFileContent, "192.168.1.1") {
		t.Errorf("OperatorIpAddressNotLogged: ExpectedContainOperatorIp")
	}
	if !strings.Contains(logFileContent, "UnknownError") {
		t.Errorf("ErrorMessageNotLogged: ExpectedContainUnknownError")
	}
}

func TestApiHandlePanic(t *testing.T) {
	originalEnv := os.Getenv(tkInfra.TrustedIpsEnvVarName)
	defer os.Setenv(tkInfra.TrustedIpsEnvVarName, originalEnv)

	testCaseStructs := []struct {
		name                    string
		panicValue              any
		requestUri              string
		operatorIp              string
		trustedIpsEnv           string
		expectedStatus          int
		shouldExpectFullDetails bool
	}{
		{
			name:                    "InfraPanic",
			panicValue:              "database connection failed",
			requestUri:              "/api/test",
			operatorIp:              "192.168.1.1",
			trustedIpsEnv:           "192.168.1.1",
			expectedStatus:          http.StatusInternalServerError,
			shouldExpectFullDetails: true,
		},
		{
			name:                    "UntrustedOperator",
			panicValue:              "database connection failed",
			requestUri:              "/api/test",
			operatorIp:              "192.168.1.2",
			trustedIpsEnv:           "192.168.1.1",
			expectedStatus:          http.StatusInternalServerError,
			shouldExpectFullDetails: false,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			os.Setenv(tkInfra.TrustedIpsEnvVarName, testCase.trustedIpsEnv)

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(http.MethodGet, testCase.requestUri, nil)
			httpRequest.RemoteAddr = testCase.operatorIp + ":12345"
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

			func() {
				defer apiHandlePanic(echoContext)
				panic(testCase.panicValue)
			}()

			if httpRecorder.Code != testCase.expectedStatus {
				t.Errorf(
					"HttpStatusMismatch: Expected=%d, Actual=%d",
					testCase.expectedStatus, httpRecorder.Code,
				)
				return
			}

			var apiResponse map[string]any
			err := json.Unmarshal(httpRecorder.Body.Bytes(), &apiResponse)
			if err != nil {
				t.Errorf("ResponseBodyUnmarshalFailed: %v", err)
				return
			}

			if apiResponse["readableMessage"] != "SomethingWentWrong" {
				t.Errorf(
					"ReadableMessageMismatch: Expected='SomethingWentWrong', Actual='%v'",
					apiResponse["readableMessage"],
				)
			}

			responseBody, ok := apiResponse["body"].(map[string]any)
			if !ok {
				t.Errorf("ResponseBodyInvalidFormat: ExpectedMap")
				return
			}

			if testCase.shouldExpectFullDetails {
				if responseBody["uri"] != testCase.requestUri {
					t.Errorf(
						"RequestUriMismatch: Expected='%s', Actual='%v'",
						testCase.requestUri, responseBody["uri"],
					)
				}
				if responseBody["exceptionTrace"] == nil {
					t.Errorf("ExceptionTraceMissing: ExpectedNonNil")
				}
			}
			if !testCase.shouldExpectFullDetails {
				if responseBody["exceptionCode"] == nil {
					t.Errorf("ExceptionCodeMissing: ExpectedNonNil")
				}
				if len(responseBody["exceptionCode"].(string)) > panicHandlerMaxErrorLength {
					t.Errorf("ExceptionCodeTooLong: ExpectedTruncated")
				}
			}
		})
	}
}

func TestApiPanicHandler(t *testing.T) {
	os.Setenv(tkInfra.TrustedIpsEnvVarName, "")

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	panicHandler := ApiPanicHandler(func(c echo.Context) error {
		panic("test panic")
	})

	err := panicHandler(echoContext)
	if err != nil {
		t.Errorf("PanicHandlerErrorUnexpected: %v", err)
	}

	if httpRecorder.Code != http.StatusInternalServerError {
		t.Errorf(
			"HttpStatusMismatch: Expected=%d, Actual=%d",
			http.StatusInternalServerError, httpRecorder.Code,
		)
	}
}

func TestCliPanicHandlerExitCode(t *testing.T) {
	tempDir := t.TempDir()
	workingDir, err := filepath.Abs(tempDir)
	if err != nil {
		t.Errorf("WorkingDirectoryResolutionFailed: %v", err)
		return
	}

	goModContent := `module testCliPanicHandler

go 1.25.3

require github.com/goinfinite/tk v0.1.2
`
	fileClerk := tkInfra.FileClerk{}
	err = fileClerk.UpdateFileContent(workingDir+"/go.mod", goModContent, true)
	if err != nil {
		t.Errorf("GoModFileCreationFailed: %v", err)
		return
	}

	mainContent := `package main

import tkPresentationMiddleware "github.com/goinfinite/tk/src/presentation/middleware"

func main() {
	defer tkPresentationMiddleware.CliPanicHandler()
	panic("test cli panic")
}
`

	programFilePath := tempDir + "/testCliPanicHandler.go"
	err = fileClerk.UpdateFileContent(programFilePath, mainContent, true)
	if err != nil {
		t.Errorf("TestProgramFileCreationFailed: %v", err)
		return
	}

	if !fileClerk.FileExists(workingDir + "/go.sum") {
		goModShell := tkInfra.NewShell(tkInfra.ShellSettings{
			Command:          "go",
			Args:             []string{"mod", "tidy"},
			WorkingDirectory: workingDir,
		})
		_, err = goModShell.Run()
		if err != nil {
			t.Errorf("GoModTidyFailed: %v", err)
			return
		}
	}

	goRunShell := tkInfra.NewShell(tkInfra.ShellSettings{
		Command:          "go",
		Args:             []string{"run", programFilePath},
		WorkingDirectory: workingDir,
		Envs:             []string{"TERM=dumb"},
	})
	_, err = goRunShell.Run()
	var exitCode int
	switch shellErr := err.(type) {
	case *tkInfra.ShellError:
		exitCode = shellErr.ExitCode
	case nil:
		t.Errorf("ProgramExitUnexpected: ExpectedExitWithCode1")
		return
	default:
		t.Errorf("CommandExecutionFailed: %v", err)
		return
	}

	if exitCode != 1 {
		t.Errorf(
			"ExitCodeMismatch: Expected=1, Actual=%d",
			exitCode,
		)
	}
}
