package tkPresentation

import (
	"log/slog"
	"os"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type honeypotSettingsParser struct{}

func (parser honeypotSettingsParser) parseBanDuration(
	settings HoneypotMiddlewareSettings,
) tkValueObject.HoneypotBanDuration {
	banDuration, _ := tkValueObject.NewHoneypotBanDuration(
		settings.BanDuration.Duration(),
	)
	return banDuration
}

func (parser honeypotSettingsParser) parseRedirectUrl(
	settings HoneypotMiddlewareSettings,
) tkValueObject.Url {
	if settings.RedirectUrl.String() != "" {
		return settings.RedirectUrl
	}
	defaultRedirectUrl, _ := tkValueObject.NewUrl(
		"https://xkcd.com/",
	)
	return defaultRedirectUrl
}

func (parser honeypotSettingsParser) parseMaxEntries(
	settings HoneypotMiddlewareSettings,
) tkValueObject.HoneypotMaxEntries {
	rawMaxEntries := os.Getenv("HONEYPOT_MAX_ENTRIES")
	if rawMaxEntries == "" {
		if settings.MaxEntries.Int() == 0 {
			maxEntries, _ := tkValueObject.NewHoneypotMaxEntries("")
			return maxEntries
		}
		return settings.MaxEntries
	}

	maxEntries, maxEntriesErr := tkValueObject.NewHoneypotMaxEntries(
		rawMaxEntries,
	)
	if maxEntriesErr != nil {
		slog.Error("HoneypotMaxEntriesInvalid",
			slog.String("err", maxEntriesErr.Error()))
		if settings.MaxEntries.Int() == 0 {
			maxEntries, _ = tkValueObject.NewHoneypotMaxEntries("")
			return maxEntries
		}
		return settings.MaxEntries
	}
	return maxEntries
}

func (parser honeypotSettingsParser) parseActivePathCount(
	settings HoneypotMiddlewareSettings,
	poolCeiling int,
) tkValueObject.HoneypotActivePathCount {
	rawActivePaths := os.Getenv("HONEYPOT_ACTIVE_PATHS")
	if rawActivePaths == "" {
		if settings.ActivePathCount.Int() == 0 {
			activePathCount, _ := tkValueObject.
				NewHoneypotActivePathCount("", poolCeiling)
			return activePathCount
		}
		if poolCeiling > 0 &&
			settings.ActivePathCount.Int() > poolCeiling {
			clamped, _ := tkValueObject.
				NewHoneypotActivePathCount(
					poolCeiling, poolCeiling,
				)
			return clamped
		}
		return settings.ActivePathCount
	}

	activePathCount, countErr := tkValueObject.
		NewHoneypotActivePathCount(rawActivePaths, poolCeiling)
	if countErr != nil {
		slog.Error("HoneypotActivePathCountInvalid",
			slog.String("err", countErr.Error()))
		return settings.ActivePathCount
	}
	return activePathCount
}

func (parser honeypotSettingsParser) parseMaxStreamSizeBytes(
	settings HoneypotMiddlewareSettings,
) tkValueObject.HoneypotMaxStreamSizeBytes {
	rawMaxStream := os.Getenv("HONEYPOT_MAX_STREAM_SIZE")
	if rawMaxStream == "" {
		return settings.MaxStreamSizeBytes
	}

	maxStreamSize, streamErr := tkValueObject.
		NewHoneypotMaxStreamSizeBytes(rawMaxStream)
	if streamErr != nil {
		slog.Error("HoneypotMaxStreamSizeInvalid",
			slog.String("err", streamErr.Error()))
		return settings.MaxStreamSizeBytes
	}
	return maxStreamSize
}

func (parser honeypotSettingsParser) parseStatsInterval(
	settings HoneypotMiddlewareSettings,
) tkValueObject.HoneypotStatsInterval {
	rawStatsInterval := os.Getenv("HONEYPOT_STATS_INTERVAL")
	if rawStatsInterval == "" {
		if settings.StatsInterval.Duration() == 0 {
			statsInterval, _ := tkValueObject.
				NewHoneypotStatsInterval("")
			return statsInterval
		}
		return settings.StatsInterval
	}

	statsInterval, intervalErr := tkValueObject.
		NewHoneypotStatsInterval(rawStatsInterval)
	if intervalErr != nil {
		slog.Error("HoneypotStatsIntervalInvalid",
			slog.String("err", intervalErr.Error()))
		if settings.StatsInterval.Duration() == 0 {
			statsInterval, _ = tkValueObject.
				NewHoneypotStatsInterval("")
			return statsInterval
		}
		return settings.StatsInterval
	}
	return statsInterval
}

func (parser honeypotSettingsParser) parseAggressivenessMode(
	settings HoneypotMiddlewareSettings,
) tkValueObject.HoneypotAggressivenessMode {
	rawMode := os.Getenv("HONEYPOT_AGGRESSIVENESS")
	if rawMode == "" {
		if settings.AggressivenessMode.String() == "" {
			return tkValueObject.
				HoneypotAggressivenessModeBalanced
		}
		return settings.AggressivenessMode
	}

	switch rawMode {
	case "standard", "lenient", "passive":
		slog.Error("AggressivenessModeDeprecatedFallback",
			slog.String("deprecated", rawMode),
			slog.String("resolved", "balanced"))
		return tkValueObject.
			HoneypotAggressivenessModeBalanced
	}

	resolvedMode, resolveErr := tkValueObject.
		NewHoneypotAggressivenessMode(rawMode)
	if resolveErr != nil {
		slog.Error("AggressivenessModeInvalidFallback",
			slog.String("invalid", rawMode),
			slog.String("resolved", "balanced"))
		return tkValueObject.
			HoneypotAggressivenessModeBalanced
	}
	return resolvedMode
}

func (parser honeypotSettingsParser) Parse(
	settings HoneypotMiddlewareSettings,
	poolCeiling int,
) HoneypotMiddlewareSettings {
	settings.BanDuration = parser.parseBanDuration(settings)
	settings.RedirectUrl = parser.parseRedirectUrl(settings)
	settings.MaxEntries = parser.parseMaxEntries(settings)
	settings.ActivePathCount = parser.parseActivePathCount(
		settings, poolCeiling,
	)
	settings.MaxStreamSizeBytes = parser.parseMaxStreamSizeBytes(
		settings,
	)
	settings.StatsInterval = parser.parseStatsInterval(settings)
	settings.AggressivenessMode = parser.parseAggressivenessMode(
		settings,
	)
	return settings
}
