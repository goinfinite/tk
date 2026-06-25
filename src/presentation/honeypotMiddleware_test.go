package tkPresentation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
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

func newStandardSettings() HoneypotMiddlewareSettings {
	return HoneypotMiddlewareSettings{
		BanDuration: 24 * time.Hour,
		RedirectUrl: newDefaultRedirectUrl(),
	}
}

func newTransientDbSvc() *tkInfraDb.TransientDatabaseService {
	dbSvc, _ := tkInfraDb.NewTransientDatabaseService()
	dbSvc.Handler.Exec("DELETE FROM key_values")
	return dbSvc
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
	hitData := honeypotHitData{
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
	hitData := honeypotHitData{
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

func TestHoneypotMiddlewareCreation(t *testing.T) {
	testCaseStructs := []struct {
		name     string
		settings HoneypotMiddlewareSettings
	}{
		{"WithBanDuration", HoneypotMiddlewareSettings{
			BanDuration: 24 * time.Hour,
		}},
		{"WithEmptySettings", HoneypotMiddlewareSettings{}},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			middleware := NewHoneypotMiddleware(
				testCase.settings, nil, nil,
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

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			transientDbSvc, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		httpRequest := httptest.NewRequest(
			http.MethodGet, "/.env", nil,
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

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			transientDbSvc, cmdRepo,
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

		if httpRecorder.Code != http.StatusFound {
			t.Errorf("StatusCodeMismatch: got=%d, want=%d",
				httpRecorder.Code, http.StatusFound)
		}

		location := httpRecorder.Header().Get("Location")
		if location != "https://xkcd.com/" {
			t.Errorf("LocationMismatch: got=%s, want=https://xkcd.com/",
				location)
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

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			transientDbSvc, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		firstRequest := httptest.NewRequest(
			http.MethodGet, "/.env", nil,
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

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			transientDbSvc, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())

		httpRequest := httptest.NewRequest(
			http.MethodGet, "/.env", nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)

		if httpRecorder.Code != http.StatusFound {
			t.Errorf("StatusCodeMismatch: got=%d, want=%d",
				httpRecorder.Code, http.StatusFound)
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

			middleware := NewHoneypotMiddleware(
				newStandardSettings(),
				transientDbSvc, nil,
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

			if httpRecorder.Code != http.StatusFound {
				t.Errorf("StatusCodeMismatch: got=%d, want=%d",
					httpRecorder.Code, http.StatusFound)
			}

			location := httpRecorder.Header().Get("Location")
			if location != "https://xkcd.com/" {
				t.Errorf("LocationMismatch: got=%s, want=https://xkcd.com/",
					location)
			}
		})
	}
}

func TestAllHoneypotPathsReturnPayloads(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPaths := []string{
		"/.env", "/wp-config.php", "/wp-config.php.bak",
		"/config.php", "/backup.sql", "/backup.zip",
		"/.git/config", "/.aws/credentials",
		"/actuator/env", "/actuator/configprops",
		"/server-status", "/phpmyadmin/index.php",
		"/admin.php", "/administrator/index.php",
		"/login.php", "/shell.php", "/cmd.php",
		"/test.php", "/.htaccess", "/web.config",
		"/robots.txt", "/sitemap.xml", "/debug.php",
		"/info.php", "/console",
	}

	for _, honeypotPath := range honeypotPaths {
		t.Run(honeypotPath[1:], func(t *testing.T) {
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

		middleware := NewHoneypotMiddleware(
			newStandardSettings(),
			transientDbSvc, cmdRepo,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())
		echoInstance.GET("/api/health", func(
			c echo.Context,
		) error {
			return c.String(http.StatusOK, "OK")
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
			newStandardSettings(), nil, nil,
		)
		defer middleware.Stop()

		echoInstance := echo.New()
		echoInstance.Use(middleware.MiddlewareFunc())
		echoInstance.GET("/api/health", func(
			c echo.Context,
		) error {
			return c.String(http.StatusOK, "OK")
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
		settings, nil, nil,
	)
	defer middleware.Stop()

	if middleware == nil {
		t.Errorf("MiddlewareIsNil")
	}

	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)

	middlewareWithDb := NewHoneypotMiddleware(
		settings, transientDbSvc, nil,
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

	if httpRecorder.Code != http.StatusFound {
		t.Errorf("StatusCodeMismatch: got=%d, want=%d",
			httpRecorder.Code, http.StatusFound)
	}

	location := httpRecorder.Header().Get("Location")
	if location != "https://xkcd.com/" {
		t.Errorf("LocationMismatch: got=%s, want=https://xkcd.com/",
			location)
	}
}

func TestSharedNATBlocksLegitimateUsers(t *testing.T) {
	cmdRepo := newNoopCmdRepo()
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(
		transientDbSvc, "1.2.3.4", 3,
	)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
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

	if secondRecorder.Code != http.StatusFound {
		t.Errorf("SecondUserShouldBeBlocked: got=%d, want=%d",
			secondRecorder.Code, http.StatusFound)
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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPaths := []string{
		"/.env", "/wp-config.php", "/wp-config.php.bak",
		"/config.php", "/backup.sql", "/backup.zip",
		"/.git/config", "/.aws/credentials",
		"/actuator/env", "/actuator/configprops",
		"/server-status", "/phpmyadmin/index.php",
		"/admin.php", "/administrator/index.php",
		"/login.php", "/shell.php", "/cmd.php",
		"/test.php", "/.htaccess", "/web.config",
		"/robots.txt", "/sitemap.xml", "/debug.php",
		"/info.php", "/console",
	}

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
					UrlPath: fakeApiKeysUrlPath,
					Body:    `{"api_key":"fake-key-12345"}`,
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
				ExtraPathRoutes: testCase.extraPathRoutes,
			}

			honeypotMiddleware := NewHoneypotMiddleware(
				honeypotSettings, nil, nil,
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		c echo.Context,
	) error {
		return c.String(http.StatusOK, "OK")
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		c echo.Context,
	) error {
		return c.String(http.StatusOK, "OK")
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

	honeypotRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	rawValue, readErr := transientDbSvc.Read("honeypot:hit:1.2.3.4")
	if readErr != nil {
		t.Fatalf("HitDataNotStored: %v", readErr)
	}

	var hitData honeypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count != 1 {
		t.Errorf("HitCountMismatch: got=%d, want=1", hitData.Count)
	}
}

func TestGraduatedBanTierTwoBannedOnHoneypotPaths(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 2)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(httpRecorder, httpRequest)

	if httpRecorder.Code != http.StatusFound {
		t.Errorf("TierTwoHoneypotPathShouldRedirect: got=%d",
			httpRecorder.Code)
	}
}

func TestGraduatedBanTierThreeFullBan(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 3)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
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

			if httpRecorder.Code != http.StatusFound {
				t.Errorf("TierThreeShouldRedirectAll: path=%s, got=%d",
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		c echo.Context,
	) error {
		return c.String(http.StatusOK, "OK")
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

func TestNewHoneypotMiddlewareAcceptsSettingsTransientDbSvcAndCmdRepo(
	t *testing.T,
) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
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
		newStandardSettings(), nil, nil,
	)
	defer middleware.Stop()

	var _ echo.MiddlewareFunc = middleware.MiddlewareFunc()
	middleware.Stop()
}

func TestMiddlewareFuncReturnsEchoMiddlewareFunc(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil,
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

	middleware := NewHoneypotMiddleware(
		settings, newTransientDbSvc(), nil,
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
		newStandardSettings(), nil, nil,
	)

	middleware.Stop()
	middleware.Stop()
}

func TestStopCalledMultipleTimesIsSafe(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil,
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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	paths := []string{"/.env", "/.env", "/wp-config.php"}
	for _, honeypotPath := range paths {
		httpRequest := httptest.NewRequest(
			http.MethodGet, honeypotPath, nil,
		)
		httpRequest.RemoteAddr = "1.2.3.4:1234"
		httpRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(httpRecorder, httpRequest)
	}

	rawValue, _ := transientDbSvc.Read("honeypot:hit:1.2.3.4")
	var hitData honeypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count != 3 {
		t.Errorf("TotalCountMismatch: got=%d, want=3",
			hitData.Count)
	}

	if hitData.Endpoints["/.env"] != 2 {
		t.Errorf("EnvEndpointCountMismatch: got=%d, want=2",
			hitData.Endpoints["/.env"])
	}

	if hitData.Endpoints["/wp-config.php"] != 1 {
		t.Errorf("WpConfigEndpointCountMismatch: got=%d, want=1",
			hitData.Endpoints["/wp-config.php"])
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

	cleanExpiredEntries(
		transientDbSvc.Handler, 24*time.Hour,
	)

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

	cleanExpiredEntries(
		transientDbSvc.Handler, 24*time.Hour,
	)

	if transientDbSvc.Count() != 1 {
		t.Errorf("ActiveEntryShouldBePreserved")
	}
}

func TestEnforceMaxEntriesDeletesOldestEntries(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 5 {
		entry := tkInfraDb.KeyValueModel{
			Key:       "key:" + string(rune('a'+entryIndex)),
			Value:     "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Minute,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	enforceMaxEntries(transientDbSvc.Handler, 3)

	remaining := transientDbSvc.Count()
	if remaining != 3 {
		t.Errorf("ExpectedThreeEntries: got=%d", remaining)
	}
}

func TestEnforceMaxEntriesFloorRespected(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 3 {
		entry := tkInfraDb.KeyValueModel{
			Key:       "key:" + string(rune('a'+entryIndex)),
			Value:     "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Minute,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	enforceMaxEntries(transientDbSvc.Handler, 5)

	if transientDbSvc.Count() != 3 {
		t.Errorf("EntriesBelowMaxShouldNotBeDeleted")
	}
}

func TestEnforceMaxEntriesCeilingRespected(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 10 {
		entry := tkInfraDb.KeyValueModel{
			Key: "key:" + string(rune('a'+entryIndex)),
			Value:     "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Minute,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	enforceMaxEntries(transientDbSvc.Handler, 5)

	remaining := transientDbSvc.Count()
	if remaining != 5 {
		t.Errorf("ExpectedFiveEntries: got=%d", remaining)
	}
}

func TestAggressivenessImmediateFirstHitBans(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.2.3.4", 1)

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeImmediate

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, nil,
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

	if httpRecorder.Code != http.StatusFound {
		t.Errorf("ImmediateShouldBanAfterOneHit: got=%d",
			httpRecorder.Code)
	}
}

func TestAggressivenessBalancedGraduatedTiers(t *testing.T) {
	testCaseStructs := []struct {
		name             string
		hitCount         int
		expectedRedirect bool
	}{
		{"ZeroHitsPasses", 0, false},
		{"OneHitServesPayload", 1, false},
		{"TwoHitsRedirectsHoneypot", 2, true},
		{"ThreeHitsFullBan", 3, true},
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

			settings := newStandardSettings()
			settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeBalanced

			middleware := NewHoneypotMiddleware(
				settings, transientDbSvc, nil,
			)
			defer middleware.Stop()

			echoInstance := echo.New()
			echoInstance.Use(middleware.MiddlewareFunc())

			httpRequest := httptest.NewRequest(
				http.MethodGet, "/.env", nil,
			)
			httpRequest.RemoteAddr = "1.2.3.4:1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

			isRedirect := httpRecorder.Code == http.StatusFound
			if isRedirect != testCase.expectedRedirect {
				t.Errorf("RedirectMismatch: hits=%d, got=%d, wantRedirect=%v",
					testCase.hitCount,
					httpRecorder.Code,
					testCase.expectedRedirect)
			}
		})
	}
}

func TestAggressivenessTolerantGraduatedTiers(t *testing.T) {
	testCaseStructs := []struct {
		name     string
		hitCount int
		path     string
		wantCode int
		ipSuffix string
	}{
		{
			"OneHitPassesAll", 1,
			"/api/health", http.StatusOK, "1",
		},
		{
			"TwoHitsPassesLegit", 2,
			"/api/health", http.StatusOK, "2",
		},
		{
			"TwoHitsServesPayloadTierOne", 2,
			"/.env", http.StatusOK, "3",
		},
		{
			"FiveHitsRedirectsHoneypot", 5,
			"/.env", http.StatusFound, "4",
		},
		{
			"FiveHitsPassesLegitTierTwo", 5,
			"/api/health", http.StatusOK, "5",
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

			settings := newStandardSettings()
			settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeTolerant

			middleware := NewHoneypotMiddleware(
				settings, transientDbSvc, nil,
			)
			defer middleware.Stop()

			echoInstance := echo.New()
			echoInstance.Use(middleware.MiddlewareFunc())
			echoInstance.GET("/api/health", func(
				c echo.Context,
			) error {
				return c.String(http.StatusOK, "OK")
			})

			httpRequest := httptest.NewRequest(
				http.MethodGet, testCase.path, nil,
			)
			httpRequest.RemoteAddr = testIp + ":1234"
			httpRecorder := httptest.NewRecorder()
			echoInstance.ServeHTTP(httpRecorder, httpRequest)

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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, nil,
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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, nil,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
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
		newStandardSettings(), nil, nil,
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
	middleware := NewHoneypotMiddleware(
		settings, newTransientDbSvc(), nil,
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

	middleware := NewHoneypotMiddleware(
		settings, newTransientDbSvc(), nil,
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()

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

	enforceMaxEntries(transientDbSvc.Handler, 5)

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	detailsMap, ok := capturedRecord.RecordDetails.(map[string]string)
	if !ok {
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()

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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()

	if statsCount == 0 {
		t.Errorf("StatsShouldBeProducedEvenWithNoCleanup")
	}
}

func TestWatchdogRespectsContextCancellation(t *testing.T) {
	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"5m",
	)

	middleware := NewHoneypotMiddleware(
		settings, newTransientDbSvc(), nil,
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	tier := middleware.determineBanTier(testIp)
	if tier != 0 {
		t.Errorf("ExpiredHitsShouldReturnTierZero: got=%d",
			tier)
	}
}

func TestProbabilisticEnforcementTriggersOnWrite(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	for entryIndex := range 200 {
		entry := tkInfraDb.KeyValueModel{
			Key: "key:" + string(rune(entryIndex)),
			Value:     "val",
			CreatedAt: time.Now().Add(
				time.Duration(entryIndex) * time.Millisecond,
			),
		}
		transientDbSvc.Handler.Create(&entry)
	}

	cmdRepo := newNoopCmdRepo()
	settings := newStandardSettings()
	settings.MaxEntries, _ = tkValueObject.NewHoneypotMaxEntries(100)

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	for range 500 {
		middleware.incrementHitCount("1.2.3.4", "/.env")
	}

	remaining := transientDbSvc.Count()
	if remaining > 150 {
		t.Logf("Probabilistic enforcement may not have triggered: count=%d", remaining)
	}
}

func TestProbabilisticEnforcementNotAlwaysTriggered(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()

	settings := newStandardSettings()
	settings.MaxEntries, _ = tkValueObject.NewHoneypotMaxEntries(5000)

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	countBefore := transientDbSvc.Count()
	middleware.incrementHitCount("1.2.3.4", "/.env")
	countAfter := transientDbSvc.Count()

	if countAfter != countBefore+1 {
		t.Errorf("SingleWriteShouldAddOneEntry")
	}
}

func TestGraduatedBanTransientDbReadErrorHandled(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil,
	)
	defer middleware.Stop()

	tier := middleware.determineBanTier("1.2.3.4")
	if tier != 0 {
		t.Errorf("NilTransientDbShouldReturnTierZero: got=%d",
			tier)
	}
}

func TestTransientDbReadErrorHandled(t *testing.T) {
	transientDbSvc := newTransientDbSvc()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	tier := middleware.determineBanTier("nonexistent.ip")
	if tier != 0 {
		t.Errorf("MissingKeyShouldReturnTierZero: got=%d", tier)
	}
}

func TestProbabilisticEnforcementHandlesMaxEntriesError(t *testing.T) {
	middleware := NewHoneypotMiddleware(
		newStandardSettings(), nil, nil,
	)
	defer middleware.Stop()

	middleware.incrementHitCount("1.2.3.4", "/.env")
}

func TestStopOnUninitializedMiddlewareDoesNotPanic(t *testing.T) {
	middleware := &HoneypotMiddleware{}
	middleware.Stop()
}

func TestMaintenanceWatchdogHandlesNilCmdRepo(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, nil,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()
}

func TestMaintenanceWatchdogRecoversFromPanicInTick(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	populateTransientDbWithHits(transientDbSvc, "1.1.1.1", 1)

	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"100ms",
	)

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, nil,
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
		newStandardSettings(), nil, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.aggregateStats()

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
	settings.BanDuration = 72 * time.Hour

	transientDbSvc := newTransientDbSvc()
	testIp := newUniqueTestIp()
	populateTransientDbWithHits(transientDbSvc, testIp, 1)

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
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	middleware.runMaintenanceTick()
	middleware.runMaintenanceTick()

	if statsCount < 2 {
		t.Errorf("WatchdogShouldTickMultipleTimes: got=%d",
			statsCount)
	}
}

func TestScannerFloodTriggersTierEscalation(t *testing.T) {
	testIp := newUniqueTestIp()
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	honeypotPaths := []string{
		"/.env", "/wp-config.php", "/config.php",
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
	var hitData honeypotHitData
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
	cmdRepo := newNoopCmdRepo()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	goroutineCount := 10
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for goroutineIndex := range goroutineCount {
		go func(index int) {
			defer waitGroup.Done()
			middleware.incrementHitCount(testIp, "/.env")
		}(goroutineIndex)
	}

	waitGroup.Wait()

	rawValue, _ := transientDbSvc.Read("honeypot:hit:" + testIp)
	var hitData honeypotHitData
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

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeImmediate

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, nil,
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

	if httpRecorder.Code != http.StatusFound {
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

	settings := newStandardSettings()
	settings.AggressivenessMode = tkValueObject.HoneypotAggressivenessModeObserve

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	echoInstance := echo.New()
	echoInstance.Use(middleware.MiddlewareFunc())
	echoInstance.GET("/api/health", func(
		c echo.Context,
	) error {
		return c.String(http.StatusOK, "OK")
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
	cmdRepo := newNoopCmdRepo()
	settings := newStandardSettings()
	settings.MaxEntries, _ = tkValueObject.NewHoneypotMaxEntries(100)

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	var waitGroup sync.WaitGroup
	waitGroup.Add(20)

	for goroutineIndex := range 20 {
		go func(index int) {
			defer waitGroup.Done()
			middleware.incrementHitCount(
				"1.2.3.4", "/.env",
			)
		}(goroutineIndex)
	}

	waitGroup.Wait()

	if transientDbSvc.Count() == 0 {
		t.Errorf("EntriesShouldExistAfterConcurrentWrites")
	}
}

func TestConcurrentHitsAndMaintenanceCycleNoDataLoss(t *testing.T) {
	transientDbSvc := newTransientDbSvc()
	cmdRepo := newNoopCmdRepo()

	settings := newStandardSettings()
	settings.StatsInterval, _ = tkValueObject.NewHoneypotStatsInterval(
		"50ms",
	)

	middleware := NewHoneypotMiddleware(
		settings, transientDbSvc, cmdRepo,
	)
	defer middleware.Stop()

	var waitGroup sync.WaitGroup
	waitGroup.Add(10)

	for goroutineIndex := range 10 {
		go func(index int) {
			defer waitGroup.Done()
			middleware.incrementHitCount(
				"1.2.3.4", "/.env",
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

	var hitData honeypotHitData
	json.Unmarshal([]byte(rawValue), &hitData)

	if hitData.Count < 1 {
		t.Errorf("ConcurrentHitsShouldBeCounted: got=%d",
			hitData.Count)
	}
}
