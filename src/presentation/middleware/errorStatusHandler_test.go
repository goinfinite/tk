package tkPresentationMiddleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkPresentation "github.com/goinfinite/tk/src/presentation"
	"github.com/labstack/echo/v4"
)

func TestErrorStatusHandlerApiHandler(test *testing.T) {
	test.Run("PassesThroughNilError", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		handler := NewErrorStatusHandler(nil)
		wrappedHandler := handler.ApiHandler(func(echoContext echo.Context) error {
			return echoContext.String(http.StatusCreated, "handler-response")
		})

		err := wrappedHandler(echoContext)
		if err != nil {
			subtest.Fatalf("UnexpectedError: %v", err)
		}

		if httpRecorder.Code != http.StatusCreated {
			subtest.Errorf(
				"StatusCodeMismatch: expected '%d', got '%d'",
				http.StatusCreated, httpRecorder.Code,
			)
		}

		if httpRecorder.Body.String() != "handler-response" {
			subtest.Errorf(
				"ResponseBodyMismatch: expected 'handler-response', got '%s'",
				httpRecorder.Body.String(),
			)
		}
	})

	test.Run("PassesThroughEchoHttpError", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		handlerErr := echo.NewHTTPError(http.StatusBadRequest, "BadRequest")
		handler := NewErrorStatusHandler(nil)
		wrappedHandler := handler.ApiHandler(func(echoContext echo.Context) error {
			return handlerErr
		})

		err := wrappedHandler(echoContext)
		if err != handlerErr {
			subtest.Fatalf("ExpectedEchoHttpErrorPassThrough: got '%v'", err)
		}

		if httpRecorder.Code != http.StatusOK {
			subtest.Errorf(
				"ResponseShouldNotBeWritten: got status '%d'", httpRecorder.Code,
			)
		}
	})

	test.Run("ReturnsErrorAfterResponseCommitted", func(subtest *testing.T) {
		echoInstance := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

		handlerErr := errors.New("ResponseCommittedError")
		handler := NewErrorStatusHandler(nil)
		wrappedHandler := handler.ApiHandler(func(echoContext echo.Context) error {
			writeErr := echoContext.String(http.StatusOK, "already-written")
			if writeErr != nil {
				subtest.Fatalf("ResponseWriteError: %v", writeErr)
			}
			return handlerErr
		})

		err := wrappedHandler(echoContext)
		if err != handlerErr {
			subtest.Fatalf("ExpectedHandlerError: got '%v'", err)
		}

		if httpRecorder.Code != http.StatusOK {
			subtest.Errorf(
				"CommittedStatusChanged: expected '%d', got '%d'",
				http.StatusOK, httpRecorder.Code,
			)
		}

		if httpRecorder.Body.String() != "already-written" {
			subtest.Errorf(
				"CommittedBodyChanged: got '%s'", httpRecorder.Body.String(),
			)
		}
	})

	test.Run("MapsHandlerErrors", func(subtest *testing.T) {
		errNotFound := errors.New("NotFound")
		errConflict := errors.New("Conflict")

		testCases := []struct {
			name                string
			errorStatusCodesMap map[error]tkValueObject.HttpStatusCode
			handlerErr          error
			expectedStatusCode  int
			expectedBody        string
		}{
			{
				name: "ExactSentinel",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{
					errNotFound: tkValueObject.HttpStatusCodeNotFound,
				},
				handlerErr:         errNotFound,
				expectedStatusCode: http.StatusNotFound,
				expectedBody:       "NotFound",
			},
			{
				name: "WrappedSentinel",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{
					errNotFound: tkValueObject.HttpStatusCodeNotFound,
				},
				handlerErr:         fmt.Errorf("WrappedSentinelError: %w", errNotFound),
				expectedStatusCode: http.StatusNotFound,
				expectedBody:       "NotFound",
			},
			{
				name: "UnmappedError",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{
					errNotFound: tkValueObject.HttpStatusCodeNotFound,
				},
				handlerErr:         errors.New("UnmappedError"),
				expectedStatusCode: http.StatusInternalServerError,
				expectedBody:       "InternalServerError",
			},
			{
				name:                "EmptyMap",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{},
				handlerErr:          errors.New("UnmappedError"),
				expectedStatusCode:  http.StatusInternalServerError,
				expectedBody:        "InternalServerError",
			},
			{
				name:                "NilMap",
				errorStatusCodesMap: nil,
				handlerErr:          errors.New("UnmappedError"),
				expectedStatusCode:  http.StatusInternalServerError,
				expectedBody:        "InternalServerError",
			},
			{
				name: "AmbiguousMatch",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{
					errNotFound: tkValueObject.HttpStatusCodeNotFound,
					errConflict: tkValueObject.HttpStatusCodeConflict,
				},
				handlerErr:         fmt.Errorf("%w: %w", errNotFound, errConflict),
				expectedStatusCode: http.StatusInternalServerError,
				expectedBody:       "InternalServerError",
			},
			{
				name:                "WrappedEchoHttpErrorIsUnmapped",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{},
				handlerErr: fmt.Errorf(
					"WrappedEchoHttpError: %w",
					echo.NewHTTPError(http.StatusBadRequest, "BadRequest"),
				),
				expectedStatusCode: http.StatusInternalServerError,
				expectedBody:       "InternalServerError",
			},
			{
				name: "InvalidMappedStatusCodeFailsClosed",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{
					errNotFound: tkValueObject.HttpStatusCode(0),
				},
				handlerErr:         errNotFound,
				expectedStatusCode: http.StatusInternalServerError,
				expectedBody:       "InternalServerError",
			},
			{
				name: "NonErrorMappedStatusCodeFailsClosed",
				errorStatusCodesMap: map[error]tkValueObject.HttpStatusCode{
					errNotFound: tkValueObject.HttpStatusCodeOk,
				},
				handlerErr:         errNotFound,
				expectedStatusCode: http.StatusInternalServerError,
				expectedBody:       "InternalServerError",
			},
		}

		for _, testCase := range testCases {
			subtest.Run(testCase.name, func(subtest *testing.T) {
				echoInstance := echo.New()
				httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
				httpRecorder := httptest.NewRecorder()
				echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

				handler := NewErrorStatusHandler(testCase.errorStatusCodesMap)
				wrappedHandler := handler.ApiHandler(
					func(echoContext echo.Context) error {
						return testCase.handlerErr
					},
				)

				err := wrappedHandler(echoContext)
				if err != nil {
					subtest.Fatalf("UnexpectedError: %v", err)
				}

				if httpRecorder.Code != testCase.expectedStatusCode {
					subtest.Errorf(
						"StatusCodeMismatch: expected '%d', got '%d'",
						testCase.expectedStatusCode, httpRecorder.Code,
					)
				}

				apiResponse := tkPresentation.ApiResponseWrapper{}
				decodeErr := json.Unmarshal(httpRecorder.Body.Bytes(), &apiResponse)
				if decodeErr != nil {
					subtest.Fatalf("ResponseDecodingError: %v", decodeErr)
				}

				if apiResponse.Status != testCase.expectedStatusCode {
					subtest.Errorf(
						"EnvelopeStatusMismatch: expected '%d', got '%d'",
						testCase.expectedStatusCode, apiResponse.Status,
					)
				}

				if apiResponse.Body != testCase.expectedBody {
					subtest.Errorf(
						"ResponseBodyMismatch: expected '%s', got '%v'",
						testCase.expectedBody, apiResponse.Body,
					)
				}

				if apiResponse.ReadableMessage != testCase.expectedBody {
					subtest.Errorf(
						"ReadableMessageMismatch: expected '%s', got '%s'",
						testCase.expectedBody, apiResponse.ReadableMessage,
					)
				}
			})
		}
	})
}
