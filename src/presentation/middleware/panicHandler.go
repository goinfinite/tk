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

type StackTrace struct {
	trace string
}

func (st *StackTrace) String() string {
	return st.trace
}

func readStackTrace() *StackTrace {
	traceBuffer := make([]byte, 1<<16)
	stackBufBytesCount := runtime.Stack(traceBuffer, true)
	stackTraceStr := string(traceBuffer[:stackBufBytesCount])
	traceLines := strings.Split(stackTraceStr, "\n")

	filteredTraceLines := []string{}
	for _, traceLine := range traceLines {
		filteredTraceLines = append(filteredTraceLines, traceLine)
		if strings.Contains(traceLine, "created by net/http") {
			break
		}
	}

	filteredStackTrace := strings.Join(filteredTraceLines, "\n")
	return &StackTrace{trace: filteredStackTrace}
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
	logsDir := "logs"
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		err = os.Mkdir(logsDir, 0755)
		if err != nil {
			slog.Error("CreateLogDirectoryError", slog.String("error", err.Error()))
			return
		}
	}

	logFilePath := filepath.Join(logsDir, "panic.log")
	logFile, errOpen := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errOpen != nil {
		slog.Error("OpenLogFileError", slog.String("error", errOpen.Error()))
		return
	}
	defer logFile.Close()

	slogger := slog.New(slog.NewTextHandler(logFile, nil))
	slogger.Error("PanicRecovered", slog.String("error", err.Error()))
	slogger.Error("StackTrace", slog.String("stackTrace", stackTrace.String()))
}

var domainLayerPathRegex = regexp.MustCompile(`domain/(valueObject|entity|useCase)`)

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
	queryParams := echoContext.QueryParams()

	statusCode := http.StatusInternalServerError
	if domainLayerPathRegex.MatchString(stackTrace.String()) {
		statusCode = http.StatusBadRequest
	}

	errStr := err.Error()
	humanReadableErrStr := "SomethingWentWrong"

	shortErrStr := "InternalServerError"
	shortErrStrIdealLength := 150
	if len(errStr) > shortErrStrIdealLength {
		shortErrStr = errStr[:shortErrStrIdealLength]
	}

	jsonResponse := map[string]any{
		"status": statusCode,
		"body": map[string]any{
			"uri":            requestUri,
			"queryParams":    queryParams,
			"exceptionCode":  errStr,
			"exceptionTrace": stackTrace.String(),
		},
		"humanReadableMessage": humanReadableErrStr,
	}
	if !isRequesterTrustworthy(echoContext) {
		jsonResponse["body"] = shortErrStr
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
