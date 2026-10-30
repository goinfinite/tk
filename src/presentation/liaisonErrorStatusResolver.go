package tkPresentation

import "errors"

type liaisonErrorStatusResolver struct {
	errorStatusesMap map[error]LiaisonResponseStatus
}

func newLiaisonErrorStatusResolver(
	errorStatusesMap map[error]LiaisonResponseStatus,
) liaisonErrorStatusResolver {
	return liaisonErrorStatusResolver{
		errorStatusesMap: errorStatusesMap,
	}
}

func (resolver liaisonErrorStatusResolver) resolve(
	liaisonErr error,
) (
	status LiaisonResponseStatus, matchedErrorsCount int,
	matchedErr error,
) {
	for knownErr, mappedStatus := range resolver.errorStatusesMap {
		if !errors.Is(liaisonErr, knownErr) {
			continue
		}

		status = mappedStatus
		matchedErr = knownErr
		matchedErrorsCount++
	}

	return status, matchedErrorsCount, matchedErr
}
