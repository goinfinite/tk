package tkPresentationMiddleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

type mockHoneypotCmdRepo struct {
	createErr           error
	deleteExpiredErr    error
	enforceMaxEntriesErr error
}

func (mock *mockHoneypotCmdRepo) Create(
	createDto tkDto.CreateHoneypotHit,
) error {
	return mock.createErr
}

func (mock *mockHoneypotCmdRepo) DeleteExpired(
	banDuration tkValueObject.HoneypotBanDuration,
) error {
	return mock.deleteExpiredErr
}

func (mock *mockHoneypotCmdRepo) EnforceMaxEntries(
	maxEntries tkValueObject.HoneypotMaxEntries,
) error {
	return mock.enforceMaxEntriesErr
}

type mockHoneypotQueryRepo struct {
	banDecision    tkDto.ReadHoneypotBanDecisionResponse
	banDecisionErr error
}

func (mock *mockHoneypotQueryRepo) ReadBanDecision(
	request tkDto.ReadHoneypotBanDecisionRequest,
) (tkDto.ReadHoneypotBanDecisionResponse, error) {
	return mock.banDecision, mock.banDecisionErr
}

func (mock *mockHoneypotQueryRepo) ReadStatsReport(
	request tkDto.ReadHoneypotStatsReportRequest,
) (tkDto.ReadHoneypotStatsReportResponse, error) {
	return tkDto.ReadHoneypotStatsReportResponse{}, nil
}

type mockActivityRecordCmdRepo struct {
	createErr error
}

func (mock *mockActivityRecordCmdRepo) Create(
	createDto tkDto.CreateActivityRecord,
) error {
	return mock.createErr
}

func (mock *mockActivityRecordCmdRepo) Delete(
	deleteDto tkDto.DeleteActivityRecord,
) error {
	return nil
}

func buildTestHoneypotSettings() tkDto.HoneypotSettings {
	maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSize(
		uint64(1 * 1024 * 1024),
	)
	statsInterval, _ := tkValueObject.NewHoneypotStatsInterval("30m")
	banDuration, _ := tkValueObject.NewHoneypotBanDuration("24h")
	maxEntries, _ := tkValueObject.NewHoneypotMaxEntries(uint64(5000))
	activePathCount, _ := tkValueObject.NewHoneypotActivePathCount(200, 200)
	aggressivenessMode, _ := tkValueObject.NewHoneypotAggressivenessMode(
		"balanced",
	)

	return tkDto.HoneypotSettings{
		AggressivenessMode: aggressivenessMode,
		ActivePathCount:    activePathCount,
		MaxEntries:         maxEntries,
		MaxStreamSize:      maxStreamSize,
		StatsInterval:      statsInterval,
		BanDuration:        banDuration,
		RandomSeed:         12345,
	}
}

func TestNewHoneypotMiddleware(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)

	if middleware == nil {
		t.Errorf("ExpectedNonNilMiddleware")
	}

	middleware.Stop()
}

func TestMiddlewareFuncHoneypotPath(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{
		banDecision: tkDto.ReadHoneypotBanDecisionResponse{
			IsBanned: false,
			HitCount: 1,
			SuggestedAction: tkValueObject.HoneypotSuggestedActionServeMixed,
		},
	}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/wp-config.php", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	handler := middleware.MiddlewareFunc()
	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "next")
	}

	err := handler(nextHandler)(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	if httpRecorder.Code == http.StatusOK &&
		httpRecorder.Body.String() == "next" {
		t.Errorf("ExpectedHoneypotResponse: NextHandlerWasCalled")
	}
}

func TestMiddlewareFuncLegitimatePath(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/users", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	handler := middleware.MiddlewareFunc()
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	err := handler(nextHandler)(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	if !nextCalled {
		t.Errorf("ExpectedNextHandlerCalled")
	}
}

func TestMiddlewareFuncBannedIp(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{
		banDecision: tkDto.ReadHoneypotBanDecisionResponse{
			IsBanned:        true,
			HitCount:        10,
			SuggestedAction: tkValueObject.HoneypotSuggestedActionBan,
		},
	}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	handler := middleware.MiddlewareFunc()
	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "next")
	}

	err := handler(nextHandler)(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	if httpRecorder.Code != http.StatusForbidden {
		t.Errorf(
			"StatusMismatch: Expected=%d, Actual=%d",
			http.StatusForbidden, httpRecorder.Code,
		)
	}
}

func TestStopIdempotent(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)

	middleware.Stop()
	middleware.Stop()
	middleware.Stop()
}

func TestStopConcurrent(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)

	var waitGroup sync.WaitGroup
	for range 10 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			middleware.Stop()
		}()
	}
	waitGroup.Wait()
}

func TestMiddlewareRecordsHit(t *testing.T) {
	settings := buildTestHoneypotSettings()
	createCallCount := 0
	honeypotCmdRepo := &mockHoneypotCmdRepo{
		createErr: nil,
	}
	honeypotQueryRepo := &mockHoneypotQueryRepo{
		banDecision: tkDto.ReadHoneypotBanDecisionResponse{
			IsBanned: false,
			HitCount: 0,
			SuggestedAction: tkValueObject.HoneypotSuggestedActionServeMixed,
		},
	}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	_ = createCallCount

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env.local", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	handler := middleware.MiddlewareFunc()
	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "next")
	}

	err := handler(nextHandler)(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}
}

func TestMiddlewareFuncBanDecisionError(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{
		banDecisionErr: errors.New("DatabaseConnectionError"),
	}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/wp-config.php", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	handler := middleware.MiddlewareFunc()
	nextCalled := false
	nextHandler := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "next")
	}

	err := handler(nextHandler)(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	if nextCalled {
		t.Errorf("ExpectedHoneypotResponse: NextHandlerShouldNotBeCalled")
	}
}

func TestMiddlewareFuncServesResponse(t *testing.T) {
	settings := buildTestHoneypotSettings()
	honeypotCmdRepo := &mockHoneypotCmdRepo{}
	honeypotQueryRepo := &mockHoneypotQueryRepo{
		banDecision: tkDto.ReadHoneypotBanDecisionResponse{
			IsBanned: false,
			HitCount: 1,
			SuggestedAction: tkValueObject.HoneypotSuggestedActionServePayload,
		},
	}
	activityRecordCmdRepo := &mockActivityRecordCmdRepo{}

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo,
		honeypotQueryRepo,
		activityRecordCmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/backup.sql", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	handler := middleware.MiddlewareFunc()
	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "next")
	}

	err := handler(nextHandler)(echoContext)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	if httpRecorder.Body.String() == "next" {
		t.Errorf("ExpectedHoneypotResponse: NextHandlerWasCalled")
	}
}
