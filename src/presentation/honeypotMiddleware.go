package tkPresentation

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"sync"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkUseCase "github.com/goinfinite/tk/src/domain/useCase"
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
	BanDuration        tkValueObject.HoneypotBanDuration
	ExtraPathRoutes    []HoneypotPathMapping
	MaxEntries         tkValueObject.HoneypotMaxEntries
	MaxStreamSizeBytes tkValueObject.HoneypotMaxStreamSizeBytes
	RedirectUrl        tkValueObject.Url
	StatsInterval      tkValueObject.HoneypotStatsInterval
}
type HoneypotMiddleware struct {
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo
	honeypotCmdRepo       tkRepository.HoneypotCmdRepo
	honeypotPayloads      map[string]HoneypotPathMapping
	honeypotQueryRepo     tkRepository.HoneypotQueryRepo
	honeypotRecordCode    tkValueObject.ActivityRecordCode
	honeypotRecordLevel   tkValueObject.ActivityRecordLevel
	ipExtractor           RequesterIpExtractor
	settings              HoneypotMiddlewareSettings
	writeMu               sync.Mutex
	cancelFunc            context.CancelFunc
}
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
			Body: string(payloadContent), MimeType: mimeType, UrlPath: urlPath,
		}
	}
	for _, extraRoute := range extraRoutes {
		payloadMap[extraRoute.UrlPath.String()] = extraRoute
	}
	return payloadMap
}
func (middleware *HoneypotMiddleware) lookupHoneypotPath(
	interceptPath string,
) (HoneypotPathMapping, bool) {
	matchedPayload, existentHoneypotPath :=
		middleware.honeypotPayloads[interceptPath]
	return matchedPayload, existentHoneypotPath
}
func (middleware *HoneypotMiddleware) serveHoneypotPayload(
	echoContext echo.Context,
	matchedPayload HoneypotPathMapping,
) error {
	echoContext.Response().Header().Set(
		"Content-Type", matchedPayload.MimeType.String(),
	)
	return echoContext.String(http.StatusOK, matchedPayload.Body)
}
func (middleware *HoneypotMiddleware) serveBanRedirect(
	echoContext echo.Context,
) error {
	return echoContext.Redirect(
		http.StatusFound, middleware.settings.RedirectUrl.String(),
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
		RecordLevel:         middleware.honeypotRecordLevel,
		RecordCode:          middleware.honeypotRecordCode,
		AffectedResources:   []tkValueObject.SystemResourceIdentifier{},
		RecordDetails:       map[string]string{"path": interceptPath},
		OperatorIpAddress:   &requesterIp,
	}
	createErr := middleware.activityRecordCmdRepo.Create(
		honeypotHitCreateRequest,
	)
	if createErr != nil {
		slog.Error("HoneypotHitRecordCreationFailed",
			slog.String("err", createErr.Error()))
	}
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
		banTier, _ := tkUseCase.ReadHoneypotBanDecision(
			middleware.honeypotQueryRepo,
			requesterIp,
			middleware.settings.BanDuration,
			middleware.settings.AggressivenessMode,
		)
		if banTier >= 3 {
			return middleware.serveBanRedirect(echoContext)
		}
		matchedPayload, existentHoneypotPath :=
			middleware.lookupHoneypotPath(httpRequest.URL.Path)
		if !existentHoneypotPath {
			return next(echoContext)
		}
		if banTier >= 2 {
			return middleware.serveBanRedirect(echoContext)
		}
		middleware.writeMu.Lock()
		tkUseCase.CreateHoneypotHit(
			middleware.honeypotCmdRepo,
			requesterIp,
			httpRequest.URL.Path,
			middleware.settings.MaxEntries,
		)
		middleware.writeMu.Unlock()
		middleware.recordHoneypotHit(requesterIp, httpRequest.URL.Path)
		return middleware.serveHoneypotPayload(
			echoContext, matchedPayload,
		)
	}
}
func (middleware *HoneypotMiddleware) MiddlewareFunc() echo.MiddlewareFunc {
	return middleware.Execute
}
func (middleware *HoneypotMiddleware) Stop() {
	if middleware.cancelFunc != nil {
		middleware.cancelFunc()
	}
}
func (middleware *HoneypotMiddleware) runMaintenance() {
	statsRecordCode, codeErr := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)
	if codeErr != nil {
		slog.Error("HoneypotStatsCodeCreationFailed",
			slog.String("err", codeErr.Error()))
	}
	tkUseCase.RunHoneypotMaintenance(
		middleware.honeypotCmdRepo,
		middleware.honeypotQueryRepo,
		middleware.activityRecordCmdRepo,
		tkDto.RunHoneypotMaintenanceRequest{
			AggressivenessMode: middleware.settings.AggressivenessMode,
			BanDuration:        middleware.settings.BanDuration,
			MaxEntries:         middleware.settings.MaxEntries,
			StatsRecordCode:    statsRecordCode,
			StatsRecordLevel:   middleware.honeypotRecordLevel,
		},
	)
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
						slog.Error(
							"HoneypotWatchdogTickPanicRecovered",
							slog.Any("panic", rv),
						)
					}
				}()
				middleware.runMaintenance()
			}()
		}
	}
}
func (middleware *HoneypotMiddleware) Start() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	middleware.cancelFunc = cancelFunc
	go middleware.honeypotMaintenanceWatchdog(ctx)
}
func NewHoneypotMiddleware(
	settings HoneypotMiddlewareSettings,
	honeypotCmdRepo tkRepository.HoneypotCmdRepo,
	honeypotQueryRepo tkRepository.HoneypotQueryRepo,
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
) *HoneypotMiddleware {
	payloadMap := buildHoneypotPayloadMap(settings.ExtraPathRoutes)
	resolvedSettings := honeypotSettingsParser{}.Parse(
		settings, len(payloadMap),
	)
	honeypotRecordCode, codeErr := tkValueObject.NewActivityRecordCode(
		"HoneypotHit",
	)
	if codeErr != nil {
		slog.Error("HoneypotCodeCreationFailed",
			slog.String("err", codeErr.Error()))
	}
	middleware := &HoneypotMiddleware{
		activityRecordCmdRepo: activityRecordCmdRepo,
		honeypotCmdRepo:       honeypotCmdRepo,
		honeypotPayloads:      payloadMap,
		honeypotQueryRepo:     honeypotQueryRepo,
		honeypotRecordCode:    honeypotRecordCode,
		honeypotRecordLevel:   tkValueObject.ActivityRecordLevelSecurity,
		ipExtractor:           NewRequesterIpExtractor(),
		settings:              resolvedSettings,
	}
	middleware.Start()
	return middleware
}
