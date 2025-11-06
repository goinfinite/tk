package tkPresentation

import (
	"log/slog"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

// This function parses time parameters from untrusted input. Time parameters are
// almost always optional, so it returns a map of pointers to tkValueObject.UnixTime.
// This allows the caller to easily determine if a time parameter was provided by
// checking if the pointer is nil. The parsed time parameters can then be directly
// set in a DTO.
func TimeParamsParser(
	timeParamNames []string,
	untrustedInput map[string]any,
) map[string]*tkValueObject.UnixTime {
	timeParamsPtr := map[string]*tkValueObject.UnixTime{}

	for _, timeParamName := range timeParamNames {
		switch untrustedInput[timeParamName].(type) {
		case string:
			if untrustedInput[timeParamName] == "" {
				continue
			}
		case nil:
			continue
		}

		timeParam, err := tkValueObject.NewUnixTime(untrustedInput[timeParamName])
		if err != nil {
			slog.Debug("InvalidTimeParam", slog.String("timeParamName", timeParamName))
			timeParamsPtr[timeParamName] = nil
			continue
		}

		timeParamsPtr[timeParamName] = &timeParam
	}

	return timeParamsPtr
}
