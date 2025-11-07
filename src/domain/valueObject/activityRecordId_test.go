package tkValueObject

import "testing"

func TestNewActivityRecordId(t *testing.T) {
	t.Run("ValidActivityRecordId", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"0", false},
			{int(0), false},
			{int8(0), false},
			{int16(0), false},
			{int32(0), false},
			{int64(0), false},
			{uint(0), false},
			{uint8(0), false},
			{uint16(0), false},
			{uint32(0), false},
			{uint64(0), false},
			{float32(0), false},
			{float64(0), false},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordId(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})

	t.Run("InvalidActivityRecordId", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue  any
			expectError bool
		}{
			{"-1", true},
			{int(-1), true},
			{int8(-1), true},
			{int16(-1), true},
			{int32(-1), true},
			{int64(-1), true},
			{float32(-1), true},
			{float64(-1), true},
		}

		for _, testCase := range testCaseStructs {
			_, err := NewActivityRecordId(testCase.inputValue)
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%v]", testCase.inputValue)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%v]", err.Error(), testCase.inputValue)
			}
		}
	})
}
