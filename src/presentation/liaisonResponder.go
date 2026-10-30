package tkPresentation

import "log/slog"

const internalServerErrorMessage = "InternalServerError"

type LiaisonResponder struct {
	errorStatusResolver liaisonErrorStatusResolver
}

func NewLiaisonResponder(
	errorStatusesMap map[error]LiaisonResponseStatus,
) LiaisonResponder {
	return LiaisonResponder{
		errorStatusResolver: newLiaisonErrorStatusResolver(errorStatusesMap),
	}
}

func (responder LiaisonResponder) Success(
	body any, readableMessage string,
) LiaisonResponse {
	return newLiaisonResponse(
		LiaisonResponseStatusSuccess, body, readableMessage,
	)
}

func (responder LiaisonResponder) Created(
	body any, readableMessage string,
) LiaisonResponse {
	return newLiaisonResponse(
		LiaisonResponseStatusCreated, body, readableMessage,
	)
}

func (responder LiaisonResponder) Accepted(
	body any, readableMessage string,
) LiaisonResponse {
	return newLiaisonResponse(
		LiaisonResponseStatusAccepted, body, readableMessage,
	)
}

func (responder LiaisonResponder) MultiStatus(
	body any, readableMessage string,
) LiaisonResponse {
	return newLiaisonResponse(
		LiaisonResponseStatusMultiStatus, body, readableMessage,
	)
}

func (responder LiaisonResponder) InputError(
	message string,
) LiaisonResponse {
	return newLiaisonResponse(
		LiaisonResponseStatusUserError, message, message,
	)
}

func (responder LiaisonResponder) AppError(
	appErr error,
) LiaisonResponse {
	fallbackResponse := newLiaisonResponse(
		LiaisonResponseStatusUnknownError,
		internalServerErrorMessage,
		internalServerErrorMessage,
	)
	if appErr == nil {
		slog.Error("LiaisonResponderNilAppError")
		return fallbackResponse
	}

	status, matchedErrorsCount, matchedErr := responder.errorStatusResolver.resolve(
		appErr,
	)
	if matchedErrorsCount == 1 && status.isError() {
		if status.isServerError() {
			slog.Error(
				"LiaisonResponderMappedServerError",
				slog.String("status", string(status)),
				slog.String("matchedError", matchedErr.Error()),
				slog.String("error", appErr.Error()),
			)
		}
		return newLiaisonResponse(
			status, matchedErr.Error(), matchedErr.Error(),
		)
	}

	logAttributes := []any{
		slog.Int("matchedErrorsCount", matchedErrorsCount),
		slog.String("error", appErr.Error()),
	}

	switch matchedErrorsCount {
	case 0:
		slog.Error(
			"LiaisonResponderUnmappedAppError",
			logAttributes...,
		)
	case 1:
		logAttributes = append(
			logAttributes, slog.String("status", string(status)),
		)
		slog.Warn(
			"LiaisonResponderInvalidMappedStatus",
			logAttributes...,
		)
	default:
		slog.Warn(
			"LiaisonResponderAmbiguousMapping",
			logAttributes...,
		)
	}

	return fallbackResponse
}
