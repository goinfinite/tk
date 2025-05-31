package tkPresentation

import (
	"log/slog"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type RequestInputSettings struct {
	RawRequestInput        map[string]any
	KnownParamConstructors map[string]func(any) (any, error)
}

// Go requires exact function signature matches, unless using generics to infer type.
// Hence, this wrapper must be used when creating the KnownParamConstructors map.
func ParamConstructorWrapper[objectInstance any](
	objectConstructor func(any) (objectInstance, error),
) func(any) (any, error) {
	return func(rawValue any) (any, error) {
		return objectConstructor(rawValue)
	}
}

const (
	// spf13/cobra (cli lib) does not support nil or pointers as default values.
	// UnsetParameterValueStr is a special value to circumvent this limitation.
	UnsetParameterValueStr string = "__UNSET__"
)

type RequestInputParsed struct {
	KnownParams      map[string]any
	KnownParamErrors map[string]error
	ClearableParams  map[string]any
	UnknownParams    map[string]any
}

func RequestInputParser(
	componentSettings RequestInputSettings,
) (requestInputParsed RequestInputParsed) {
	requestInputParsed = RequestInputParsed{
		KnownParams:      make(map[string]any),
		KnownParamErrors: make(map[string]error),
		ClearableParams:  make(map[string]any),
		UnknownParams:    make(map[string]any),
	}

	for rawKey, rawValue := range componentSettings.RawRequestInput {
		isClearableParam := false
		switch rawValueAsserted := rawValue.(type) {
		case string:
			if rawValueAsserted == UnsetParameterValueStr {
				continue
			}

			if len(rawValueAsserted) == 0 || rawValueAsserted == " " {
				isClearableParam = true
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			rawValueAssertedInt64, err := tkVoUtil.InterfaceToInt64(rawValueAsserted)
			if err != nil {
				slog.Debug(
					err.Error(),
					slog.String("rawKey", rawKey),
					slog.Any("rawValue", rawValue),
				)
				continue
			}
			if rawValueAssertedInt64 == 0 {
				isClearableParam = true
			}
		case bool:
			if !rawValueAsserted {
				isClearableParam = true
			}
		case nil:
			isClearableParam = true
		}

		if isClearableParam {
			requestInputParsed.ClearableParams[rawKey] = rawValue
			continue
		}

		if _, isKnownKey := componentSettings.KnownParamConstructors[rawKey]; !isKnownKey {
			requestInputParsed.UnknownParams[rawKey] = rawValue
			continue
		}

		parsedKnownParam, err := componentSettings.KnownParamConstructors[rawKey](rawValue)
		if err != nil {
			slog.Debug(
				err.Error(),
				slog.String("rawKey", rawKey),
				slog.Any("rawValue", rawValue),
			)

			requestInputParsed.KnownParamErrors[rawKey] = err
			continue
		}

		requestInputParsed.KnownParams[rawKey] = parsedKnownParam
	}

	return requestInputParsed
}
