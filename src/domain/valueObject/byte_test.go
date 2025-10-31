package tkValueObject

import (
	"testing"
)

func TestNewByte(t *testing.T) {
	t.Run("NewByte", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Byte
			expectError    bool
		}{
			{0, Byte(0), false},
			{1024, Byte(1024), false},
			{1048576, Byte(1048576), false},
			{"1024", Byte(1024), false},
			{int64(2048), Byte(2048), false},
			// Invalid inputs
			{"invalid", Byte(0), true},
			{true, Byte(0), true},
			{[]string{"1024"}, Byte(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewByte(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("NewKibibyte", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Byte
			expectError    bool
		}{
			{0, Byte(0), false},
			{1, Byte(1024), false},
			{2, Byte(2048), false},
			{"1", Byte(1024), false},
			{int64(3), Byte(3072), false},
			// Invalid inputs
			{"invalid", Byte(0), true},
			{true, Byte(0), true},
			{[]string{"1"}, Byte(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewKibibyte(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("NewMebibyte", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Byte
			expectError    bool
		}{
			{0, Byte(0), false},
			{1, Byte(1048576), false},
			{2, Byte(2097152), false},
			{"1", Byte(1048576), false},
			{int64(3), Byte(3145728), false},
			// Invalid inputs
			{"invalid", Byte(0), true},
			{true, Byte(0), true},
			{[]string{"1"}, Byte(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewMebibyte(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("NewGibibyte", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Byte
			expectError    bool
		}{
			{0, Byte(0), false},
			{1, Byte(1073741824), false},
			{2, Byte(2147483648), false},
			{"1", Byte(1073741824), false},
			{int64(3), Byte(3221225472), false},
			// Invalid inputs
			{"invalid", Byte(0), true},
			{true, Byte(0), true},
			{[]string{"1"}, Byte(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewGibibyte(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("NewTebibyte", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput Byte
			expectError    bool
		}{
			{0, Byte(0), false},
			{1, Byte(1099511627776), false},
			{2, Byte(2199023255552), false},
			{"1", Byte(1099511627776), false},
			{int64(3), Byte(3298534883328), false},
			// Invalid inputs
			{"invalid", Byte(0), true},
			{true, Byte(0), true},
			{[]string{"1"}, Byte(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewTebibyte(testCase.inputValue)
			if testCase.expectError && conversionErr == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && conversionErr != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", conversionErr.Error(), testCase.inputValue)
			}
			if !testCase.expectError && actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("Int64Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput int64
		}{
			{Byte(0), 0},
			{Byte(1024), 1024},
			{Byte(1048576), 1048576},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Int64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("Uint64Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput uint64
		}{
			{Byte(0), 0},
			{Byte(1024), 1024},
			{Byte(1048576), 1048576},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Uint64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("Float64Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput float64
		}{
			{Byte(0), 0.0},
			{Byte(1024), 1024.0},
			{Byte(1048576), 1048576.0},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Float64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ToKiBMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput int64
		}{
			{Byte(0), 0},
			{Byte(1024), 1},
			{Byte(2048), 2},
			{Byte(1536), 2}, // Rounded up
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ToKiB()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ToMiBMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput int64
		}{
			{Byte(0), 0},
			{Byte(1048576), 1},
			{Byte(2097152), 2},
			{Byte(1572864), 2}, // Rounded up
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ToMiB()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ToGiBMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput int64
		}{
			{Byte(0), 0},
			{Byte(1073741824), 1},
			{Byte(2147483648), 2},
			{Byte(1610612736), 2}, // Rounded up
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ToGiB()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ToTiBMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput int64
		}{
			{Byte(0), 0},
			{Byte(1099511627776), 1},
			{Byte(2199023255552), 2},
			{Byte(1649267441664), 2}, // Rounded up
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ToTiB()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput string
		}{
			{Byte(0), "0"},
			{Byte(1024), "1024"},
			{Byte(1048576), "1048576"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("StringWithSuffixMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     Byte
			expectedOutput string
		}{
			{Byte(0), "0 B"},
			{Byte(512), "512 B"},      // < 1024
			{Byte(1024), "1 KiB"},     // >= 1024
			{Byte(1048576), "1 MiB"},  // >= 1048576
			{Byte(1073741824), "1 GiB"}, // >= 1073741824
			{Byte(1099511627776), "1 TiB"}, // >= 1099511627776
			{Byte(1125899906842624), "1125899906842624 B"}, // >= 1125899906842624
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.StringWithSuffix()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
