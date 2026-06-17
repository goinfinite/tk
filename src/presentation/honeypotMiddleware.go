package tkPresentation

import (
	"embed"
	"errors"
	"log/slog"
	"net/http"
	"os"
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
	ContentType string
	Body        string
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
				"HoneypotPayloadLoadFailed: " + fileMetadata.embeddedFilename,
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
	activityRecordCmdRepo   tkRepository.ActivityRecordCmdRepo
	activityRecordQueryRepo tkRepository.ActivityRecordQueryRepo
	honeypotHttpPayloads    honeypotPayloadLoader
	honeypotRecordCode      tkValueObject.ActivityRecordCode
	honeypotRecordLevel     tkValueObject.ActivityRecordLevel
	ipExtractor             RequesterIpExtractor
	settings                HoneypotMiddlewareSettings
}

func (middleware HoneypotMiddleware) isIpBanned(
	requesterIp tkValueObject.IpAddress,
) bool {
	if middleware.activityRecordQueryRepo == nil {
		return false
	}

	banWindowStart := time.Now().Add(-middleware.settings.BanDuration)
	banWindowStartVo, _ := tkValueObject.NewUnixTime(banWindowStart)

	banCheckQueryRequest := tkDto.ReadActivityRecordsRequest{
		RecordCode:        &middleware.honeypotRecordCode,
		OperatorIpAddress: &requesterIp,
		CreatedAfterAt:    &banWindowStartVo,
	}

	honeypotBanRecord, banCheckQueryErr := middleware.activityRecordQueryRepo.ReadFirst(
		banCheckQueryRequest,
	)
	if banCheckQueryErr != nil {
		slog.Debug("BanCheckQueryFailed",
			slog.String("err", banCheckQueryErr.Error()))
		return false
	}

	return honeypotBanRecord.RecordId != 0
}

func (middleware HoneypotMiddleware) serveBanRedirect(
	echoContext echo.Context,
) error {
	return echoContext.Redirect(
		http.StatusFound, middleware.settings.RedirectUrl.String(),
	)
}

func (middleware HoneypotMiddleware) lookupHoneypotPath(
	interceptPath string,
) (HoneypotHttpPayload, bool) {
	matchedPayload, isDefaultHoneypotPath := middleware.honeypotHttpPayloads.ReadPayload(
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

func (middleware HoneypotMiddleware) recordHoneypotHit(
	requesterIp tkValueObject.IpAddress,
	interceptPath string,
) {
	if middleware.activityRecordCmdRepo == nil {
		return
	}

	honeypotHitCreateRequest := tkDto.CreateActivityRecord{
		RecordLevel:       middleware.honeypotRecordLevel,
		RecordCode:        middleware.honeypotRecordCode,
		AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		RecordDetails:     map[string]string{"path": interceptPath},
		OperatorIpAddress: &requesterIp,
	}

	honeypotHitCreateErr := middleware.activityRecordCmdRepo.Create(
		honeypotHitCreateRequest,
	)
	if honeypotHitCreateErr != nil {
		slog.Debug("HoneypotHitRecordCreationFailed",
			slog.String("err", honeypotHitCreateErr.Error()))
	}
}

func (middleware HoneypotMiddleware) serveHoneypotPayload(
	echoContext echo.Context,
	matchedPayload HoneypotHttpPayload,
) error {
	echoContext.Response().Header().Set(
		"Content-Type", matchedPayload.ContentType,
	)
	return echoContext.String(http.StatusOK, matchedPayload.Body)
}

func (middleware HoneypotMiddleware) Execute(
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

		if middleware.isIpBanned(requesterIp) {
			return middleware.serveBanRedirect(echoContext)
		}

		matchedPayload, isHoneypotPath := middleware.lookupHoneypotPath(
			httpRequest.URL.Path,
		)
		if !isHoneypotPath {
			return next(echoContext)
		}

		middleware.recordHoneypotHit(
			requesterIp, httpRequest.URL.Path,
		)
		return middleware.serveHoneypotPayload(
			echoContext, matchedPayload,
		)
	}
}

func NewHoneypotMiddleware(
	settings HoneypotMiddlewareSettings,
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	activityRecordQueryRepo tkRepository.ActivityRecordQueryRepo,
) echo.MiddlewareFunc {
	if settings.BanDuration <= 0 {
		settings.BanDuration = 24 * time.Hour
	}

	if settings.RedirectUrl.String() == "" {
		defaultRedirectUrl, _ := tkValueObject.NewUrl("https://xkcd.com/")
		settings.RedirectUrl = defaultRedirectUrl
	}

	maxEntries, maxEntriesErr := tkValueObject.NewHoneypotMaxEntries(
		os.Getenv("HONEYPOT_MAX_ENTRIES"),
	)
	if maxEntriesErr != nil {
		slog.Debug("HoneypotMaxEntriesInvalid",
			slog.String("err", maxEntriesErr.Error()))
	}
	settings.MaxEntries = maxEntries

	payloadLoader := &honeypotPayloadLoader{
		payloadMap: make(map[string]HoneypotHttpPayload),
	}
	loaderErr := payloadLoader.loadEmbeddedPayloads()
	if loaderErr != nil {
		slog.Debug("HoneypotPayloadLoaderCreationFailed",
			slog.String("err", loaderErr.Error()))
	}

	poolCeiling := payloadLoader.totalCandidatePoolSize()
	activePathCount, activePathCountErr := tkValueObject.NewHoneypotActivePathCount(
		os.Getenv("HONEYPOT_ACTIVE_PATHS"), poolCeiling,
	)
	if activePathCountErr != nil {
		slog.Debug("HoneypotActivePathCountInvalid",
			slog.String("err", activePathCountErr.Error()))
	}
	settings.ActivePathCount = activePathCount

	maxStreamSize, maxStreamSizeErr := tkValueObject.NewHoneypotMaxStreamSizeBytes(
		os.Getenv("HONEYPOT_MAX_STREAM_SIZE"),
	)
	if maxStreamSizeErr != nil {
		slog.Debug("HoneypotMaxStreamSizeInvalid",
			slog.String("err", maxStreamSizeErr.Error()))
	}
	settings.MaxStreamSizeBytes = maxStreamSize

	statsInterval, statsIntervalErr := tkValueObject.NewHoneypotStatsInterval(
		os.Getenv("HONEYPOT_STATS_INTERVAL"),
	)
	if statsIntervalErr != nil {
		slog.Debug("HoneypotStatsIntervalInvalid",
			slog.String("err", statsIntervalErr.Error()))
	}
	settings.StatsInterval = statsInterval

	rawAggressivenessMode := os.Getenv("HONEYPOT_AGGRESSIVENESS")
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeBalanced
	switch rawAggressivenessMode {
	case "standard", "lenient", "passive":
		slog.Debug("AggressivenessModeDeprecatedFallback",
			slog.String("deprecated", rawAggressivenessMode),
			slog.String("resolved", "balanced"))
	case "":
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

	honeypotRecordCode, codeErr := tkValueObject.NewActivityRecordCode(
		"HoneypotHit",
	)
	if codeErr != nil {
		slog.Debug("HoneypotCodeCreationFailed",
			slog.String("err", codeErr.Error()))
	}

	middleware := &HoneypotMiddleware{
		activityRecordCmdRepo:   activityRecordCmdRepo,
		activityRecordQueryRepo: activityRecordQueryRepo,
		honeypotHttpPayloads:    *payloadLoader,
		honeypotRecordCode:      honeypotRecordCode,
		honeypotRecordLevel:     tkValueObject.ActivityRecordLevelSecurity,
		ipExtractor:             NewRequesterIpExtractor(),
		settings:                settings,
	}

	return middleware.Execute
}
