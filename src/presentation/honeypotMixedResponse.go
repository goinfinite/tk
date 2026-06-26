package tkPresentation

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
)

const defaultRedirectUrlString string = "https://xkcd.com/"

var lawEnforcementRedirectUrls = []string{
	"https://www.fbi.gov/",
	"https://www.nsa.gov/",
	"https://www.interpol.int/",
}

var securityQueryStringPool = []string{
	"?ref=suspicious-activity-investigate-ip",
	"?source=security-alert-botnet-suspect",
	"?utm=investigate-this-ip-threat",
	"?ref=botnet-activity-security-risk",
	"?source=ip-needs-investigation-suspicious",
}

func fake503Body() string {
	return "<!DOCTYPE html>\n<html>\n<head>" +
		"<title>503 Service Temporarily Unavailable" +
		"</title>\n</head>\n<body>\n" +
		"<center><h1>503 Service Temporarily " +
		"Unavailable</h1></center>\n<hr>" +
		"<center>nginx</center>\n</body>\n</html>"
}

func fake502Body() string {
	return "<!DOCTYPE html>\n<html>\n<head>" +
		"<title>502 Bad Gateway</title>\n</head>\n" +
		"<body>\n<center><h1>502 Bad Gateway" +
		"</h1></center>\n<hr><center>nginx" +
		"</center>\n</body>\n</html>"
}

func fake429Body() string {
	bodyBytes, _ := json.Marshal(map[string]string{
		"error": "Too Many Requests",
	})
	return string(bodyBytes)
}

func isMixedResponseStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusFound,
		http.StatusTemporaryRedirect,
		http.StatusServiceUnavailable,
		http.StatusBadGateway,
		http.StatusTooManyRequests:
		return true
	}
	return false
}

func (middleware *HoneypotMiddleware) hasCustomRedirect() bool {
	return middleware.settings.RedirectUrl.String() !=
		defaultRedirectUrlString
}

func (middleware *HoneypotMiddleware) serveLawEnforcementRedirect(
	echoContext echo.Context,
) error {
	if middleware.hasCustomRedirect() {
		return echoContext.Redirect(
			http.StatusFound,
			middleware.settings.RedirectUrl.String(),
		)
	}
	urlIndex := rand.Intn(
		len(lawEnforcementRedirectUrls),
	)
	redirectUrl := lawEnforcementRedirectUrls[urlIndex]
	queryIndex := rand.Intn(
		len(securityQueryStringPool),
	)
	queryString := securityQueryStringPool[queryIndex]
	statusCode := http.StatusFound
	if rand.Intn(2) == 0 {
		statusCode = http.StatusTemporaryRedirect
	}
	return echoContext.Redirect(
		statusCode, redirectUrl+queryString,
	)
}

func (middleware *HoneypotMiddleware) serveFake503(
	echoContext echo.Context,
) error {
	echoContext.Response().Header().Set(
		"Content-Type", "text/html",
	)
	return echoContext.String(
		http.StatusServiceUnavailable, fake503Body(),
	)
}

func (middleware *HoneypotMiddleware) serveFake502(
	echoContext echo.Context,
) error {
	echoContext.Response().Header().Set(
		"Content-Type", "text/html",
	)
	return echoContext.String(
		http.StatusBadGateway, fake502Body(),
	)
}

func (middleware *HoneypotMiddleware) serveFake429(
	echoContext echo.Context,
) error {
	echoContext.Response().Header().Set(
		"Content-Type", "application/json",
	)
	echoContext.Response().Header().Set(
		"Retry-After", "3600",
	)
	return echoContext.String(
		http.StatusTooManyRequests, fake429Body(),
	)
}

func (middleware *HoneypotMiddleware) serveMixedResponse(
	echoContext echo.Context,
) error {
	if middleware.hasCustomRedirect() {
		return echoContext.Redirect(
			http.StatusFound,
			middleware.settings.RedirectUrl.String(),
		)
	}
	randomValue := rand.Intn(100)
	switch {
	case randomValue < 40:
		return middleware.serveLawEnforcementRedirect(
			echoContext,
		)
	case randomValue < 70:
		return middleware.serveFake503(echoContext)
	case randomValue < 90:
		return middleware.serveFake502(echoContext)
	default:
		return middleware.serveFake429(echoContext)
	}
}
