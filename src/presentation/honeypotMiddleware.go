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
	"sync"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

//go:embed honeypot/payloads/*
var honeypotPayloadsFs embed.FS

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
	Body        string
	ContentType string
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
	honeypotCmdRepo       tkRepository.HoneypotCmdRepo
	honeypotHttpPayloads  honeypotPayloadLoader
	honeypotQueryRepo     tkRepository.HoneypotQueryRepo
	honeypotRecordCode    tkValueObject.ActivityRecordCode
	honeypotRecordLevel   tkValueObject.ActivityRecordLevel
	ipExtractor           RequesterIpExtractor
	settings              HoneypotMiddlewareSettings
	writeMu               sync.Mutex
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
	if middleware.honeypotCmdRepo == nil {
		return
	}

	middleware.writeMu.Lock()
	defer middleware.writeMu.Unlock()

	ipAddr, ipErr := tkValueObject.NewIpAddress(ipString)
	if ipErr != nil {
		return
	}

	middleware.honeypotCmdRepo.IncrementHit(ipAddr, interceptPath)

	if rand.Float64() < 0.02 {
		middleware.honeypotCmdRepo.EnforceMaxEntries(
			middleware.settings.MaxEntries.Int(),
		)
	}
}

func (middleware *HoneypotMiddleware) determineBanTier(
	ipString string,
) int {
	if middleware.honeypotQueryRepo == nil {
		return 0
	}

	ipAddr, ipErr := tkValueObject.NewIpAddress(ipString)
	if ipErr != nil {
		return 0
	}

	hitData, readErr := middleware.honeypotQueryRepo.ReadHitRecord(
		ipAddr,
	)
	if readErr != nil {
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

func (middleware *HoneypotMiddleware) reportStats() {
	if middleware.honeypotQueryRepo == nil {
		return
	}

	statsReport, reportErr := middleware.honeypotQueryRepo.ReadReport(
		middleware.settings.BanDuration,
		middleware.settings.AggressivenessMode,
	)
	if reportErr != nil {
		slog.Debug("HoneypotStatsReadReportFailed",
			slog.String("err", reportErr.Error()))
		return
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
	if middleware.honeypotCmdRepo != nil {
		middleware.honeypotCmdRepo.CleanExpiredEntries(
			middleware.settings.BanDuration,
		)
		middleware.honeypotCmdRepo.EnforceMaxEntries(
			middleware.settings.MaxEntries.Int(),
		)
	}

	if middleware.honeypotQueryRepo == nil {
		return
	}

	entryCount := middleware.honeypotQueryRepo.Count()
	if entryCount == 0 {
		return
	}

	middleware.reportStats()
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
	honeypotCmdRepo tkRepository.HoneypotCmdRepo,
	honeypotQueryRepo tkRepository.HoneypotQueryRepo,
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
		honeypotCmdRepo:       honeypotCmdRepo,
		honeypotHttpPayloads:  *payloadLoader,
		honeypotQueryRepo:     honeypotQueryRepo,
		honeypotRecordCode:    honeypotRecordCode,
		honeypotRecordLevel:   tkValueObject.ActivityRecordLevelSecurity,
		ipExtractor:           NewRequesterIpExtractor(),
		settings:              settings,
	}

	go middleware.honeypotMaintenanceWatchdog(ctx)

	return middleware
}
