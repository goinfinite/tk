package tkPresentation

import (
	"log/slog"
	"reflect"
	"strings"
)

// StringSliceValueObjectParser converts various input formats into a slice of typed objects.
// It accepts:
// - nil (returns empty slice)
// - string (splits by ";" or "," and parses each element)
// - slice (parses each element)
// - single value (parses as one element)
//
// The valueObjectConstructor function is used to convert each raw value into the desired type.
// Invalid values are logged and skipped.
func StringSliceValueObjectParser[TypedObject any](
	rawInputValues any,
	valueObjectConstructor func(any) (TypedObject, error),
) []TypedObject {
	parsedObjects := make([]TypedObject, 0)

	if rawInputValues == nil {
		return parsedObjects
	}

	rawReflectedSlice := make([]any, 0)

	reflectedRawValues := reflect.ValueOf(rawInputValues)
	switch rawInputValuesKind := reflectedRawValues.Kind(); rawInputValuesKind {
	case reflect.String:
		reflectedRawValuesStr := reflectedRawValues.String()
		rawSeparatedValues := strings.Split(reflectedRawValuesStr, ";")
		if len(rawSeparatedValues) <= 1 {
			rawSeparatedValues = strings.Split(reflectedRawValuesStr, ",")
		}

		for _, rawValue := range rawSeparatedValues {
			rawReflectedSlice = append(rawReflectedSlice, rawValue)
		}
	case reflect.Slice:
		for valueIndex := range reflectedRawValues.Len() {
			rawReflectedSlice = append(
				rawReflectedSlice, reflectedRawValues.Index(valueIndex).Interface(),
			)
		}
	default:
		rawReflectedSlice = append(rawReflectedSlice, rawInputValues)
	}

	for _, rawValue := range rawReflectedSlice {
		valueObject, err := valueObjectConstructor(rawValue)
		if err != nil {
			slog.Debug(err.Error(), slog.Any("rawValue", rawValue))
			continue
		}

		parsedObjects = append(parsedObjects, valueObject)
	}

	return parsedObjects
}
