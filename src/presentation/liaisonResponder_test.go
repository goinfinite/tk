package tkPresentation

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestLiaisonResponderSuccessResponses(test *testing.T) {
	responder := NewLiaisonResponder(nil)
	responseBody := map[string]string{"id": "1"}
	readableMessage := "OperationCompletedSuccessfully"

	testCases := []struct {
		name           string
		response       LiaisonResponse
		expectedStatus LiaisonResponseStatus
	}{
		{
			name:           "Success",
			response:       responder.Success(responseBody, readableMessage),
			expectedStatus: LiaisonResponseStatusSuccess,
		},
		{
			name:           "Created",
			response:       responder.Created(responseBody, readableMessage),
			expectedStatus: LiaisonResponseStatusCreated,
		},
		{
			name:           "Accepted",
			response:       responder.Accepted(responseBody, readableMessage),
			expectedStatus: LiaisonResponseStatusAccepted,
		},
		{
			name:           "MultiStatus",
			response:       responder.MultiStatus(responseBody, readableMessage),
			expectedStatus: LiaisonResponseStatusMultiStatus,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			if testCase.response.Status != testCase.expectedStatus {
				subtest.Errorf(
					"StatusMismatch: expected '%s', got '%s'",
					testCase.expectedStatus, testCase.response.Status,
				)
			}

			if !reflect.DeepEqual(testCase.response.Body, responseBody) {
				subtest.Errorf(
					"BodyMismatch: expected '%v', got '%v'",
					responseBody, testCase.response.Body,
				)
			}

			if testCase.response.ReadableMessage != readableMessage {
				subtest.Errorf(
					"ReadableMessageMismatch: expected '%s', got '%s'",
					readableMessage, testCase.response.ReadableMessage,
				)
			}
		})
	}
}

func TestLiaisonResponderInputError(test *testing.T) {
	responder := NewLiaisonResponder(nil)
	errorMessage := "InvalidUsername"

	response := responder.InputError(errorMessage)

	if response.Status != LiaisonResponseStatusUserError {
		test.Errorf(
			"StatusMismatch: expected '%s', got '%s'",
			LiaisonResponseStatusUserError, response.Status,
		)
	}

	if response.Body != errorMessage {
		test.Errorf(
			"BodyMismatch: expected '%s', got '%v'",
			errorMessage, response.Body,
		)
	}

	if response.ReadableMessage != errorMessage {
		test.Errorf(
			"ReadableMessageMismatch: expected '%s', got '%s'",
			errorMessage, response.ReadableMessage,
		)
	}
}

func TestLiaisonResponderAppError(test *testing.T) {
	errRecordNotFound := errors.New("RecordNotFound")
	errAccessForbidden := errors.New("AccessForbidden")

	errorStatusesMap := map[error]LiaisonResponseStatus{
		errRecordNotFound:  LiaisonResponseStatusNotFound,
		errAccessForbidden: LiaisonResponseStatusForbidden,
	}

	testCases := []struct {
		name                    string
		errorStatusesMap        map[error]LiaisonResponseStatus
		appErr                  error
		expectedStatus          LiaisonResponseStatus
		expectedBody            any
		expectedReadableMessage string
	}{
		{
			name:                    "ExactSentinel",
			errorStatusesMap:        errorStatusesMap,
			appErr:                  errRecordNotFound,
			expectedStatus:          LiaisonResponseStatusNotFound,
			expectedBody:            "RecordNotFound",
			expectedReadableMessage: "RecordNotFound",
		},
		{
			name:             "WrappedSentinel",
			errorStatusesMap: errorStatusesMap,
			appErr: fmt.Errorf(
				"ReadRecordError: %w", errRecordNotFound,
			),
			expectedStatus:          LiaisonResponseStatusNotFound,
			expectedBody:            "RecordNotFound",
			expectedReadableMessage: "RecordNotFound",
		},
		{
			name:                    "UnmappedError",
			errorStatusesMap:        errorStatusesMap,
			appErr:                  errors.New("UnexpectedError"),
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
		{
			name:             "AmbiguousMatch",
			errorStatusesMap: errorStatusesMap,
			appErr: fmt.Errorf(
				"%w: %w", errRecordNotFound, errAccessForbidden,
			),
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
		{
			name: "InvalidMappedStatus",
			errorStatusesMap: map[error]LiaisonResponseStatus{
				errRecordNotFound: LiaisonResponseStatusSuccess,
			},
			appErr:                  errRecordNotFound,
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
		{
			name: "EmptyMappedStatus",
			errorStatusesMap: map[error]LiaisonResponseStatus{
				errRecordNotFound: LiaisonResponseStatus(""),
			},
			appErr:                  errRecordNotFound,
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
		{
			name: "MappedServerError",
			errorStatusesMap: map[error]LiaisonResponseStatus{
				errRecordNotFound: LiaisonResponseStatusInfraError,
			},
			appErr:                  errRecordNotFound,
			expectedStatus:          LiaisonResponseStatusInfraError,
			expectedBody:            "RecordNotFound",
			expectedReadableMessage: "RecordNotFound",
		},
		{
			name:                    "EmptyMap",
			errorStatusesMap:        map[error]LiaisonResponseStatus{},
			appErr:                  errRecordNotFound,
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
		{
			name:                    "NilMap",
			errorStatusesMap:        nil,
			appErr:                  errRecordNotFound,
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
		{
			name:                    "NilError",
			errorStatusesMap:        errorStatusesMap,
			appErr:                  nil,
			expectedStatus:          LiaisonResponseStatusUnknownError,
			expectedBody:            "InternalServerError",
			expectedReadableMessage: "InternalServerError",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			responder := NewLiaisonResponder(testCase.errorStatusesMap)
			response := responder.AppError(testCase.appErr)

			if response.Status != testCase.expectedStatus {
				subtest.Errorf(
					"StatusMismatch: expected '%s', got '%s'",
					testCase.expectedStatus, response.Status,
				)
			}

			if response.Body != testCase.expectedBody {
				subtest.Errorf(
					"BodyMismatch: expected '%v', got '%v'",
					testCase.expectedBody, response.Body,
				)
			}

			if response.ReadableMessage != testCase.expectedReadableMessage {
				subtest.Errorf(
					"ReadableMessageMismatch: expected '%s', got '%s'",
					testCase.expectedReadableMessage,
					response.ReadableMessage,
				)
			}
		})
	}
}
