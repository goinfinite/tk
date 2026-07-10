package tkPresentationMiddlewareHoneypot

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	"github.com/labstack/echo/v4"
)

var lawEnforcementRedirectPool = []string{
	"https://www.ic3.gov/ComplaintChoice.aspx",
	"https://www.interpol.int/How-to-report-a-crime",
	"https://www.europol.europa.eu/report-a-crime",
	"https://reportfraud.ftc.gov/",
	"https://www.actionfraud.police.uk/",
	"https://www.bka.de/EN/CurrentInformation/Reporting/reporting_node.html",
	"https://www.rcmp-grc.gc.ca/en/reporting-crime",
}

type MixedResponseHandler struct {
	mutex           sync.Mutex
	randomGen       *rand.Rand
	redirectPoolIdx int
}

func NewMixedResponseHandler() *MixedResponseHandler {
	return &MixedResponseHandler{
		randomGen:       rand.New(rand.NewSource(time.Now().UnixNano())),
		redirectPoolIdx: 0,
	}
}

func (handler *MixedResponseHandler) Serve(
	echoContext echo.Context,
	pathClass tkValueObject.HoneypotPathClass,
) error {
	handler.mutex.Lock()
	selectionRoll := handler.randomGen.Intn(100)
	handler.mutex.Unlock()

	if selectionRoll < 60 {
		return handler.serveStaticPayload(echoContext, pathClass)
	}

	if selectionRoll < 85 {
		return handler.serveStreamRedirect(echoContext)
	}

	return handler.serveLawEnforcementRedirect(echoContext)
}

func (handler *MixedResponseHandler) resolveLawEnforcementRedirect() string {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()
	redirectUrl := lawEnforcementRedirectPool[handler.redirectPoolIdx]
	handler.redirectPoolIdx++
	if handler.redirectPoolIdx >= len(lawEnforcementRedirectPool) {
		handler.redirectPoolIdx = 0
	}
	return redirectUrl
}

func (handler *MixedResponseHandler) serveStaticPayload(
	echoContext echo.Context,
	pathClass tkValueObject.HoneypotPathClass,
) error {
	payloadByClass := map[tkValueObject.HoneypotPathClass]string{
		tkValueObject.HoneypotPathClassStaticVulnerability: "DB_PASSWORD=supersecret\nAPI_KEY=sk-fake-12345\n",
		tkValueObject.HoneypotPathClassBandwidthExhaust:    "{\"status\":\"ok\",\"data\":{}}",
		tkValueObject.HoneypotPathClassAiTrap:              "{\"model\":\"gpt-fake\",\"choices\":[]}",
	}

	payload, exists := payloadByClass[pathClass]
	if !exists {
		payload = payloadByClass[tkValueObject.HoneypotPathClassStaticVulnerability]
	}

	return echoContext.String(http.StatusOK, payload)
}

func (handler *MixedResponseHandler) serveStreamRedirect(
	echoContext echo.Context,
) error {
	return echoContext.Redirect(
		http.StatusFound, "/api/v1/stream/logs",
	)
}

func (handler *MixedResponseHandler) serveLawEnforcementRedirect(
	echoContext echo.Context,
) error {
	redirectUrl := handler.resolveLawEnforcementRedirect()
	return echoContext.Redirect(http.StatusFound, redirectUrl)
}
