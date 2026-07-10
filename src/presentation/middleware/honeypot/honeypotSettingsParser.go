package tkPresentationMiddlewareHoneypot

import (
	"os"
	"strconv"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	honeypotAggressivenessEnvVarName string = "HONEYPOT_AGGRESSIVENESS"
	honeypotActivePathsEnvVarName    string = "HONEYPOT_ACTIVE_PATHS"
	honeypotMaxEntriesEnvVarName     string = "HONEYPOT_MAX_ENTRIES"
	honeypotMaxStreamSizeEnvVarName  string = "HONEYPOT_MAX_STREAM_SIZE"
	honeypotStatsIntervalEnvVarName  string = "HONEYPOT_STATS_INTERVAL"
	honeypotBanDurationEnvVarName    string = "HONEYPOT_BAN_DURATION"
	honeypotRandomSeedEnvVarName     string = "HONEYPOT_RANDOM_SEED"

	honeypotDefaultAggressiveness string = "balanced"
	honeypotDefaultActivePaths    int    = 30
	honeypotDefaultMaxEntries     uint64 = 5000
	honeypotDefaultMaxStreamSize  uint64 = 20 * 1024 * 1024
	honeypotDefaultStatsInterval  string = "30m"
	honeypotDefaultBanDuration    string = "24h"
	honeypotDefaultRandomSeed     int64  = 0
)

func readHoneypotAggressivenessMode() (
	tkValueObject.HoneypotAggressivenessMode, error,
) {
	rawValue := os.Getenv(honeypotAggressivenessEnvVarName)
	if rawValue == "" {
		rawValue = honeypotDefaultAggressiveness
	}
	return tkValueObject.NewHoneypotAggressivenessMode(rawValue)
}

func readHoneypotActivePathCount(ceiling int) (
	tkValueObject.HoneypotActivePathCount, error,
) {
	rawValue := os.Getenv(honeypotActivePathsEnvVarName)
	if rawValue == "" {
		return tkValueObject.NewHoneypotActivePathCount(
			honeypotDefaultActivePaths, ceiling,
		)
	}
	intValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return tkValueObject.NewHoneypotActivePathCount(rawValue, ceiling)
	}
	return tkValueObject.NewHoneypotActivePathCount(intValue, ceiling)
}

func readHoneypotMaxEntries() (tkValueObject.HoneypotMaxEntries, error) {
	rawValue := os.Getenv(honeypotMaxEntriesEnvVarName)
	if rawValue == "" {
		return tkValueObject.NewHoneypotMaxEntries(honeypotDefaultMaxEntries)
	}
	uintValue, err := strconv.ParseUint(rawValue, 10, 64)
	if err != nil {
		return tkValueObject.NewHoneypotMaxEntries(rawValue)
	}
	return tkValueObject.NewHoneypotMaxEntries(uintValue)
}

func readHoneypotMaxStreamSize() (tkValueObject.HoneypotMaxStreamSize, error) {
	rawValue := os.Getenv(honeypotMaxStreamSizeEnvVarName)
	if rawValue == "" {
		return tkValueObject.NewHoneypotMaxStreamSize(
			honeypotDefaultMaxStreamSize,
		)
	}
	uintValue, err := strconv.ParseUint(rawValue, 10, 64)
	if err != nil {
		return tkValueObject.NewHoneypotMaxStreamSize(rawValue)
	}
	return tkValueObject.NewHoneypotMaxStreamSize(uintValue)
}

func readHoneypotStatsInterval() (tkValueObject.HoneypotStatsInterval, error) {
	rawValue := os.Getenv(honeypotStatsIntervalEnvVarName)
	if rawValue == "" {
		rawValue = honeypotDefaultStatsInterval
	}
	return tkValueObject.NewHoneypotStatsInterval(rawValue)
}

func readHoneypotBanDuration() (tkValueObject.HoneypotBanDuration, error) {
	rawValue := os.Getenv(honeypotBanDurationEnvVarName)
	if rawValue == "" {
		rawValue = honeypotDefaultBanDuration
	}
	return tkValueObject.NewHoneypotBanDuration(rawValue)
}

func readHoneypotRandomSeed() (int64, error) {
	rawValue := os.Getenv(honeypotRandomSeedEnvVarName)
	if rawValue == "" {
		return honeypotDefaultRandomSeed, nil
	}
	return strconv.ParseInt(rawValue, 10, 64)
}

func ParseHoneypotSettings(pathPoolSize int) (
	tkDto.HoneypotSettings, error,
) {
	aggressivenessMode, err := readHoneypotAggressivenessMode()
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	activePathCount, err := readHoneypotActivePathCount(pathPoolSize)
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	maxEntries, err := readHoneypotMaxEntries()
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	maxStreamSize, err := readHoneypotMaxStreamSize()
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	statsInterval, err := readHoneypotStatsInterval()
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	banDuration, err := readHoneypotBanDuration()
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	randomSeed, err := readHoneypotRandomSeed()
	if err != nil {
		return tkDto.HoneypotSettings{}, err
	}

	return tkDto.HoneypotSettings{
		AggressivenessMode: aggressivenessMode,
		ActivePathCount:    activePathCount,
		MaxEntries:         maxEntries,
		MaxStreamSize:      maxStreamSize,
		StatsInterval:      statsInterval,
		BanDuration:        banDuration,
		RandomSeed:         randomSeed,
	}, nil
}
