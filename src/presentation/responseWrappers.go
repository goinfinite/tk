package tkPresentation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/labstack/echo/v4"

	tkInfra "github.com/goinfinite/tk/src/infra"
)

type ApiResponseWrapper struct {
	Status          int    `json:"status"`
	Body            any    `json:"body"`
	ReadableMessage string `json:"readableMessage"`
}

func NewApiResponseWrapper(
	status int,
	body any,
	readableMessage string,
) ApiResponseWrapper {
	return ApiResponseWrapper{
		Status:          status,
		Body:            body,
		ReadableMessage: readableMessage,
	}
}

type LiaisonResponseStatus string

const (
	LiaisonResponseStatusSuccess            LiaisonResponseStatus = "success"
	LiaisonResponseStatusCreated            LiaisonResponseStatus = "created"
	LiaisonResponseStatusAccepted           LiaisonResponseStatus = "accepted"
	LiaisonResponseStatusMultiStatus        LiaisonResponseStatus = "multiStatus"
	LiaisonResponseStatusUserError          LiaisonResponseStatus = "userError"
	LiaisonResponseStatusUnauthorized       LiaisonResponseStatus = "unauthorized"
	LiaisonResponseStatusForbidden          LiaisonResponseStatus = "forbidden"
	LiaisonResponseStatusNotFound           LiaisonResponseStatus = "notFound"
	LiaisonResponseStatusTimeout            LiaisonResponseStatus = "timeout"
	LiaisonResponseStatusConflict           LiaisonResponseStatus = "conflict"
	LiaisonResponseStatusRateLimited        LiaisonResponseStatus = "rateLimited"
	LiaisonResponseStatusInfraError         LiaisonResponseStatus = "infraError"
	LiaisonResponseStatusUnknownError       LiaisonResponseStatus = "unknownError"
	LiaisonResponseStatusServiceUnavailable LiaisonResponseStatus = "serviceUnavailable"
)

func (status LiaisonResponseStatus) isError() bool {
	switch status {
	case LiaisonResponseStatusUserError,
		LiaisonResponseStatusUnauthorized,
		LiaisonResponseStatusForbidden,
		LiaisonResponseStatusNotFound,
		LiaisonResponseStatusTimeout,
		LiaisonResponseStatusConflict,
		LiaisonResponseStatusRateLimited,
		LiaisonResponseStatusInfraError,
		LiaisonResponseStatusUnknownError,
		LiaisonResponseStatusServiceUnavailable:
		return true
	}

	return false
}

func (status LiaisonResponseStatus) isServerError() bool {
	switch status {
	case LiaisonResponseStatusInfraError,
		LiaisonResponseStatusUnknownError,
		LiaisonResponseStatusServiceUnavailable:
		return true
	}

	return false
}

type LiaisonResponse struct {
	Status          LiaisonResponseStatus `json:"status"`
	Body            any                   `json:"body"`
	ReadableMessage string                `json:"readableMessage"`
}

func newLiaisonResponse(
	status LiaisonResponseStatus,
	body any,
	readableMessage string,
) LiaisonResponse {
	return LiaisonResponse{
		Status:          status,
		Body:            body,
		ReadableMessage: readableMessage,
	}
}

func LiaisonApiResponseEmitter(
	echoContext echo.Context,
	liaisonResponse LiaisonResponse,
) error {
	httpStatus := http.StatusOK
	switch liaisonResponse.Status {
	case LiaisonResponseStatusCreated:
		httpStatus = http.StatusCreated
	case LiaisonResponseStatusAccepted:
		httpStatus = http.StatusAccepted
	case LiaisonResponseStatusMultiStatus:
		httpStatus = http.StatusMultiStatus
	case LiaisonResponseStatusUserError:
		httpStatus = http.StatusBadRequest
	case LiaisonResponseStatusUnauthorized:
		httpStatus = http.StatusUnauthorized
	case LiaisonResponseStatusForbidden:
		httpStatus = http.StatusForbidden
	case LiaisonResponseStatusNotFound:
		httpStatus = http.StatusNotFound
	case LiaisonResponseStatusTimeout:
		httpStatus = http.StatusRequestTimeout
	case LiaisonResponseStatusConflict:
		httpStatus = http.StatusConflict
	case LiaisonResponseStatusRateLimited:
		httpStatus = http.StatusTooManyRequests
	case LiaisonResponseStatusInfraError, LiaisonResponseStatusUnknownError:
		httpStatus = http.StatusInternalServerError
	case LiaisonResponseStatusServiceUnavailable:
		httpStatus = http.StatusServiceUnavailable
	}

	return echoContext.JSON(httpStatus, NewApiResponseWrapper(
		httpStatus, liaisonResponse.Body, liaisonResponse.ReadableMessage,
	))
}

func LiaisonCliResponseRenderer(liaisonResponse LiaisonResponse) {
	const (
		exitCodeOk          = 0
		exitCodeUsage       = 64
		exitCodeDataErr     = 65
		exitCodeNoInput     = 66
		exitCodeUnavailable = 69
		exitCodeSoftware    = 70
		exitCodeTempFail    = 75
		exitCodeNoPerm      = 77
	)

	var exitCode int
	switch liaisonResponse.Status {
	case LiaisonResponseStatusSuccess:
		exitCode = exitCodeOk
	case LiaisonResponseStatusCreated:
		exitCode = exitCodeOk
	case LiaisonResponseStatusAccepted:
		exitCode = exitCodeOk
	case LiaisonResponseStatusUserError:
		exitCode = exitCodeUsage
	case LiaisonResponseStatusMultiStatus:
		exitCode = exitCodeDataErr
	case LiaisonResponseStatusNotFound:
		exitCode = exitCodeNoInput
	case LiaisonResponseStatusConflict:
		exitCode = exitCodeDataErr
	case LiaisonResponseStatusInfraError:
		exitCode = exitCodeUnavailable
	case LiaisonResponseStatusServiceUnavailable:
		exitCode = exitCodeUnavailable
	case LiaisonResponseStatusUnknownError:
		exitCode = exitCodeSoftware
	case LiaisonResponseStatusTimeout:
		exitCode = exitCodeTempFail
	case LiaisonResponseStatusRateLimited:
		exitCode = exitCodeTempFail
	case LiaisonResponseStatusUnauthorized:
		exitCode = exitCodeNoPerm
	case LiaisonResponseStatusForbidden:
		exitCode = exitCodeNoPerm
	default:
		exitCode = exitCodeSoftware
	}

	if !tkInfra.IsStdoutTerminal() {
		jsonBytes, err := json.Marshal(liaisonResponse)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ResponseEncodingError: %v\n", err)
			os.Exit(exitCodeSoftware)
		}

		fmt.Println(string(jsonBytes))
		os.Exit(exitCode)
	}

	prettyJsonBytes, err := json.MarshalIndent(liaisonResponse, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ResponseEncodingError: %v\n", err)
		os.Exit(exitCodeSoftware)
	}

	syntaxHighlightingLexer := lexers.Get("json")
	if syntaxHighlightingLexer == nil {
		syntaxHighlightingLexer = lexers.Fallback
	}

	shIterator, err := syntaxHighlightingLexer.Tokenise(nil, string(prettyJsonBytes))
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "SyntaxHighlightingTokenizingError: %v\n", err,
		)
		os.Exit(exitCodeSoftware)
	}

	shFormatter := formatters.Get("terminal256")
	if shFormatter == nil {
		shFormatter = formatters.Fallback
	}

	err = shFormatter.Format(os.Stdout, styles.Vulcan, shIterator)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "SyntaxHighlightingFormatError: %v\n", err,
		)
		os.Exit(exitCodeSoftware)
	}
	fmt.Println()
	os.Exit(exitCode)
}

func SimpleCliResponseRenderer(
	isSuccess bool, message string,
) {
	status := LiaisonResponseStatusUserError
	if isSuccess {
		status = LiaisonResponseStatusSuccess
	}

	liaisonResponse := newLiaisonResponse(
		status, message, "",
	)
	LiaisonCliResponseRenderer(liaisonResponse)
}
