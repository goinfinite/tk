package tkPresentation

import (
	"embed"
	"encoding/base64"
	"log/slog"
	"math/rand"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

//go:embed honeypot/payloads/*.bin
var honeypotPayloadsFs embed.FS

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

func decodePayloadSpec(
	spec honeypotPayloadSpec,
) (HoneypotPathMapping, error) {
	binContent, readErr := honeypotPayloadsFs.ReadFile(
		"honeypot/payloads/" + spec.binFileName,
	)
	if readErr != nil {
		return HoneypotPathMapping{}, readErr
	}
	decodedBytes, decodeErr := base64.StdEncoding.DecodeString(
		string(binContent),
	)
	if decodeErr != nil {
		return HoneypotPathMapping{}, decodeErr
	}
	urlPath, pathErr := tkValueObject.NewUrlPath(
		spec.urlPath,
	)
	if pathErr != nil {
		return HoneypotPathMapping{}, pathErr
	}
	mimeType, mimeErr := tkValueObject.NewMimeType(
		spec.mimeType,
	)
	if mimeErr != nil {
		return HoneypotPathMapping{}, mimeErr
	}
	return HoneypotPathMapping{
		Body:     string(decodedBytes),
		MimeType: mimeType,
		UrlPath:  urlPath,
	}, nil
}

func extractStaticPathKeys() []string {
	staticPathKeys := make(
		[]string, 0, len(honeypotPayloadSpecs),
	)
	for _, spec := range honeypotPayloadSpecs {
		staticPathKeys = append(
			staticPathKeys, spec.urlPath,
		)
	}
	return staticPathKeys
}

func buildHoneypotPayloadMap(
	activePathMap map[string]HoneypotPathClass,
	extraRoutes []HoneypotPathMapping,
) map[string]HoneypotPathMapping {
	payloadMap := make(map[string]HoneypotPathMapping)
	failedPaths := make([]string, 0)
	for activePath, pathClass := range activePathMap {
		if pathClass != HoneypotPathClassStaticVuln {
			continue
		}
		spec := findPayloadSpec(activePath)
		if spec == nil {
			continue
		}
		mapping, decodeErr := decodePayloadSpec(*spec)
		if decodeErr != nil {
			slog.Error(
				"HoneypotPayloadDecodeFailed",
				slog.String("path", activePath),
				slog.String("err",
					decodeErr.Error()),
			)
			failedPaths = append(
				failedPaths, activePath,
			)
			continue
		}
		payloadMap[activePath] = mapping
	}
	resolveDecodeFailures(
		activePathMap, payloadMap, failedPaths,
	)
	for _, extraRoute := range extraRoutes {
		payloadMap[extraRoute.UrlPath.String()] =
			extraRoute
	}
	return payloadMap
}

func resolveDecodeFailures(
	activePathMap map[string]HoneypotPathClass,
	payloadMap map[string]HoneypotPathMapping,
	failedPaths []string,
) {
	dormantPaths := make([]string, 0)
	for _, spec := range honeypotPayloadSpecs {
		if _, isActive := activePathMap[spec.urlPath]; !isActive {
			dormantPaths = append(
				dormantPaths, spec.urlPath,
			)
		}
	}
	replacementIdx := 0
	for _, failedPath := range failedPaths {
		delete(activePathMap, failedPath)
		replacementFound := false
		for replacementIdx < len(dormantPaths) {
			replacementPath := dormantPaths[replacementIdx]
			replacementIdx++
			spec := findPayloadSpec(replacementPath)
			if spec == nil {
				continue
			}
			mapping, decodeErr := decodePayloadSpec(*spec)
			if decodeErr != nil {
				slog.Error(
					"HoneypotPayloadDecodeFailed",
					slog.String("path",
						replacementPath),
					slog.String("err",
						decodeErr.Error()),
				)
				continue
			}
			activePathMap[replacementPath] =
				HoneypotPathClassStaticVuln
			payloadMap[replacementPath] = mapping
			replacementFound = true
			break
		}
		if !replacementFound {
			slog.Error(
				"HoneypotReplacementExhausted",
				slog.String("failedPath",
					failedPath),
			)
		}
	}
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
