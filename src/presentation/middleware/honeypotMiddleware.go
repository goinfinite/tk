package tkPresentationMiddleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkUseCase "github.com/goinfinite/tk/src/domain/useCase"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkPresentation "github.com/goinfinite/tk/src/presentation"
	tkPresentationMiddlewareHoneypot "github.com/goinfinite/tk/src/presentation/middleware/honeypot"
	"github.com/labstack/echo/v4"
)

type HoneypotMiddleware struct {
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo
	honeypotCmdRepo       tkRepository.HoneypotCmdRepo
	honeypotQueryRepo     tkRepository.HoneypotQueryRepo
	pathMapping           *tkPresentationMiddlewareHoneypot.HoneypotPathMapping
	streamHandler         *tkPresentationMiddlewareHoneypot.StreamHandler
	mixedResponseHandler  *tkPresentationMiddlewareHoneypot.MixedResponseHandler
	aiTrapGenerator       *tkPresentationMiddlewareHoneypot.AiTrapGenerator
	settings              tkDto.HoneypotSettings
	stopOnce              sync.Once
	stopChannel           chan struct{}
}

func NewHoneypotMiddleware(
	settings tkDto.HoneypotSettings,
	honeypotCmdRepo tkRepository.HoneypotCmdRepo,
	honeypotQueryRepo tkRepository.HoneypotQueryRepo,
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
) *HoneypotMiddleware {
	pathPool := tkPresentationMiddlewareHoneypot.NewHoneypotPathPool()
	pathSelector := tkPresentationMiddlewareHoneypot.NewHoneypotPathSelector(
		settings, pathPool,
	)
	activePaths := pathSelector.Select()
	pathMapping := tkPresentationMiddlewareHoneypot.NewHoneypotPathMapping(
		activePaths,
	)
	streamHandler := tkPresentationMiddlewareHoneypot.NewStreamHandler(
		settings.MaxStreamSize,
	)
	mixedResponseHandler := tkPresentationMiddlewareHoneypot.NewMixedResponseHandler()
	aiTrapGenerator := tkPresentationMiddlewareHoneypot.NewAiTrapGenerator()

	stopChannel := make(chan struct{})
	middleware := &HoneypotMiddleware{
		activityRecordCmdRepo: activityRecordCmdRepo,
		honeypotCmdRepo:       honeypotCmdRepo,
		honeypotQueryRepo:     honeypotQueryRepo,
		pathMapping:           pathMapping,
		streamHandler:         streamHandler,
		mixedResponseHandler:  mixedResponseHandler,
		aiTrapGenerator:       aiTrapGenerator,
		settings:              settings,
		stopChannel:           stopChannel,
	}

	go middleware.runWatchdog(stopChannel)

	return middleware
}

func (mw *HoneypotMiddleware) Stop() {
	mw.stopOnce.Do(func() {
		close(mw.stopChannel)
	})
}

func (mw *HoneypotMiddleware) runWatchdog(stopChannel chan struct{}) {
	ticker := time.NewTicker(mw.settings.StatsInterval.Duration())
	defer ticker.Stop()

	for {
		select {
		case <-stopChannel:
			return
		case <-ticker.C:
			maintenanceRequest := tkDto.RunHoneypotMaintenanceRequest{
				MaxEntries:  mw.settings.MaxEntries,
				BanDuration: mw.settings.BanDuration,
			}
			maintenanceErr := tkUseCase.RunHoneypotMaintenance(
				mw.honeypotCmdRepo, maintenanceRequest,
			)
			if maintenanceErr != nil {
				slog.Error(
					"HoneypotWatchdogMaintenanceError",
					slog.String("err", maintenanceErr.Error()),
				)
			}
		}
	}
}

func (mw *HoneypotMiddleware) MiddlewareFunc() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoContext echo.Context) error {
			requestPath := echoContext.Request().URL.Path
			urlPath, pathErr := tkValueObject.NewUrlPath(requestPath)
			if pathErr != nil {
				return next(echoContext)
			}

			pathClass, isHoneypot := mw.pathMapping.Resolve(urlPath)
			if !isHoneypot {
				return next(echoContext)
			}

			operatorIpAddress := mw.extractOperatorIp(echoContext)
			banDecision := mw.readBanDecision(operatorIpAddress)

			mw.recordHit(echoContext, urlPath, pathClass, banDecision)

			if banDecision.IsBanned {
				return echoContext.NoContent(http.StatusForbidden)
			}

			return mw.serveHoneypotResponse(
				echoContext, pathClass, banDecision,
			)
		}
	}
}

func (mw *HoneypotMiddleware) extractOperatorIp(
	echoContext echo.Context,
) tkValueObject.IpAddress {
	extractor := tkPresentation.NewRequesterIpExtractor()
	operatorIpAddress, extractionErr := extractor.Execute(
		echoContext.Request(),
	)
	if extractionErr != nil {
		return tkValueObject.IpAddressLocal
	}
	return operatorIpAddress
}

func (mw *HoneypotMiddleware) readBanDecision(
	operatorIpAddress tkValueObject.IpAddress,
) tkDto.ReadHoneypotBanDecisionResponse {
	banDecisionRequest := tkDto.ReadHoneypotBanDecisionRequest{
		RequesterIpAddress: operatorIpAddress,
	}
	banDecision, banErr := tkUseCase.ReadHoneypotBanDecision(
		mw.honeypotQueryRepo, banDecisionRequest, mw.settings,
	)
	if banErr != nil {
		return tkDto.ReadHoneypotBanDecisionResponse{
			IsBanned:        false,
			HitCount:        0,
			SuggestedAction: tkValueObject.HoneypotSuggestedActionServeMixed,
		}
	}
	return banDecision
}

func (mw *HoneypotMiddleware) recordHit(
	echoContext echo.Context,
	urlPath tkValueObject.UrlPath,
	pathClass tkValueObject.HoneypotPathClass,
	banDecision tkDto.ReadHoneypotBanDecisionResponse,
) {
	operatorIpAddress := mw.extractOperatorIp(echoContext)
	createHitDto := tkDto.CreateHoneypotHit{
		RequesterIpAddress: operatorIpAddress,
		HoneypotPath:       urlPath,
		HitClass:           pathClass,
		HitCount:           banDecision.HitCount + 1,
	}
	createHitErr := mw.honeypotCmdRepo.Create(createHitDto)
	if createHitErr != nil {
		slog.Error(
			"HoneypotHitRecordError",
			slog.String("err", createHitErr.Error()),
		)
	}

	activityRecordCode, codeErr := tkValueObject.NewActivityRecordCode(
		"HoneypotHitDetected",
	)
	if codeErr != nil {
		return
	}
	activityRecordLevel := tkValueObject.ActivityRecordLevelSecurity
	createRecordDto := tkDto.CreateActivityRecord{
		RecordLevel:   activityRecordLevel,
		RecordCode:    activityRecordCode,
		OperatorIpAddress: &operatorIpAddress,
		RecordDetails: map[string]any{
			"path":       urlPath.String(),
			"pathClass":  pathClass.String(),
			"hitCount":   banDecision.HitCount + 1,
			"isBanned":   banDecision.IsBanned,
			"requestUri": echoContext.Request().RequestURI,
		},
	}
	tkUseCase.CreateActivityRecord(
		mw.activityRecordCmdRepo, createRecordDto,
	)
}

func (mw *HoneypotMiddleware) serveHoneypotResponse(
	echoContext echo.Context,
	pathClass tkValueObject.HoneypotPathClass,
	banDecision tkDto.ReadHoneypotBanDecisionResponse,
) error {
	suggestedAction := banDecision.SuggestedAction

	if suggestedAction == tkValueObject.HoneypotSuggestedActionBan {
		return echoContext.NoContent(http.StatusForbidden)
	}

	if suggestedAction == tkValueObject.HoneypotSuggestedActionServePayload {
		return mw.mixedResponseHandler.Serve(echoContext, pathClass)
	}

	if suggestedAction == tkValueObject.HoneypotSuggestedActionServeStream {
		return mw.streamHandler.ServeBandwidthExhaust(echoContext)
	}

	if suggestedAction == tkValueObject.HoneypotSuggestedActionServeAiTrap {
		return mw.streamHandler.ServeAiTrap(
			echoContext, mw.aiTrapGenerator,
		)
	}

	if suggestedAction == tkValueObject.HoneypotSuggestedActionServeMixed {
		return mw.mixedResponseHandler.Serve(echoContext, pathClass)
	}

	return mw.mixedResponseHandler.Serve(echoContext, pathClass)
}
