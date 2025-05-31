package tkPresentation

import (
	"errors"
	"strings"
)

func RequiredParamsInspector(paramsReceived map[string]any, paramsRequired []string) error {
	paramsMissing := []string{}
	for _, paramName := range paramsRequired {
		if _, exists := paramsReceived[paramName]; !exists {
			paramsMissing = append(paramsMissing, paramName)
		}
	}

	if len(paramsMissing) == 0 {
		return nil
	}

	return errors.New("RequiredParamsMissing: " + strings.Join(paramsMissing, ", "))
}
