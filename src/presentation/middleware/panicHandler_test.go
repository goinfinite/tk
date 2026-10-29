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

func TestIsOperatorTrustworthy(test *testing.T) {
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
		test.Run(testCase.name, func(subtest *testing.T) {
			subtest.Setenv(tkInfra.TrustedIpsEnvVarName, testCase.trustedIpsEnv)

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
			httpRequest.RemoteAddr = testCase.requestIp + ":12345"
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

			isOperatorTrustworthy := isOperatorTrustworthy(echoContext)
			if isOperatorTrustworthy != testCase.shouldBeTrustworthy {
				subtest.Errorf(
					"OperatorTrustworthinessMismatch: Expected=%v, Actual=%v",
					testCase.shouldBeTrustworthy, isOperatorTrustworthy,
				)
			}
		})
	}
}

func TestLogPanic(test *testing.T) {
	tempDir := test.TempDir()
	test.Chdir(tempDir)

	fileClerk := tkInfra.FileClerk{}
	panicReport := &PanicReport{
		StackTrace:        "test stack trace",
		RequestUri:        "/test",
		OperatorIpAddress: "192.168.1.1",
	}

	logPanic(panicReport)

	logFilePath := filepath.Join("logs", "panic.log")
	if !fileClerk.FileExists(logFilePath) {
		test.Errorf("LogFileCreationFailed: %s", logFilePath)
		return
	}

	logFileInfo, statErr := os.Stat(logFilePath)
	if statErr != nil {
		test.Errorf("LogFileStatFailed: %v", statErr)
		return
	}
	if logFileInfo.Mode().Perm() != 0600 {
		test.Errorf(
			"LogFilePermissionMismatch: Expected=0600, Actual=%o",
			logFileInfo.Mode().Perm(),
		)
	}

	logFileContent, err := fileClerk.ReadFileContent(logFilePath, nil)
	if err != nil {
		test.Errorf("LogFileReadFailed: %v", err)
		return
	}

	if !strings.Contains(logFileContent, "test stack trace") {
		test.Errorf("StackTraceNotLogged: ExpectedContainTestStackTrace")
	}
	if !strings.Contains(logFileContent, "/test") {
		test.Errorf("RequestUriNotLogged: ExpectedContainTestUri")
	}
	if !strings.Contains(logFileContent, "192.168.1.1") {
		test.Errorf("OperatorIpAddressNotLogged: ExpectedContainOperatorIp")
	}
	if !strings.Contains(logFileContent, "UnknownError") {
		test.Errorf("ErrorMessageNotLogged: ExpectedContainUnknownError")
	}
}

func TestLogPanicSecuresExistingFile(test *testing.T) {
	tempDir := test.TempDir()
	test.Chdir(tempDir)

	logsDir := filepath.Join(tempDir, "logs")
	err := os.Mkdir(logsDir, 0755)
	if err != nil {
		test.Fatalf("LogDirectoryCreationFailed: %v", err)
	}
	logFilePath := filepath.Join(logsDir, "panic.log")
	err = os.WriteFile(logFilePath, []byte("existing"), 0644)
	if err != nil {
		test.Fatalf("LogFileCreationFailed: %v", err)
	}

	logPanic(&PanicReport{})

	logFileInfo, err := os.Stat(logFilePath)
	if err != nil {
		test.Fatalf("LogFileStatFailed: %v", err)
	}
	if logFileInfo.Mode().Perm() != 0600 {
		test.Errorf(
			"LogFilePermissionMismatch: Expected=0600, Actual=%o",
			logFileInfo.Mode().Perm(),
		)
	}
}

func TestApiHandlePanic(test *testing.T) {
	testCaseStructs := []struct {
		name                    string
		panicValue              any
		requestUri              string
		operatorIp              string
		forwardedFor            string
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
		{
			name:                    "TrustedProxyWithUntrustedClient",
			panicValue:              "database connection failed",
			requestUri:              "/api/test",
			operatorIp:              "192.168.1.1",
			forwardedFor:            "203.0.113.5",
			trustedIpsEnv:           "192.168.1.1",
			expectedStatus:          http.StatusInternalServerError,
			shouldExpectFullDetails: false,
		},
	}

	for _, testCase := range testCaseStructs {
		test.Run(testCase.name, func(subtest *testing.T) {
			subtest.Setenv(tkInfra.TrustedIpsEnvVarName, testCase.trustedIpsEnv)

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(http.MethodGet, testCase.requestUri, nil)
			httpRequest.RemoteAddr = testCase.operatorIp + ":12345"
			if testCase.forwardedFor != "" {
				httpRequest.Header.Set("X-Forwarded-For", testCase.forwardedFor)
			}
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

			func() {
				defer apiHandlePanic(echoContext)
				panic(testCase.panicValue)
			}()

			if httpRecorder.Code != testCase.expectedStatus {
				subtest.Errorf(
					"HttpStatusMismatch: Expected=%d, Actual=%d",
					testCase.expectedStatus, httpRecorder.Code,
				)
				return
			}

			var apiResponse map[string]any
			err := json.Unmarshal(httpRecorder.Body.Bytes(), &apiResponse)
			if err != nil {
				subtest.Errorf("ResponseBodyUnmarshalFailed: %v", err)
				return
			}

			if apiResponse["readableMessage"] != "SomethingWentWrong" {
				subtest.Errorf(
					"ReadableMessageMismatch: Expected='SomethingWentWrong', Actual='%v'",
					apiResponse["readableMessage"],
				)
			}

			responseBody, ok := apiResponse["body"].(map[string]any)
			if !ok {
				subtest.Errorf("ResponseBodyInvalidFormat: ExpectedMap")
				return
			}

			if testCase.shouldExpectFullDetails {
				if responseBody["uri"] != testCase.requestUri {
					subtest.Errorf(
						"RequestUriMismatch: Expected='%s', Actual='%v'",
						testCase.requestUri, responseBody["uri"],
					)
				}
				if responseBody["exceptionTrace"] == nil {
					subtest.Errorf("ExceptionTraceMissing: ExpectedNonNil")
				}
			}
			if !testCase.shouldExpectFullDetails {
				if responseBody["exceptionCode"] == nil {
					subtest.Errorf("ExceptionCodeMissing: ExpectedNonNil")
				}
				if len(responseBody["exceptionCode"].(string)) > panicHandlerMaxErrorLength {
					subtest.Errorf("ExceptionCodeTooLong: ExpectedTruncated")
				}
				if responseBody["uri"] != nil {
					subtest.Errorf("RequestUriLeaked: ExpectedRedacted")
				}
				if responseBody["queryParams"] != nil {
					subtest.Errorf("QueryParamsLeaked: ExpectedRedacted")
				}
				if responseBody["exceptionTrace"] != nil {
					subtest.Errorf("ExceptionTraceLeaked: ExpectedRedacted")
				}
			}
		})
	}
}

func TestApiPanicHandler(test *testing.T) {
	test.Setenv(tkInfra.TrustedIpsEnvVarName, "")

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	panicHandler := ApiPanicHandler(func(_ echo.Context) error {
		panic("test panic")
	})

	err := panicHandler(echoContext)
	if err != nil {
		test.Errorf("PanicHandlerErrorUnexpected: %v", err)
	}

	if httpRecorder.Code != http.StatusInternalServerError {
		test.Errorf(
			"HttpStatusMismatch: Expected=%d, Actual=%d",
			http.StatusInternalServerError, httpRecorder.Code,
		)
	}
}

func TestCliPanicHandlerExitCode(test *testing.T) {
	tempDir := test.TempDir()
	workingDir, err := filepath.Abs(tempDir)
	if err != nil {
		test.Errorf("WorkingDirectoryResolutionFailed: %v", err)
		return
	}

	goModContent := `module testCliPanicHandler

go 1.25.3

require github.com/goinfinite/tk v0.1.2
`
	fileClerk := tkInfra.FileClerk{}
	err = os.WriteFile(workingDir+"/go.mod", []byte(goModContent), 0644)
	if err != nil {
		test.Errorf("GoModFileCreationFailed: %v", err)
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
	err = os.WriteFile(programFilePath, []byte(mainContent), 0644)
	if err != nil {
		test.Errorf("TestProgramFileCreationFailed: %v", err)
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
			test.Errorf("GoModTidyFailed: %v", err)
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
		test.Errorf("ProgramExitUnexpected: ExpectedExitWithCode1")
		return
	default:
		test.Errorf("CommandExecutionFailed: %v", err)
		return
	}

	if exitCode != 1 {
		test.Errorf(
			"ExitCodeMismatch: Expected=1, Actual=%d",
			exitCode,
		)
	}
}
