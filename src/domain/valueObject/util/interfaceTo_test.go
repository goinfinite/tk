package tkVoUtil

import (
	"math"
	"testing"
)

func TestInterfaceToBool(t *testing.T) {
	t.Run("BoolInput", func(t *testing.T) {
		actualOutput, conversionErr := InterfaceToBool(true)
		if conversionErr != nil {
			t.Errorf("UnexpectedError: '%s'", conversionErr.Error())
		}
		if actualOutput != true {
			t.Errorf("UnexpectedOutputValue: '%v' vs 'true'", actualOutput)
		}

		actualOutput, conversionErr = InterfaceToBool(false)
		if conversionErr != nil {
			t.Errorf("UnexpectedError: '%s'", conversionErr.Error())
		}
		if actualOutput != false {
			t.Errorf("UnexpectedOutputValue: '%v' vs 'false'", actualOutput)
		}
	})

	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput bool
		}{
			{"true", true},
			{"false", false},
			{"TRUE", true},
			{"FALSE", false},
			{"True", true},
			{"False", false},
			{"on", true},
			{"off", false},
			{"1", true},
			{"0", false},
			{"  true  ", true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToBool(testCase.inputString)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}

		invalidInputStrings := []string{"yes", "no", "invalid", "2", "-1"}
		for _, invalidString := range invalidInputStrings {
			_, conversionErr := InterfaceToBool(invalidString)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", invalidString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		integerInputs := []any{
			int(1), int8(1), int16(1), int32(1), int64(1),
			int(0), int8(0), int16(0), int32(0), int64(0),
			int(-1), int8(-1), int16(-1), int32(-1), int64(-1),
		}
		for _, numericInput := range integerInputs {
			expectedOutput := false
			switch typedValue := numericInput.(type) {
			case int:
				expectedOutput = typedValue != 0
			case int8:
				expectedOutput = typedValue != 0
			case int16:
				expectedOutput = typedValue != 0
			case int32:
				expectedOutput = typedValue != 0
			case int64:
				expectedOutput = typedValue != 0
			}
			actualOutput, conversionErr := InterfaceToBool(numericInput)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), numericInput)
			}
			if actualOutput != expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, expectedOutput, numericInput)
			}
		}

		unsignedInputs := []any{
			uint(1), uint8(1), uint16(1), uint32(1), uint64(1),
			uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		}
		for _, numericInput := range unsignedInputs {
			expectedOutput := false
			switch typedValue := numericInput.(type) {
			case uint:
				expectedOutput = typedValue != 0
			case uint8:
				expectedOutput = typedValue != 0
			case uint16:
				expectedOutput = typedValue != 0
			case uint32:
				expectedOutput = typedValue != 0
			case uint64:
				expectedOutput = typedValue != 0
			}
			actualOutput, conversionErr := InterfaceToBool(numericInput)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), numericInput)
			}
			if actualOutput != expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, expectedOutput, numericInput)
			}
		}

		floatInputs := []any{
			float32(1.0), float64(1.0),
			float32(0.0), float64(0.0),
			float32(-1.0), float64(-1.0),
		}
		for _, numericInput := range floatInputs {
			expectedOutput := false
			switch typedValue := numericInput.(type) {
			case float32:
				expectedOutput = typedValue != 0
			case float64:
				expectedOutput = typedValue != 0
			}
			actualOutput, conversionErr := InterfaceToBool(numericInput)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), numericInput)
			}
			if actualOutput != expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, expectedOutput, numericInput)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToBool(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToString(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput string
		}{
			{"test", "test"},
			{"123", "123"},
			{"  trimmed  ", "trimmed"},
			{"", ""},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToString(testCase.inputString)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%s' vs '%s' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput string
		}{
			{int(123), "123"},
			{int8(127), "127"},
			{int16(-32768), "-32768"},
			{int32(2147483647), "2147483647"},
			{int64(-9223372036854775808), "-9223372036854775808"},
			{uint(123), "123"},
			{uint8(255), "255"},
			{uint16(65535), "65535"},
			{uint32(4294967295), "4294967295"},
			{uint64(18446744073709551615), "18446744073709551615"},
			{float32(123.456), "123.45600128173828"},
			{float64(-987.654), "-987.654"},
			{float64(1.0), "1"},
			{bool(true), "true"},
			{bool(false), "false"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToString(testCase.inputValue)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%s' vs '%s' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToString(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToInt(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput int
			expectError    bool
		}{
			{"123", 123, false},
			{"-456", -456, false},
			{"0", 0, false},
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput int
		}{
			{int(123), 123},
			{int8(127), 127},
			{int16(-32768), -32768},
			{int32(2147483647), 2147483647},
			{int64(-123), -123},
			{uint(123), 123},
			{uint8(255), 255},
			{uint16(65535), 65535},
			{float32(123.45), 123},
			{float64(-987.65), -987},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt(testCase.inputValue)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToInt(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToInt8(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput int8
			expectError    bool
		}{
			{"127", 127, false},
			{"-128", -128, false},
			{"0", 0, false},
			{"128", 0, true},  // Overflow
			{"-129", 0, true}, // Underflow
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt8(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput int8
			expectError    bool
		}{
			{int(127), 127, false},
			{int8(127), 127, false},
			{int16(100), 100, false},
			{int32(-128), -128, false},
			{int64(0), 0, false},
			{uint(127), 127, false},
			{uint8(127), 127, false},
			{uint16(100), 100, false},
			{float32(100.45), 100, false},
			{float64(-100.65), -100, false},
			// Values outside int8 range
			{int(128), 0, true},
			{int16(32767), 0, true},
			{int32(-129), 0, true},
			{uint(255), 0, true},
			{float64(200.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt8(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToInt8(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToInt16(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput int16
			expectError    bool
		}{
			{"32767", 32767, false},
			{"-32768", -32768, false},
			{"0", 0, false},
			{"32768", 0, true},  // Overflow
			{"-32769", 0, true}, // Underflow
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt16(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput int16
			expectError    bool
		}{
			{int(32767), 32767, false},
			{int8(127), 127, false},
			{int16(32767), 32767, false},
			{int32(-32768), -32768, false},
			{int64(10000), 10000, false},
			{uint(32767), 32767, false},
			{uint8(255), 255, false},
			{uint16(32767), 32767, false},
			{float32(10000.45), 10000, false},
			{float64(-10000.65), -10000, false},
			// Values outside int16 range
			{int(32768), 0, true},
			{int32(2147483647), 0, true},
			{int64(-32769), 0, true},
			{uint(65535), 0, true},
			{float64(40000.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt16(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToInt16(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToInt32(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput int32
			expectError    bool
		}{
			{"2147483647", 2147483647, false},
			{"-2147483648", -2147483648, false},
			{"0", 0, false},
			{"2147483648", 0, true},  // Overflow
			{"-2147483649", 0, true}, // Underflow
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt32(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput int32
			expectError    bool
		}{
			{int(2147483647), 2147483647, false},
			{int8(127), 127, false},
			{int16(32767), 32767, false},
			{int32(-2147483648), -2147483648, false},
			{int64(1000000), 1000000, false},
			{uint(2147483647), 2147483647, false},
			{uint8(255), 255, false},
			{uint16(65535), 65535, false},
			{uint32(2147483647), 2147483647, false},
			{float32(1000000.45), 1000000, false},
			{float64(-1000000.65), -1000000, false},
			// Values outside int32 range
			{int64(2147483648), 0, true},
			{int64(-2147483649), 0, true},
			{uint64(4294967295), 0, true},
			{float64(3000000000.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt32(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToInt32(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToInt64(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput int64
			expectError    bool
		}{
			{"9223372036854775807", 9223372036854775807, false},
			{"-9223372036854775808", -9223372036854775808, false},
			{"0", 0, false},
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt64(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput int64
		}{
			{int(2147483647), 2147483647},
			{int8(127), 127},
			{int16(32767), 32767},
			{int32(-2147483648), -2147483648},
			{int64(9223372036854775807), 9223372036854775807},
			{int64(-9223372036854775808), -9223372036854775808},
			{uint(2147483647), 2147483647},
			{uint8(255), 255},
			{uint16(65535), 65535},
			{uint32(4294967295), 4294967295},
			{uint64(9223372036854775807), 9223372036854775807},
			{float32(1000000.45), 1000000},
			{float64(-1000000.65), -1000000},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToInt64(testCase.inputValue)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToInt64(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToUint8(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput uint8
			expectError    bool
		}{
			{"255", 255, false},
			{"0", 0, false},
			{"256", 0, true}, // Overflow
			{"-1", 0, true},  // Negative
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint8(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput uint8
			expectError    bool
		}{
			{int(255), 255, false},
			{int8(127), 127, false},
			{int16(100), 100, false},
			{int32(0), 0, false},
			{int64(200), 200, false},
			{uint(255), 255, false},
			{uint8(255), 255, false},
			{uint16(100), 100, false},
			{uint32(200), 200, false},
			{uint64(100), 100, false},
			{float32(100.45), 100, false},
			{float64(200.65), 200, false},
			// Values outside uint8 range or negative
			{int(256), 0, true},
			{int16(32767), 0, true},
			{int32(-1), 0, true},
			{int64(-100), 0, true},
			{uint(256), 0, true},
			{uint16(65535), 0, true},
			{float64(-0.1), 0, true},
			{float64(300.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint8(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToUint8(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToUint16(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput uint16
			expectError    bool
		}{
			{"65535", 65535, false},
			{"0", 0, false},
			{"65536", 0, true}, // Overflow
			{"-1", 0, true},    // Negative
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint16(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput uint16
			expectError    bool
		}{
			{int(65535), 65535, false},
			{int8(127), 127, false},
			{int16(32767), 32767, false},
			{int32(0), 0, false},
			{int64(10000), 10000, false},
			{uint(65535), 65535, false},
			{uint8(255), 255, false},
			{uint16(65535), 65535, false},
			{uint32(10000), 10000, false},
			{uint64(65535), 65535, false},
			{float32(10000.45), 10000, false},
			{float64(65535.0), 65535, false},
			// Values outside uint16 range or negative
			{int(65536), 0, true},
			{int32(-1), 0, true},
			{int64(-100), 0, true},
			{uint(65536), 0, true},
			{uint32(70000), 0, true},
			{float64(-0.1), 0, true},
			{float64(70000.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint16(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToUint16(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToUint32(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput uint32
			expectError    bool
		}{
			{"4294967295", 4294967295, false},
			{"0", 0, false},
			{"4294967296", 0, true}, // Overflow
			{"-1", 0, true},         // Negative
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint32(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput uint32
			expectError    bool
		}{
			{int(2147483647), 2147483647, false},
			{int8(127), 127, false},
			{int16(32767), 32767, false},
			{int32(2147483647), 2147483647, false},
			{int64(1000000), 1000000, false},
			{uint(4294967295), 4294967295, false},
			{uint8(255), 255, false},
			{uint16(65535), 65535, false},
			{uint32(4294967295), 4294967295, false},
			{uint64(4294967295), 4294967295, false},
			{float32(1000000.45), 1000000, false},
			{float64(4294967295.0), 4294967295, false},
			// Values outside uint32 range or negative
			{int32(-1), 0, true},
			{int64(-100), 0, true},
			{uint64(4294967296), 0, true},
			{float64(-0.1), 0, true},
			{float64(5000000000.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint32(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToUint32(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToUint64(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput uint64
			expectError    bool
		}{
			{"18446744073709551615", 18446744073709551615, false},
			{"0", 0, false},
			{"-1", 0, true}, // Negative
			{"invalid", 0, true},
			{"123.45", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint64(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput uint64
			expectError    bool
		}{
			{int(2147483647), 2147483647, false},
			{int8(127), 127, false},
			{int16(32767), 32767, false},
			{int32(2147483647), 2147483647, false},
			{int64(9223372036854775807), 9223372036854775807, false},
			{uint(4294967295), 4294967295, false},
			{uint8(255), 255, false},
			{uint16(65535), 65535, false},
			{uint32(4294967295), 4294967295, false},
			{uint64(18446744073709551615), 18446744073709551615, false},
			{float32(1000000.45), 1000000, false},
			{float64(9223372036854775807.0), 9223372036854775808, false}, // Float64 precision issue
			// Negative values
			{int32(-1), 0, true},
			{int64(-100), 0, true},
			{float64(-0.1), 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToUint64(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%d' vs '%d' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToUint64(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToFloat32(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput float32
			expectError    bool
		}{
			{"123.456", 123.456, false},
			{"-987.654", -987.654, false},
			{"0", 0, false},
			{"1e10", 1e10, false},
			{"invalid", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToFloat32(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%f' vs '%f' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput float32
		}{
			{int(123), 123.0},
			{int8(-127), -127.0},
			{int16(32767), 32767.0},
			{int32(-2147483648), -2147483648.0},
			{int64(1000000), 1000000.0},
			{uint(123), 123.0},
			{uint8(255), 255.0},
			{uint16(65535), 65535.0},
			{uint32(4294967295), 4294967295.0},
			{uint64(1000000), 1000000.0},
			{float32(123.456), 123.456},
			{float64(-987.654), -987.654},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToFloat32(testCase.inputValue)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			// Use a small epsilon for floating point comparison
			if math.Abs(float64(actualOutput-testCase.expectedOutput)) > 0.0001 {
				t.Errorf("UnexpectedOutputValue: '%f' vs '%f' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToFloat32(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}

func TestInterfaceToFloat64(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputString    string
			expectedOutput float64
			expectError    bool
		}{
			{"123.456789", 123.456789, false},
			{"-987.654321", -987.654321, false},
			{"0", 0, false},
			{"1e15", 1e15, false},
			{"invalid", 0, true},
			{"", 0, true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToFloat64(testCase.inputString)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%s]", testCase.inputString)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%s]", conversionErr.Error(), testCase.inputString)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%f' vs '%f' [%s]", actualOutput, testCase.expectedOutput, testCase.inputString)
			}
		}
	})

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput float64
		}{
			{int(123), 123.0},
			{int8(-127), -127.0},
			{int16(32767), 32767.0},
			{int32(-2147483648), -2147483648.0},
			{int64(9223372036854775807), 9223372036854775807.0},
			{uint(123), 123.0},
			{uint8(255), 255.0},
			{uint16(65535), 65535.0},
			{uint32(4294967295), 4294967295.0},
			{uint64(18446744073709551615), 18446744073709551615.0},
			{float32(123.456), 123.456},
			{float64(-987.654321), -987.654321},
			{float64(1.7976931348623157e+308), 1.7976931348623157e+308}, // Max float64
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := InterfaceToFloat64(testCase.inputValue)
			if conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			// Use a small epsilon for floating point comparison
			if math.Abs(actualOutput-testCase.expectedOutput) > 0.0001 {
				t.Errorf("UnexpectedOutputValue: '%f' vs '%f' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("UnsupportedInput", func(t *testing.T) {
		unsupportedInputs := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			nil,
			bool(true),
			bool(false),
		}
		for _, invalidInput := range unsupportedInputs {
			_, conversionErr := InterfaceToFloat64(invalidInput)
			if conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", invalidInput)
			}
		}
	})
}
