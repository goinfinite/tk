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

type PanicReport struct {
	RecoverErr error
	StackTrace string
}

func readPanicReport() *PanicReport {
	var recoverErr error
	recoverFunc := recover()

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

func logPanic(panicReport *PanicReport) {
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
	slogger.Error("PanicRecovered", slog.String("error", panicReport.RecoverErr.Error()))
	slogger.Error("StackTrace", slog.String("stackTrace", panicReport.StackTrace))
}

func apiHandlePanic(echoContext echo.Context) {
	panicReport := readPanicReport()
	stackTraceStr := panicReport.StackTrace

	statusCode := http.StatusInternalServerError
	if panicHandlerDomainLayerPathRegex.MatchString(stackTraceStr) {
		statusCode = http.StatusBadRequest
	}

	recoverErrStr := panicReport.RecoverErr.Error()

	shortErrStr := "InternalServerError"
	if len(recoverErrStr) > panicHandlerMaxErrorLength {
		shortErrStr = recoverErrStr[:panicHandlerMaxErrorLength] + "..."
	}

	jsonResponse := map[string]any{
		"status": statusCode,
		"body": map[string]any{
			"uri":            echoContext.Request().RequestURI,
			"queryParams":    echoContext.QueryParams(),
			"exceptionCode":  recoverErrStr,
			"exceptionTrace": stackTraceStr,
		},
		"humanReadableMessage": "SomethingWentWrong",
	}

	if !isRequesterTrustworthy(echoContext) {
		jsonResponse["body"] = map[string]any{
			"exceptionCode": shortErrStr,
		}
	}

	echoContext.JSON(statusCode, jsonResponse)

	logPanic(panicReport)
}

func ApiPanicHandler(subsequentHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(echoContext echo.Context) error {
		defer apiHandlePanic(echoContext)
		return subsequentHandler(echoContext)
	}
}

func cliHandlePanic() {
	panicReport := readPanicReport()
	logPanic(panicReport)

	fmt.Println("FatalError. Please check the panic.log file for more details.")
	os.Exit(1)
}

func CliPanicHandler() {
	defer cliHandlePanic()
}
