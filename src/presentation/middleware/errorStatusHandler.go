package tkPresentationMiddleware

import (
	"errors"
	"log/slog"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkPresentation "github.com/goinfinite/tk/src/presentation"
	"github.com/labstack/echo/v4"
)

const highestHttpStatusCode tkValueObject.HttpStatusCode = 599

type ErrorStatusHandler struct {
	errorStatusCodesMap  map[error]tkValueObject.HttpStatusCode
	requesterIpExtractor tkPresentation.RequesterIpExtractor
}

func NewErrorStatusHandler(
	errorStatusCodesMap map[error]tkValueObject.HttpStatusCode,
) ErrorStatusHandler {
	return ErrorStatusHandler{
		errorStatusCodesMap:  errorStatusCodesMap,
		requesterIpExtractor: tkPresentation.NewRequesterIpExtractor(),
	}
}

func (handler ErrorStatusHandler) errorStatusResolver(
	handlerErr error,
) (
	statusCode tkValueObject.HttpStatusCode, matchedErrorsCount int,
	matchedErr error,
) {
	for knownErr, mappedStatusCode := range handler.errorStatusCodesMap {
		if !errors.Is(handlerErr, knownErr) {
			continue
		}
		statusCode = mappedStatusCode
		matchedErr = knownErr
		matchedErrorsCount++
	}

	if matchedErrorsCount == 1 {
		return statusCode, matchedErrorsCount, matchedErr
	}

	return tkValueObject.HttpStatusCodeInternalServerError, matchedErrorsCount, nil
}

func (handler ErrorStatusHandler) buildRequestLogAttributes(
	echoContext echo.Context,
) []any {
	requestId := echoContext.Response().Header().Get(echo.HeaderXRequestID)
	logAttributes := []any{
		slog.String("method", echoContext.Request().Method),
		slog.String("path", echoContext.Request().URL.Path),
		slog.String("route", echoContext.Path()),
		slog.String("requestId", requestId),
	}

	operatorIpAddress, extractionErr := handler.requesterIpExtractor.Execute(
		echoContext.Request(),
	)
	if extractionErr == nil {
		logAttributes = append(
			logAttributes,
			slog.String("operatorIpAddress", operatorIpAddress.String()),
		)
	}

	operatorAccountId, hasOperatorAccountId := echoContext.Get(
		"operatorAccountId",
	).(tkValueObject.AccountId)
	if hasOperatorAccountId {
		logAttributes = append(
			logAttributes,
			slog.String("operatorAccountId", operatorAccountId.String()),
		)
	}

	return logAttributes
}

func (handler ErrorStatusHandler) emitResponse(
	echoContext echo.Context,
	statusCode tkValueObject.HttpStatusCode,
	readableMessage string,
	handlerErr error,
) error {
	httpStatusCode := int(statusCode.Uint16())
	if statusCode < tkValueObject.HttpStatusCodeBadRequest ||
		statusCode > highestHttpStatusCode {
		logAttributes := handler.buildRequestLogAttributes(echoContext)
		logAttributes = append(
			logAttributes,
			slog.String("statusCode", statusCode.String()),
			slog.String("matchedError", readableMessage),
			slog.String("error", handlerErr.Error()),
		)
		slog.Warn(
			"ErrorStatusHandlerInvalidStatusCode",
			logAttributes...,
		)
		httpStatusCode = int(tkValueObject.HttpStatusCodeInternalServerError.Uint16())
		readableMessage = "InternalServerError"
	}

	return echoContext.JSON(
		httpStatusCode,
		tkPresentation.NewApiResponseWrapper(
			httpStatusCode, readableMessage, readableMessage,
		),
	)
}

func (handler ErrorStatusHandler) ApiHandler(
	subsequentHandler echo.HandlerFunc,
) echo.HandlerFunc {
	return func(echoContext echo.Context) error {
		handlerErr := subsequentHandler(echoContext)
		if handlerErr == nil {
			return nil
		}

		if echoContext.Response().Committed {
			logAttributes := handler.buildRequestLogAttributes(echoContext)
			logAttributes = append(
				logAttributes,
				slog.Int("responseStatus", echoContext.Response().Status),
				slog.String("error", handlerErr.Error()),
			)
			slog.Error(
				"ErrorStatusHandlerErrorAfterResponseCommitted",
				logAttributes...,
			)
			return handlerErr
		}

		_, isEchoHttpError := handlerErr.(*echo.HTTPError)
		if isEchoHttpError {
			return handlerErr
		}

		statusCode, matchedErrorsCount, matchedErr := handler.errorStatusResolver(
			handlerErr,
		)
		if matchedErrorsCount == 1 {
			isServerError := statusCode >= tkValueObject.HttpStatusCodeInternalServerError &&
				statusCode <= highestHttpStatusCode
			if isServerError {
				logAttributes := handler.buildRequestLogAttributes(echoContext)
				logAttributes = append(
					logAttributes,
					slog.Int("statusCode", int(statusCode.Uint16())),
					slog.String("matchedError", matchedErr.Error()),
					slog.String("error", handlerErr.Error()),
				)
				slog.Error(
					"ErrorStatusHandlerMappedServerError",
					logAttributes...,
				)
			}
			return handler.emitResponse(
				echoContext, statusCode, matchedErr.Error(), handlerErr,
			)
		}

		logAttributes := handler.buildRequestLogAttributes(echoContext)
		logAttributes = append(
			logAttributes,
			slog.Int("matchedErrorsCount", matchedErrorsCount),
			slog.String("error", handlerErr.Error()),
		)

		switch matchedErrorsCount {
		case 0:
			slog.Error(
				"ErrorStatusHandlerUnmappedError",
				logAttributes...,
			)
		default:
			slog.Warn(
				"ErrorStatusHandlerAmbiguousMapping",
				logAttributes...,
			)
		}

		return handler.emitResponse(
			echoContext, statusCode, "InternalServerError", handlerErr,
		)
	}
}
