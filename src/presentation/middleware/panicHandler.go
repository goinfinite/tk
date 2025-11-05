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

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

const (
	panicHandlerMaxStackTraceSize    int    = 1 << 16
	panicHandlerMaxErrorLength       int    = 150
	panicHandlerLogsDir              string = "logs"
	panicHandlerLogFileName          string = "panic.log"
	PanicHandlerTrustedIpsEnvVarName string = "TRUSTED_IPS"
)

var panicHandlerDomainLayerPathRegex = regexp.MustCompile(`domain/(valueObject|entity|useCase)`)

type PanicReport struct {
	RecoverErr         error
	StackTrace         string
	RequestUri         string
	RequesterIpAddress string
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

func isRequesterTrustworthy(echoContext echo.Context) bool {
	rawTrustedIps := os.Getenv(PanicHandlerTrustedIpsEnvVarName)
	if rawTrustedIps == "" {
		return false
	}

	rawRequesterIpAddress := echoContext.RealIP()
	if rawRequesterIpAddress == "" {
		return false
	}

	requesterIpAddress, ipErr := tkValueObject.NewIpAddress(rawRequesterIpAddress)
	if ipErr != nil {
		return false
	}
	requesterIpAddressStr := requesterIpAddress.String()

	for staffIp := range strings.SplitSeq(rawTrustedIps, ",") {
		if strings.TrimSpace(staffIp) == requesterIpAddressStr {
			return true
		}
	}

	return false
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
		slog.String("requesterIpAddress", panicReportPtr.RequesterIpAddress),
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
		"humanReadableMessage": "SomethingWentWrong",
	}

	if !isRequesterTrustworthy(echoContext) {
		jsonResponse["body"] = map[string]any{
			"exceptionCode": shortRecoverErrStr,
		}
	}

	echoContext.JSON(statusCode, jsonResponse)

	panicReportPtr.RequestUri = echoContext.Request().RequestURI
	requesterIpAddress, ipErr := tkValueObject.NewIpAddress(echoContext.RealIP())
	if ipErr == nil {
		panicReportPtr.RequesterIpAddress = requesterIpAddress.String()
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
