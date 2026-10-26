package tkPresentationMiddleware

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLogLevelParser(test *testing.T) {
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
		test.Run(testCase.configuredLevel, func(subtest *testing.T) {
			parsedLevel := logHandler.logLevelParser(testCase.configuredLevel)
			if parsedLevel != testCase.expectedLevel {
				subtest.Errorf(
					"LevelMismatch: '%s' expected %d, got %d",
					testCase.configuredLevel, testCase.expectedLevel, parsedLevel,
				)
			}
		})
	}
}
