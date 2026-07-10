package tkPresentationMiddlewareHoneypot

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

func TestNewMixedResponseHandler(t *testing.T) {
	handler := NewMixedResponseHandler()
	if handler == nil {
		t.Errorf("ExpectedNonNilHandler")
	}
}

func TestMixedResponseHandlerServe(t *testing.T) {
	testCaseStructs := []struct {
		name      string
		pathClass tkValueObject.HoneypotPathClass
	}{
		{
			name:      "StaticVulnerabilityClass",
			pathClass: tkValueObject.HoneypotPathClassStaticVulnerability,
		},
		{
			name:      "BandwidthExhaustClass",
			pathClass: tkValueObject.HoneypotPathClassBandwidthExhaust,
		},
		{
			name:      "AiTrapClass",
			pathClass: tkValueObject.HoneypotPathClassAiTrap,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewMixedResponseHandler()

			echoInstance := echo.New()
			httpRequest := httptest.NewRequest(
				http.MethodGet, "/wp-config.php", nil,
			)
			httpRecorder := httptest.NewRecorder()
			echoContext := echoInstance.NewContext(
				httpRequest, httpRecorder,
			)

			err := handler.Serve(echoContext, testCase.pathClass)
			if err != nil {
				t.Errorf("UnexpectedError: %v", err)
			}

			if httpRecorder.Code == 0 {
				t.Errorf("ExpectedValidStatusCode")
			}
		})
	}
}

func TestMixedResponseHandlerServeReturnsValidStatusCodes(t *testing.T) {
	handler := NewMixedResponseHandler()
	validStatuses := map[int]bool{
		http.StatusOK:    true,
		http.StatusFound: true,
	}

	echoInstance := echo.New()
	httpRequest := httptest.NewRequest(
		http.MethodGet, "/.env", nil,
	)
	httpRecorder := httptest.NewRecorder()
	echoContext := echoInstance.NewContext(httpRequest, httpRecorder)

	err := handler.Serve(
		echoContext,
		tkValueObject.HoneypotPathClassStaticVulnerability,
	)
	if err != nil {
		t.Errorf("UnexpectedError: %v", err)
	}

	if !validStatuses[httpRecorder.Code] {
		t.Errorf(
			"UnexpectedStatusCode: Actual=%d",
			httpRecorder.Code,
		)
	}
}

func TestResolveLawEnforcementRedirect(t *testing.T) {
	handler := NewMixedResponseHandler()

	firstRedirect := handler.resolveLawEnforcementRedirect()
	if firstRedirect == "" {
		t.Errorf("ExpectedNonEmptyRedirectUrl")
	}

	secondRedirect := handler.resolveLawEnforcementRedirect()
	if secondRedirect == "" {
		t.Errorf("ExpectedNonEmptyRedirectUrl")
	}

	if firstRedirect == secondRedirect {
		t.Errorf(
			"ExpectedRotatingRedirects: First='%s', Second='%s'",
			firstRedirect, secondRedirect,
		)
	}
}

func TestResolveLawEnforcementRedirectRotatesFullPool(t *testing.T) {
	handler := NewMixedResponseHandler()
	poolSize := len(lawEnforcementRedirectPool)

	seenRedirects := make(map[string]bool)
	for range poolSize {
		redirect := handler.resolveLawEnforcementRedirect()
		seenRedirects[redirect] = true
	}

	if len(seenRedirects) != poolSize {
		t.Errorf(
			"ExpectedFullRotation: Unique=%d, PoolSize=%d",
			len(seenRedirects), poolSize,
		)
	}
}

func TestResolveLawEnforcementRedirectWrapsAround(t *testing.T) {
	handler := NewMixedResponseHandler()
	poolSize := len(lawEnforcementRedirectPool)

	for range poolSize {
		handler.resolveLawEnforcementRedirect()
	}

	wrapRedirect := handler.resolveLawEnforcementRedirect()
	firstRedirect := lawEnforcementRedirectPool[0]

	if wrapRedirect != firstRedirect {
		t.Errorf(
			"ExpectedWrapAround: Wrap='%s', First='%s'",
			wrapRedirect, firstRedirect,
		)
	}
}

func TestMixedResponseHandlerServeProducesDifferentResponses(t *testing.T) {
	handler := NewMixedResponseHandler()
	echoInstance := echo.New()

	statusCodes := make(map[int]bool)
	for range 50 {
		httpRequest := httptest.NewRequest(
			http.MethodGet, "/wp-config.php", nil,
		)
		httpRecorder := httptest.NewRecorder()
		echoContext := echoInstance.NewContext(
			httpRequest, httpRecorder,
		)

		handler.Serve(
			echoContext,
			tkValueObject.HoneypotPathClassStaticVulnerability,
		)
		statusCodes[httpRecorder.Code] = true
	}

	if len(statusCodes) < 2 {
		t.Errorf(
			"ExpectedVariedResponses: UniqueStatusCodes=%d",
			len(statusCodes),
		)
	}
}
