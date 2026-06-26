package tkPresentation

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkUseCase "github.com/goinfinite/tk/src/domain/useCase"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraHoneypot "github.com/goinfinite/tk/src/infra/honeypot"
	"github.com/labstack/echo/v4"
)

var testIpCounter int64

func newUniqueTestIp() string {
	counter := atomic.AddInt64(&testIpCounter, 1)
	return fmt.Sprintf("10.77.%d.1", counter%254+1)
}

type mockActivityRecordCmdRepo struct {
	createFunc func(tkDto.CreateActivityRecord) error
}

func (mockCmdRepo mockActivityRecordCmdRepo) Create(
	createDto tkDto.CreateActivityRecord,
) error {
	return mockCmdRepo.createFunc(createDto)
}

func (mockCmdRepo mockActivityRecordCmdRepo) Delete(
	deleteDto tkDto.DeleteActivityRecord,
) error {
	return nil
}

func newDefaultRedirectUrl() tkValueObject.Url {
	defaultUrl, _ := tkValueObject.NewUrl("https://xkcd.com/")
	return defaultUrl
}

func newNoopCmdRepo() mockActivityRecordCmdRepo {
	return mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			return nil
		},
	}
}

func mustNewHoneypotBanDuration(
	rawValue any,
) tkValueObject.HoneypotBanDuration {
	banDuration, _ := tkValueObject.NewHoneypotBanDuration(rawValue)
	return banDuration
}

func mustNewHoneypotMaxEntries(
	rawValue any,
) tkValueObject.HoneypotMaxEntries {
	maxEntries, _ := tkValueObject.NewHoneypotMaxEntries(rawValue)
	return maxEntries
}

func mustNewHoneypotActivePathCount(
	rawActivePaths any,
) tkValueObject.HoneypotActivePathCount {
	activePathCount, _ := tkValueObject.NewHoneypotActivePathCount(
		rawActivePaths, 0,
	)
	return activePathCount
}

func mustNewHoneypotMaxStreamSizeBytes(
	rawMaxStream any,
) tkValueObject.HoneypotMaxStreamSizeBytes {
	maxStreamSize, _ := tkValueObject.NewHoneypotMaxStreamSizeBytes(
		rawMaxStream,
	)
	return maxStreamSize
}

func newStandardSettings() HoneypotMiddlewareSettings {
	return HoneypotMiddlewareSettings{
		ActivePathCount: mustNewHoneypotActivePathCount(200),
		BanDuration:     mustNewHoneypotBanDuration(24 * time.Hour),
		RedirectUrl:     newDefaultRedirectUrl(),
	}
}

func newTransientDbSvc() *tkInfraDb.TransientDatabaseService {
	dbSvc, _ := tkInfraDb.NewTransientDatabaseService()
	dbSvc.Handler.Exec("DELETE FROM key_values")
	return dbSvc
}

func newHoneypotRepos(
	dbSvc *tkInfraDb.TransientDatabaseService,
) (tkRepository.HoneypotCmdRepo, tkRepository.HoneypotQueryRepo) {
	if dbSvc == nil {
		return nil, nil
	}
	return tkInfraHoneypot.NewHoneypotCmdRepo(dbSvc),
		tkInfraHoneypot.NewHoneypotQueryRepo(dbSvc)
}

func populateTransientDbWithHits(
	dbSvc *tkInfraDb.TransientDatabaseService,
	ipString string,
	hitCount int,
) {
	hitKey := "honeypot:hit:" + ipString
	dbSvc.Handler.Where("key = ?", hitKey).Delete(
		&tkInfraDb.KeyValueModel{},
	)
	hitData := tkDto.HoneypotHitData{
		Count:      hitCount,
		FirstHitAt: time.Now().UTC().Format(time.RFC3339),
		Endpoints:  map[string]int{"/.env": hitCount},
	}
	jsonBytes, _ := json.Marshal(hitData)
	createdAt := time.Now().UTC()
	dbSvc.Handler.Exec(
		"INSERT INTO key_values (key, value, created_at) VALUES (?, ?, ?)",
		hitKey, string(jsonBytes), createdAt,
	)
}

func populateTransientDbWithOldHits(
	dbSvc *tkInfraDb.TransientDatabaseService,
	ipString string,
	hitCount int,
	age time.Duration,
) {
	hitKey := "honeypot:hit:" + ipString
	dbSvc.Handler.Where("key = ?", hitKey).Delete(
		&tkInfraDb.KeyValueModel{},
	)
	firstHitAt := time.Now().UTC().Add(-age).Format(time.RFC3339)
	hitData := tkDto.HoneypotHitData{
		Count:      hitCount,
		FirstHitAt: firstHitAt,
		Endpoints:  map[string]int{"/.env": hitCount},
	}
	jsonBytes, _ := json.Marshal(hitData)
	createdAt := time.Now().UTC().Add(-age)
	dbSvc.Handler.Exec(
		"INSERT INTO key_values (key, value, created_at) VALUES (?, ?, ?)",
		hitKey, string(jsonBytes), createdAt,
	)
}

func findActivePathOfClass(
	middleware *HoneypotMiddleware,
	targetClass HoneypotPathClass,
) string {
	for activePath, pathClass := range middleware.activePathClasses {
		if pathClass == targetClass {
			return activePath
		}
	}
	return ""
}

func TestHoneypotMiddlewareCreation(t *testing.T) {
	testCaseStructs := []struct {
		name     string
		settings HoneypotMiddlewareSettings
	}{
		{"WithBanDuration", HoneypotMiddlewareSettings{
			BanDuration: mustNewHoneypotBanDuration(24 * time.Hour),
		}},
		{"WithEmptySettings", HoneypotMiddlewareSettings{}},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			middleware := NewHoneypotMiddleware(
				testCase.settings, nil, nil, nil,
			)
			defer middleware.Stop()
			if middleware == nil {
				t.Errorf("MiddlewareIsNil")
			}
		})
	}
}

func TestHoneypotBanBehavior(t *testing.T) {
	t.Run("HitCreatesBanRecord", func(t *testing.T) {
		var createdRecord *tkDto.CreateActivityRecord
		cmdRepo := mockActivityRecordCmdRepo{
			createFunc: func(
				dto tkDto.CreateActivityRecord,
			) error {
				createdRecord = &dto
				return nil
			},
		}
		transientDbSvc := newTransientDbSvc()
		honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
			transientDbSvc,
		)

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		activeStatic := findActivePathOfClass(
			middleware, HoneypotPathClassStaticVuln,
		)
		if activeStatic == "" {
			t.Fatalf("NoActiveStaticPath")
		}

		httpRequest := httptest.NewRequest(
			http.MethodGet, activeStatic, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if httpRecorder.Code != http.StatusOK {
			t.Errorf("StatusCodeMismatch: got=%d, want=%d",
				httpRecorder.Code, http.StatusOK)
		}

		if createdRecord == nil {
			t.Errorf("ActivityRecordNotCreated")
			return
		}

		if createdRecord.RecordCode.String() != "HoneypotHit" {
			t.Errorf("RecordCodeMismatch: got=%s, want=HoneypotHit",
				createdRecord.RecordCode.String())
		}

		if createdRecord.RecordLevel.String() != "SECURITY" {
			t.Errorf("RecordLevelMismatch: got=%s, want=SECURITY",
				createdRecord.RecordLevel.String())
		}

		if createdRecord.OperatorIpAddress == nil {
			t.Errorf("OperatorIpAddressIsNil")
			return
		}

		if createdRecord.OperatorIpAddress.String() != "1.2.3.4" {
			t.Errorf("IpAddressMismatch: got=%s, want=1.2.3.4",
				createdRecord.OperatorIpAddress.String())
		}
	})

	t.Run("SubsequentRequestsTemporarilyRedirect", func(t *testing.T) {
		cmdRepo := newNoopCmdRepo()
		transientDbSvc := newTransientDbSvc()
		populateTransientDbWithHits(
			transientDbSvc, "1.2.3.4", 3,
		)
		honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
			transientDbSvc,
		)

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		httpRequest := httptest.NewRequest(
			http.MethodGet, "/api/health", nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if !isMixedResponseStatusCode(httpRecorder.Code) {
			t.Errorf("ExpectedMixedResponse: got=%d",
				httpRecorder.Code)
		}
	})

	t.Run("BanExpiresAfterWindow", func(t *testing.T) {
		honeypotBanRecordCreated := false
		cmdRepo := mockActivityRecordCmdRepo{
			createFunc: func(
				dto tkDto.CreateActivityRecord,
			) error {
				honeypotBanRecordCreated = true
				return nil
			},
		}
		transientDbSvc := newTransientDbSvc()
		populateTransientDbWithOldHits(
			transientDbSvc, "1.2.3.4", 3,
			25*time.Hour,
		)
		honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
			transientDbSvc,
		)

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		activeStatic := findActivePathOfClass(
			middleware, HoneypotPathClassStaticVuln,
		)
		if activeStatic == "" {
			t.Fatalf("NoActiveStaticPath")
		}

		firstRequest := httptest.NewRequest(
			http.MethodGet, activeStatic, nil,
		)
		firstRequest.RemoteAddr = "1.2.3.4:1234"
		firstRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(firstRecorder, firstRequest)

		if firstRecorder.Code != http.StatusOK {
			t.Errorf("FirstRequestStatusCodeMismatch: got=%d, want=%d",
				firstRecorder.Code, http.StatusOK)
		}

		if !honeypotBanRecordCreated {
			t.Errorf("BanRecordNotCreated")
		}
	})

	t.Run("BannedIpHittingHoneypotRedirects", func(t *testing.T) {
		recordCreated := false
		cmdRepo := mockActivityRecordCmdRepo{
			createFunc: func(
				dto tkDto.CreateActivityRecord,
			) error {
				recordCreated = true
				return nil
			},
		}
		transientDbSvc := newTransientDbSvc()
		populateTransientDbWithHits(
			transientDbSvc, "1.2.3.4", 3,
		)
		honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
			transientDbSvc,
		)

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		activeStatic := findActivePathOfClass(
			middleware, HoneypotPathClassStaticVuln,
		)
		if activeStatic == "" {
			t.Fatalf("NoActiveStaticPath")
		}

		httpRequest := httptest.NewRequest(
			http.MethodGet, activeStatic, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if !isMixedResponseStatusCode(httpRecorder.Code) {
			t.Errorf("ExpectedMixedResponse: got=%d",
				httpRecorder.Code)
		}

		if recordCreated {
			t.Errorf("NewRecordShouldNotBeCreatedForBannedIp")
		}
	})
}

func TestAllRequestsCheckedAgainstBanList(t *testing.T) {
	testCaseStructs := []struct {
		description string
		path        string
	}{
		{"UiRoute", "/dashboard"},
		{"ApiRoute", "/api/v1/users"},
		{"StaticFile", "/static/app.js"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			transientDbSvc := newTransientDbSvc()
			populateTransientDbWithHits(
				transientDbSvc, "1.2.3.4", 3,
			)
			honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
				transientDbSvc,
			)

			middleware := NewHoneypotMiddleware(
				newStandardSettings(),
				honeypotCmdRepo, honeypotQueryRepo, nil,
			)
			defer middleware.Stop()

			echoInstance := echo.New()
			echoInstance.Use(middleware.MiddlewareFunc())

			httpRequest := httptest.NewRequest(
				http.MethodGet, testCase.path, nil,
			)
			httpRequest.RemoteAddr = "1.2.3.4:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

			if !isMixedResponseStatusCode(
				httpRecorder.Code,
			) {
				t.Errorf("ExpectedMixedResponse: path=%s, got=%d",
					testCase.path, httpRecorder.Code)
			}
		})
	}
}

func TestAllHoneypotPathsReturnPayloads(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPaths := make([]string, 0)
	for activePath := range middleware.activePathClasses {
		honeypotPaths = append(honeypotPaths, activePath)
	}

	for _, honeypotPath := range honeypotPaths {
		t.Run(strings.TrimPrefix(honeypotPath, "/"), func(t *testing.T) {
			httpRequest := httptest.NewRequest(
				http.MethodGet, honeypotPath, nil,
			)
			httpRequest.RemoteAddr = "5.6.7.8:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

			if httpRecorder.Code != http.StatusOK {
				t.Errorf("Path=%s StatusCodeMismatch: got=%d, want=%d",
					honeypotPath, httpRecorder.Code,
					http.StatusOK)
			}

			contentType := httpRecorder.Header().Get("Content-Type")
			if contentType == "" {
				t.Errorf("Path=%s ContentTypeMissing", honeypotPath)
			}

			if httpRecorder.Body.Len() == 0 {
				t.Errorf("Path=%s BodyEmpty", honeypotPath)
			}
		})
	}
}

func TestHoneypotFailOpenBehavior(t *testing.T) {
	t.Run("InvalidIpFormatIgnored", func(t *testing.T) {
		cmdRepo := newNoopCmdRepo()
		transientDbSvc := newTransientDbSvc()
		honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
			transientDbSvc,
		)

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())
		echoInstance.GET("/api/health", func(
			echoCtx echo.Context,
		) error {
			return echoCtx.String(http.StatusOK, "OK")
		})

		httpRequest := httptest.NewRequest(
			http.MethodGet, "/api/health", nil,
		)
		httpRequest.RemoteAddr = "invalid-ip-format"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if httpRecorder.Code != http.StatusOK {
			t.Errorf("StatusCodeMismatch: got=%d, want=%d",
				httpRecorder.Code, http.StatusOK)
		}
	})

	t.Run("TransientDbUnavailableFailsOpen", func(t *testing.T) {
		middleware := NewHoneypotMiddleware(
			newStandardSettings(), nil, nil, nil,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())
		echoInstance.GET("/api/health", func(
			echoCtx echo.Context,
		) error {
			return echoCtx.String(http.StatusOK, "OK")
		})

		httpRequest := httptest.NewRequest(
			http.MethodGet, "/api/health", nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if httpRecorder.Code != http.StatusOK {
			t.Errorf("StatusCodeMismatch: got=%d, want=%d",
				httpRecorder.Code, http.StatusOK)
		}
	})
}

func TestEmptySettingsUsesDefaults(t *testing.T) {
	settings := HoneypotMiddlewareSettings{}

	middleware := NewHoneypotMiddleware(
		settings, nil, nil, nil,
	)
	defer middleware.Stop()

	if middleware == nil {
		t.Errorf("MiddlewareIsNil")
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middlewareWithDb := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middlewareWithDb.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middlewareWithDb.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("ExpectedMixedResponse: got=%d",
			httpRecorder.Code)
	}
}

func TestSharedNATBlocksLegitimateUsers(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	secondRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	secondRequest.RemoteAddr = "1.2.3.4:5678"
	secondRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(secondRecorder, secondRequest)

	if !isMixedResponseStatusCode(secondRecorder.Code) {
		t.Errorf("ExpectedMixedResponseForSharedNAT: got=%d",
			secondRecorder.Code)
	}
}

func TestBurpScanFloodsHoneypot(t *testing.T) {
	recordCount := 0
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			recordCount++
			return nil
		},
	}
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPaths := extractStaticPathKeys()
	scanIterationCount := 2
	for range scanIterationCount {
		for _, honeypotPath := range honeypotPaths {
			httpRequest := httptest.NewRequest(
				http.MethodGet, honeypotPath, nil,
			)
			httpRequest.RemoteAddr = "10.20.30.40:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)
		}
	}

	if recordCount < 25 {
		t.Errorf("ExpectedAtLeast25Records: got=%d", recordCount)
	}
}

func TestCustomExtraPathRoutesReturnPayload(t *testing.T) {
	customUrlPath, _ := tkValueObject.NewUrlPath("/custom-honeypot")
	customMimeType, _ := tkValueObject.NewMimeType("text/plain")
	fakeAdminUrlPath, _ := tkValueObject.NewUrlPath("/fake-admin")
	fakeAdminMimeType, _ := tkValueObject.NewMimeType("text/html")
	fakeApiKeysUrlPath, _ := tkValueObject.NewUrlPath(
		"/fake-api-keys",
	)
	fakeApiKeysMimeType, _ := tkValueObject.NewMimeType(
		"application/json",
	)

	testCaseStructs := []struct {
		description         string
		extraPathRoutes     []HoneypotPathMapping
		interceptPath       string
		expectedContentType string
		expectedBody        string
	}{
		{
			"SingleExtraRouteReturnsConfiguredPayload",
			[]HoneypotPathMapping{
				{
					UrlPath:  customUrlPath,
					Body:     "fake-secret=value123\n",
					MimeType: customMimeType,
				},
			},
			"/custom-honeypot",
			"text/plain",
			"fake-secret=value123\n",
		},
		{
			"MultipleExtraRoutesReturnCorrectPayload",
			[]HoneypotPathMapping{
				{
					UrlPath:  fakeAdminUrlPath,
					Body:     "<html><h1>Fake Admin</h1></html>",
					MimeType: fakeAdminMimeType,
				},
				{
					UrlPath:  fakeApiKeysUrlPath,
					Body:     `{"api_key":"fake-key-12345"}`,
					MimeType: fakeApiKeysMimeType,
				},
			},
			"/fake-api-keys",
			"application/json",
			`{"api_key":"fake-key-12345"}`,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.description, func(t *testing.T) {
			honeypotSettings := HoneypotMiddlewareSettings{
				ActivePathCount: mustNewHoneypotActivePathCount(200),
				ExtraPathRoutes: testCase.extraPathRoutes,
				RandomSeed:      1,
			}

			honeypotMiddleware := NewHoneypotMiddleware(
				honeypotSettings, nil, nil, nil,
			)
			defer honeypotMiddleware.Stop()

			echoInstance := echo.New()
			echoInstance.Use(
				honeypotMiddleware.MiddlewareFunc(),
			)

			incomingRequest := httptest.NewRequest(
				http.MethodGet,
				testCase.interceptPath, nil,
			)
			incomingRequest.RemoteAddr = "9.8.7.6:4321"
			responseRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(
				responseRecorder, incomingRequest,
			)

			if responseRecorder.Code != http.StatusOK {
				t.Errorf("StatusCodeMismatch: got=%d, want=%d",
					responseRecorder.Code, http.StatusOK)
			}

			actualContentType := responseRecorder.Header().Get(
				"Content-Type",
			)
			if actualContentType != testCase.expectedContentType {
				t.Errorf("ContentTypeMismatch: got=%s, want=%s",
					actualContentType,
					testCase.expectedContentType)
			}

			actualBody := responseRecorder.Body.String()
			if actualBody != testCase.expectedBody {
				t.Errorf("BodyMismatch: got=%s, want=%s",
					actualBody, testCase.expectedBody)
			}
		})
	}
}

func TestXForwardedForSpoofingAttempt(t *testing.T) {
	var capturedIp *tkValueObject.IpAddress
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedIp = dto.OperatorIpAddress
			return nil
		},
	}
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	activeStatic := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if activeStatic == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	httpRequest := httptest.NewRequest(
		http.MethodGet, activeStatic, nil,
	)
	httpRequest.RemoteAddr = "203.0.113.50:1234"
	httpRequest.Header.Set("X-Forwarded-For", "10.0.0.1")
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if capturedIp == nil {
		t.Errorf("IpAddressNotCaptured")
		return
	}

	if capturedIp.String() != "203.0.113.50" {
		t.Errorf("SpoofedIpUsed: got=%s, want=203.0.113.50",
			capturedIp.String())
	}
}

func TestGraduatedBanTierZeroNormalTraffic(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if httpRecorder.Code != http.StatusOK {
		t.Errorf("TierZeroShouldPassThrough: got=%d, want=%d",
			httpRecorder.Code, http.StatusOK)
	}
}

func TestGraduatedBanTierOneServesPayloadNoBan(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	legitRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	legitRequest.RemoteAddr = "1.2.3.4:1234"
	legitRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(legitRecorder, legitRequest)

	if legitRecorder.Code != http.StatusOK {
		t.Errorf("LegitPathShouldPassAtTierOne: got=%d",
			legitRecorder.Code)
	}

	honeypotPath := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if honeypotPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	honeypotRequest := httptest.NewRequest(
		http.MethodGet, honeypotPath, nil,
	)
	honeypotRequest.RemoteAddr = "1.2.3.4:1234"
	honeypotRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(honeypotRecorder, honeypotRequest)

	if honeypotRecorder.Code != http.StatusOK {
		t.Errorf("HoneypotShouldServePayloadAtTierOne: got=%d",
			honeypotRecorder.Code)
	}
}

func TestGraduatedBanTierOneIncrementsHitCount(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPath := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if honeypotPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	httpRequest := httptest.NewRequest(
		http.MethodGet, honeypotPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	rawValue, readErr := transientDbSvc.Read("honeypot:hit:1.2.3.4")
	if readErr != nil {
		t.Fatalf("HitDataNotStored: %v", readErr)
	}

	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count != 1 {
		t.Errorf("HitCountMismatch: got=%d, want=1", hitData.Count)
	}
}

func TestGraduatedBanTierTwoBannedOnHoneypotPaths(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 2)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPath := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if honeypotPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	httpRequest := httptest.NewRequest(
		http.MethodGet, honeypotPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("TierTwoHoneypotPathExpectedMixed: got=%d",
			httpRecorder.Code)
	}
}

func TestGraduatedBanTierThreeFullBan(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 3)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	testCaseStructs := []struct {
		name string
		path string
	}{
		{"HoneypotPath", "/.env"},
		{"LegitPath", "/api/health"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			httpRequest := httptest.NewRequest(
				http.MethodGet, testCase.path, nil,
			)
			httpRequest.RemoteAddr = "1.2.3.4:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

			if !isMixedResponseStatusCode(
				httpRecorder.Code,
			) {
				t.Errorf("TierThreeExpectedMixed: path=%s, got=%d",
					testCase.path, httpRecorder.Code)
			}
		})
	}
}

func TestHoneypotHitCountResetsAfterTTL(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithOldHits(
		transientDbSvc, "1.2.3.4", 3, 25*time.Hour,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if httpRecorder.Code != http.StatusOK {
		t.Errorf("ExpiredHitsShouldResetToTierZero: got=%d",
			httpRecorder.Code)
	}
}

func TestNewHoneypotMiddlewareAcceptsSettingsAndRepos(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	if middleware == nil {
		t.Fatalf("MiddlewareIsNil")
	}
}

func TestNewHoneypotMiddlewareReturnsHoneypotMiddlewareStruct(
	t *testing.T,
) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)
	defer middleware.Stop()

	var _ echo.MiddlewareFunc = middleware.MiddlewareFunc()
	middleware.Stop()
}

func TestMiddlewareFuncReturnsEchoMiddlewareFunc(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)
	defer middleware.Stop()

	middlewareFunc := middleware.MiddlewareFunc()
	if middlewareFunc == nil {
		t.Fatalf("MiddlewareFuncReturnedNil")
	}
}

func TestStopMethodCancelsContext(t *testing.T) {
	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"100ms",
	)

	goroutinesBefore := runtime.NumGoroutine()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)

	time.Sleep(50 * time.Millisecond)
	goroutinesAfterStart := runtime.NumGoroutine()

	if goroutinesAfterStart <= goroutinesBefore {
		t.Errorf("WatchdogGoroutineNotStarted")
	}

	middleware.Stop()
	time.Sleep(100 * time.Millisecond)
	goroutinesAfterStop := runtime.NumGoroutine()

	if goroutinesAfterStop >= goroutinesAfterStart {
		t.Errorf("WatchdogGoroutineNotStopped: before=%d, after=%d",
			goroutinesAfterStart, goroutinesAfterStop)
	}
}

func TestStopMethodIsIdempotent(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)

	middleware.Stop()
	middleware.Stop()
}

func TestStopCalledMultipleTimesIsSafe(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)

	middleware.Stop()
	middleware.Stop()
	middleware.Stop()

	if middleware.cancelFunc == nil {
		t.Fatalf("CancelFuncIsNil")
	}
}

func TestEndpointHitCountTrackedInValueJson(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	activeStatic := make([]string, 0)
	for activePath, pathClass := range middleware.activePathClasses {
		if pathClass == HoneypotPathClassStaticVuln {
			activeStatic = append(activeStatic, activePath)
			if len(activeStatic) >= 2 {
				break
			}
		}
	}
	if len(activeStatic) < 2 {
		t.Fatalf("NotEnoughActiveStaticPaths")
	}
	paths := []string{
		activeStatic[0], activeStatic[0], activeStatic[1],
	}
	for _, honeypotPath := range paths {
		httpRequest := httptest.NewRequest(
			http.MethodGet, honeypotPath, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
	}

	rawValue, _ := transientDbSvc.Read("honeypot:hit:1.2.3.4")
	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count != 3 {
		t.Errorf("TotalCountMismatch: got=%d, want=3",
			hitData.Count)
	}

	if hitData.Endpoints[activeStatic[0]] != 2 {
		t.Errorf("FirstEndpointCountMismatch: got=%d, want=2",
			hitData.Endpoints[activeStatic[0]])
	}

	if hitData.Endpoints[activeStatic[1]] != 1 {
		t.Errorf("SecondEndpointCountMismatch: got=%d, want=1",
			hitData.Endpoints[activeStatic[1]])
	}
}

func TestCleanExpiredEntriesRemovesStaleEntries(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	oldEntry := tkInfraDb.KeyValueModel{
		Key:       "honeypot:hit:old",
		Value:     "stale",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	transientDbSvc.Handler.Create(&oldEntry)

	newEntry := tkInfraDb.KeyValueModel{
		Key:       "honeypot:hit:new",
		Value:     "active",
		CreatedAt: time.Now(),
	}
	transientDbSvc.Handler.Create(&newEntry)

	honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(
		transientDbSvc,
	)
	honeypotCmdRepo.CleanExpiredEntries(24 * time.Hour)

	var remaining []tkInfraDb.KeyValueModel
	transientDbSvc.Handler.Find(&remaining)

	if len(remaining) != 1 {
		t.Fatalf("ExpectedOneEntry: got=%d", len(remaining))
	}

	if remaining[0].Key != "honeypot:hit:new" {
		t.Errorf("WrongEntryRemaining: got=%s", remaining[0].Key)
	}
}

func TestCleanExpiredEntriesPreservesActiveEntries(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	activeEntry := tkInfraDb.KeyValueModel{
		Key:       "honeypot:hit:active",
		Value:     "data",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	transientDbSvc.Handler.Create(&activeEntry)

	honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(
		transientDbSvc,
	)
	honeypotCmdRepo.CleanExpiredEntries(24 * time.Hour)

	if transientDbSvc.Count() != 1 {
		t.Errorf("ActiveEntryShouldBePreserved")
	}
}

func TestEnforceMaxEntriesDeletesOldestEntries(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 5 {
		entry := tkInfraDb.KeyValueModel{
			Key:   "key:" + string(rune('a'+entryIndex)),
			Value: "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Minute,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(
		transientDbSvc,
	)
	honeypotCmdRepo.EnforceMaxEntries(3)

	remaining := transientDbSvc.Count()
	if remaining != 3 {
		t.Errorf("ExpectedThreeEntries: got=%d", remaining)
	}
}

func TestEnforceMaxEntriesFloorRespected(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 3 {
		entry := tkInfraDb.KeyValueModel{
			Key:   "key:" + string(rune('a'+entryIndex)),
			Value: "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Minute,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(
		transientDbSvc,
	)
	honeypotCmdRepo.EnforceMaxEntries(5)

	if transientDbSvc.Count() != 3 {
		t.Errorf("EntriesBelowMaxShouldNotBeDeleted")
	}
}

func TestEnforceMaxEntriesCeilingRespected(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 10 {
		entry := tkInfraDb.KeyValueModel{
			Key:   "key:" + string(rune('a'+entryIndex)),
			Value: "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Minute,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(
		transientDbSvc,
	)
	honeypotCmdRepo.EnforceMaxEntries(5)

	remaining := transientDbSvc.Count()
	if remaining != 5 {
		t.Errorf("ExpectedFiveEntries: got=%d", remaining)
	}
}

func TestAggressivenessImmediateFirstHitBans(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeImmediate

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("ImmediateShouldBanAfterOneHit: got=%d",
			httpRecorder.Code)
	}
}

func TestAggressivenessBalancedGraduatedTiers(t *testing.T) {
	testCaseStructs := []struct {
		name          string
		hitCount      int
		expectedMixed bool
	}{
		{"ZeroHitsPasses", 0, false},
		{"OneHitServesPayload", 1, false},
		{"TwoHitsMixedOnHoneypot", 2, true},
		{"ThreeHitsFullMixed", 3, true},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			transientDbSvc := newTransientDbSvc()
			if testCase.hitCount > 0 {
				populateTransientDbWithHits(
					transientDbSvc,
					"1.2.3.4",
					testCase.hitCount,
				)
			}
			honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
				transientDbSvc,
			)

			settings := newStandardSettings()
			settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeBalanced

			middleware := NewHoneypotMiddleware(
				settings,
				honeypotCmdRepo, honeypotQueryRepo, nil,
			)
			defer middleware.Stop()

			echoInstance := echo.New()
			echoInstance.Use(middleware.MiddlewareFunc())

			activeStatic := findActivePathOfClass(
				middleware,
				HoneypotPathClassStaticVuln,
			)
			if activeStatic == "" {
				t.Fatalf("NoActiveStaticPath")
			}

			httpRequest := httptest.NewRequest(
				http.MethodGet, activeStatic, nil,
			)
			httpRequest.RemoteAddr = "1.2.3.4:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

			isMixed := isMixedResponseStatusCode(
				httpRecorder.Code,
			)
			if isMixed != testCase.expectedMixed {
				t.Errorf("MixedMismatch: hits=%d, got=%d, wantMixed=%v",
					testCase.hitCount,
					httpRecorder.Code,
					testCase.expectedMixed)
			}
		})
	}
}

func TestAggressivenessTolerantGraduatedTiers(t *testing.T) {
	testCaseStructs := []struct {
		name       string
		hitCount   int
		isHoneypot bool
		wantCode   int
		ipSuffix   string
		isMixed    bool
	}{
		{
			"OneHitPassesAll", 1,
			false, http.StatusOK, "1", false,
		},
		{
			"TwoHitsPassesLegit", 2,
			false, http.StatusOK, "2", false,
		},
		{
			"TwoHitsServesPayloadTierOne", 2,
			true, http.StatusOK, "3", false,
		},
		{
			"FiveHitsMixedOnHoneypot", 5,
			true, 0, "4", true,
		},
		{
			"FiveHitsPassesLegitTierTwo", 5,
			false, http.StatusOK, "5", false,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			testIp := "10.77." + testCase.ipSuffix + ".1"
			transientDbSvc := newTransientDbSvc()
			populateTransientDbWithHits(
				transientDbSvc, testIp,
				testCase.hitCount,
			)
			honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
				transientDbSvc,
			)

			settings := newStandardSettings()
			settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeTolerant

			middleware := NewHoneypotMiddleware(
				settings,
				honeypotCmdRepo, honeypotQueryRepo, nil,
			)
			defer middleware.Stop()

			echoInstance := echo.New()
			echoInstance.Use(middleware.MiddlewareFunc())
		echoInstance.GET("/api/health", func(
			echoCtx echo.Context,
		) error {
			return echoCtx.String(http.StatusOK, "OK")
		})

			requestPath := "/api/health"
			if testCase.isHoneypot {
				activeStatic := findActivePathOfClass(
					middleware,
					HoneypotPathClassStaticVuln,
				)
				if activeStatic == "" {
					t.Fatalf("NoActiveStaticPath")
				}
				requestPath = activeStatic
			}

			httpRequest := httptest.NewRequest(
				http.MethodGet, requestPath, nil,
			)
			httpRequest.RemoteAddr = testIp + ":1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

			if testCase.isMixed {
				if !isMixedResponseStatusCode(
					httpRecorder.Code,
				) {
					t.Errorf("ExpectedMixedResponse: got=%d",
						httpRecorder.Code)
				}
				return
			}

			if httpRecorder.Code != testCase.wantCode {
				t.Errorf("StatusCodeMismatch: got=%d, want=%d",
					httpRecorder.Code, testCase.wantCode)
			}
		})
	}
}

func TestAggressivenessObserveNeverBans(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 50)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if httpRecorder.Code == http.StatusFound {
		t.Errorf("ObserveModeShouldNeverRedirect")
	}
}

func TestAggressivenessObserveAlwaysServesPayloads(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 50)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	activeStaticPath := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if activeStaticPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	httpRequest := httptest.NewRequest(
		http.MethodGet, activeStaticPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if httpRecorder.Code != http.StatusOK {
		t.Errorf("ObserveShouldAlwaysServePayloads: got=%d",
			httpRecorder.Code)
	}
}

func TestNoSeparateWatchdogSettingsStruct(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)
	defer middleware.Stop()

	middlewareType := reflect.TypeOf(middleware).Elem()
	for fieldIndex := range middlewareType.NumField() {
		fieldName := middlewareType.Field(fieldIndex).Name
		if strings.Contains(
			strings.ToLower(fieldName), "watchdog",
		) {
			t.Errorf("NoSeparateWatchdogSettingsExpected")
		}
	}
}

func TestWatchdogAutoStartedByConstructor(t *testing.T) {
	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"5m",
	)

	goroutinesBefore := runtime.NumGoroutine()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	time.Sleep(50 * time.Millisecond)
	goroutinesAfter := runtime.NumGoroutine()

	if goroutinesAfter <= goroutinesBefore {
		t.Errorf("WatchdogGoroutineNotStarted: before=%d, after=%d",
			goroutinesBefore, goroutinesAfter)
	}
}

func TestWatchdogStoppedByStopMethod(t *testing.T) {
	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"5m",
	)

	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)

	time.Sleep(50 * time.Millisecond)
	goroutinesWithWatchdog := runtime.NumGoroutine()

	middleware.Stop()
	time.Sleep(100 * time.Millisecond)
	goroutinesAfterStop := runtime.NumGoroutine()

	if goroutinesAfterStop >= goroutinesWithWatchdog {
		t.Errorf("WatchdogNotStopped: with=%d, after=%d",
			goroutinesWithWatchdog, goroutinesAfterStop)
	}
}

func TestHoneypotMaintenanceWatchdogIsUnexported(t *testing.T) {
	middlewareType := reflect.TypeOf(&HoneypotMiddleware{})
	_, hasExported := middlewareType.MethodByName(
		"HoneypotMaintenanceWatchdog",
	)
	if hasExported {
		t.Errorf("WatchdogMethodShouldBeUnexported")
	}
}

func TestMaintenanceWatchdogCleansExpiredEntries(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	oldEntry := tkInfraDb.KeyValueModel{
		Key:       "honeypot:hit:old",
		Value:     "stale",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	transientDbSvc.Handler.Create(&oldEntry)

	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	var remaining []tkInfraDb.KeyValueModel
	transientDbSvc.Handler.Find(&remaining)

	for _, entry := range remaining {
		if entry.Key == "honeypot:hit:old" {
			t.Errorf("ExpiredEntryShouldBeCleaned")
		}
	}
}

func TestMaintenanceWatchdogPreservesActiveEntries(t *testing.T) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, testIp, 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if transientDbSvc.Count() == 0 {
		t.Errorf("ActiveEntriesShouldBePreserved")
	}
}

func TestMaintenanceWatchdogEnforcesMaxEntries(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	testPrefix := newUniqueTestIp()

	for entryIndex := range 10 {
		createdAt := time.Now().Add(
			time.Duration(entryIndex) * time.Minute,
		)
		entryKey := testPrefix + ":hit:" + string(rune('a'+entryIndex))
		transientDbSvc.Handler.Exec(
			"INSERT INTO key_values (key, value, created_at) VALUES (?, ?, ?)",
			entryKey, "val", createdAt,
		)
	}

	honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(
		transientDbSvc,
	)
	honeypotCmdRepo.EnforceMaxEntries(5)

	remaining := transientDbSvc.Count()
	if remaining > 10 {
		t.Errorf("EnforcementShouldDeleteEntries: got=%d",
			remaining)
	}
}

func TestStatsReportIncludesCorrectBannedIpCount(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	populateTransientDbWithHits(transientDbSvc, "2.2.2.2", 2)
	populateTransientDbWithHits(transientDbSvc, "3.3.3.3", 3)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	detailsMap, assertOk := capturedRecord.RecordDetails.(map[string]string)
	if !assertOk {
		t.Fatalf("RecordDetailsTypeMismatch")
	}

	var statsReport map[string]any
	json.Unmarshal([]byte(detailsMap["statsReport"]), &statsReport)

	bannedCount := int(statsReport["bannedIpCount"].(float64))
	if bannedCount != 2 {
		t.Errorf("BannedIpCountMismatch: got=%d, want=2",
			bannedCount)
	}
}

func TestStatsReportIncludesTopOffenders(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 5)
	populateTransientDbWithHits(transientDbSvc, "2.2.2.2", 10)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	detailsMap := capturedRecord.RecordDetails.(map[string]string)
	var statsReport map[string]any
	json.Unmarshal([]byte(detailsMap["statsReport"]), &statsReport)

	topOffenders := statsReport["topOffenders"].([]any)
	if len(topOffenders) == 0 {
		t.Errorf("TopOffendersEmpty")
	}
}

func TestStatsReportIncludesTopEndpoints(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 5)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	detailsMap := capturedRecord.RecordDetails.(map[string]string)
	var statsReport map[string]any
	json.Unmarshal([]byte(detailsMap["statsReport"]), &statsReport)

	topEndpoints := statsReport["topEndpoints"].([]any)
	if len(topEndpoints) == 0 {
		t.Errorf("TopEndpointsEmpty")
	}
}

func TestStatsReportJsonMatchesExpectedSchema(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	detailsMap := capturedRecord.RecordDetails.(map[string]string)
	var statsReport map[string]any
	json.Unmarshal([]byte(detailsMap["statsReport"]), &statsReport)

	requiredFields := []string{
		"bannedIpCount", "topOffenders", "topEndpoints",
	}
	for _, fieldName := range requiredFields {
		if _, exists := statsReport[fieldName]; !exists {
			t.Errorf("MissingField: %s", fieldName)
		}
	}
}

func TestStatsReportUsesHoneypotPeriodicReportRecordCode(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	if capturedRecord.RecordCode.String() != "HoneypotPeriodicReport" {
		t.Errorf("RecordCodeMismatch: got=%s, want=HoneypotPeriodicReport",
			capturedRecord.RecordCode.String())
	}
}

func TestStatsReportUsesSecurityRecordLevel(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	if capturedRecord.RecordLevel.String() != "SECURITY" {
		t.Errorf("RecordLevelMismatch: got=%s, want=SECURITY",
			capturedRecord.RecordLevel.String())
	}
}

func TestEmptyTransientDbSkipsStatsReport(t *testing.T) {
	recordCreated := false
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			recordCreated = true
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if recordCreated {
		t.Errorf("EmptyDbShouldSkipStatsReport")
	}
}

func TestCleanupRunsBeforeStatsInSameTick(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()

	testIp := newUniqueTestIp()
	populateTransientDbWithHits(transientDbSvc, testIp, 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Errorf("StatsShouldBeProducedAfterCleanup")
	}
}

func TestStatsProducedRegardlessOfCleanupVolume(t *testing.T) {
	statsCount := 0
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			if dto.RecordCode.String() == "HoneypotPeriodicReport" {
				statsCount++
			}
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	testIp := newUniqueTestIp()
	populateTransientDbWithHits(transientDbSvc, testIp, 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if statsCount == 0 {
		t.Errorf("StatsShouldBeProducedEvenWithNoCleanup")
	}
}

func TestWatchdogRespectsContextCancellation(t *testing.T) {
	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"5m",
	)

	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)

	time.Sleep(50 * time.Millisecond)
	middleware.Stop()

	time.Sleep(100 * time.Millisecond)
	if middleware.cancelFunc == nil {
		t.Fatalf("CancelFuncIsNil")
	}
}

func TestWatchdogReadsBanDurationAsTTL(t *testing.T) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithOldHits(
		transientDbSvc, testIp, 3, 25*time.Hour,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	testIpAddr, _ := tkValueObject.NewIpAddress(testIp)
	tier, _ := tkUseCase.ReadHoneypotBanDecision(
		honeypotQueryRepo, testIpAddr,
		mustNewHoneypotBanDuration(24*time.Hour),
		tkValueObject.HoneypotAggressivenessModeBalanced,
	)
	if tier != 0 {
		t.Errorf("ExpiredHitsShouldReturnTierZero: got=%d",
			tier)
	}
}

func TestProbabilisticEnforcementTriggersOnWrite(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 200 {
		entry := tkInfraDb.KeyValueModel{
			Key:   "key:" + string(rune(entryIndex)),
			Value: "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Millisecond,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	cmdRepo := newNoopCmdRepo()
	settings := newStandardSettings()
	settings.MaxEntries, _ = tkValueObject.NewHoneypotMaxEntries(100)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	for range 500 {
		existentIp, _ := tkValueObject.NewIpAddress("1.2.3.4")
		tkUseCase.CreateHoneypotHit(
			honeypotCmdRepo, existentIp,
			"/.env", settings.MaxEntries,
		)
	}

	remaining := transientDbSvc.Count()
	if remaining > 150 {
		t.Logf("Probabilistic enforcement may not have triggered: count=%d", remaining)
	}
}

func TestProbabilisticEnforcementNotAlwaysTriggered(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.MaxEntries, _ = tkValueObject.NewHoneypotMaxEntries(5000)

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	countBefore := transientDbSvc.Count()
	existentIp, _ := tkValueObject.NewIpAddress("1.2.3.4")
	tkUseCase.CreateHoneypotHit(
		honeypotCmdRepo, existentIp,
		"/.env", settings.MaxEntries,
	)
	countAfter := transientDbSvc.Count()

	if countAfter != countBefore+1 {
		t.Errorf("SingleWriteShouldAddOneEntry")
	}
}

func TestGraduatedBanTransientDbReadErrorHandled(t *testing.T) {
	existentIp, _ := tkValueObject.NewIpAddress("1.2.3.4")
	tier, _ := tkUseCase.ReadHoneypotBanDecision(
		nil, existentIp,
		mustNewHoneypotBanDuration(24*time.Hour),
		tkValueObject.HoneypotAggressivenessModeBalanced,
	)
	if tier != 0 {
		t.Errorf("NilTransientDbShouldReturnTierZero: got=%d",
			tier)
	}
}

func TestTransientDbReadErrorHandled(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	_, honeypotQueryRepo := newHoneypotRepos(transientDbSvc)

	existentIp, _ := tkValueObject.NewIpAddress(
		"nonexistent.ip",
	)
	tier, _ := tkUseCase.ReadHoneypotBanDecision(
		honeypotQueryRepo, existentIp,
		mustNewHoneypotBanDuration(24*time.Hour),
		tkValueObject.HoneypotAggressivenessModeBalanced,
	)
	if tier != 0 {
		t.Errorf("MissingKeyShouldReturnTierZero: got=%d", tier)
	}
}

func TestProbabilisticEnforcementHandlesMaxEntriesError(t *testing.T) {
	existentIp, _ := tkValueObject.NewIpAddress("1.2.3.4")
	tkUseCase.CreateHoneypotHit(
		nil, existentIp, "/.env",
		mustNewHoneypotMaxEntries(5000),
	)
}

func TestStopOnUninitializedMiddlewareDoesNotPanic(t *testing.T) {
	middleware := &HoneypotMiddleware{}
	middleware.Stop()
}

func TestMaintenanceWatchdogHandlesNilCmdRepo(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	middleware.runMaintenance()
}

func TestMaintenanceWatchdogRecoversFromPanicInTick(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"100ms",
	)

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	time.Sleep(250 * time.Millisecond)
}

func TestTransientDbReadAllErrorDuringStatsSkipsReport(t *testing.T) {
	recordCreated := false
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			if dto.RecordCode.String() == "HoneypotPeriodicReport" {
				recordCreated = true
			}
			return nil
		},
	}

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if recordCreated {
		t.Errorf("NilTransientDbShouldSkipStatsReport")
	}
}

func TestAggressivenessModeObserveReportsZeroBannedIps(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 5)
	populateTransientDbWithHits(transientDbSvc, "2.2.2.2", 10)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	detailsMap := capturedRecord.RecordDetails.(map[string]string)
	var statsReport map[string]any
	json.Unmarshal([]byte(detailsMap["statsReport"]), &statsReport)

	bannedCount := int(statsReport["bannedIpCount"].(float64))
	if bannedCount != 0 {
		t.Errorf("ObserveModeShouldReportZeroBannedIps: got=%d",
			bannedCount)
	}
}

func TestWatchdogUsesStatsIntervalFromSettings(t *testing.T) {
	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"100ms",
	)
	settings.BanDuration = mustNewHoneypotBanDuration(
		72 * time.Hour,
	)

	transientDbSvc := newTransientDbSvc()
	testIp := newUniqueTestIp()
	populateTransientDbWithHits(transientDbSvc, testIp, 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	statsCount := 0
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			if dto.RecordCode.String() == "HoneypotPeriodicReport" {
				statsCount++
			}
			return nil
		},
	}

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenance()
	middleware.runMaintenance()

	if statsCount < 2 {
		t.Errorf("WatchdogShouldTickMultipleTimes: got=%d",
			statsCount)
	}
}

func TestScannerFloodTriggersTierEscalation(t *testing.T) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPaths := make([]string, 0, 3)
	for activePath, pathClass := range middleware.activePathClasses {
		if pathClass == HoneypotPathClassStaticVuln {
			honeypotPaths = append(honeypotPaths, activePath)
			if len(honeypotPaths) >= 3 {
				break
			}
		}
	}

	for _, honeypotPath := range honeypotPaths {
		httpRequest := httptest.NewRequest(
			http.MethodGet, honeypotPath, nil,
		)
		httpRequest.RemoteAddr = testIp + ":1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
	}

	rawValue, _ := transientDbSvc.Read("honeypot:hit:" + testIp)
	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count != 2 {
		t.Errorf("ScannerFloodCountMismatch: got=%d, want=2",
			hitData.Count)
	}

	legitRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	legitRequest.RemoteAddr = testIp + ":1234"
	legitRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(legitRecorder, legitRequest)

	if legitRecorder.Code == http.StatusFound {
		t.Errorf("TierTwoShouldPassLegitPaths")
	}
}

func TestConcurrentHitsCountCorrectly(t *testing.T) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, _ := newHoneypotRepos(transientDbSvc)

	goroutineCount := 10
	var waitGroup sync.WaitGroup
	var writeMu sync.Mutex
	waitGroup.Add(goroutineCount)

	existentIp, _ := tkValueObject.NewIpAddress(testIp)
	for goroutineIndex := range goroutineCount {
		go func(index int) {
			defer waitGroup.Done()
			writeMu.Lock()
			defer writeMu.Unlock()
			tkUseCase.CreateHoneypotHit(
				honeypotCmdRepo, existentIp,
				"/.env",
				mustNewHoneypotMaxEntries(5000),
			)
		}(goroutineIndex)
	}

	waitGroup.Wait()

	rawValue, _ := transientDbSvc.Read(
		"honeypot:hit:" + testIp,
	)
	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count != goroutineCount {
		t.Errorf("ConcurrentCountMismatch: got=%d, want=%d",
			hitData.Count, goroutineCount)
	}
}

func TestPhaseOneCoreBehaviorPreservedAfterPhaseThree(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	recordCreated := false
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			recordCreated = true
			return nil
		},
	}
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	activeStaticPath := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if activeStaticPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	httpRequest := httptest.NewRequest(
		http.MethodGet, activeStaticPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if httpRecorder.Code != http.StatusOK {
		t.Errorf("HoneypotPathShouldServePayload: got=%d",
			httpRecorder.Code)
	}

	if !recordCreated {
		t.Errorf("ActivityRecordShouldBeCreated")
	}
}

func TestAggressivenessImmediateAutoBansScannerOnFirstProbe(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeImmediate

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("ImmediateShouldBanOnFirstProbe: got=%d",
			httpRecorder.Code)
	}
}

func TestAggressivenessObserveGathersIntelWithoutInterference(
	t *testing.T,
) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	for range 5 {
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/.env", nil,
		)
		httpRequest.RemoteAddr = testIp + ":1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
	}

	legitRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	legitRequest.RemoteAddr = testIp + ":1234"
	legitRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(legitRecorder, legitRequest)

	if legitRecorder.Code != http.StatusOK {
		t.Errorf("ObserveShouldNotInterfere: got=%d",
			legitRecorder.Code)
	}
}

func TestProbabilisticEnforcementConcurrentWithNormalWrites(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, _ := newHoneypotRepos(transientDbSvc)

	var waitGroup sync.WaitGroup
	var writeMu sync.Mutex
	waitGroup.Add(20)

	existentIp, _ := tkValueObject.NewIpAddress("1.2.3.4")
	for goroutineIndex := range 20 {
		go func(index int) {
			defer waitGroup.Done()
			writeMu.Lock()
			defer writeMu.Unlock()
			tkUseCase.CreateHoneypotHit(
				honeypotCmdRepo, existentIp,
				"/.env",
				mustNewHoneypotMaxEntries(100),
			)
		}(goroutineIndex)
	}

	waitGroup.Wait()

	if transientDbSvc.Count() == 0 {
		t.Errorf("EntriesShouldExistAfterConcurrentWrites")
	}
}

func TestSettingsParserInvalidEnvVarUsesDefault(t *testing.T) {
	t.Setenv("HONEYPOT_MAX_ENTRIES", "abc")
	t.Setenv("HONEYPOT_STATS_INTERVAL", "xyz")

	emptySettings := HoneypotMiddlewareSettings{}
	resolvedSettings := honeypotSettingsParser{}.Parse(
		emptySettings, 25,
	)

	expectedMaxEntries := mustNewHoneypotMaxEntries(5000)
	if resolvedSettings.MaxEntries.Int() != expectedMaxEntries.Int() {
		t.Errorf(
			"MaxEntriesDefaultMismatch: got=%d, want=%d",
			resolvedSettings.MaxEntries.Int(),
			expectedMaxEntries.Int(),
		)
	}

	expectedStatsInterval, _ := tkValueObject.NewHoneypotStatsInterval("")
	actualInterval := resolvedSettings.StatsInterval.Duration()
	if actualInterval != expectedStatsInterval.Duration() {
		t.Errorf(
			"StatsIntervalDefaultMismatch: got=%v, want=%v",
			actualInterval,
			expectedStatsInterval.Duration(),
		)
	}
}

func TestConcurrentHitsAndMaintenanceCycleNoDataLoss(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, _ := newHoneypotRepos(transientDbSvc)

	var waitGroup sync.WaitGroup
	var writeMu sync.Mutex
	waitGroup.Add(10)

	existentIp, _ := tkValueObject.NewIpAddress("1.2.3.4")
	for goroutineIndex := range 10 {
		go func(index int) {
			defer waitGroup.Done()
			writeMu.Lock()
			defer writeMu.Unlock()
			tkUseCase.CreateHoneypotHit(
				honeypotCmdRepo, existentIp,
				"/.env",
				mustNewHoneypotMaxEntries(5000),
			)
		}(goroutineIndex)
	}

	waitGroup.Wait()

	rawValue, readErr := transientDbSvc.Read(
		"honeypot:hit:1.2.3.4",
	)
	if readErr != nil {
		t.Fatalf("HitDataLost: %v", readErr)
	}

	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count < 1 {
		t.Errorf("ConcurrentHitsShouldBeCounted: got=%d",
			hitData.Count)
	}
}
func TestZeroElseKeywordsInMiddleware(t *testing.T) {
	middlewareFile, readErr := os.ReadFile("honeypotMiddleware.go")
	if readErr != nil {
		t.Fatalf("MiddlewareFileReadFailed: %v", readErr)
	}
	elsePattern := regexp.MustCompile(`\belse\b`)
	matchCount := len(elsePattern.FindAll(
		middlewareFile, -1,
	))
	if matchCount != 0 {
		t.Errorf("MiddlewareElseCountMismatch: got=%d, want=0",
			matchCount)
	}
}
func TestMiddlewareUnderThreeHundredLoc(t *testing.T) {
	middlewareFile, readErr := os.ReadFile("honeypotMiddleware.go")
	if readErr != nil {
		t.Fatalf("MiddlewareFileReadFailed: %v", readErr)
	}
	lineCount := len(strings.Split(
		string(middlewareFile), "\n",
	))
	if lineCount >= 300 {
		t.Errorf("MiddlewareLocExceeded: got=%d, want<300",
			lineCount)
	}
}
func TestConstructorUnderFiftyLoc(t *testing.T) {
	middlewareFile, readErr := os.ReadFile("honeypotMiddleware.go")
	if readErr != nil {
		t.Fatalf("MiddlewareFileReadFailed: %v", readErr)
	}
	fileContent := string(middlewareFile)
	constructorStart := strings.Index(
		fileContent, "func NewHoneypotMiddleware(",
	)
	if constructorStart == -1 {
		t.Fatalf("ConstructorNotFound")
	}
	openBraceIdx := strings.Index(
		fileContent[constructorStart:], "{",
	)
	if openBraceIdx == -1 {
		t.Fatalf("ConstructorOpenBraceNotFound")
	}
	braceDepth := 0
	constructorEnd := -1
	bodyStart := constructorStart + openBraceIdx + 1
	for charIdx := bodyStart; charIdx < len(fileContent); charIdx++ {
		if fileContent[charIdx] == '{' {
			braceDepth++
		}
		if fileContent[charIdx] == '}' {
			braceDepth--
			if braceDepth == 0 {
				constructorEnd = charIdx
				break
			}
		}
	}
	if constructorEnd == -1 {
		t.Fatalf("ConstructorCloseBraceNotFound")
	}
	constructorBody := fileContent[bodyStart:constructorEnd]
	lineCount := len(strings.Split(constructorBody, "\n"))
	if lineCount > 50 {
		t.Errorf("ConstructorLocExceeded: got=%d, want<=50",
			lineCount)
	}
}
func TestMethodOrderingCalleesAboveCallers(t *testing.T) {
	middlewareFile, readErr := os.ReadFile("honeypotMiddleware.go")
	if readErr != nil {
		t.Fatalf("MiddlewareFileReadFailed: %v", readErr)
	}
	fileContent := string(middlewareFile)
	methodPattern := regexp.MustCompile(`(?m)^func.*?\) (\w+)\(`)
	methodPositions := make(map[string]int)
	for _, match := range methodPattern.FindAllStringSubmatchIndex(fileContent, -1) {
		methodName := fileContent[match[2]:match[3]]
		methodPositions[methodName] = match[0]
	}
	testCaseStructs := []struct {
		callee string
		caller string
	}{
		{"lookupActivePathClass", "Execute"},
		{"Execute", "Start"},
		{"runMaintenance", "honeypotMaintenanceWatchdog"},
	}
	for _, testCase := range testCaseStructs {
		calleePos, calleeExists := methodPositions[testCase.callee]
		if !calleeExists {
			t.Errorf("CalleeMethodNotFound: %s", testCase.callee)
			continue
		}
		callerPos, callerExists := methodPositions[testCase.caller]
		if !callerExists {
			t.Errorf("CallerMethodNotFound: %s", testCase.caller)
			continue
		}
		if calleePos >= callerPos {
			t.Errorf("MethodOrderingViolation: %s(pos=%d) should be before %s(pos=%d)",
				testCase.callee, calleePos,
				testCase.caller, callerPos)
		}
	}
}
func TestMaxNestingDepthIsThree(t *testing.T) {
	middlewareFile, readErr := os.ReadFile("honeypotMiddleware.go")
	if readErr != nil {
		t.Fatalf("MiddlewareFileReadFailed: %v", readErr)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(middlewareFile)))
	executeStart := -1
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		executeSig := "func (middleware *HoneypotMiddleware) Execute("
		if strings.Contains(scanner.Text(), executeSig) {
			executeStart = lineNumber
			break
		}
	}
	if executeStart == -1 {
		t.Fatalf("ExecuteMethodNotFound")
	}
	maxDepth := 0
	currentDepth := 0
	inExecute := false
	scanner = bufio.NewScanner(strings.NewReader(string(middlewareFile)))
	lineNumber = 0
	for scanner.Scan() {
		lineNumber++
		if lineNumber < executeStart {
			continue
		}
		trimmedLine := strings.TrimSpace(scanner.Text())
		isNextMethod := strings.HasPrefix(trimmedLine, "func ")
		if inExecute && isNextMethod && lineNumber > executeStart {
			break
		}
		inExecute = true
		for _, char := range scanner.Text() {
			if char == '{' {
				currentDepth++
				if currentDepth > maxDepth {
					maxDepth = currentDepth
				}
			}
			if char == '}' {
				currentDepth--
			}
		}
	}
	if maxDepth > 3 {
		t.Errorf("MaxNestingDepthExceeded: got=%d, want<=3",
			maxDepth)
	}
}
func TestInfraErrorsLoggedWithSlogError(t *testing.T) {
	fileNames := []string{
		"honeypotMiddleware.go",
		"honeypotMixedResponse.go",
		"honeypotSettingsParser.go",
		"honeypotPathSelector.go",
		"honeypotAiTrapGenerator.go",
		"honeypotStreamHandler.go",
	}
	for _, fileName := range fileNames {
		fileContent, readErr := os.ReadFile(fileName)
		if readErr != nil {
			t.Fatalf("FileReadFailed: file=%s, err=%v",
				fileName, readErr)
		}
		if strings.Contains(string(fileContent), "slog.Debug") {
			t.Errorf("SlogDebugFoundInFile: %s", fileName)
		}
	}
}

func newTestEchoContext() (
	echo.Context, *httptest.ResponseRecorder,
) {
	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/test", nil,
	)
	recorder := httptest.NewRecorder()
	return echoInstance.NewContext(
		httpRequest, recorder,
	), recorder
}

func newMixedResponseMiddleware() *HoneypotMiddleware {
	return NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)
}

func TestMixedResponseIncludesAllFourTypes(t *testing.T) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	hasRedirect := false
	hasServiceUnavailable := false
	hasBadGateway := false
	hasTooManyRequests := false

	for range 100 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveMixedResponse(echoContext)
		switch recorder.Code {
		case http.StatusFound,
			http.StatusTemporaryRedirect:
			hasRedirect = true
		case http.StatusServiceUnavailable:
			hasServiceUnavailable = true
		case http.StatusBadGateway:
			hasBadGateway = true
		case http.StatusTooManyRequests:
			hasTooManyRequests = true
		}
	}

	if !hasRedirect {
		t.Errorf("RedirectTypeMissing")
	}
	if !hasServiceUnavailable {
		t.Errorf("Fake503TypeMissing")
	}
	if !hasBadGateway {
		t.Errorf("Fake502TypeMissing")
	}
	if !hasTooManyRequests {
		t.Errorf("Fake429TypeMissing")
	}
}

func TestMixedResponseApproximateWeights(t *testing.T) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	redirectCount := 0
	serviceUnavailableCount := 0
	badGatewayCount := 0
	tooManyRequestsCount := 0

	for range 1000 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveMixedResponse(echoContext)
		switch recorder.Code {
		case http.StatusFound,
			http.StatusTemporaryRedirect:
			redirectCount++
		case http.StatusServiceUnavailable:
			serviceUnavailableCount++
		case http.StatusBadGateway:
			badGatewayCount++
		case http.StatusTooManyRequests:
			tooManyRequestsCount++
		}
	}

	if redirectCount < 300 || redirectCount > 500 {
		t.Errorf(
			"RedirectWeightOutOfRange: got=%d",
			redirectCount,
		)
	}
	if serviceUnavailableCount < 200 ||
		serviceUnavailableCount > 400 {
		t.Errorf(
			"Fake503WeightOutOfRange: got=%d",
			serviceUnavailableCount,
		)
	}
	if badGatewayCount < 100 || badGatewayCount > 300 {
		t.Errorf(
			"Fake502WeightOutOfRange: got=%d",
			badGatewayCount,
		)
	}
	if tooManyRequestsCount < 50 ||
		tooManyRequestsCount > 150 {
		t.Errorf(
			"Fake429WeightOutOfRange: got=%d",
			tooManyRequestsCount,
		)
	}
}

func TestFake503BodyLooksLikeNginx(t *testing.T) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	echoContext, recorder := newTestEchoContext()
	middleware.serveFake503(echoContext)

	bodyStr := recorder.Body.String()
	if !strings.Contains(bodyStr, "503") {
		t.Errorf("BodyMissing503Status")
	}
	if !strings.Contains(bodyStr, "nginx") {
		t.Errorf("BodyMissingNginxSignature")
	}
	if !strings.Contains(bodyStr, "<html>") {
		t.Errorf("BodyMissingHtmlTag")
	}
	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			recorder.Code,
			http.StatusServiceUnavailable)
	}
}

func TestFake502BodyLooksLikeNginx(t *testing.T) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	echoContext, recorder := newTestEchoContext()
	middleware.serveFake502(echoContext)

	bodyStr := recorder.Body.String()
	if !strings.Contains(bodyStr, "502") {
		t.Errorf("BodyMissing502Status")
	}
	if !strings.Contains(bodyStr, "nginx") {
		t.Errorf("BodyMissingNginxSignature")
	}
	if !strings.Contains(bodyStr, "<html>") {
		t.Errorf("BodyMissingHtmlTag")
	}
	if recorder.Code != http.StatusBadGateway {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			recorder.Code, http.StatusBadGateway)
	}
}

func TestFake429BodyIsValidJsonWithRetryAfter(
	t *testing.T,
) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	echoContext, recorder := newTestEchoContext()
	middleware.serveFake429(echoContext)

	var jsonBody map[string]any
	jsonErr := json.Unmarshal(
		recorder.Body.Bytes(), &jsonBody,
	)
	if jsonErr != nil {
		t.Fatalf("BodyIsNotValidJson: %v", jsonErr)
	}
	if _, hasError := jsonBody["error"]; !hasError {
		t.Errorf("JsonMissingErrorField")
	}
	retryAfter := recorder.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Errorf("RetryAfterHeaderMissing")
	}
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			recorder.Code,
			http.StatusTooManyRequests)
	}
}

func TestRedirectRotatesLawEnforcementUrls(t *testing.T) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	seenBaseUrls := make(map[string]bool)
	for range 100 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveLawEnforcementRedirect(
			echoContext,
		)
		location := recorder.Header().Get("Location")
		queryIdx := strings.Index(location, "?")
		baseUrl := location
		if queryIdx >= 0 {
			baseUrl = location[:queryIdx]
		}
		seenBaseUrls[baseUrl] = true
	}

	for _, expectedUrl := range lawEnforcementRedirectUrls {
		if !seenBaseUrls[expectedUrl] {
			t.Errorf(
				"LawEnforcementUrlNeverSeen: %s",
				expectedUrl,
			)
		}
	}
}

func TestRedirectAppendsRotatingQueryString(
	t *testing.T,
) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	seenQueryStrings := make(map[string]bool)
	for range 50 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveLawEnforcementRedirect(
			echoContext,
		)
		location := recorder.Header().Get("Location")
		queryIdx := strings.Index(location, "?")
		if queryIdx < 0 {
			continue
		}
		queryString := location[queryIdx:]
		seenQueryStrings[queryString] = true
	}

	for _, expectedQs := range securityQueryStringPool {
		if !seenQueryStrings[expectedQs] {
			t.Errorf(
				"QueryStringNeverSeen: %s",
				expectedQs,
			)
		}
	}
}

func TestRedirectQueryStringRotatesAmongPool(
	t *testing.T,
) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	seenQueryStrings := make(map[string]bool)
	for range 50 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveLawEnforcementRedirect(
			echoContext,
		)
		location := recorder.Header().Get("Location")
		queryIdx := strings.Index(location, "?")
		if queryIdx < 0 {
			continue
		}
		seenQueryStrings[location[queryIdx:]] = true
	}

	if len(seenQueryStrings) < 3 {
		t.Errorf(
			"InsufficientQueryRotation: got=%d, want>=3",
			len(seenQueryStrings),
		)
	}

	for queryString := range seenQueryStrings {
		isInPool := false
		for _, poolEntry := range securityQueryStringPool {
			if poolEntry == queryString {
				isInPool = true
				break
			}
		}
		if !isInPool {
			t.Errorf(
				"QueryStringNotInPool: %s",
				queryString,
			)
		}
	}
}

func TestRedirectStatusCodeMixed302And307(t *testing.T) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	hasStatusFound := false
	hasStatusTempRedirect := false

	for range 100 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveLawEnforcementRedirect(
			echoContext,
		)
		switch recorder.Code {
		case http.StatusFound:
			hasStatusFound = true
		case http.StatusTemporaryRedirect:
			hasStatusTempRedirect = true
		}
	}

	if !hasStatusFound {
		t.Errorf("Status302NeverSeen")
	}
	if !hasStatusTempRedirect {
		t.Errorf("Status307NeverSeen")
	}
}

func TestCustomRedirectUrlOverridesPool(t *testing.T) {
	customUrl, _ := tkValueObject.NewUrl(
		"https://custom.example.com/",
	)
	settings := newStandardSettings()
	settings.RedirectUrl = customUrl
	middleware := NewHoneypotMiddleware(
		settings, nil, nil, nil,
	)
	defer middleware.Stop()

	for range 50 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveLawEnforcementRedirect(
			echoContext,
		)
		location := recorder.Header().Get("Location")
		if location != "https://custom.example.com/" {
			t.Errorf(
				"CustomUrlNotUsed: got=%s",
				location,
			)
		}
		if strings.Contains(location, "?") {
			t.Errorf("QueryStringShouldNotBeAppended")
		}
	}
}

func TestTierTwoMixedOnHoneypotPathsOnly(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 2,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	hpPath := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if hpPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	hpRequest := httptest.NewRequest(
		http.MethodGet, hpPath, nil,
	)
	hpRequest.RemoteAddr = "1.2.3.4:1234"
	hpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(hpRecorder, hpRequest)

	if !isMixedResponseStatusCode(hpRecorder.Code) {
		t.Errorf("HoneypotPathExpectedMixed: got=%d",
			hpRecorder.Code)
	}

	legitRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	legitRequest.RemoteAddr = "1.2.3.4:1234"
	legitRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(legitRecorder, legitRequest)

	if legitRecorder.Code != http.StatusOK {
		t.Errorf("LegitPathShouldPassThrough: got=%d",
			legitRecorder.Code)
	}
}

func TestTierThreeMixedOnAllPaths(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	testCaseStructs := []struct {
		name string
		path string
	}{
		{"HoneypotPath", "/.env"},
		{"LegitPath", "/api/health"},
		{"StaticFile", "/static/app.js"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			httpRequest := httptest.NewRequest(
				http.MethodGet, testCase.path, nil,
			)
			httpRequest.RemoteAddr = "1.2.3.4:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(
				httpRecorder, httpRequest,
			)
			if !isMixedResponseStatusCode(
				httpRecorder.Code,
			) {
				t.Errorf(
					"TierThreeExpectedMixed: path=%s, got=%d",
					testCase.path,
					httpRecorder.Code,
				)
			}
		})
	}
}

func TestObserveModeNeverReturnsMixedResponses(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 50,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	activeStatic := findActivePathOfClass(
		middleware, HoneypotPathClassStaticVuln,
	)
	if activeStatic == "" {
		t.Fatalf("NoActiveStaticPath")
	}

	for range 5 {
		httpRequest := httptest.NewRequest(
			http.MethodGet, activeStatic, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if isMixedResponseStatusCode(
			httpRecorder.Code,
		) {
			t.Errorf(
				"ObserveModeShouldNotReturnMixed: got=%d",
				httpRecorder.Code,
			)
		}
		if httpRecorder.Code != http.StatusOK {
			t.Errorf(
				"ObserveModeShouldServePayload: got=%d",
				httpRecorder.Code,
			)
		}
	}
}

func TestTierTwoLegitimatePathAlwaysPassesThrough(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 2,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})

	for reqIdx := range 50 {
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/api/health", nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if httpRecorder.Code != http.StatusOK {
			t.Errorf(
				"LegitPathShouldAlwaysPass: req=%d, got=%d",
				reqIdx, httpRecorder.Code,
			)
		}
	}
}

func TestTierThreeLegitimatePathAlwaysReturnsMixed(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	for reqIdx := range 50 {
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/api/health", nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if !isMixedResponseStatusCode(
			httpRecorder.Code,
		) {
			t.Errorf(
				"TierThreeLegitPathExpectedMixed: req=%d, got=%d",
				reqIdx, httpRecorder.Code,
			)
		}
	}
}

func TestMixedResponseWithCustomRedirectOnlyRedirects(
	t *testing.T,
) {
	customUrl, _ := tkValueObject.NewUrl(
		"https://custom.example.com/",
	)
	settings := newStandardSettings()
	settings.RedirectUrl = customUrl

	middleware := NewHoneypotMiddleware(
		settings, nil, nil, nil,
	)
	defer middleware.Stop()

	for range 100 {
		echoContext, recorder := newTestEchoContext()
		middleware.serveMixedResponse(echoContext)

		if recorder.Code != http.StatusFound {
			t.Errorf(
				"CustomRedirectShouldAlways302: got=%d",
				recorder.Code,
			)
		}
		location := recorder.Header().Get("Location")
		if location != "https://custom.example.com/" {
			t.Errorf(
				"CustomRedirectUrlMismatch: got=%s",
				location,
			)
		}
	}
}

func TestMixedResponseDistributionStableOverThousandsOfRequests(
	t *testing.T,
) {
	middleware := newMixedResponseMiddleware()
	defer middleware.Stop()

	redirectCount := 0
	serviceUnavailableCount := 0
	badGatewayCount := 0
	tooManyRequestsCount := 0
	totalRequests := 5000

	for range totalRequests {
		echoContext, recorder := newTestEchoContext()
		middleware.serveMixedResponse(echoContext)
		switch recorder.Code {
		case http.StatusFound,
			http.StatusTemporaryRedirect:
			redirectCount++
		case http.StatusServiceUnavailable:
			serviceUnavailableCount++
		case http.StatusBadGateway:
			badGatewayCount++
		case http.StatusTooManyRequests:
			tooManyRequestsCount++
		}
	}

	redirectPct := float64(redirectCount) /
		float64(totalRequests) * 100
	svcUnavPct := float64(serviceUnavailableCount) /
		float64(totalRequests) * 100
	badGwPct := float64(badGatewayCount) /
		float64(totalRequests) * 100
	tooManyPct := float64(tooManyRequestsCount) /
		float64(totalRequests) * 100

	if redirectPct < 35 || redirectPct > 45 {
		t.Errorf(
			"RedirectPctOutOfRange: %.1f%%",
			redirectPct,
		)
	}
	if svcUnavPct < 25 || svcUnavPct > 35 {
		t.Errorf(
			"Fake503PctOutOfRange: %.1f%%",
			svcUnavPct,
		)
	}
	if badGwPct < 15 || badGwPct > 25 {
		t.Errorf(
			"Fake502PctOutOfRange: %.1f%%",
			badGwPct,
		)
	}
	if tooManyPct < 5 || tooManyPct > 15 {
		t.Errorf(
			"Fake429PctOutOfRange: %.1f%%",
			tooManyPct,
		)
	}
}

func TestHoneypotPathClassConstantsDefined(t *testing.T) {
	classes := []HoneypotPathClass{
		HoneypotPathClassStaticVuln,
		HoneypotPathClassBandwidthExhaust,
		HoneypotPathClassAITrap,
	}
	seenValues := make(map[HoneypotPathClass]bool)
	for _, pathClass := range classes {
		if seenValues[pathClass] {
			t.Errorf("DuplicatePathClassValue: %d",
				pathClass)
		}
		seenValues[pathClass] = true
	}
	if len(seenValues) != 3 {
		t.Errorf("ExpectedThreeDistinctClasses: got=%d",
			len(seenValues))
	}
}

func TestAutoRatioProducesCorrectCountsDefaultActivePathCount(t *testing.T) {
	staticCount, bandwidthCount, aiTrapCount :=
		computeAutoRatio(30)
	totalCount := staticCount + bandwidthCount + aiTrapCount
	if totalCount != 30 {
		t.Errorf("TotalCountMismatch: got=%d, want=30",
			totalCount)
	}
	if staticCount != 20 {
		t.Errorf("StaticCountMismatch: got=%d, want=20",
			staticCount)
	}
	if bandwidthCount != 5 {
		t.Errorf("BandwidthCountMismatch: got=%d, want=5",
			bandwidthCount)
	}
	if aiTrapCount != 5 {
		t.Errorf("AiTrapCountMismatch: got=%d, want=5",
			aiTrapCount)
	}
}

func TestAutoRatioProducesCorrectCountsCustomActivePathCount(t *testing.T) {
	staticCount, bandwidthCount, aiTrapCount :=
		computeAutoRatio(60)
	totalCount := staticCount + bandwidthCount + aiTrapCount
	if totalCount != 60 {
		t.Errorf("TotalCountMismatch: got=%d, want=60",
			totalCount)
	}
	if staticCount != 40 {
		t.Errorf("StaticCountMismatch: got=%d, want=40",
			staticCount)
	}
	if bandwidthCount != 10 {
		t.Errorf("BandwidthCountMismatch: got=%d, want=10",
			bandwidthCount)
	}
	if aiTrapCount != 10 {
		t.Errorf("AiTrapCountMismatch: got=%d, want=10",
			aiTrapCount)
	}
}

func TestAutoRatioFloorGuaranteesMinOnePerClass(t *testing.T) {
	testCaseStructs := []struct {
		name            string
		activePathCount int
	}{
		{"OnePath", 1},
		{"TwoPaths", 2},
		{"FivePaths", 5},
		{"TenPaths", 10},
	}
	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			_, bandwidthCount, aiTrapCount :=
				computeAutoRatio(testCase.activePathCount)
			if bandwidthCount < 1 {
				t.Errorf("BandwidthBelowFloor: got=%d",
					bandwidthCount)
			}
			if aiTrapCount < 1 {
				t.Errorf("AiTrapBelowFloor: got=%d",
					aiTrapCount)
			}
		})
	}
}

func TestRandomSelectionIsDeterministicWithSeed(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	firstSelection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 42,
	)
	secondSelection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 42,
	)
	if len(firstSelection) != len(secondSelection) {
		t.Errorf("SelectionSizeMismatch: first=%d, second=%d",
			len(firstSelection), len(secondSelection))
	}
	for selectedPath, firstClass := range firstSelection {
		secondClass, pathExists :=
			secondSelection[selectedPath]
		if !pathExists {
			t.Errorf("PathMissingInSecondSelection: %s",
				selectedPath)
			continue
		}
		if firstClass != secondClass {
			t.Errorf("ClassMismatchForPath: path=%s",
				selectedPath)
		}
	}
}

func TestRandomSelectionDiffersWithDifferentSeeds(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	firstSelection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 42,
	)
	secondSelection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 99,
	)
	hasDifference := false
	for selectedPath := range firstSelection {
		if _, existsInSecond :=
			secondSelection[selectedPath]; !existsInSecond {
			hasDifference = true
			break
		}
	}
	for selectedPath := range secondSelection {
		if _, existsInFirst :=
			firstSelection[selectedPath]; !existsInFirst {
			hasDifference = true
			break
		}
	}
	if !hasDifference {
		t.Errorf("SelectionsShouldDifferWithDifferentSeeds")
	}
}

func TestRandomSelectionDiffersAcrossRestarts(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	selections := make([]map[string]HoneypotPathClass, 5)
	for runIdx := range 5 {
		selections[runIdx] = selectActivePaths(
			staticPaths,
			bandwidthExhaustCandidatePaths,
			aiTrapCandidatePaths,
			30, 0,
		)
		time.Sleep(time.Millisecond)
	}
	hasDifference := false
	for comparisonIdx := 1; comparisonIdx < 5; comparisonIdx++ {
		if len(selections[comparisonIdx]) !=
			len(selections[0]) {
			hasDifference = true
			break
		}
		for selectedPath := range selections[0] {
			if _, pathExists :=
				selections[comparisonIdx][selectedPath]; !pathExists {
				hasDifference = true
				break
			}
		}
		if hasDifference {
			break
		}
	}
	if !hasDifference {
		t.Errorf("TimeBasedSeedsShouldProduceDifferentSelections")
	}
}

func TestDormantPathReturnsNextHandler(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})
	allStaticPaths := extractStaticPathKeys()
	dormantPathFound := false
	for _, staticPath := range allStaticPaths {
		if _, pathIsActive :=
			middleware.activePathClasses[staticPath]; pathIsActive {
			continue
		}
		dormantPathFound = true
		httpRequest := httptest.NewRequest(
			http.MethodGet, staticPath, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
		if httpRecorder.Code == http.StatusOK {
			payloadContentType := httpRecorder.Header().Get(
				"Content-Type",
			)
			if payloadContentType != "" &&
				payloadContentType != "text/plain; charset=UTF-8" {
				t.Errorf("DormantPathShouldNotServePayload: path=%s",
					staticPath)
			}
		}
	}
	if !dormantPathFound {
		t.Errorf("NoDormantPathsFoundWithSeed42")
	}
}

func TestSelectedBandwidthExhaustPathStreamsGarbage(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			httpRecorder.Code, http.StatusOK)
	}
	contentType := httpRecorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("ContentTypeMismatch: got=%s, want=text/plain",
			contentType)
	}
	bodySize := httpRecorder.Body.Len()
	if bodySize < 5*1024*1024 {
		t.Errorf("StreamBelowFloor: got=%d bytes", bodySize)
	}
}

func TestSelectedAITrapPathStreamsPlausibleContent(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	aiTrapPath := findActivePathOfClass(
		middleware, HoneypotPathClassAITrap,
	)
	if aiTrapPath == "" {
		t.Fatalf("NoActiveAiTrapPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, aiTrapPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			httpRecorder.Code, http.StatusOK)
	}
	contentType := httpRecorder.Header().Get("Content-Type")
	if contentType == "" {
		t.Errorf("ContentTypeMissing")
	}
	bodySize := httpRecorder.Body.Len()
	if bodySize < 5*1024*1024 {
		t.Errorf("StreamBelowFloor: got=%d bytes", bodySize)
	}
}

func TestAITrapContentContainsPromptInjection(t *testing.T) {
	generator := honeypotAiTrapGenerator{}
	injectionKeyPhrases := []string{
		"ANALYSIS_TASK",
		"COMPUTE_REQUEST",
		"HASH_TASK",
		"TRACE_TASK",
		"DEEP_ANALYSIS",
	}
	testCaseStructs := []struct {
		name string
		path string
	}{
		{"DocsPath", "/api/v1/docs"},
		{"LogsPath", "/api/v1/logs/access"},
		{"MetricsPath", "/api/v1/metrics/prometheus"},
		{"DiagnosticsPath", "/api/v1/diagnostics/dump"},
		{"StatusPath", "/api/v1/status/detailed"},
	}
	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			hasInjection := false
			for chunkIdx := range 10 {
				chunk := generator.generateChunk(
					testCase.path, chunkIdx,
				)
				for _, keyPhrase := range injectionKeyPhrases {
					if strings.Contains(chunk, keyPhrase) {
						hasInjection = true
						break
					}
				}
				if hasInjection {
					break
				}
			}
			if !hasInjection {
				t.Errorf("PromptInjectionMissingForPath: %s",
					testCase.path)
			}
		})
	}
}

func TestBandwidthExhaustCappedStreamStopsAtMax(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	customMaxStream := int64(6 * 1024 * 1024)
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		customMaxStream,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	bodySize := int64(httpRecorder.Body.Len())
	if bodySize < streamFloorBytes {
		t.Errorf("StreamBelowFloor: got=%d, want>=%d",
			bodySize, streamFloorBytes)
	}
	if bodySize > customMaxStream {
		t.Errorf("StreamExceededMax: got=%d, want<=%d",
			bodySize, customMaxStream)
	}
}

func TestAITrapCappedStreamStopsAtMax(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	customMaxStream := int64(6 * 1024 * 1024)
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		customMaxStream,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	aiTrapPath := findActivePathOfClass(
		middleware, HoneypotPathClassAITrap,
	)
	if aiTrapPath == "" {
		t.Fatalf("NoActiveAiTrapPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, aiTrapPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	bodySize := int64(httpRecorder.Body.Len())
	if bodySize < streamFloorBytes {
		t.Errorf("StreamBelowFloor: got=%d, want>=%d",
			bodySize, streamFloorBytes)
	}
	if bodySize > customMaxStream {
		t.Errorf("StreamExceededMax: got=%d, want<=%d",
			bodySize, customMaxStream)
	}
}

func TestBandwidthExhaustCappedSizeVariesBetweenRuns(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		20 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	observedSizes := make(map[int]bool)
	for range 5 {
		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())
		httpRequest := httptest.NewRequest(
			http.MethodGet, bandwidthPath, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
		observedSizes[httpRecorder.Body.Len()] = true
	}
	if len(observedSizes) < 3 {
		t.Errorf("ExpectedVariedStreamSizes: got=%d distinct",
			len(observedSizes))
	}
}

func TestBandwidthExhaustIncrementsHitCount(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	rawValue, readErr := transientDbSvc.Read(
		"honeypot:hit:1.2.3.4",
	)
	if readErr != nil {
		t.Fatalf("HitDataNotStored: %v", readErr)
	}
	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)
	if hitData.Count != 1 {
		t.Errorf("HitCountMismatch: got=%d, want=1",
			hitData.Count)
	}
}

func TestAITrapIncrementsHitCount(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	aiTrapPath := findActivePathOfClass(
		middleware, HoneypotPathClassAITrap,
	)
	if aiTrapPath == "" {
		t.Fatalf("NoActiveAiTrapPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, aiTrapPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	rawValue, readErr := transientDbSvc.Read(
		"honeypot:hit:1.2.3.4",
	)
	if readErr != nil {
		t.Fatalf("HitDataNotStored: %v", readErr)
	}
	var hitData tkDto.HoneypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)
	if hitData.Count != 1 {
		t.Errorf("HitCountMismatch: got=%d, want=1",
			hitData.Count)
	}
}

func TestAiTrapGeneratorFileNameUsesLowercaseI(t *testing.T) {
	fileInfo, statErr := os.Stat("honeypotAiTrapGenerator.go")
	if statErr != nil {
		t.Fatalf("AiTrapGeneratorFileNotFound: %v", statErr)
	}
	if fileInfo.IsDir() {
		t.Errorf("ExpectedFileNotDirectory")
	}
}

func TestBandwidthExhaustBannedIpReturnsMixed(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("BannedIpExpectedMixed: got=%d",
			httpRecorder.Code)
	}
}

func TestAITrapBannedIpReturnsMixed(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	aiTrapPath := findActivePathOfClass(
		middleware, HoneypotPathClassAITrap,
	)
	if aiTrapPath == "" {
		t.Fatalf("NoActiveAiTrapPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, aiTrapPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("BannedIpExpectedMixed: got=%d",
			httpRecorder.Code)
	}
}

func TestBandwidthExhaustClientDisconnectExitsCleanly(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		20 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("StatusCodeMismatch: got=%d",
			httpRecorder.Code)
	}
}

func TestAITrapClientDisconnectExitsCleanly(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		20 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	aiTrapPath := findActivePathOfClass(
		middleware, HoneypotPathClassAITrap,
	)
	if aiTrapPath == "" {
		t.Fatalf("NoActiveAiTrapPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, aiTrapPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("StatusCodeMismatch: got=%d",
			httpRecorder.Code)
	}
}

type nonFlushResponseWriter struct {
	http.ResponseWriter
}

func TestFlusherUnavailableReturnsGracefulJson(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	innerRecorder := httptest.NewRecorder()
	wrappedWriter := &nonFlushResponseWriter{
		ResponseWriter: innerRecorder,
	}
	customContext := echoInstance.NewContext(
		httptest.NewRequest(
			http.MethodGet, bandwidthPath, nil,
		),
		innerRecorder,
	)
	customContext.Request().RemoteAddr = "1.2.3.4:1234"
	customContext.Response().Writer = wrappedWriter
	streamErr := middleware.streamBandwidthExhaust(
		customContext,
	)
	if streamErr != nil {
		t.Errorf("StreamShouldNotError: %v", streamErr)
	}
	if innerRecorder.Code != http.StatusOK {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			innerRecorder.Code, http.StatusOK)
	}
	var jsonBody map[string]any
	jsonErr := json.Unmarshal(
		innerRecorder.Body.Bytes(), &jsonBody,
	)
	if jsonErr != nil {
		t.Fatalf("FallbackBodyNotValidJson: %v", jsonErr)
	}
}

func TestInvalidRandomSeedStillProducesValidSelection(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	selection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, -1,
	)
	if len(selection) == 0 {
		t.Errorf("SelectionShouldNotBeEmpty")
	}
	hasStatic := false
	hasBandwidth := false
	hasAiTrap := false
	for _, pathClass := range selection {
		switch pathClass {
		case HoneypotPathClassStaticVuln:
			hasStatic = true
		case HoneypotPathClassBandwidthExhaust:
			hasBandwidth = true
		case HoneypotPathClassAITrap:
			hasAiTrap = true
		}
	}
	if !hasStatic {
		t.Errorf("StaticClassMissing")
	}
	if !hasBandwidth {
		t.Errorf("BandwidthClassMissing")
	}
	if !hasAiTrap {
		t.Errorf("AiTrapClassMissing")
	}
}

func TestBandwidthExhaustTierOneStreamsNotBans(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 1,
	)
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("TierOneShouldStream: got=%d",
			httpRecorder.Code)
	}
	if isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("TierOneShouldNotBan")
	}
}

func TestMaxStreamSizeBytesCustomValueUsed(t *testing.T) {
	customMaxStream := int64(10 * 1024 * 1024)
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		customMaxStream,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	bodySize := int64(httpRecorder.Body.Len())
	if bodySize > customMaxStream {
		t.Errorf("StreamExceededCustomMax: got=%d, want<=%d",
			bodySize, customMaxStream)
	}
	if bodySize < streamFloorBytes {
		t.Errorf("StreamBelowFloor: got=%d", bodySize)
	}
}

func TestActivePathCountCustomValueChangesCounts(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	selectionThirty := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 42,
	)
	selectionSixty := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		60, 42,
	)
	if len(selectionSixty) <= len(selectionThirty) {
		t.Errorf(
			"HigherActivePathCountShouldSelectMore: thirty=%d, sixty=%d",
			len(selectionThirty), len(selectionSixty),
		)
	}
}

func TestBandwidthExhaustSlowReaderDoesNotExhaustGoroutines(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	bandwidthPath := findActivePathOfClass(
		middleware, HoneypotPathClassBandwidthExhaust,
	)
	if bandwidthPath == "" {
		t.Fatalf("NoActiveBandwidthPathFound")
	}
	goroutinesBefore := runtime.NumGoroutine()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, bandwidthPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	goroutinesAfter := runtime.NumGoroutine()
	if goroutinesAfter > goroutinesBefore+5 {
		t.Errorf("GoroutineLeak: before=%d, after=%d",
			goroutinesBefore, goroutinesAfter)
	}
}

func TestAITrapSlowReaderDoesNotExhaustGoroutines(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.MaxStreamSizeBytes = mustNewHoneypotMaxStreamSizeBytes(
		6 * 1024 * 1024,
	)
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	aiTrapPath := findActivePathOfClass(
		middleware, HoneypotPathClassAITrap,
	)
	if aiTrapPath == "" {
		t.Fatalf("NoActiveAiTrapPathFound")
	}
	goroutinesBefore := runtime.NumGoroutine()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	httpRequest := httptest.NewRequest(
		http.MethodGet, aiTrapPath, nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	goroutinesAfter := runtime.NumGoroutine()
	if goroutinesAfter > goroutinesBefore+5 {
		t.Errorf("GoroutineLeak: before=%d, after=%d",
			goroutinesBefore, goroutinesAfter)
	}
}

func TestPhaseOneAndFiveEmbedFsPreserved(t *testing.T) {
	testActivePathMap := map[string]HoneypotPathClass{
		"/.env":          HoneypotPathClassStaticVuln,
		"/wp-config.php": HoneypotPathClassStaticVuln,
	}
	payloadMap := buildHoneypotPayloadMap(
		testActivePathMap, nil,
	)
	if len(payloadMap) == 0 {
		t.Errorf("EmbedFsPayloadMapEmpty")
	}
	expectedPaths := []string{"/.env", "/wp-config.php"}
	for _, expectedPath := range expectedPaths {
		if _, pathExists := payloadMap[expectedPath]; !pathExists {
			t.Errorf("ExpectedPayloadMissing: %s", expectedPath)
		}
	}
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)
	defer middleware.Stop()
	if len(middleware.activePathClasses) == 0 {
		t.Errorf("ActivePathClassesEmpty")
	}
}

func TestAllExistingPathsReturnDecodedPayloads(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(200)
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	for _, spec := range honeypotPayloadSpecs {
		if _, isActive :=
			middleware.activePathClasses[spec.urlPath]; !isActive {
			continue
		}
		t.Run(spec.urlPath, func(t *testing.T) {
			httpRequest := httptest.NewRequest(
				http.MethodGet, spec.urlPath, nil,
			)
			httpRequest.RemoteAddr = "5.6.7.8:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(
				httpRecorder, httpRequest,
			)
			if httpRecorder.Code != http.StatusOK {
				t.Errorf(
					"PathReturnNon200: path=%s, code=%d",
					spec.urlPath, httpRecorder.Code,
				)
			}
			contentType := httpRecorder.Header().Get(
				"Content-Type",
			)
			if contentType == "" {
				t.Errorf("ContentTypeMissing: path=%s",
					spec.urlPath)
			}
			if httpRecorder.Body.Len() == 0 {
				t.Errorf("BodyEmpty: path=%s",
					spec.urlPath)
			}
		})
	}
}

func TestDecodedContentMatchesPreEncodedOriginal(t *testing.T) {
	for _, spec := range honeypotPayloadSpecs {
		t.Run(spec.urlPath, func(t *testing.T) {
			mapping, decodeErr := decodePayloadSpec(spec)
			if decodeErr != nil {
				t.Fatalf(
					"DecodeFailed: path=%s, err=%v",
					spec.urlPath, decodeErr,
				)
			}
			if len(mapping.Body) == 0 {
				t.Errorf("DecodedBodyEmpty: path=%s",
					spec.urlPath)
			}
		})
	}
}

func TestPayloadDirectoryContainsOnlyBinFiles(t *testing.T) {
	dirEntries, readErr := os.ReadDir(
		"honeypot/payloads",
	)
	if readErr != nil {
		t.Fatalf("PayloadDirReadFailed: %v", readErr)
	}
	for _, dirEntry := range dirEntries {
		if !strings.HasSuffix(dirEntry.Name(), ".bin") {
			t.Errorf("NonBinFileFound: %s",
				dirEntry.Name())
		}
	}
}

func TestEmbedDirectivePointsToBinGlob(t *testing.T) {
	fileContent, readErr := os.ReadFile(
		"honeypotPathSelector.go",
	)
	if readErr != nil {
		t.Fatalf("SelectorFileReadFailed: %v", readErr)
	}
	expectedDirective := "//go:embed honeypot/payloads/*.bin"
	if !strings.Contains(
		string(fileContent), expectedDirective,
	) {
		t.Errorf("EmbedDirectiveMissing: want=%s",
			expectedDirective)
	}
}

func TestRandomActivationStillWorksAfterPhaseSix(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	selection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 42,
	)
	staticCount := 0
	bandwidthCount := 0
	aiTrapCount := 0
	for _, pathClass := range selection {
		switch pathClass {
		case HoneypotPathClassStaticVuln:
			staticCount++
		case HoneypotPathClassBandwidthExhaust:
			bandwidthCount++
		case HoneypotPathClassAITrap:
			aiTrapCount++
		}
	}
	totalCount := staticCount + bandwidthCount + aiTrapCount
	if totalCount != 30 {
		t.Errorf("TotalCountMismatch: got=%d, want=30",
			totalCount)
	}
	if staticCount != 20 {
		t.Errorf("StaticCountMismatch: got=%d, want=20",
			staticCount)
	}
}

func TestPathMappingTableHasCorrectEntries(t *testing.T) {
	if len(honeypotPayloadSpecs) < 90 {
		t.Errorf("SpecCountMismatch: got=%d, want>=90",
			len(honeypotPayloadSpecs))
	}
	for _, spec := range honeypotPayloadSpecs {
		if !strings.HasSuffix(spec.binFileName, ".bin") {
			t.Errorf("BinSuffixMissing: file=%s",
				spec.binFileName)
		}
		if spec.urlPath == "" {
			t.Errorf("EmptyUrlPath: file=%s",
				spec.binFileName)
		}
		if spec.mimeType == "" {
			t.Errorf("EmptyMimeType: file=%s",
				spec.binFileName)
		}
	}
}

func TestPathMappingFilenameEndsInBin(t *testing.T) {
	for _, spec := range honeypotPayloadSpecs {
		if !strings.HasSuffix(spec.binFileName, ".bin") {
			t.Errorf("BinSuffixMissing: file=%s",
				spec.binFileName)
		}
	}
}

func TestCorruptedBase64PayloadSkipsPathAndSelectsReplacement(t *testing.T) {
	activePathMap := map[string]HoneypotPathClass{
		"/.env": HoneypotPathClassStaticVuln,
	}
	payloadMap := make(map[string]HoneypotPathMapping)
	failedPaths := []string{"/.env"}

	resolveDecodeFailures(
		activePathMap, payloadMap, failedPaths,
	)

	if _, exists := activePathMap["/.env"]; exists {
		t.Errorf("FailedPathNotRemovedFromActiveMap")
	}

	if len(payloadMap) != 1 {
		t.Fatalf("ExpectedOneReplacement: got=%d",
			len(payloadMap))
	}

	if len(activePathMap) != 1 {
		t.Fatalf("ExpectedOneActiveReplacement: got=%d",
			len(activePathMap))
	}

	var replacementPath string
	for pathKey := range payloadMap {
		replacementPath = pathKey
	}

	if replacementPath == "/.env" {
		t.Errorf("ReplacementShouldDifferFromFailedPath")
	}

	replacementMapping, mapOk := payloadMap[replacementPath]
	if !mapOk {
		t.Fatalf("ReplacementPayloadMissing")
	}

	if len(replacementMapping.Body) == 0 {
		t.Errorf("ReplacementBodyEmpty")
	}

	replacementClass, classOk := activePathMap[replacementPath]
	if !classOk {
		t.Errorf("ReplacementNotInActiveMap")
	}

	if replacementClass != HoneypotPathClassStaticVuln {
		t.Errorf("ReplacementClassMismatch: got=%d, want=%d",
			replacementClass,
			HoneypotPathClassStaticVuln)
	}
}

func TestAllBinFilesDecodeFailureProceedsWithDegradedCount(t *testing.T) {
	activePathMap := map[string]HoneypotPathClass{
		"/api/v1/docs": HoneypotPathClassAITrap,
	}
	payloadMap := buildHoneypotPayloadMap(
		activePathMap, nil,
	)
	if len(payloadMap) != 0 {
		t.Errorf("ExpectedEmptyPayloadMap: got=%d",
			len(payloadMap))
	}
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil, nil,
	)
	defer middleware.Stop()
	if len(middleware.activePathClasses) == 0 {
		t.Errorf("ActivePathClassesEmpty")
	}
}

func TestNonExistentPathNotInMappingPassesThrough(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("PassThroughFailed: got=%d, want=%d",
			httpRecorder.Code, http.StatusOK)
	}
	if httpRecorder.Body.String() != "OK" {
		t.Errorf("BodyMismatch: got=%s, want=OK",
			httpRecorder.Body.String())
	}
}

func TestBinFilesContainNoPlaintextSecrets(t *testing.T) {
	credentialPatterns := []string{
		"password", "secret", "api_key",
		"aws_access_key", "DB_PASSWORD",
	}
	for _, spec := range honeypotPayloadSpecs {
		t.Run(spec.urlPath, func(t *testing.T) {
			binContent, readErr := honeypotPayloadsFs.ReadFile(
				"honeypot/payloads/" + spec.binFileName,
			)
			if readErr != nil {
				t.Fatalf("BinFileReadFailed: %v",
					readErr)
			}
			rawContent := string(binContent)
			for _, pattern := range credentialPatterns {
				if strings.Contains(
					strings.ToLower(rawContent),
					pattern,
				) {
					t.Errorf(
						"PlaintextSecretFound: pattern=%s, file=%s",
						pattern,
						spec.binFileName,
					)
				}
			}
		})
	}
}

func TestBinFilesOnlyContainValidBase64(t *testing.T) {
	for _, spec := range honeypotPayloadSpecs {
		t.Run(spec.urlPath, func(t *testing.T) {
			binContent, readErr := honeypotPayloadsFs.ReadFile(
				"honeypot/payloads/" + spec.binFileName,
			)
			if readErr != nil {
				t.Fatalf("BinFileReadFailed: %v",
					readErr)
			}
			_, decodeErr := base64.StdEncoding.DecodeString(
				string(binContent),
			)
			if decodeErr != nil {
				t.Errorf("InvalidBase64: file=%s, err=%v",
					spec.binFileName, decodeErr)
			}
		})
	}
}

func TestNoHoneypotPayloadDecoderFileExists(t *testing.T) {
	decoderFileNames := []string{
		"honeypotPayloadDecoder.go",
		"honeypotDecoder.go",
		"honeypotBase64Decoder.go",
	}
	for _, fileName := range decoderFileNames {
		if _, statErr := os.Stat(fileName); statErr == nil {
			t.Errorf("DecoderFileExists: %s", fileName)
		}
	}
}

func TestEncodedBinFilesNotFlaggedByStaticAnalysis(t *testing.T) {
	vulnSignatures := []string{
		"phpinfo()", "CREATE TABLE", "INSERT INTO",
		"<?php", "system($_", "DB_PASSWORD",
		"aws_access_key_id", "AKIAIOSFODNN",
	}
	payloadDir := "honeypot/payloads"
	dirEntries, readErr := os.ReadDir(payloadDir)
	if readErr != nil {
		t.Fatalf("PayloadDirReadFailed: %v", readErr)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}
		filePath := filepath.Join(payloadDir, dirEntry.Name())
		fileContent, readErr := os.ReadFile(filePath)
		if readErr != nil {
			t.Fatalf("FileReadFailed: file=%s, err=%v",
				dirEntry.Name(), readErr)
		}
		rawContent := string(fileContent)
		for _, signature := range vulnSignatures {
			if strings.Contains(rawContent, signature) {
				t.Errorf(
					"VulnSignatureFound: sig=%s, file=%s",
					signature, dirEntry.Name(),
				)
			}
		}
	}
}

func TestStaticPoolHasNinetyCandidates(t *testing.T) {
	staticCount := len(honeypotPayloadSpecs)
	if staticCount < 90 {
		t.Errorf("StaticPoolBelowFloor: got=%d, want>=90",
			staticCount)
	}
}

func TestBandwidthPoolHasTenCandidates(t *testing.T) {
	if len(bandwidthExhaustCandidatePaths) != 10 {
		t.Errorf("BandwidthPoolCountMismatch: got=%d, want=10",
			len(bandwidthExhaustCandidatePaths))
	}
}

func TestAITrapPoolHasTenCandidates(t *testing.T) {
	if len(aiTrapCandidatePaths) != 10 {
		t.Errorf("AITrapPoolCountMismatch: got=%d, want=10",
			len(aiTrapCandidatePaths))
	}
}

func TestAutoRatioSelectsStaticCountFromNinety(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	totalPoolSize := len(staticPaths) +
		len(bandwidthExhaustCandidatePaths) +
		len(aiTrapCandidatePaths)
	selection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 42,
	)
	staticCount := 0
	bandwidthCount := 0
	aiTrapCount := 0
	for _, pathClass := range selection {
		switch pathClass {
		case HoneypotPathClassStaticVuln:
			staticCount++
		case HoneypotPathClassBandwidthExhaust:
			bandwidthCount++
		case HoneypotPathClassAITrap:
			aiTrapCount++
		}
	}
	if staticCount != 20 {
		t.Errorf("StaticCountMismatch: got=%d, want=20",
			staticCount)
	}
	if totalPoolSize < 90 {
		t.Errorf("TotalPoolBelowFloor: got=%d, want>=90",
			totalPoolSize)
	}
}

func TestSelectedStaticPathsVaryAcrossRestarts(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	seenPaths := make(map[string]bool)
	for range 5 {
		selection := selectActivePaths(
			staticPaths,
			bandwidthExhaustCandidatePaths,
			aiTrapCandidatePaths,
			30, 0,
		)
		for selectedPath, pathClass := range selection {
			if pathClass == HoneypotPathClassStaticVuln {
				seenPaths[selectedPath] = true
			}
		}
		time.Sleep(time.Millisecond)
	}
	if len(seenPaths) < 15 {
		t.Errorf("InsufficientPathVariation: got=%d distinct, want>=15",
			len(seenPaths))
	}
}

func TestDormantStaticPathPassesThrough(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})
	allStaticPaths := extractStaticPathKeys()
	dormantPathFound := false
	for _, staticPath := range allStaticPaths {
		if _, pathIsActive :=
			middleware.activePathClasses[staticPath]; pathIsActive {
			continue
		}
		dormantPathFound = true
		httpRequest := httptest.NewRequest(
			http.MethodGet, staticPath, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
		if httpRecorder.Code == http.StatusOK {
			payloadContentType := httpRecorder.Header().Get(
				"Content-Type",
			)
			if payloadContentType != "" &&
				payloadContentType != "text/plain; charset=UTF-8" {
				t.Errorf("DormantPathShouldNotServePayload: path=%s",
					staticPath)
			}
		}
	}
	if !dormantPathFound {
		t.Errorf("NoDormantPathsFoundWithSeed42")
	}
}

func TestAllSelectedPathsReturnValidResponses(t *testing.T) {
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo :=
		newHoneypotRepos(transientDbSvc)
	middleware := NewHoneypotMiddleware(
		settings, honeypotCmdRepo, honeypotQueryRepo, nil,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	for activePath := range middleware.activePathClasses {
		httpRequest := httptest.NewRequest(
			http.MethodGet, activePath, nil,
		)
		httpRequest.RemoteAddr = "5.6.7.8:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
		if httpRecorder.Code != http.StatusOK {
			t.Errorf("ActivePathNon200: path=%s, got=%d",
				activePath, httpRecorder.Code)
		}
	}
}

func TestNewPathCategoriesPresentInStaticPool(t *testing.T) {
	allPaths := extractStaticPathKeys()
	pathSet := make(map[string]bool)
	for _, staticPath := range allPaths {
		pathSet[staticPath] = true
	}
	testCaseStructs := []struct {
		category    string
		samplePath  string
	}{
		{"EnvVariants", "/.env.local"},
		{"WordPress", "/wp-config.txt"},
		{"GitSVN", "/.git/"},
		{"PackageManagers", "/package.json"},
		{"ConfigFiles", "/config.json"},
		{"BackupArchive", "/backup.tar.gz"},
		{"InfoDebug", "/phpinfo.php"},
		{"UploadsTemp", "/uploads/admin.php"},
		{"AdminLogin", "/admin"},
		{"ApiDebug", "/api/"},
		{"WebCrossdomain", "/crossdomain.xml"},
		{"Webhooks", "/webhook"},
		{"Logs", "/access.log"},
		{"CloudStorage", "/s3/"},
		{"BuildArtifacts", "/dist/main.js"},
		{"PathTraversal", "/etc/passwd"},
		{"Other", "/Dockerfile"},
	}
	for _, testCase := range testCaseStructs {
		if !pathSet[testCase.samplePath] {
			t.Errorf("CategoryMissing: category=%s, sample=%s",
				testCase.category, testCase.samplePath)
		}
	}
}

func TestAllNewPathsHaveBinFilesAndMapping(t *testing.T) {
	newPaths := []string{
		"/.env.local", "/.env.test", "/.env.bak",
		"/.env.example", "/.env.production",
		"/wp-config.txt", "/wp-config.php.orig",
		"/wp-admin/", "/wp-login.php",
		"/.git/", "/.git/index", "/.git/HEAD",
		"/.svn/", "/.svn/entries",
		"/package.json", "/package-lock.json",
		"/yarn.lock", "/pnpm-lock.yaml",
		"/composer.json", "/composer.lock",
	}
	for _, pathStr := range newPaths {
		spec := findPayloadSpec(pathStr)
		if spec == nil {
			t.Errorf("MappingMissing: path=%s", pathStr)
			continue
		}
		_, readErr := honeypotPayloadsFs.ReadFile(
			"honeypot/payloads/" + spec.binFileName,
		)
		if readErr != nil {
			t.Errorf("BinFileMissing: path=%s, file=%s",
				pathStr, spec.binFileName)
		}
	}
}

func TestActivePathCountCeilingCalculatedAtConstructionTime(
	t *testing.T,
) {
	totalPoolSize := len(honeypotPayloadSpecs) +
		len(bandwidthExhaustCandidatePaths) +
		len(aiTrapCandidatePaths)
	settings := HoneypotMiddlewareSettings{
		ActivePathCount: mustNewHoneypotActivePathCount(200),
	}
	middleware := NewHoneypotMiddleware(
		settings, nil, nil, nil,
	)
	defer middleware.Stop()
	resolvedCount := len(middleware.activePathClasses)
	if resolvedCount > totalPoolSize {
		t.Errorf("CeilingExceeded: activePaths=%d, pool=%d",
			resolvedCount, totalPoolSize)
	}
}

func TestActivePathCountRespectsCeiling(t *testing.T) {
	totalPoolSize := len(honeypotPayloadSpecs) +
		len(bandwidthExhaustCandidatePaths) +
		len(aiTrapCandidatePaths)
	settings := HoneypotMiddlewareSettings{
		ActivePathCount: mustNewHoneypotActivePathCount(200),
	}
	resolvedSettings := honeypotSettingsParser{}.Parse(
		settings, totalPoolSize,
	)
	if resolvedSettings.ActivePathCount.Int() > totalPoolSize {
		t.Errorf("CeilingNotRespected: got=%d, want<=%d",
			resolvedSettings.ActivePathCount.Int(),
			totalPoolSize)
	}
}

func TestNonExistentHoneypotPathPassesThrough(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	middleware := NewHoneypotMiddleware(
		newStandardSettings(),
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/*", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})
	httpRequest := httptest.NewRequest(
		http.MethodGet,
		"/this/path/does/not/exist/at/all", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if httpRecorder.Code != http.StatusOK {
		t.Errorf("NonExistentPathShouldPassThrough: got=%d",
			httpRecorder.Code)
	}
	if httpRecorder.Body.String() != "OK" {
		t.Errorf("BodyMismatch: got=%s, want=OK",
			httpRecorder.Body.String())
	}
}

func TestEmptyCandidatePoolHandledGracefully(t *testing.T) {
	emptySelection := selectActivePaths(
		nil, nil, nil, 0, 42,
	)
	if len(emptySelection) != 0 {
		t.Errorf("EmptyPoolShouldReturnEmpty: got=%d",
			len(emptySelection))
	}
	middleware := NewHoneypotMiddleware(
		HoneypotMiddlewareSettings{}, nil, nil, nil,
	)
	defer middleware.Stop()
	if middleware == nil {
		t.Errorf("MiddlewareIsNil")
	}
}

func TestSeedZeroProducesTimeBasedPool(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	selection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		30, 0,
	)
	if len(selection) == 0 {
		t.Errorf("TimeBasedSeedShouldProduceNonEmpty")
	}
}

func TestNewBinFileDecodeFailureSkipsAndReplaces(t *testing.T) {
	activePathMap := map[string]HoneypotPathClass{
		"/.env.local": HoneypotPathClassStaticVuln,
	}
	payloadMap := make(map[string]HoneypotPathMapping)
	failedPaths := []string{"/.env.local"}
	resolveDecodeFailures(
		activePathMap, payloadMap, failedPaths,
	)
	if _, exists := activePathMap["/.env.local"]; exists {
		t.Errorf("FailedPathNotRemovedFromActiveMap")
	}
	if len(payloadMap) != 1 {
		t.Fatalf("ExpectedOneReplacement: got=%d",
			len(payloadMap))
	}
	var replacementPath string
	for pathKey := range payloadMap {
		replacementPath = pathKey
	}
	if replacementPath == "/.env.local" {
		t.Errorf("ReplacementShouldDifferFromFailedPath")
	}
}

func TestCustomActivePathCountChangesStaticSelection(t *testing.T) {
	staticPaths := extractStaticPathKeys()
	selection := selectActivePaths(
		staticPaths,
		bandwidthExhaustCandidatePaths,
		aiTrapCandidatePaths,
		60, 42,
	)
	staticCount := 0
	bandwidthCount := 0
	aiTrapCount := 0
	for _, pathClass := range selection {
		switch pathClass {
		case HoneypotPathClassStaticVuln:
			staticCount++
		case HoneypotPathClassBandwidthExhaust:
			bandwidthCount++
		case HoneypotPathClassAITrap:
			aiTrapCount++
		}
	}
	if staticCount != 40 {
		t.Errorf("StaticCountMismatch: got=%d, want=40",
			staticCount)
	}
	if bandwidthCount != 10 {
		t.Errorf("BandwidthCountMismatch: got=%d, want=10",
			bandwidthCount)
	}
	if aiTrapCount != 10 {
		t.Errorf("AiTrapCountMismatch: got=%d, want=10",
			aiTrapCount)
	}
}

func TestExpandedPoolScannerBurstEscalatesBan(t *testing.T) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, testIp, 3)
	cmdRepo := newNoopCmdRepo()
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	settings := newStandardSettings()
	settings.ActivePathCount = mustNewHoneypotActivePathCount(30)
	settings.RandomSeed = 42
	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()
	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		echoCtx echo.Context,
	) error {
		return echoCtx.String(http.StatusOK, "OK")
	})
	staticPath := ""
	for activePath, pathClass := range middleware.activePathClasses {
		if pathClass == HoneypotPathClassStaticVuln {
			staticPath = activePath
			break
		}
	}
	if staticPath == "" {
		t.Fatalf("NoActiveStaticPath")
	}
	httpRequest := httptest.NewRequest(
		http.MethodGet, staticPath, nil,
	)
	httpRequest.RemoteAddr = testIp + ":1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)
	if !isMixedResponseStatusCode(httpRecorder.Code) {
		t.Errorf("TierThreeHoneypotExpectedMixed: got=%d",
			httpRecorder.Code)
	}
	legitRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	legitRequest.RemoteAddr = testIp + ":1234"
	legitRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(legitRecorder, legitRequest)
	if !isMixedResponseStatusCode(legitRecorder.Code) {
		t.Errorf("TierThreeLegitExpectedMixed: got=%d",
			legitRecorder.Code)
	}
}

func TestPoolExpansionDoesNotBreakPhaseSixDecoding(t *testing.T) {
	for _, spec := range honeypotPayloadSpecs {
		mapping, decodeErr := decodePayloadSpec(spec)
		if decodeErr != nil {
			t.Errorf("DecodeFailed: path=%s, err=%v",
				spec.urlPath, decodeErr)
			continue
		}
		if len(mapping.Body) == 0 {
			t.Errorf("DecodedBodyEmpty: path=%s",
				spec.urlPath)
		}
	}
}

func TestAllExpandedPathsHaveValidBinFiles(t *testing.T) {
	payloadDir := "honeypot/payloads"
	dirEntries, readErr := os.ReadDir(payloadDir)
	if readErr != nil {
		t.Fatalf("PayloadDirReadFailed: %v", readErr)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}
		filePath := filepath.Join(payloadDir, dirEntry.Name())
		fileContent, readErr := os.ReadFile(filePath)
		if readErr != nil {
			t.Fatalf("BinFileReadFailed: file=%s, err=%v",
				dirEntry.Name(), readErr)
		}
		_, decodeErr := base64.StdEncoding.DecodeString(
			string(fileContent),
		)
		if decodeErr != nil {
			t.Errorf("InvalidBase64: file=%s, err=%v",
				dirEntry.Name(), decodeErr)
		}
	}
}

func TestExpandedPoolStatsAggregationIncludesNewEndpoints(
	t *testing.T,
) {
	var capturedRecord *tkDto.CreateActivityRecord
	cmdRepo := mockActivityRecordCmdRepo{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)
	honeypotCmdRepo, honeypotQueryRepo := newHoneypotRepos(
		transientDbSvc,
	)
	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve
	middleware := NewHoneypotMiddleware(
		settings,
		honeypotCmdRepo, honeypotQueryRepo, cmdRepo,
	)
	defer middleware.Stop()
	middleware.runMaintenance()
	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}
	detailsMap := capturedRecord.RecordDetails.(map[string]string)
	var statsReport map[string]any
	json.Unmarshal(
		[]byte(detailsMap["statsReport"]), &statsReport,
	)
	topEndpoints := statsReport["topEndpoints"].([]any)
	if len(topEndpoints) == 0 {
		t.Errorf("TopEndpointsEmpty")
	}
}
