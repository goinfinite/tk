package tkPresentationMiddlewareHoneypot

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

func TestNewStreamHandler(t *testing.T) {
	maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
		uint64(10 * 1024 * 1024),
	)
	handler := NewStreamHandler(maxStreamSize)
	if handler == nil {
		t.Errorf("ExpectedNonNilHandler")
	}
}

func TestStreamHandlerServeBandwidthExhaust(t *testing.T) {
	testCaseStructs := []struct {
		name           string
		maxStreamSize  uint64
		expectedStatus int
	}{
		{
			name:           "ValidMaxStreamSize1MB",
			maxStreamSize:  1 * 1024 * 1024,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ZeroMaxStreamSize",
			maxStreamSize:  0,
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
				testCase.maxStreamSize,
			)
			handler := NewStreamHandler(maxStreamSize)

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(
				http.MethodGet, "/api/v1/stream/logs", nil,
			)
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(
				httpRequest, httpRecorder,
			)

			err := handler.ServeBandwidthExhaust(echoContext)
			if err != nil {
				t.Errorf("UnexpectedError: %v", err)
			}

			if httpRecorder.Code != testCase.expectedStatus {
				t.Errorf(
					"StatusMismatch: Expected=%d, Actual=%d",
					testCase.expectedStatus, httpRecorder.Code,
				)
			}
		})
	}
}

func TestStreamHandlerServeBandwidthExhaustContentSize(t *testing.T) {
	maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
		uint64(1 * 1024 * 1024),
	)
	handler := NewStreamHandler(maxStreamSize)

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/v1/stream/logs", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	err := handler.ServeBandwidthExhaust(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	bodySize := httpRecorder.Body.Len()
	if bodySize == 0 {
		t.Errorf("ExpectedNonEmptyBody")
	}
}

func TestStreamHandlerServeAiTrap(t *testing.T) {
	testCaseStructs := []struct {
		name           string
		maxStreamSize  uint64
		expectedStatus int
	}{
		{
			name:           "ValidMaxStreamSize1MB",
			maxStreamSize:  1 * 1024 * 1024,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ZeroMaxStreamSize",
			maxStreamSize:  0,
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
				testCase.maxStreamSize,
			)
			handler := NewStreamHandler(maxStreamSize)
			generator := NewAiTrapGenerator()

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(
				http.MethodGet, "/api/v1/ai/models", nil,
			)
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(
				httpRequest, httpRecorder,
			)

			err := handler.ServeAiTrap(echoContext, generator)
			if err != nil {
				t.Errorf("UnexpectedError: %v", err)
			}

			if httpRecorder.Code != testCase.expectedStatus {
				t.Errorf(
					"StatusMismatch: Expected=%d, Actual=%d",
					testCase.expectedStatus, httpRecorder.Code,
				)
			}
		})
	}
}

func TestStreamHandlerServeAiTrapContainsPlausibleText(t *testing.T) {
	maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
		uint64(1 * 1024 * 1024),
	)
	handler := NewStreamHandler(maxStreamSize)
	generator := NewAiTrapGenerator()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/v1/completions", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	err := handler.ServeAiTrap(echoContext, generator)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	bodyContent := httpRecorder.Body.String()
	if len(bodyContent) == 0 {
		t.Errorf("ExpectedNonEmptyBody")
	}

	contentType := httpRecorder.Header().Get(echo.HeaderContentType)
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf(
			"ContentTypeMismatch: Expected='text/plain; charset=utf-8', Actual='%s'",
			contentType,
		)
	}
}

func TestStreamHandlerResolveStreamSize(t *testing.T) {
	testCaseStructs := []struct {
		name          string
		maxStreamSize uint64
	}{
		{
			name:          "MaxStreamSizeSmallerThanRange",
			maxStreamSize: 1 * 1024 * 1024,
		},
		{
			name:          "MaxStreamSizeLargerThanRange",
			maxStreamSize: 50 * 1024 * 1024,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
				testCase.maxStreamSize,
			)
			handler := NewStreamHandler(maxStreamSize)

			resolvedSize := handler.resolveStreamSize()

			if resolvedSize == 0 {
				t.Errorf("ExpectedNonZeroSize")
			}

			if resolvedSize > testCase.maxStreamSize {
				t.Errorf(
					"SizeExceedsMax: Resolved=%d, Max=%d",
					resolvedSize, testCase.maxStreamSize,
				)
			}
		})
	}
}
