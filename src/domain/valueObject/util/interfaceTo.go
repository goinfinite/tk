package tkVoUtil

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func InterfaceToBool(input any) (output bool, conversionErr error) {
	switch inputValue := input.(type) {
	case bool:
		return inputValue, nil
	case string:
		stringValue := strings.TrimSpace(inputValue)
		stringValue = strings.ToLower(stringValue)
		if stringValue == "on" {
			return true, nil
		}

		if stringValue == "off" {
			return false, nil
		}

		return strconv.ParseBool(stringValue)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		return intValue != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		return uintValue != 0, nil
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		return floatValue != 0, nil
	default:
		return false, errors.New("CannotConvertToBool")
	}
}

func InterfaceToString(input any) (output string, conversionErr error) {
	switch inputValue := input.(type) {
	case string:
		output = inputValue
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		output = strconv.FormatInt(intValue, 10)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		output = strconv.FormatUint(uintValue, 10)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		output = strconv.FormatFloat(floatValue, 'f', -1, 64)
	case bool:
		boolValue := reflect.ValueOf(inputValue).Bool()
		output = strconv.FormatBool(boolValue)
	default:
		return "", errors.New("CannotConvertToString")
	}

	return strings.TrimSpace(output), nil
}

func InterfaceToInt(input any) (output int, conversionErr error) {
	switch inputValue := input.(type) {
	case string:
		return strconv.Atoi(inputValue)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		return int(intValue), nil
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		return int(uintValue), nil
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		return int(floatValue), nil
	default:
		return 0, errors.New("CannotConvertToInt")
	}
}

func InterfaceToInt8(input any) (output int8, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToInt8")

	switch inputValue := input.(type) {
	case string:
		int64Value, parseErr := strconv.ParseInt(inputValue, 10, 8)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		output = int8(int64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < -128 || intValue > 127 {
			return 0, conversionErrMsg
		}
		output = int8(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		if uintValue > 127 {
			return 0, conversionErrMsg
		}
		output = int8(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < -128 || floatValue > 127 {
			return 0, conversionErrMsg
		}
		output = int8(floatValue)
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToInt16(input any) (output int16, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToInt16")

	switch inputValue := input.(type) {
	case string:
		int64Value, parseErr := strconv.ParseInt(inputValue, 10, 16)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		output = int16(int64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < -32768 || intValue > 32767 {
			return 0, conversionErrMsg
		}
		output = int16(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		if uintValue > 32767 {
			return 0, conversionErrMsg
		}
		output = int16(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < -32768 || floatValue > 32767 {
			return 0, conversionErrMsg
		}
		output = int16(floatValue)
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToInt32(input any) (output int32, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToInt32")

	switch inputValue := input.(type) {
	case string:
		int64Value, parseErr := strconv.ParseInt(inputValue, 10, 32)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		output = int32(int64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < -2147483648 || intValue > 2147483647 {
			return 0, conversionErrMsg
		}
		output = int32(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		if uintValue > 2147483647 {
			return 0, conversionErrMsg
		}
		output = int32(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < -2147483648 || floatValue > 2147483647 {
			return 0, conversionErrMsg
		}
		output = int32(floatValue)
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToInt64(input any) (output int64, conversionErr error) {
	switch inputValue := input.(type) {
	case string:
		return strconv.ParseInt(inputValue, 10, 64)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		return int64(intValue), nil
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		return int64(uintValue), nil
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		return int64(floatValue), nil
	default:
		return 0, errors.New("CannotConvertToInt64")
	}
}

func InterfaceToUint(input any) (output uint, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToUint")

	switch inputValue := input.(type) {
	case string:
		uint64Value, parseErr := strconv.ParseUint(inputValue, 10, 64)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		if uint64Value > 4294967295 {
			return 0, conversionErrMsg
		}
		output = uint(uint64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < 0 || intValue > 4294967295 {
			return 0, conversionErrMsg
		}
		output = uint(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := uint(reflect.ValueOf(inputValue).Uint())
		if uintValue > 4294967295 {
			return 0, conversionErrMsg
		}
		output = uintValue
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < 0 || floatValue > 4294967295 {
			return 0, conversionErrMsg
		}
		output = uint(floatValue)
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToUint8(input any) (output uint8, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToUint8")

	switch inputValue := input.(type) {
	case string:
		uint64Value, parseErr := strconv.ParseUint(inputValue, 10, 8)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		output = uint8(uint64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < 0 || intValue > 255 {
			return 0, conversionErrMsg
		}
		output = uint8(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		if uintValue > 255 {
			return 0, conversionErrMsg
		}
		output = uint8(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < 0 || floatValue > 255 {
			return 0, conversionErrMsg
		}
		output = uint8(floatValue)
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToUint16(input any) (output uint16, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToUint16")

	switch inputValue := input.(type) {
	case string:
		uint64Value, parseErr := strconv.ParseUint(inputValue, 10, 64)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		if uint64Value > 65535 {
			return 0, conversionErrMsg
		}
		output = uint16(uint64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < 0 || intValue > 65535 {
			conversionErr = conversionErrMsg
		}
		output = uint16(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		if uintValue > 65535 {
			conversionErr = conversionErrMsg
		}
		output = uint16(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < 0 || floatValue > 65535 {
			conversionErr = conversionErrMsg
		}
		output = uint16(floatValue)
	default:
		conversionErr = conversionErrMsg
	}

	return output, conversionErr
}

func InterfaceToUint32(input any) (output uint32, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToUint32")
	switch inputValue := input.(type) {
	case string:
		uint64Value, parseErr := strconv.ParseUint(inputValue, 10, 64)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		if uint64Value > 4294967295 {
			return 0, conversionErrMsg
		}
		output = uint32(uint64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < 0 || intValue > 4294967295 {
			conversionErr = conversionErrMsg
		}
		output = uint32(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		if uintValue > 4294967295 {
			conversionErr = conversionErrMsg
		}
		output = uint32(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < 0 || floatValue > 4294967295 {
			conversionErr = conversionErrMsg
		}
		output = uint32(floatValue)
	default:
		conversionErr = conversionErrMsg
	}

	return output, conversionErr
}

func InterfaceToUint64(input any) (output uint64, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToUint64")

	switch inputValue := input.(type) {
	case string:
		output, conversionErr = strconv.ParseUint(inputValue, 10, 64)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		if intValue < 0 {
			conversionErr = conversionErrMsg
		}
		output = uint64(intValue)
	case uint, uint8, uint16, uint32, uint64:
		output = uint64(reflect.ValueOf(inputValue).Uint())
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		if floatValue < 0 {
			conversionErr = conversionErrMsg
		}
		output = uint64(floatValue)
	default:
		conversionErr = conversionErrMsg
	}

	if conversionErr != nil {
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToFloat32(input any) (output float32, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToFloat32")

	switch inputValue := input.(type) {
	case string:
		float64Value, parseErr := strconv.ParseFloat(inputValue, 32)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		output = float32(float64Value)
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		output = float32(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		output = float32(uintValue)
	case float32, float64:
		floatValue := reflect.ValueOf(inputValue).Float()
		output = float32(floatValue)
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}

func InterfaceToFloat64(input any) (output float64, conversionErr error) {
	conversionErrMsg := errors.New("CannotConvertToFloat64")

	switch inputValue := input.(type) {
	case string:
		parsedValue, parseErr := strconv.ParseFloat(inputValue, 64)
		if parseErr != nil {
			return 0, conversionErrMsg
		}
		output = parsedValue
	case int, int8, int16, int32, int64:
		intValue := reflect.ValueOf(inputValue).Int()
		output = float64(intValue)
	case uint, uint8, uint16, uint32, uint64:
		uintValue := reflect.ValueOf(inputValue).Uint()
		output = float64(uintValue)
	case float32, float64:
		output = reflect.ValueOf(inputValue).Float()
	default:
		return 0, conversionErrMsg
	}

	return output, nil
}
