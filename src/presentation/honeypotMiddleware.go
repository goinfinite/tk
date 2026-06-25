package tkPresentation

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

//go:embed honeypot/payloads/*
var honeypotPayloadsFs embed.FS

const honeypotHitKeyPrefix = "honeypot:hit:"

type HoneypotPathMapping struct {
	Body     string
	MimeType tkValueObject.MimeType
	UrlPath  tkValueObject.UrlPath
}

type HoneypotMiddlewareSettings struct {
	ActivePathCount    tkValueObject.HoneypotActivePathCount
	AggressivenessMode tkValueObject.HoneypotAggressivenessMode
	BanDuration        time.Duration
	ExtraPathRoutes    []HoneypotPathMapping
	MaxEntries         tkValueObject.HoneypotMaxEntries
	MaxStreamSizeBytes tkValueObject.HoneypotMaxStreamSizeBytes
	RedirectUrl        tkValueObject.Url
	StatsInterval      tkValueObject.HoneypotStatsInterval
}

type HoneypotHttpPayload struct {
	ContentType string
	Body        string
}

type honeypotHitData struct {
	Count      int            `json:"count"`
	FirstHitAt string         `json:"firstHitAt"`
	Endpoints  map[string]int `json:"endpoints"`
}

type honeypotStatsOffender struct {
	IpAddress string `json:"ipAddress"`
	HitCount  int    `json:"hitCount"`
}

type honeypotStatsEndpoint struct {
	Path     string `json:"path"`
	HitCount int    `json:"hitCount"`
}

type honeypotPayloadLoader struct {
	payloadMap map[string]HoneypotHttpPayload
}

func (loader *honeypotPayloadLoader) loadEmbeddedPayloads() error {
	type embeddedFileMetadata struct {
		embeddedFilename string
		contentType      string
	}

	embeddedFileMap := map[string]embeddedFileMetadata{
		"/.env":                    {"env", "text/plain"},
		"/wp-config.php":           {"wp-config.php", "application/x-httpd-php"},
		"/wp-config.php.bak":       {"wp-config.php.bak", "application/x-httpd-php"},
		"/config.php":              {"config.php", "application/x-httpd-php"},
		"/backup.sql":              {"backup.sql", "application/sql"},
		"/backup.zip":              {"backup.zip", "application/zip"},
		"/.git/config":             {"git-config", "text/plain"},
		"/.aws/credentials":        {"aws-credentials", "text/plain"},
		"/actuator/env":            {"actuator-env.json", "application/json"},
		"/actuator/configprops":    {"actuator-configprops.json", "application/json"},
		"/server-status":           {"server-status.html", "text/html"},
		"/phpmyadmin/index.php":    {"phpmyadmin-index.php", "text/html"},
		"/admin.php":               {"admin.php", "text/html"},
		"/administrator/index.php": {"administrator-index.php", "text/html"},
		"/login.php":               {"login.php", "text/html"},
		"/shell.php":               {"shell.php", "application/x-httpd-php"},
		"/cmd.php":                 {"cmd.php", "application/x-httpd-php"},
		"/test.php":                {"test.php", "application/x-httpd-php"},
		"/.htaccess":               {"htaccess", "text/plain"},
		"/web.config":              {"web.config", "text/xml"},
		"/robots.txt":              {"robots.txt", "text/plain"},
		"/sitemap.xml":             {"sitemap.xml", "application/xml"},
		"/debug.php":               {"debug.php", "application/x-httpd-php"},
		"/info.php":                {"info.php", "application/x-httpd-php"},
		"/console":                 {"console.html", "text/html"},
	}

	for interceptPath, fileMetadata := range embeddedFileMap {
		payloadContent, payloadReadErr := honeypotPayloadsFs.ReadFile(
			"honeypot/payloads/" + fileMetadata.embeddedFilename,
		)
		if payloadReadErr != nil {
			slog.Debug("HoneypotPayloadReadFailed",
				slog.String("path", fileMetadata.embeddedFilename),
				slog.String("err", payloadReadErr.Error()))
			return errors.New(
				"HoneypotPayloadLoadFailed: " +
					fileMetadata.embeddedFilename,
			)
		}
		loader.payloadMap[interceptPath] = HoneypotHttpPayload{
			ContentType: fileMetadata.contentType,
			Body:        string(payloadContent),
		}
	}

	return nil
}

func (loader honeypotPayloadLoader) ReadPayload(
	interceptPath string,
) (HoneypotHttpPayload, bool) {
	matchedPayload, isDefaultHoneypotPath := loader.payloadMap[interceptPath]
	return matchedPayload, isDefaultHoneypotPath
}

func (loader honeypotPayloadLoader) totalCandidatePoolSize() int {
	return len(loader.payloadMap)
}

type HoneypotMiddleware struct {
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo
	cancelFunc            context.CancelFunc
	honeypotHttpPayloads  honeypotPayloadLoader
	honeypotRecordCode    tkValueObject.ActivityRecordCode
	honeypotRecordLevel   tkValueObject.ActivityRecordLevel
	ipExtractor           RequesterIpExtractor
	settings              HoneypotMiddlewareSettings
	transientDbSvc        *tkInfraDb.TransientDatabaseService
	writeMu               sync.Mutex
}

func cleanExpiredEntries(
	handler *gorm.DB,
	ttlDuration time.Duration,
) {
	cutoff := time.Now().Add(-ttlDuration)
	handler.Where(
		"created_at < ?", cutoff,
	).Delete(&tkInfraDb.KeyValueModel{})
}

func enforceMaxEntries(
	handler *gorm.DB,
	maxEntries int,
) {
	var totalCount int64
	handler.Model(
		&tkInfraDb.KeyValueModel{},
	).Count(&totalCount)

	if int(totalCount) <= maxEntries {
		return
	}

	excessCount := int(totalCount) - maxEntries
	keysToDelete := make([]string, 0, excessCount)
	handler.Model(
		&tkInfraDb.KeyValueModel{},
	).Order("created_at ASC").Limit(
		excessCount,
	).Pluck("key", &keysToDelete)

	if len(keysToDelete) > 0 {
		handler.Where(
			"key IN ?", keysToDelete,
		).Delete(&tkInfraDb.KeyValueModel{})
	}
}

func (middleware *HoneypotMiddleware) lookupHoneypotPath(
	interceptPath string,
) (HoneypotHttpPayload, bool) {
	matchedPayload, isDefaultHoneypotPath :=
		middleware.honeypotHttpPayloads.ReadPayload(
			interceptPath,
		)
	if isDefaultHoneypotPath {
		return matchedPayload, true
	}

	if middleware.settings.ExtraPathRoutes == nil {
		return HoneypotHttpPayload{}, false
	}

	for _, extraRoute := range middleware.settings.ExtraPathRoutes {
		if extraRoute.UrlPath.String() == interceptPath {
			return HoneypotHttpPayload{
				ContentType: extraRoute.MimeType.String(),
				Body:        extraRoute.Body,
			}, true
		}
	}

	return HoneypotHttpPayload{}, false
}

func (middleware *HoneypotMiddleware) serveHoneypotPayload(
	echoContext echo.Context,
	matchedPayload HoneypotHttpPayload,
) error {
	echoContext.Response().Header().Set(
		"Content-Type", matchedPayload.ContentType,
	)
	return echoContext.String(http.StatusOK, matchedPayload.Body)
}

func (middleware *HoneypotMiddleware) serveBanRedirect(
	echoContext echo.Context,
) error {
	return echoContext.Redirect(
		http.StatusFound,
		middleware.settings.RedirectUrl.String(),
	)
}

func (middleware *HoneypotMiddleware) recordHoneypotHit(
	requesterIp tkValueObject.IpAddress,
	interceptPath string,
) {
	if middleware.activityRecordCmdRepo == nil {
		return
	}

	honeypotHitCreateRequest := tkDto.CreateActivityRecord{
		RecordLevel: middleware.honeypotRecordLevel,
		RecordCode:  middleware.honeypotRecordCode,
		AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		RecordDetails:     map[string]string{"path": interceptPath},
		OperatorIpAddress: &requesterIp,
	}

	createErr := middleware.activityRecordCmdRepo.Create(
		honeypotHitCreateRequest,
	)
	if createErr != nil {
		slog.Debug("HoneypotHitRecordCreationFailed",
			slog.String("err", createErr.Error()))
	}
}

func (middleware *HoneypotMiddleware) incrementHitCount(
	ipString string,
	interceptPath string,
) {
	if middleware.transientDbSvc == nil {
		return
	}

	middleware.writeMu.Lock()
	defer middleware.writeMu.Unlock()

	handler := middleware.transientDbSvc.Handler
	hitKey := honeypotHitKeyPrefix + ipString

	rawValue, readErr := middleware.transientDbSvc.Read(hitKey)

	var hitData honeypotHitData
	if readErr == nil {
		parseErr := json.Unmarshal(
			[]byte(rawValue), &hitData,
		)
		if parseErr != nil {
			slog.Debug("HoneypotHitDataParseFailed",
				slog.String("err", parseErr.Error()))
		}
	}

	if hitData.Count == 0 {
		hitData.FirstHitAt = time.Now().UTC().Format(
			time.RFC3339,
		)
		hitData.Endpoints = make(map[string]int)
	}

	hitData.Count++
	hitData.Endpoints[interceptPath]++

	jsonBytes, marshalErr := json.Marshal(hitData)
	if marshalErr != nil {
		slog.Debug("HoneypotHitDataMarshalFailed",
			slog.String("err", marshalErr.Error()))
		return
	}

	setErr := middleware.transientDbSvc.Set(
		hitKey, string(jsonBytes),
	)
	if setErr != nil {
		slog.Debug("HoneypotHitCountSetFailed",
			slog.String("err", setErr.Error()))
		return
	}

	if rand.Float64() < 0.02 {
		enforceMaxEntries(
			handler,
			middleware.settings.MaxEntries.Int(),
		)
	}
}

func (middleware *HoneypotMiddleware) determineBanTier(
	ipString string,
) int {
	if middleware.transientDbSvc == nil {
		return 0
	}

	hitKey := honeypotHitKeyPrefix + ipString
	rawValue, readErr := middleware.transientDbSvc.Read(hitKey)
	if readErr != nil {
		return 0
	}

	var hitData honeypotHitData
	parseErr := json.Unmarshal([]byte(rawValue), &hitData)
	if parseErr != nil {
		return 0
	}

	firstHitAt, timeParseErr := time.Parse(
		time.RFC3339, hitData.FirstHitAt,
	)
	if timeParseErr != nil {
		return 0
	}

	banWindowStart := time.Now().Add(
		-middleware.settings.BanDuration,
	)
	if firstHitAt.Before(banWindowStart) {
		return 0
	}

	return middleware.settings.AggressivenessMode.ResolveTier(
		hitData.Count,
	)
}

func buildTopOffenders(
	ipHitCounts map[string]int,
	limit int,
) []honeypotStatsOffender {
	offenders := make(
		[]honeypotStatsOffender, 0, len(ipHitCounts),
	)
	for ipAddr, hitCount := range ipHitCounts {
		offenders = append(offenders, honeypotStatsOffender{
			IpAddress: ipAddr,
			HitCount:  hitCount,
		})
	}
	sort.Slice(offenders, func(a, b int) bool {
		return offenders[a].HitCount > offenders[b].HitCount
	})
	if len(offenders) > limit {
		offenders = offenders[:limit]
	}
	return offenders
}

func buildTopEndpoints(
	endpointCounts map[string]int,
	limit int,
) []honeypotStatsEndpoint {
	endpoints := make(
		[]honeypotStatsEndpoint, 0, len(endpointCounts),
	)
	for endpointPath, hitCount := range endpointCounts {
		endpoints = append(endpoints, honeypotStatsEndpoint{
			Path:     endpointPath,
			HitCount: hitCount,
		})
	}
	sort.Slice(endpoints, func(a, b int) bool {
		return endpoints[a].HitCount > endpoints[b].HitCount
	})
	if len(endpoints) > limit {
		endpoints = endpoints[:limit]
	}
	return endpoints
}

func (middleware *HoneypotMiddleware) aggregateStats() {
	if middleware.transientDbSvc == nil {
		return
	}

	entries, readAllErr := middleware.transientDbSvc.ReadAll()
	if readAllErr != nil {
		slog.Debug("HoneypotStatsReadAllFailed",
			slog.String("err", readAllErr.Error()))
		return
	}

	ipHitCounts := make(map[string]int)
	endpointCounts := make(map[string]int)
	bannedIpCount := 0

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Key, honeypotHitKeyPrefix) {
			continue
		}

		var hitData honeypotHitData
		parseErr := json.Unmarshal(
			[]byte(entry.Value), &hitData,
		)
		if parseErr != nil {
			continue
		}

		ipAddr := strings.TrimPrefix(
			entry.Key, honeypotHitKeyPrefix,
		)
		resolvedTier := middleware.settings.AggressivenessMode.ResolveTier(
			hitData.Count,
		)
		if resolvedTier >= 2 {
			bannedIpCount++
		}

		ipHitCounts[ipAddr] = hitData.Count
		for endpointPath, count := range hitData.Endpoints {
			endpointCounts[endpointPath] += count
		}
	}

	topOffenders := buildTopOffenders(ipHitCounts, 10)
	topEndpoints := buildTopEndpoints(endpointCounts, 10)

	statsReport := map[string]any{
		"bannedIpCount": bannedIpCount,
		"topOffenders":  topOffenders,
		"topEndpoints":  topEndpoints,
	}

	reportJson, marshalErr := json.Marshal(statsReport)
	if marshalErr != nil {
		slog.Debug("HoneypotStatsMarshalFailed",
			slog.String("err", marshalErr.Error()))
		return
	}

	if middleware.activityRecordCmdRepo == nil {
		return
	}

	statsRecordCode, _ := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)

	createRequest := tkDto.CreateActivityRecord{
		RecordLevel: tkValueObject.ActivityRecordLevelSecurity,
		RecordCode:  statsRecordCode,
		AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		RecordDetails: map[string]string{
			"statsReport": string(reportJson),
		},
	}

	createErr := middleware.activityRecordCmdRepo.Create(
		createRequest,
	)
	if createErr != nil {
		slog.Debug("HoneypotStatsReportCreationFailed",
			slog.String("err", createErr.Error()))
	}
}

func (middleware *HoneypotMiddleware) runMaintenanceTick() {
	if middleware.transientDbSvc == nil {
		return
	}

	handler := middleware.transientDbSvc.Handler
	ttlDuration := middleware.settings.BanDuration

	cleanExpiredEntries(handler, ttlDuration)
	enforceMaxEntries(
		handler, middleware.settings.MaxEntries.Int(),
	)

	entryCount := middleware.transientDbSvc.Count()
	if entryCount == 0 {
		return
	}

	middleware.aggregateStats()
}

func (middleware *HoneypotMiddleware) honeypotMaintenanceWatchdog(
	ctx context.Context,
) {
	tickInterval := middleware.settings.StatsInterval.Duration()
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				defer func() {
					if rv := recover(); rv != nil {
						slog.Debug(
							"HoneypotWatchdogTickPanicRecovered",
							slog.Any("panic", rv),
						)
					}
				}()
				middleware.runMaintenanceTick()
			}()
		}
	}
}

func (middleware *HoneypotMiddleware) MiddlewareFunc() echo.MiddlewareFunc {
	return middleware.Execute
}

func (middleware *HoneypotMiddleware) Execute(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(echoContext echo.Context) error {
		httpRequest := echoContext.Request()

		requesterIp, ipExtractionErr := middleware.ipExtractor.Execute(
			httpRequest,
		)
		if ipExtractionErr != nil {
			return next(echoContext)
		}

		ipString := requesterIp.String()
		banTier := middleware.determineBanTier(ipString)

		if banTier >= 3 {
			return middleware.serveBanRedirect(echoContext)
		}

		matchedPayload, isHoneypotPath := middleware.lookupHoneypotPath(
			httpRequest.URL.Path,
		)
		if !isHoneypotPath {
			return next(echoContext)
		}

		if banTier >= 2 {
			return middleware.serveBanRedirect(echoContext)
		}

		middleware.incrementHitCount(
			ipString, httpRequest.URL.Path,
		)
		middleware.recordHoneypotHit(
			requesterIp, httpRequest.URL.Path,
		)
		return middleware.serveHoneypotPayload(
			echoContext, matchedPayload,
		)
	}
}

func (middleware *HoneypotMiddleware) Stop() {
	if middleware.cancelFunc != nil {
		middleware.cancelFunc()
	}
}

func NewHoneypotMiddleware(
	settings HoneypotMiddlewareSettings,
	transientDbSvc *tkInfraDb.TransientDatabaseService,
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
) *HoneypotMiddleware {
	if settings.BanDuration <= 0 {
		settings.BanDuration = 24 * time.Hour
	}

	if settings.RedirectUrl.String() == "" {
		defaultRedirectUrl, _ := tkValueObject.NewUrl(
			"https://xkcd.com/",
		)
		settings.RedirectUrl = defaultRedirectUrl
	}

	if rawMaxEntries := os.Getenv("HONEYPOT_MAX_ENTRIES"); rawMaxEntries != "" {
		maxEntries, maxEntriesErr := tkValueObject.NewHoneypotMaxEntries(
			rawMaxEntries,
		)
		if maxEntriesErr != nil {
			slog.Debug("HoneypotMaxEntriesInvalid",
				slog.String("err", maxEntriesErr.Error()))
		} else {
			settings.MaxEntries = maxEntries
		}
	}
	if settings.MaxEntries.Int() == 0 {
		settings.MaxEntries, _ = tkValueObject.NewHoneypotMaxEntries("")
	}

	payloadLoader := &honeypotPayloadLoader{
		payloadMap: make(map[string]HoneypotHttpPayload),
	}
	loaderErr := payloadLoader.loadEmbeddedPayloads()
	if loaderErr != nil {
		slog.Debug("HoneypotPayloadLoaderCreationFailed",
			slog.String("err", loaderErr.Error()))
	}

	poolCeiling := payloadLoader.totalCandidatePoolSize()
	if rawActivePaths := os.Getenv("HONEYPOT_ACTIVE_PATHS"); rawActivePaths != "" {
		activePathCount, activePathCountErr := tkValueObject.NewHoneypotActivePathCount(
			rawActivePaths, poolCeiling,
		)
		if activePathCountErr != nil {
			slog.Debug("HoneypotActivePathCountInvalid",
				slog.String("err", activePathCountErr.Error()))
		} else {
			settings.ActivePathCount = activePathCount
		}
	}

	if rawMaxStream := os.Getenv("HONEYPOT_MAX_STREAM_SIZE"); rawMaxStream != "" {
		maxStreamSize, maxStreamSizeErr := tkValueObject.NewHoneypotMaxStreamSizeBytes(
			rawMaxStream,
		)
		if maxStreamSizeErr != nil {
			slog.Debug("HoneypotMaxStreamSizeInvalid",
				slog.String("err", maxStreamSizeErr.Error()))
		} else {
			settings.MaxStreamSizeBytes = maxStreamSize
		}
	}

	rawStatsInterval := os.Getenv(
		"HONEYPOT_STATS_INTERVAL",
	)
	if rawStatsInterval != "" {
		statsInterval, statsIntervalErr := tkValueObject.NewHoneypotStatsInterval(
			rawStatsInterval,
		)
		if statsIntervalErr != nil {
			slog.Debug("HoneypotStatsIntervalInvalid",
				slog.String("err", statsIntervalErr.Error()))
		} else {
			settings.StatsInterval = statsInterval
		}
	}
	if settings.StatsInterval.Duration() == 0 {
		settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval("")
	}

	rawAggressivenessMode := os.Getenv("HONEYPOT_AGGRESSIVENESS")
	if rawAggressivenessMode != "" {
		settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeBalanced
		switch rawAggressivenessMode {
		case "standard", "lenient", "passive":
			slog.Debug("AggressivenessModeDeprecatedFallback",
				slog.String("deprecated", rawAggressivenessMode),
				slog.String("resolved", "balanced"))
		default:
			resolvedMode, resolveErr := tkValueObject.NewHoneypotAggressivenessMode(
				rawAggressivenessMode,
			)
			if resolveErr != nil {
				slog.Debug("AggressivenessModeInvalidFallback",
					slog.String("invalid", rawAggressivenessMode),
					slog.String("resolved", "balanced"))
			} else {
				settings.AggressivenessMode = resolvedMode
			}
		}
	} else if settings.AggressivenessMode.String() == "" {
		settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeBalanced
	}

	honeypotRecordCode, codeErr := tkValueObject.NewActivityRecordCode(
		"HoneypotHit",
	)
	if codeErr != nil {
		slog.Debug("HoneypotCodeCreationFailed",
			slog.String("err", codeErr.Error()))
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	middleware := &HoneypotMiddleware{
		activityRecordCmdRepo: activityRecordCmdRepo,
		cancelFunc:            cancelFunc,
		honeypotHttpPayloads:  *payloadLoader,
		honeypotRecordCode:    honeypotRecordCode,
		honeypotRecordLevel:   tkValueObject.ActivityRecordLevelSecurity,
		ipExtractor:           NewRequesterIpExtractor(),
		settings:              settings,
		transientDbSvc:        transientDbSvc,
	}

	go middleware.honeypotMaintenanceWatchdog(ctx)

	return middleware
}
