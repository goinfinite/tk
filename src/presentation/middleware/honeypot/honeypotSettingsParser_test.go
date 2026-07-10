package tkPresentationMiddlewareHoneypot

import (
	"os"
	"testing"
)

func TestParseHoneypotSettings(t *testing.T) {
	testCaseStructs := []struct {
		testName           string
		envVars            map[string]string
		inputPathPoolSize  int
		expectError        bool
		checkDefaults      bool
	}{
		{
			testName: "AllEnvVarsSet",
			envVars: map[string]string{
				"HONEYPOT_AGGRESSIVENESS": "immediate",
				"HONEYPOT_ACTIVE_PATHS":   "50",
				"HONEYPOT_MAX_ENTRIES":    "10000",
				"HONEYPOT_MAX_STREAM_SIZE": "10485760",
				"HONEYPOT_STATS_INTERVAL":  "1h",
				"HONEYPOT_BAN_DURATION":    "48h",
				"HONEYPOT_RANDOM_SEED":     "42",
			},
			inputPathPoolSize: 114,
			expectError:       false,
		},
		{
			testName:          "NoEnvVarsSet_AllDefaults",
			envVars:           map[string]string{},
			inputPathPoolSize: 114,
			expectError:       false,
			checkDefaults:     true,
		},
		{
			testName: "ActivePathsAbovePoolSize_ClampedToCeiling",
			envVars: map[string]string{
				"HONEYPOT_ACTIVE_PATHS": "999",
			},
			inputPathPoolSize: 114,
			expectError:       false,
		},
		{
			testName: "InvalidAggressivenessMode",
			envVars: map[string]string{
				"HONEYPOT_AGGRESSIVENESS": "invalid",
			},
			inputPathPoolSize: 114,
			expectError:       true,
		},
		{
			testName: "InvalidMaxEntries",
			envVars: map[string]string{
				"HONEYPOT_MAX_ENTRIES": "notanumber",
			},
			inputPathPoolSize: 114,
			expectError:       true,
		},
		{
			testName: "InvalidBanDuration",
			envVars: map[string]string{
				"HONEYPOT_BAN_DURATION": "invalid",
			},
			inputPathPoolSize: 114,
			expectError:       true,
		},
		{
			testName: "InvalidRandomSeed",
			envVars: map[string]string{
				"HONEYPOT_RANDOM_SEED": "notanumber",
			},
			inputPathPoolSize: 114,
			expectError:       true,
		},
	}

	envVarNames := []string{
		"HONEYPOT_AGGRESSIVENESS",
		"HONEYPOT_ACTIVE_PATHS",
		"HONEYPOT_MAX_ENTRIES",
		"HONEYPOT_MAX_STREAM_SIZE",
		"HONEYPOT_STATS_INTERVAL",
		"HONEYPOT_BAN_DURATION",
		"HONEYPOT_RANDOM_SEED",
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			originalValues := make(map[string]string)
			for _, envName := range envVarNames {
				originalValues[envName] = os.Getenv(envName)
				os.Unsetenv(envName)
			}

			for envName, envValue := range testCase.envVars {
				os.Setenv(envName, envValue)
			}

			settings, err := ParseHoneypotSettings(testCase.inputPathPoolSize)

			for _, envName := range envVarNames {
				os.Setenv(envName, originalValues[envName])
			}

			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError")
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s'", err.Error())
			}

			if testCase.checkDefaults && !testCase.expectError {
				if settings.AggressivenessMode.String() != "balanced" {
					t.Errorf(
						"DefaultAggressivenessMismatch: Expected='balanced', Actual='%s'",
						settings.AggressivenessMode.String(),
					)
				}
				if settings.ActivePathCount.Int() != 30 {
					t.Errorf(
						"DefaultActivePathCountMismatch: Expected=30, Actual=%d",
						settings.ActivePathCount.Int(),
					)
				}
				if settings.MaxEntries.Uint64() != 5000 {
					t.Errorf(
						"DefaultMaxEntriesMismatch: Expected=5000, Actual=%d",
						settings.MaxEntries.Uint64(),
					)
				}
				if settings.MaxStreamSize.Uint64() != 20*1024*1024 {
					t.Errorf(
						"DefaultMaxStreamSizeMismatch: Expected=%d, Actual=%d",
						20*1024*1024, settings.MaxStreamSize.Uint64(),
					)
				}
				if string(settings.StatsInterval) != "30m" {
					t.Errorf(
						"DefaultStatsIntervalMismatch: Expected='30m', Actual='%s'",
						string(settings.StatsInterval),
					)
				}
				if string(settings.BanDuration) != "24h" {
					t.Errorf(
						"DefaultBanDurationMismatch: Expected='24h', Actual='%s'",
						string(settings.BanDuration),
					)
				}
				if settings.RandomSeed != 0 {
					t.Errorf(
						"DefaultRandomSeedMismatch: Expected=0, Actual=%d",
						settings.RandomSeed,
					)
				}
			}

			if testCase.testName == "ActivePathsAbovePoolSize_ClampedToCeiling" {
				if settings.ActivePathCount.Int() != 114 {
					t.Errorf(
						"ActivePathCountNotClamped: Expected=114, Actual=%d",
						settings.ActivePathCount.Int(),
					)
				}
			}
		})
	}
}
