package tkValueObject

import (
	"errors"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	HttpStatusCodeContinue                      HttpStatusCode = 100
	HttpStatusCodeSwitchingProtocols            HttpStatusCode = 101
	HttpStatusCodeProcessing                    HttpStatusCode = 102
	HttpStatusCodeEarlyHints                    HttpStatusCode = 103
	HttpStatusCodeOk                            HttpStatusCode = 200
	HttpStatusCodeCreated                       HttpStatusCode = 201
	HttpStatusCodeAccepted                      HttpStatusCode = 202
	HttpStatusCodeNonAuthoritativeInfo          HttpStatusCode = 203
	HttpStatusCodeNoContent                     HttpStatusCode = 204
	HttpStatusCodeResetContent                  HttpStatusCode = 205
	HttpStatusCodePartialContent                HttpStatusCode = 206
	HttpStatusCodeMultiStatus                   HttpStatusCode = 207
	HttpStatusCodeAlreadyReported               HttpStatusCode = 208
	HttpStatusCodeImUsed                        HttpStatusCode = 226
	HttpStatusCodeMultipleChoices               HttpStatusCode = 300
	HttpStatusCodeMovedPermanently              HttpStatusCode = 301
	HttpStatusCodeFound                         HttpStatusCode = 302
	HttpStatusCodeSeeOther                      HttpStatusCode = 303
	HttpStatusCodeNotModified                   HttpStatusCode = 304
	HttpStatusCodeTemporaryRedirect             HttpStatusCode = 307
	HttpStatusCodePermanentRedirect             HttpStatusCode = 308
	HttpStatusCodeBadRequest                    HttpStatusCode = 400
	HttpStatusCodeUnauthorized                  HttpStatusCode = 401
	HttpStatusCodePaymentRequired               HttpStatusCode = 402
	HttpStatusCodeForbidden                     HttpStatusCode = 403
	HttpStatusCodeNotFound                      HttpStatusCode = 404
	HttpStatusCodeMethodNotAllowed              HttpStatusCode = 405
	HttpStatusCodeNotAcceptable                 HttpStatusCode = 406
	HttpStatusCodeProxyAuthenticationRequired   HttpStatusCode = 407
	HttpStatusCodeRequestTimeout                HttpStatusCode = 408
	HttpStatusCodeConflict                      HttpStatusCode = 409
	HttpStatusCodeGone                          HttpStatusCode = 410
	HttpStatusCodeLengthRequired                HttpStatusCode = 411
	HttpStatusCodePreconditionFailed            HttpStatusCode = 412
	HttpStatusCodePayloadTooLarge               HttpStatusCode = 413
	HttpStatusCodeUriTooLong                    HttpStatusCode = 414
	HttpStatusCodeUnsupportedMediaType          HttpStatusCode = 415
	HttpStatusCodeRangeNotSatisfiable           HttpStatusCode = 416
	HttpStatusCodeExpectationFailed             HttpStatusCode = 417
	HttpStatusCodeImATeapot                     HttpStatusCode = 418
	HttpStatusCodeMisdirectedRequest            HttpStatusCode = 421
	HttpStatusCodeUnprocessableEntity           HttpStatusCode = 422
	HttpStatusCodeLocked                        HttpStatusCode = 423
	HttpStatusCodeFailedDependency              HttpStatusCode = 424
	HttpStatusCodeTooEarly                      HttpStatusCode = 425
	HttpStatusCodeUpgradeRequired               HttpStatusCode = 426
	HttpStatusCodePreconditionRequired          HttpStatusCode = 428
	HttpStatusCodeTooManyRequests               HttpStatusCode = 429
	HttpStatusCodeRequestHeaderFieldsTooLarge   HttpStatusCode = 431
	HttpStatusCodeUnavailableForLegalReasons    HttpStatusCode = 451
	HttpStatusCodeInternalServerError           HttpStatusCode = 500
	HttpStatusCodeNotImplemented                HttpStatusCode = 501
	HttpStatusCodeBadGateway                    HttpStatusCode = 502
	HttpStatusCodeServiceUnavailable            HttpStatusCode = 503
	HttpStatusCodeGatewayTimeout                HttpStatusCode = 504
	HttpStatusCodeHttpVersionNotSupported       HttpStatusCode = 505
	HttpStatusCodeVariantAlsoNegotiates         HttpStatusCode = 506
	HttpStatusCodeInsufficientStorage           HttpStatusCode = 507
	HttpStatusCodeLoopDetected                  HttpStatusCode = 508
	HttpStatusCodeNotExtended                   HttpStatusCode = 510
	HttpStatusCodeNetworkAuthenticationRequired HttpStatusCode = 511
)

type HttpStatusCode uint16

func NewHttpStatusCode(value any) (statusCode HttpStatusCode, err error) {
	uint16Value, err := tkVoUtil.InterfaceToUint16(value)
	if err != nil {
		return statusCode, errors.New("HttpStatusCodeMustBeUint16")
	}

	if uint16Value < 100 || uint16Value > 599 {
		return statusCode, errors.New("InvalidHttpStatusCode")
	}

	return HttpStatusCode(uint16Value), nil
}

func (vo HttpStatusCode) Uint16() uint16 {
	return uint16(vo)
}

func (vo HttpStatusCode) String() string {
	return strconv.FormatUint(uint64(vo), 10)
}
