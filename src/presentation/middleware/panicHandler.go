package tkPresentationMiddleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	panicHandlerMaxStackTraceSize int    = 1 << 16
	panicHandlerMaxErrorLength    int    = 150
	panicHandlerLogsDir           string = "logs"
	panicHandlerLogFileName       string = "panic.log"
)

var panicHandlerDomainLayerPathRegex = regexp.MustCompile(`domain/(valueObject|entity|useCase)`)

type StackTrace struct {
	trace string
}

func (st *StackTrace) String() string {
	return st.trace
}

func readStackTrace() *StackTrace {
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
	return &StackTrace{trace: filteredStackTraceStr}
}

func isRequesterTrustworthy(echoContext echo.Context) bool {
	currentIp := echoContext.RealIP()
	if currentIp == "" {
		return false
	}

	trustedIps := strings.SplitSeq(os.Getenv("TRUSTED_IPS"), ",")
	for staffIp := range trustedIps {
		if currentIp == strings.TrimSpace(staffIp) {
			return true
		}
	}

	return false
}

func logPanic(err error, stackTrace *StackTrace) {
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
	slogger.Error("PanicRecovered", slog.String("error", err.Error()))
	slogger.Error("StackTrace", slog.String("stackTrace", stackTrace.String()))
}

func handlePanic(echoContext echo.Context) {
	recoverFunc := recover()
	if recoverFunc == nil {
		return
	}

	err, isError := recoverFunc.(error)
	if !isError {
		err = fmt.Errorf("%v", recoverFunc)
	}

	stackTrace := readStackTrace()
	requestUri := echoContext.Request().RequestURI

	statusCode := http.StatusInternalServerError
	if panicHandlerDomainLayerPathRegex.MatchString(stackTrace.String()) {
		statusCode = http.StatusBadRequest
	}

	errStr := err.Error()
	humanReadableErrStr := "SomethingWentWrong"

	shortErrStr := "InternalServerError"
	if len(errStr) > panicHandlerMaxErrorLength {
		shortErrStr = errStr[:panicHandlerMaxErrorLength] + "..."
	}

	jsonResponse := map[string]any{
		"status": statusCode,
		"body": map[string]any{
			"uri":            requestUri,
			"queryParams":    echoContext.QueryParams(),
			"exceptionCode":  errStr,
			"exceptionTrace": stackTrace.String(),
		},
		"humanReadableMessage": humanReadableErrStr,
	}

	if !isRequesterTrustworthy(echoContext) {
		jsonResponse["body"] = map[string]any{
			"exceptionCode": shortErrStr,
		}
	}

	echoContext.JSON(statusCode, jsonResponse)

	logPanic(err, stackTrace)
}

func PanicHandler(subsequentHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(echoContext echo.Context) error {
		defer handlePanic(echoContext)
		return subsequentHandler(echoContext)
	}
}
