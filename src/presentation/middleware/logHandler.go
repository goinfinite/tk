package tkPresentationMiddleware

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	slogZerolog "github.com/samber/slog-zerolog/v2"

	tkInfra "github.com/goinfinite/tk/src/infra"
)

const (
	LogHandlerLogLevelEnvVarName string = "LOG_LEVEL"
)

type LogHandler struct {
}

func (LogHandler) ReadLevel() string {
	return os.Getenv(LogHandlerLogLevelEnvVarName)
}

func (LogHandler) SetLevel(logLevel string) {
	setEnvErr := os.Setenv(LogHandlerLogLevelEnvVarName, logLevel)
	if setEnvErr != nil {
		slog.Error(
			"LogLevelSetEnvError",
			slog.String("error", setEnvErr.Error()),
		)
	}
}

func (LogHandler) logLevelParser(configuredLevel string) zerolog.Level {
	switch strings.ToLower(configuredLevel) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.WarnLevel
	}
}

func (logHandler LogHandler) Init() {
	logLevel := zerolog.WarnLevel
	if configuredLevel := logHandler.ReadLevel(); configuredLevel != "" {
		logLevel = logHandler.logLevelParser(configuredLevel)
	}

	var logWriter io.Writer = os.Stderr
	if tkInfra.IsStdoutTerminal() && logLevel == zerolog.DebugLevel {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	zerologLogger := zerolog.New(logWriter).Level(logLevel)

	zerologHandler := slogZerolog.Option{Logger: &zerologLogger}.NewZerologHandler()
	slog.SetDefault(slog.New(zerologHandler))
}
