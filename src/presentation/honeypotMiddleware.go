package tkPresentation

import (
	"context"
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
	RandomSeed         int64
	RedirectUrl        tkValueObject.Url
	StatsInterval      tkValueObject.HoneypotStatsInterval
}
type HoneypotMiddleware struct {
	activePathClasses     map[string]HoneypotPathClass
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo
	cancelFunc            context.CancelFunc
	honeypotCmdRepo       tkRepository.HoneypotCmdRepo
	honeypotPayloads      map[string]HoneypotPathMapping
	honeypotQueryRepo     tkRepository.HoneypotQueryRepo
	honeypotRecordCode    tkValueObject.ActivityRecordCode
	honeypotRecordLevel   tkValueObject.ActivityRecordLevel
	ipExtractor           RequesterIpExtractor
	settings              HoneypotMiddlewareSettings
	writeMu               sync.Mutex
}
func (middleware *HoneypotMiddleware) lookupActivePathClass(
	interceptPath string,
) (HoneypotPathClass, bool) {
	pathClass, pathIsActive :=
		middleware.activePathClasses[interceptPath]
	return pathClass, pathIsActive
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
func (middleware *HoneypotMiddleware) dispatchByPathClass(
	echoContext echo.Context,
	pathClass HoneypotPathClass,
	interceptPath string,
) error {
	switch pathClass {
	case HoneypotPathClassBandwidthExhaust:
		return middleware.streamBandwidthExhaust(echoContext)
	case HoneypotPathClassAITrap:
		return middleware.streamAiTrap(
			echoContext, interceptPath,
		)
	default:
		matchedPayload :=
			middleware.honeypotPayloads[interceptPath]
		return middleware.serveHoneypotPayload(
			echoContext, matchedPayload,
		)
	}
}
func (middleware *HoneypotMiddleware) recordHoneypotHit(
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
		requesterIp, ipExtractionErr :=
			middleware.ipExtractor.Execute(httpRequest)
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
			return middleware.serveMixedResponse(echoContext)
		}
		pathClass, pathIsActive :=
			middleware.lookupActivePathClass(
				httpRequest.URL.Path,
			)
		if !pathIsActive {
			return next(echoContext)
		}
		if banTier >= 2 {
			return middleware.serveMixedResponse(echoContext)
		}
		middleware.writeMu.Lock()
		tkUseCase.CreateHoneypotHit(
			middleware.honeypotCmdRepo,
			requesterIp,
			httpRequest.URL.Path,
			middleware.settings.MaxEntries,
		)
		middleware.writeMu.Unlock()
		middleware.recordHoneypotHit(
			requesterIp, httpRequest.URL.Path,
		)
		return middleware.dispatchByPathClass(
			echoContext, pathClass, httpRequest.URL.Path,
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
	totalPoolSize := len(payloadMap) +
		len(bandwidthExhaustCandidatePaths) +
		len(aiTrapCandidatePaths)
	resolvedSettings := honeypotSettingsParser{}.Parse(
		settings, totalPoolSize,
	)
	staticPathKeys := extractStaticPathKeys(honeypotPayloadEntries)
	for _, extraRoute := range settings.ExtraPathRoutes {
		staticPathKeys = append(
			staticPathKeys, extraRoute.UrlPath.String(),
		)
	}
	activePathMap := selectActivePaths(
		staticPathKeys,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		resolvedSettings.ActivePathCount.Int(),
		resolvedSettings.RandomSeed,
	)
	honeypotRecordCode, codeErr :=
		tkValueObject.NewActivityRecordCode("HoneypotHit")
	if codeErr != nil {
		slog.Error("HoneypotCodeCreationFailed",
			slog.String("err", codeErr.Error()))
	}
	middleware := &HoneypotMiddleware{
		activePathClasses:     activePathMap,
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
