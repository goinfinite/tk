package tkPresentation

import (
	"embed"
	"log/slog"
	"math/rand"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

//go:embed honeypot/payloads/*
var honeypotPayloadsFs embed.FS

var honeypotPayloadEntries = []string{
	"/.env", "env", "text/plain",
	"/wp-config.php", "wp-config.php", "application/x-httpd-php",
	"/wp-config.php.bak", "wp-config.php.bak", "application/x-httpd-php",
	"/config.php", "config.php", "application/x-httpd-php",
	"/backup.sql", "backup.sql", "application/sql",
	"/backup.zip", "backup.zip", "application/zip",
	"/.git/config", "git-config", "text/plain",
	"/.aws/credentials", "aws-credentials", "text/plain",
	"/actuator/env", "actuator-env.json", "application/json",
	"/actuator/configprops", "actuator-configprops.json", "application/json",
	"/server-status", "server-status.html", "text/html",
	"/phpmyadmin/index.php", "phpmyadmin-index.php", "text/html",
	"/admin.php", "admin.php", "text/html",
	"/administrator/index.php", "administrator-index.php", "text/html",
	"/login.php", "login.php", "text/html",
	"/shell.php", "shell.php", "application/x-httpd-php",
	"/cmd.php", "cmd.php", "application/x-httpd-php",
	"/test.php", "test.php", "application/x-httpd-php",
	"/.htaccess", "htaccess", "text/plain",
	"/web.config", "web.config", "text/xml",
	"/robots.txt", "robots.txt", "text/plain",
	"/sitemap.xml", "sitemap.xml", "application/xml",
	"/debug.php", "debug.php", "application/x-httpd-php",
	"/info.php", "info.php", "application/x-httpd-php",
	"/console", "console.html", "text/html",
}

type HoneypotPathClass int

const (
	HoneypotPathClassStaticVuln HoneypotPathClass = iota
	HoneypotPathClassBandwidthExhaust
	HoneypotPathClassAITrap
)

var bandwidthExhaustCandidatePaths = []string{
	"/api/v1/stream/logs",
	"/api/v1/events",
	"/api/v1/feed",
	"/api/v1/export/data",
	"/api/v1/debug/trace",
	"/api/v1/reports/generate",
	"/api/v1/data/export",
	"/api/v2/analytics/stream",
	"/api/v1/monitor/telemetry",
	"/api/v1/ingest/metrics",
}

var aiTrapCandidatePaths = []string{
	"/api/v1/docs",
	"/api/v1/status/detailed",
	"/api/v1/logs/access",
	"/api/v1/metrics/prometheus",
	"/api/v1/diagnostics/dump",
	"/api/v1/audit/trail",
	"/api/v1/events/log",
	"/api/v1/reports/summary",
	"/api/server-info",
	"/api/status",
}

func buildHoneypotPayloadMap(
	extraRoutes []HoneypotPathMapping,
) map[string]HoneypotPathMapping {
	payloadMap := make(map[string]HoneypotPathMapping)
	for entryIdx := 0; entryIdx < len(honeypotPayloadEntries); entryIdx += 3 {
		interceptPath := honeypotPayloadEntries[entryIdx]
		payloadContent, readErr := honeypotPayloadsFs.ReadFile(
			"honeypot/payloads/" + honeypotPayloadEntries[entryIdx+1],
		)
		if readErr != nil {
			slog.Error("HoneypotPayloadReadFailed",
				slog.String("path",
					honeypotPayloadEntries[entryIdx+1]),
				slog.String("err", readErr.Error()))
			continue
		}
		urlPath, pathErr := tkValueObject.NewUrlPath(interceptPath)
		if pathErr != nil {
			slog.Error("HoneypotPathConstructionFailed",
				slog.String("path", interceptPath),
				slog.String("err", pathErr.Error()))
			continue
		}
		mimeType, mimeErr := tkValueObject.NewMimeType(
			honeypotPayloadEntries[entryIdx+2],
		)
		if mimeErr != nil {
			slog.Error("HoneypotMimeConstructionFailed",
				slog.String("err", mimeErr.Error()))
			continue
		}
		payloadMap[interceptPath] = HoneypotPathMapping{
			Body: string(payloadContent),
			MimeType: mimeType,
			UrlPath: urlPath,
		}
	}
	for _, extraRoute := range extraRoutes {
		payloadMap[extraRoute.UrlPath.String()] = extraRoute
	}
	return payloadMap
}

func extractStaticPathKeys(
	payloadEntries []string,
) []string {
	staticPathKeys := make(
		[]string, 0, len(payloadEntries)/3,
	)
	for entryIdx := 0; entryIdx < len(payloadEntries); entryIdx += 3 {
		staticPathKeys = append(
			staticPathKeys, payloadEntries[entryIdx],
		)
	}
	return staticPathKeys
}

func computeAutoRatio(
	activePathCount int,
) (staticCount, bandwidthCount, aiTrapCount int) {
	bandwidthCount = activePathCount / 6
	if bandwidthCount < 1 {
		bandwidthCount = 1
	}
	aiTrapCount = activePathCount / 6
	if aiTrapCount < 1 {
		aiTrapCount = 1
	}
	staticCount = activePathCount -
		bandwidthCount - aiTrapCount
	if staticCount < 0 {
		staticCount = 0
	}
	return
}

func shuffleAndTake(
	candidatePaths []string,
	takeCount int,
	rng *rand.Rand,
) []string {
	if takeCount <= 0 {
		return nil
	}
	if takeCount >= len(candidatePaths) {
		return candidatePaths
	}
	shuffledPaths := make([]string, len(candidatePaths))
	copy(shuffledPaths, candidatePaths)
	rng.Shuffle(len(shuffledPaths), func(i, j int) {
		shuffledPaths[i], shuffledPaths[j] =
			shuffledPaths[j], shuffledPaths[i]
	})
	return shuffledPaths[:takeCount]
}

func newRng(seed int64) *rand.Rand {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return rand.New(rand.NewSource(seed))
}

func selectActivePaths(
	staticPaths []string,
	bandwidthPaths []string,
	aiTrapPaths []string,
	activePathCount int,
	randomSeed int64,
) map[string]HoneypotPathClass {
	staticCount, bandwidthCount, aiTrapCount :=
		computeAutoRatio(activePathCount)
	rng := newRng(randomSeed)
	activePathMap := make(map[string]HoneypotPathClass)
	for _, selectedPath := range shuffleAndTake(
		staticPaths, staticCount, rng,
	) {
		activePathMap[selectedPath] =
			HoneypotPathClassStaticVuln
	}
	for _, selectedPath := range shuffleAndTake(
		bandwidthPaths, bandwidthCount, rng,
	) {
		activePathMap[selectedPath] =
			HoneypotPathClassBandwidthExhaust
	}
	for _, selectedPath := range shuffleAndTake(
		aiTrapPaths, aiTrapCount, rng,
	) {
		activePathMap[selectedPath] =
			HoneypotPathClassAITrap
	}
	return activePathMap
}
