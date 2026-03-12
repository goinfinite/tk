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
	"golang.org/x/term"
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
	LiaisonResponseStatusSuccess      LiaisonResponseStatus = "success"
	LiaisonResponseStatusCreated      LiaisonResponseStatus = "created"
	LiaisonResponseStatusMultiStatus  LiaisonResponseStatus = "multiStatus"
	LiaisonResponseStatusUserError    LiaisonResponseStatus = "userError"
	LiaisonResponseStatusUnauthorized LiaisonResponseStatus = "unauthorized"
	LiaisonResponseStatusForbidden    LiaisonResponseStatus = "forbidden"
	LiaisonResponseStatusNotFound     LiaisonResponseStatus = "notFound"
	LiaisonResponseStatusTimeout      LiaisonResponseStatus = "timeout"
	LiaisonResponseStatusRateLimited  LiaisonResponseStatus = "rateLimited"
	LiaisonResponseStatusInfraError   LiaisonResponseStatus = "infraError"
	LiaisonResponseStatusUnknownError LiaisonResponseStatus = "unknownError"
)

type LiaisonResponse struct {
	Status          LiaisonResponseStatus `json:"status"`
	Body            any                   `json:"body"`
	ReadableMessage string                `json:"readableMessage"`
}

// The NewLiaisonResponse function is used when the response body contains complex data
// that requires a more informative message to be displayed to the user. The readable message
// field should describe the outcome of the operation in a clear and understandable manner,
// providing valuable information about the result of the request. This is particularly handy in
// scenarios where the body field contains details about the created resource, such as in POST
// requests.
func NewLiaisonResponse(
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

// The NewLiaisonResponseNoMessage function is used when the response body is the
// message (string) that should be displayed to the user, so the readable message field
// is not needed.
func NewLiaisonResponseNoMessage(
	status LiaisonResponseStatus,
	body any,
) LiaisonResponse {
	return NewLiaisonResponse(status, body, "")
}

func LiaisonApiResponseEmitter(
	echoContext echo.Context,
	liaisonResponse LiaisonResponse,
) error {
	httpStatus := http.StatusOK
	switch liaisonResponse.Status {
	case LiaisonResponseStatusCreated:
		httpStatus = http.StatusCreated
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
	case LiaisonResponseStatusRateLimited:
		httpStatus = http.StatusTooManyRequests
	case LiaisonResponseStatusInfraError, LiaisonResponseStatusUnknownError:
		httpStatus = http.StatusInternalServerError
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

	exitCode := exitCodeOk
	switch liaisonResponse.Status {
	case LiaisonResponseStatusSuccess:
		exitCode = exitCodeOk
	case LiaisonResponseStatusCreated:
		exitCode = exitCodeOk
	case LiaisonResponseStatusUserError:
		exitCode = exitCodeUsage
	case LiaisonResponseStatusMultiStatus:
		exitCode = exitCodeDataErr
	case LiaisonResponseStatusNotFound:
		exitCode = exitCodeNoInput
	case LiaisonResponseStatusInfraError:
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

	stdoutFileDescriptor := int(os.Stdout.Fd())
	isNonInteractiveSession := !term.IsTerminal(stdoutFileDescriptor)
	if isNonInteractiveSession {
		jsonBytes, err := json.Marshal(liaisonResponse)
		if err != nil {
			fmt.Println("ResponseEncodingError")
			os.Exit(exitCodeSoftware)
		}

		fmt.Println(string(jsonBytes))
		os.Exit(exitCode)
	}

	prettyJsonBytes, err := json.MarshalIndent(liaisonResponse, "", "  ")
	if err != nil {
		fmt.Println("ResponseEncodingError")
		os.Exit(exitCodeSoftware)
	}

	syntaxHighlightingLexer := lexers.Get("json")
	if syntaxHighlightingLexer == nil {
		syntaxHighlightingLexer = lexers.Fallback
	}

	shIterator, err := syntaxHighlightingLexer.Tokenise(nil, string(prettyJsonBytes))
	if err != nil {
		fmt.Println("SyntaxHighlightingTokenizingError")
		os.Exit(exitCodeSoftware)
	}

	shFormatter := formatters.Get("terminal256")
	if shFormatter == nil {
		shFormatter = formatters.Fallback
	}

	err = shFormatter.Format(os.Stdout, styles.Vulcan, shIterator)
	if err != nil {
		fmt.Println("SyntaxHighlightingFormatError")
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

	liaisonResponse := NewLiaisonResponseNoMessage(
		status, message,
	)
	LiaisonCliResponseRenderer(liaisonResponse)
}
