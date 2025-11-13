package tkPresentationMiddleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfra "github.com/goinfinite/tk/src/infra"
	"github.com/labstack/echo/v4"
)

const (
	panicHandlerMaxStackTraceSize int    = 1 << 16
	panicHandlerMaxErrorLength    int    = 150
	panicHandlerLogsDir           string = "logs"
	panicHandlerLogFileName       string = "panic.log"
)

var panicHandlerDomainLayerPathRegex = regexp.MustCompile(`domain/(valueObject|entity|useCase)`)

type PanicReport struct {
	RecoverErr        error
	StackTrace        string
	RequestUri        string
	OperatorIpAddress string
}

func readPanicReport() *PanicReport {
	var recoverErr error
	recoverFunc := recover()
	if recoverFunc == nil {
		return nil
	}

	recoverErr, isError := recoverFunc.(error)
	if !isError {
		recoverErr = fmt.Errorf("%v", recoverFunc)
	}

	traceBuffer := make([]byte, panicHandlerMaxStackTraceSize)
	stackBufBytesCount := runtime.Stack(traceBuffer, true)
	stackTraceStr := string(traceBuffer[:stackBufBytesCount])

	filteredTraceLines := []string{}
	for traceLine := range strings.SplitSeq(stackTraceStr, "\n") {
		if strings.Contains(traceLine, "created by net/http") {
			break
		}
		filteredTraceLines = append(filteredTraceLines, traceLine)
	}
	filteredStackTraceStr := strings.Join(filteredTraceLines, "\n")

	return &PanicReport{
		StackTrace: filteredStackTraceStr,
		RecoverErr: recoverErr,
	}
}

func isOperatorTrustworthy(echoContext echo.Context) bool {
	trustedIpAddresses, err := tkInfra.TrustedIpsReader()
	if err != nil {
		return false
	}

	rawOperatorIpAddress := echoContext.RealIP()
	if rawOperatorIpAddress == "" {
		return false
	}

	operatorIpAddress, ipErr := tkValueObject.NewIpAddress(rawOperatorIpAddress)
	if ipErr != nil {
		return false
	}

	return slices.Contains(trustedIpAddresses, operatorIpAddress)
}

func logPanic(panicReportPtr *PanicReport) {
	if _, statErr := os.Stat(panicHandlerLogsDir); os.IsNotExist(statErr) {
		if mkdirErr := os.Mkdir(panicHandlerLogsDir, 0755); mkdirErr != nil {
			slog.Error("CreateLogDirectoryError", slog.String("error", mkdirErr.Error()))
			return
		}
	}

	logFilePath := filepath.Join(panicHandlerLogsDir, panicHandlerLogFileName)
	logFile, openErr := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		slog.Error("OpenLogFileError", slog.String("error", openErr.Error()))
		return
	}
	defer logFile.Close()

	slogger := slog.New(slog.NewTextHandler(logFile, nil))
	slogger.Error(
		"PanicRecovered",
		slog.String("error", panicReportPtr.RecoverErr.Error()),
		slog.String("stackTrace", panicReportPtr.StackTrace),
		slog.String("requestUri", panicReportPtr.RequestUri),
		slog.String("operatorIpAddress", panicReportPtr.OperatorIpAddress),
	)
}

func apiHandlePanic(echoContext echo.Context) {
	panicReportPtr := readPanicReport()
	if panicReportPtr == nil {
		return
	}

	stackTraceStr := panicReportPtr.StackTrace

	// Finding the exact path of the panic in the stack trace is tricky, so we only check
	// the beginning of the stack trace. If the domain layer path is present, the panic is
	// (most likely) caused by a business logic error due to invalid input.
	stackTraceStrBeginning := stackTraceStr
	if len(stackTraceStr) > 1000 {
		stackTraceStrBeginning = stackTraceStr[:1000]
	}
	statusCode := http.StatusInternalServerError
	if panicHandlerDomainLayerPathRegex.MatchString(stackTraceStrBeginning) {
		statusCode = http.StatusBadRequest
	}

	fullRecoverErrStr := panicReportPtr.RecoverErr.Error()
	if fullRecoverErrStr == "" {
		fullRecoverErrStr = "InternalServerError"
	}

	shortRecoverErrStr := fullRecoverErrStr
	if len(shortRecoverErrStr) > panicHandlerMaxErrorLength {
		shortRecoverErrStr = shortRecoverErrStr[:panicHandlerMaxErrorLength] + "..."
	}

	jsonResponse := map[string]any{
		"status": statusCode,
		"body": map[string]any{
			"uri":            echoContext.Request().RequestURI,
			"queryParams":    echoContext.QueryParams(),
			"exceptionCode":  fullRecoverErrStr,
			"exceptionTrace": stackTraceStr,
		},
		"readableMessage": "SomethingWentWrong",
	}

	if !isOperatorTrustworthy(echoContext) {
		jsonResponse["body"] = map[string]any{
			"exceptionCode": shortRecoverErrStr,
		}
	}

	echoContext.JSON(statusCode, jsonResponse)

	panicReportPtr.RequestUri = echoContext.Request().RequestURI
	operatorIpAddress, ipErr := tkValueObject.NewIpAddress(echoContext.RealIP())
	if ipErr == nil {
		panicReportPtr.OperatorIpAddress = operatorIpAddress.String()
	}

	logPanic(panicReportPtr)
}

func ApiPanicHandler(subsequentHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(echoContext echo.Context) error {
		defer apiHandlePanic(echoContext)
		return subsequentHandler(echoContext)
	}
}

// @attention CliPanicHandler MUST be used as a defer statement.
func CliPanicHandler() {
	panicReport := readPanicReport()
	if panicReport == nil {
		return
	}
	logPanic(panicReport)

	fmt.Println("FatalError. Please check the 'logs/panic.log' file for more details.")
	os.Exit(1)
}
