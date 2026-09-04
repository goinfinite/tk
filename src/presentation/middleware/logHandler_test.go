package tkPresentationMiddleware

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLogLevelParser(t *testing.T) {
	logHandler := LogHandler{}

	testCases := []struct {
		configuredLevel string
		expectedLevel   zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"WARNING", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"panic", zerolog.PanicLevel},
		{"garbage", zerolog.WarnLevel},
	}

	for _, testCase := range testCases {
		t.Run(testCase.configuredLevel, func(t *testing.T) {
			parsedLevel := logHandler.logLevelParser(testCase.configuredLevel)
			if parsedLevel != testCase.expectedLevel {
				t.Errorf(
					"LevelMismatch: '%s' expected %d, got %d",
					testCase.configuredLevel, testCase.expectedLevel, parsedLevel,
				)
			}
		})
	}
}
