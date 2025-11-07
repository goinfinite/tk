package tkPresentation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"

	tkInfra "github.com/goinfinite/tk/src/infra"
)

func TestNewApiResponseWrapper(t *testing.T) {
	testCases := []struct {
		name                    string
		status                  int
		readableMessage         string
		body                    any
		expectedStatus          int
		expectedReadableMessage string
		expectedBody            any
	}{
		{
			name:                    "SuccessResponse",
			status:                  http.StatusOK,
			readableMessage:         "OperationCompletedSuccessfully",
			body:                    map[string]string{"result": "ok"},
			expectedStatus:          http.StatusOK,
			expectedReadableMessage: "OperationCompletedSuccessfully",
			expectedBody:            map[string]string{"result": "ok"},
		},
		{
			name:                    "ErrorResponse",
			status:                  http.StatusBadRequest,
			readableMessage:         "InvalidInputProvided",
			body:                    map[string]string{"error": "validation failed"},
			expectedStatus:          http.StatusBadRequest,
			expectedReadableMessage: "InvalidInputProvided",
			expectedBody:            map[string]string{"error": "validation failed"},
		},
		{
			name:                    "NilBody",
			status:                  http.StatusNotFound,
			readableMessage:         "ResourceNotFound",
			body:                    nil,
			expectedStatus:          http.StatusNotFound,
			expectedReadableMessage: "ResourceNotFound",
			expectedBody:            nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := NewApiResponseWrapper(
				testCase.status,
				testCase.body,
				testCase.readableMessage,
			)

			if result.Status != testCase.expectedStatus {
				t.Errorf(
					"StatusMismatch: expected %d, got %d",
					testCase.expectedStatus, result.Status,
				)
			}

			if result.ReadableMessage != testCase.expectedReadableMessage {
				t.Errorf(
					"ReadableMessageMismatch: expected %s, got %s",
					testCase.expectedReadableMessage, result.ReadableMessage,
				)
			}

			actualJson, err := json.Marshal(result.Body)
			if err != nil {
				t.Fatalf("ActualJsonMarshalingFailed: %v", err)
			}

			expectedJson, err := json.Marshal(testCase.expectedBody)
			if err != nil {
				t.Fatalf("ExpectedJsonMarshalingFailed: %v", err)
			}

			if string(actualJson) != string(expectedJson) {
				t.Errorf(
					"BodyMismatch: expected '%s', got '%s'",
					string(expectedJson), string(actualJson),
				)
			}
		})
	}
}

func TestLiaisonApiResponseEmitter(t *testing.T) {
	testCases := []struct {
		name                 string
		liaisonStatus        LiaisonResponseStatus
		readableMessage      string
		body                 any
		expectedHttpStatus   int
		expectedResponseBody ApiResponseWrapper
	}{
		{
			name:               "SuccessStatus",
			liaisonStatus:      LiaisonResponseStatusSuccess,
			readableMessage:    "OperationCompletedSuccessfully",
			body:               map[string]string{"result": "ok"},
			expectedHttpStatus: http.StatusOK,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusOK,
				ReadableMessage: "OperationCompletedSuccessfully",
				Body:            map[string]string{"result": "ok"},
			},
		},
		{
			name:               "CreatedStatus",
			liaisonStatus:      LiaisonResponseStatusCreated,
			readableMessage:    "ResourceCreated",
			body:               map[string]int{"id": 456},
			expectedHttpStatus: http.StatusCreated,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusCreated,
				ReadableMessage: "ResourceCreated",
				Body:            map[string]int{"id": 456},
			},
		},
		{
			name:               "MultiStatus",
			liaisonStatus:      LiaisonResponseStatusMultiStatus,
			readableMessage:    "MultipleOperations",
			body:               []string{"op1", "op2"},
			expectedHttpStatus: http.StatusMultiStatus,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusMultiStatus,
				ReadableMessage: "MultipleOperations",
				Body:            []string{"op1", "op2"},
			},
		},
		{
			name:               "UserErrorStatus",
			liaisonStatus:      LiaisonResponseStatusUserError,
			readableMessage:    "InvalidInputProvided",
			body:               map[string]string{"error": "bad input"},
			expectedHttpStatus: http.StatusBadRequest,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusBadRequest,
				ReadableMessage: "InvalidInputProvided",
				Body:            map[string]string{"error": "bad input"},
			},
		},
		{
			name:               "UnauthorizedStatus",
			liaisonStatus:      LiaisonResponseStatusUnauthorized,
			readableMessage:    "NotAuthorized",
			body:               nil,
			expectedHttpStatus: http.StatusUnauthorized,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusUnauthorized,
				ReadableMessage: "NotAuthorized",
				Body:            nil,
			},
		},
		{
			name:               "ForbiddenStatus",
			liaisonStatus:      LiaisonResponseStatusForbidden,
			readableMessage:    "AccessForbidden",
			body:               nil,
			expectedHttpStatus: http.StatusForbidden,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusForbidden,
				ReadableMessage: "AccessForbidden",
				Body:            nil,
			},
		},
		{
			name:               "InfraErrorStatus",
			liaisonStatus:      LiaisonResponseStatusInfraError,
			readableMessage:    "InternalServerErrorCode",
			body:               map[string]string{"error": "db connection failed"},
			expectedHttpStatus: http.StatusInternalServerError,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusInternalServerError,
				ReadableMessage: "InternalServerErrorCode",
				Body:            map[string]string{"error": "db connection failed"},
			},
		},
		{
			name:               "UnknownErrorStatus",
			liaisonStatus:      LiaisonResponseStatusUnknownError,
			readableMessage:    "UnknownErrorCode",
			body:               nil,
			expectedHttpStatus: http.StatusInternalServerError,
			expectedResponseBody: ApiResponseWrapper{
				Status:          http.StatusInternalServerError,
				ReadableMessage: "UnknownErrorCode",
				Body:            nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

			liaisonResponse := NewLiaisonResponse(
				testCase.liaisonStatus, testCase.body, testCase.readableMessage,
			)

			err := LiaisonApiResponseEmitter(echoContext, liaisonResponse)
			if err != nil {
				t.Fatalf("LiaisonApiResponseEmitterFailed: %v", err)
			}

			if httpRecorder.Code != testCase.expectedHttpStatus {
				t.Errorf(
					"HttpStatusMismatch: expected %d, got %d",
					testCase.expectedHttpStatus, httpRecorder.Code,
				)
			}

			var actualResponseBody ApiResponseWrapper
			err = json.Unmarshal(httpRecorder.Body.Bytes(), &actualResponseBody)
			if err != nil {
				t.Fatalf("ResponseBodyUnmarshalFailed: %v", err)
			}

			if actualResponseBody.Status != testCase.expectedResponseBody.Status {
				t.Errorf(
					"ResponseStatusMismatch: expected %d, got %d",
					testCase.expectedResponseBody.Status, actualResponseBody.Status,
				)
			}

			if actualResponseBody.ReadableMessage != testCase.expectedResponseBody.ReadableMessage {
				t.Errorf(
					"ResponseReadableMessageMismatch: expected %s, got %s",
					testCase.expectedResponseBody.ReadableMessage, actualResponseBody.ReadableMessage,
				)
			}

			actualBodyJson, err := json.Marshal(actualResponseBody.Body)
			if err != nil {
				t.Fatalf("ActualBodyMarshalFailed: %v", err)
			}

			expectedBodyJson, err := json.Marshal(testCase.expectedResponseBody.Body)
			if err != nil {
				t.Fatalf("ExpectedBodyMarshalFailed: %v", err)
			}

			if string(actualBodyJson) != string(expectedBodyJson) {
				t.Errorf(
					"ResponseBodyMismatch: expected %s, got %s",
					string(expectedBodyJson), string(actualBodyJson),
				)
			}
		})
	}
}

func TestNewLiaisonResponse(t *testing.T) {
	testCases := []struct {
		name                    string
		status                  LiaisonResponseStatus
		readableMessage         string
		body                    any
		expectedStatus          LiaisonResponseStatus
		expectedReadableMessage string
		expectedBody            any
	}{
		{
			name:                    "SuccessStatus",
			status:                  LiaisonResponseStatusSuccess,
			readableMessage:         "DataRetrievedSuccessfully",
			body:                    []string{"item1", "item2"},
			expectedStatus:          LiaisonResponseStatusSuccess,
			expectedReadableMessage: "DataRetrievedSuccessfully",
			expectedBody:            []string{"item1", "item2"},
		},
		{
			name:                    "CreatedStatus",
			status:                  LiaisonResponseStatusCreated,
			readableMessage:         "ResourceCreated",
			body:                    map[string]int{"id": 123},
			expectedStatus:          LiaisonResponseStatusCreated,
			expectedReadableMessage: "ResourceCreated",
			expectedBody:            map[string]int{"id": 123},
		},
		{
			name:                    "UserErrorStatus",
			status:                  LiaisonResponseStatusUserError,
			readableMessage:         "ValidationFailed",
			body:                    nil,
			expectedStatus:          LiaisonResponseStatusUserError,
			expectedReadableMessage: "ValidationFailed",
			expectedBody:            nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := NewLiaisonResponse(
				testCase.status,
				testCase.body,
				testCase.readableMessage,
			)

			if result.Status != testCase.expectedStatus {
				t.Errorf(
					"StatusMismatch: expected %s, got %s",
					testCase.expectedStatus, result.Status,
				)
			}

			if result.ReadableMessage != testCase.expectedReadableMessage {
				t.Errorf(
					"ReadableMessageMismatch: expected %s, got %s",
					testCase.expectedReadableMessage, result.ReadableMessage,
				)
			}

			actualJson, err := json.Marshal(result.Body)
			if err != nil {
				t.Fatalf("ActualJsonMarshalingFailed: %v", err)
			}

			expectedJson, err := json.Marshal(testCase.expectedBody)
			if err != nil {
				t.Fatalf("ExpectedJsonMarshalingFailed: %v", err)
			}

			if string(actualJson) != string(expectedJson) {
				t.Errorf(
					"BodyMismatch: expected '%s', got '%s'",
					string(expectedJson), string(actualJson),
				)
			}
		})
	}
}

func TestLiaisonCliResponseRendererExitCodes(t *testing.T) {
	testCases := []struct {
		name             string
		status           LiaisonResponseStatus
		expectedExitCode int
	}{
		{
			name:             "SuccessStatus",
			status:           LiaisonResponseStatusSuccess,
			expectedExitCode: 0,
		},
		{
			name:             "CreatedStatus",
			status:           LiaisonResponseStatusCreated,
			expectedExitCode: 0,
		},
		{
			name:             "MultiStatus",
			status:           LiaisonResponseStatusMultiStatus,
			expectedExitCode: 1,
		},
		{
			name:             "UserErrorStatus",
			status:           LiaisonResponseStatusUserError,
			expectedExitCode: 1,
		},
		{
			name:             "UnauthorizedStatus",
			status:           LiaisonResponseStatusUnauthorized,
			expectedExitCode: 1,
		},
		{
			name:             "ForbiddenStatus",
			status:           LiaisonResponseStatusForbidden,
			expectedExitCode: 1,
		},
		{
			name:             "InfraErrorStatus",
			status:           LiaisonResponseStatusInfraError,
			expectedExitCode: 1,
		},
		{
			name:             "UnknownErrorStatus",
			status:           LiaisonResponseStatusUnknownError,
			expectedExitCode: 1,
		},
	}

	// Since LiaisonCliResponseRenderer calls os.Exit, we test the exit code logic
	// by creating a test program that we run in a subprocess.
	tempDir := t.TempDir()
	testProgramGoMod := `module testLiaisonCliResponseRenderer

go 1.25.3

require github.com/goinfinite/tk v0.1.1
`
	workingDir, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatalf("ResolveWorkingDirectoryFailed: %v", err)
	}

	fileClerk := tkInfra.FileClerk{}
	err = fileClerk.UpdateFileContent(workingDir+"/go.mod", testProgramGoMod, true)
	if err != nil {
		t.Fatalf("WriteGoModFailed: %v", err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testProgramMain := `package main

import tkPresentation "github.com/goinfinite/tk/src/presentation"

func main() {
	liaisonResponse := tkPresentation.NewLiaisonResponse(
		"` + string(testCase.status) + `", "Test message", map[string]string{"test": "data"},
	)
	tkPresentation.LiaisonCliResponseRenderer(liaisonResponse)
}`

			testFile := tempDir + "/testLiaisonCliResponseRenderer.go"
			err = fileClerk.UpdateFileContent(testFile, testProgramMain, true)
			if err != nil {
				t.Fatalf("WriteTestProgramFailed: %v", err)
			}

			if !fileClerk.FileExists(workingDir + "/go.sum") {
				goModShell := tkInfra.NewShell(tkInfra.ShellSettings{
					Command:          "go",
					Args:             []string{"mod", "tidy"},
					WorkingDirectory: workingDir,
				})
				_, err = goModShell.Run()
				if err != nil {
					t.Fatalf("RunGoModTidyFailed: %v", err)
				}
			}

			goRunShell := tkInfra.NewShell(tkInfra.ShellSettings{
				Command:          "go",
				Args:             []string{"run", testFile},
				WorkingDirectory: workingDir,
				Envs:             []string{"TERM=dumb"},
			})
			stdOut, err := goRunShell.Run()
			var stdErr string
			var exitCode int
			switch shellErr := err.(type) {
			case *tkInfra.ShellError:
				exitCode = shellErr.ExitCode
				stdErr = string(shellErr.StdErr)
			case nil:
				exitCode = 0
			default:
				t.Fatalf("CommandExecutionFailed: %v", err)
			}

			if exitCode != testCase.expectedExitCode {
				t.Errorf(
					"ExitCodeMismatch: expected '%d', got '%d'. StdOut: '%s' // StdErr: '%s'",
					testCase.expectedExitCode, exitCode, stdOut, stdErr,
				)
			}
		})
	}
}
