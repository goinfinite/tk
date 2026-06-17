package tkPresentation

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

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

type mockActivityRecordQueryRepo struct {
	readFunc      func(tkDto.ReadActivityRecordsRequest) (tkDto.ReadActivityRecordsResponse, error)
	readFirstFunc func(tkDto.ReadActivityRecordsRequest) (tkEntity.ActivityRecord, error)
}

func (mockQueryRepo mockActivityRecordQueryRepo) Read(
	req tkDto.ReadActivityRecordsRequest,
) (tkDto.ReadActivityRecordsResponse, error) {
	return mockQueryRepo.readFunc(req)
}

func (mockQueryRepo mockActivityRecordQueryRepo) ReadFirst(
	req tkDto.ReadActivityRecordsRequest,
) (tkEntity.ActivityRecord, error) {
	return mockQueryRepo.readFirstFunc(req)
}

func newDefaultRedirectUrl() tkValueObject.Url {
	defaultUrl, _ := tkValueObject.NewUrl("https://xkcd.com/")
	return defaultUrl
}

func newAlwaysEmptyQueryRepo() mockActivityRecordQueryRepo {
	return mockActivityRecordQueryRepo{
		readFunc: func(
			req tkDto.ReadActivityRecordsRequest,
		) (tkDto.ReadActivityRecordsResponse, error) {
			return tkDto.ReadActivityRecordsResponse{}, nil
		},
		readFirstFunc: func(
			req tkDto.ReadActivityRecordsRequest,
		) (tkEntity.ActivityRecord, error) {
			return tkEntity.ActivityRecord{}, nil
		},
	}
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

func newBannedIpQueryRepo() mockActivityRecordQueryRepo {
	return mockActivityRecordQueryRepo{
		readFunc: func(
			req tkDto.ReadActivityRecordsRequest,
		) (tkDto.ReadActivityRecordsResponse, error) {
			return tkDto.ReadActivityRecordsResponse{}, nil
		},
		readFirstFunc: func(
			req tkDto.ReadActivityRecordsRequest,
		) (tkEntity.ActivityRecord, error) {
			if req.RecordCode != nil &&
				req.RecordCode.String() == "HoneypotHit" {
				return tkEntity.ActivityRecord{
					RecordId:   tkValueObject.ActivityRecordId(1),
					RecordCode: tkValueObject.ActivityRecordCode("HoneypotHit"),
				}, nil
			}
			return tkEntity.ActivityRecord{}, nil
		},
	}
}

func newStandardSettings() HoneypotMiddlewareSettings {
	return HoneypotMiddlewareSettings{
		BanDuration: 24 * time.Hour,
		RedirectUrl: newDefaultRedirectUrl(),
	}
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
		queryRepo := newAlwaysEmptyQueryRepo()

		middleware := NewHoneypotMiddleware(
			newStandardSettings(), cmdRepo, queryRepo,
		)
		echoInstance := echo.New()
		echoInstance.Use(middleware)

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
		queryRepo := newBannedIpQueryRepo()

		middleware := NewHoneypotMiddleware(
			newStandardSettings(), cmdRepo, queryRepo,
		)
		echoInstance := echo.New()
		echoInstance.Use(middleware)

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
		banCheckQueryCount := 0

		cmdRepo := mockActivityRecordCmdRepo{
			createFunc: func(
				dto tkDto.CreateActivityRecord,
			) error {
				honeypotBanRecordCreated = true
				return nil
			},
		}

		queryRepo := mockActivityRecordQueryRepo{
			readFunc: func(
				req tkDto.ReadActivityRecordsRequest,
			) (tkDto.ReadActivityRecordsResponse, error) {
				return tkDto.ReadActivityRecordsResponse{}, nil
			},
			readFirstFunc: func(
				req tkDto.ReadActivityRecordsRequest,
			) (tkEntity.ActivityRecord, error) {
				banCheckQueryCount++
				if banCheckQueryCount == 1 {
					return tkEntity.ActivityRecord{}, nil
				}
				return tkEntity.ActivityRecord{
					RecordId:   tkValueObject.ActivityRecordId(1),
					RecordCode: tkValueObject.ActivityRecordCode("HoneypotHit"),
				}, nil
			},
		}

		middleware := NewHoneypotMiddleware(
			newStandardSettings(), cmdRepo, queryRepo,
		)
		echoInstance := echo.New()
		echoInstance.Use(middleware)

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

		secondRequest := httptest.NewRequest(
			http.MethodGet, "/api/health", nil,
		)
		secondRequest.RemoteAddr = "1.2.3.4:1234"
		secondRecorder := httptest.NewRecorder()
		echoInstance.ServeHTTP(secondRecorder, secondRequest)

		if secondRecorder.Code != http.StatusFound {
			t.Errorf("SecondRequestStatusCodeMismatch: got=%d, want=%d",
				secondRecorder.Code, http.StatusFound)
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
		queryRepo := newBannedIpQueryRepo()

		middleware := NewHoneypotMiddleware(
			newStandardSettings(), cmdRepo, queryRepo,
		)
		echoInstance := echo.New()
		echoInstance.Use(middleware)

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
			queryRepo := newBannedIpQueryRepo()

			middleware := NewHoneypotMiddleware(
				newStandardSettings(), nil, queryRepo,
			)
			echoInstance := echo.New()
			echoInstance.Use(middleware)

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
	queryRepo := newAlwaysEmptyQueryRepo()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), cmdRepo, queryRepo,
	)
	echoInstance := echo.New()
	echoInstance.Use(middleware)

	honeypotPaths := []string{
		"/.env",
		"/wp-config.php",
		"/wp-config.php.bak",
		"/config.php",
		"/backup.sql",
		"/backup.zip",
		"/.git/config",
		"/.aws/credentials",
		"/actuator/env",
		"/actuator/configprops",
		"/server-status",
		"/phpmyadmin/index.php",
		"/admin.php",
		"/administrator/index.php",
		"/login.php",
		"/shell.php",
		"/cmd.php",
		"/test.php",
		"/.htaccess",
		"/web.config",
		"/robots.txt",
		"/sitemap.xml",
		"/debug.php",
		"/info.php",
		"/console",
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
					honeypotPath, httpRecorder.Code, http.StatusOK)
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
		queryRepo := newAlwaysEmptyQueryRepo()

		middleware := NewHoneypotMiddleware(
			newStandardSettings(), cmdRepo, queryRepo,
		)
		echoInstance := echo.New()
		echoInstance.Use(middleware)
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

	t.Run("TrailDbUnavailableFailsOpen", func(t *testing.T) {
		queryRepo := newAlwaysEmptyQueryRepo()

		middleware := NewHoneypotMiddleware(
			newStandardSettings(), nil, queryRepo,
		)
		echoInstance := echo.New()
		echoInstance.Use(middleware)
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

	middleware := NewHoneypotMiddleware(settings, nil, nil)
	if middleware == nil {
		t.Errorf("MiddlewareIsNil")
	}

	echoInstance := echo.New()
	echoInstance.Use(middleware)

	queryRepo := newBannedIpQueryRepo()
	middlewareWithQuery := NewHoneypotMiddleware(
		settings, nil, queryRepo,
	)
	echoInstance2 := echo.New()
	echoInstance2.Use(middlewareWithQuery)

	httpRequest := httptest.NewRequest(
		http.MethodGet, "/api/health", nil,
	)
	httpRequest.RemoteAddr = "1.2.3.4:1234"
	httpRecorder := httptest.NewRecorder()
	echoInstance2.ServeHTTP(httpRecorder, httpRequest)

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
	queryRepo := newBannedIpQueryRepo()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), cmdRepo, queryRepo,
	)
	echoInstance := echo.New()
	echoInstance.Use(middleware)

	firstRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
	)
	firstRequest.RemoteAddr = "1.2.3.4:1234"
	firstRecorder := httptest.NewRecorder()
	echoInstance.ServeHTTP(firstRecorder, firstRequest)

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

	queryRepo := mockActivityRecordQueryRepo{
		readFunc: func(
			req tkDto.ReadActivityRecordsRequest,
		) (tkDto.ReadActivityRecordsResponse, error) {
			return tkDto.ReadActivityRecordsResponse{}, nil
		},
		readFirstFunc: func(
			req tkDto.ReadActivityRecordsRequest,
		) (tkEntity.ActivityRecord, error) {
			if recordCount >= 50 {
				return tkEntity.ActivityRecord{
					RecordId:   tkValueObject.ActivityRecordId(1),
					RecordCode: tkValueObject.ActivityRecordCode("HoneypotHit"),
				}, nil
			}
			return tkEntity.ActivityRecord{}, nil
		},
	}

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), cmdRepo, queryRepo,
	)
	echoInstance := echo.New()
	echoInstance.Use(middleware)

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
	fakeApiKeysUrlPath, _ := tkValueObject.NewUrlPath("/fake-api-keys")
	fakeApiKeysMimeType, _ := tkValueObject.NewMimeType("application/json")

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
				ExtraPathRoutes: testCase.extraPathRoutes,
			}

			honeypotMiddleware := NewHoneypotMiddleware(
				honeypotSettings, nil, nil,
			)
			echoInstance := echo.New()
			echoInstance.Use(honeypotMiddleware)

			incomingRequest := httptest.NewRequest(
				http.MethodGet, testCase.interceptPath, nil,
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
	queryRepo := newAlwaysEmptyQueryRepo()

	middleware := NewHoneypotMiddleware(
		newStandardSettings(), cmdRepo, queryRepo,
	)
	echoInstance := echo.New()
	echoInstance.Use(middleware)

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


