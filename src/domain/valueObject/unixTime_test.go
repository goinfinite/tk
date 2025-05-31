package tkValueObject

import (
	"testing"
	"time"
)

func TestNewUnixTime(t *testing.T) {
	t.Run("StringInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixTime
			expectError    bool
		}{
			{"0", UnixTime(0), false},
			{"1234567890", UnixTime(1234567890), false},
			{"-1", UnixTime(-1), false},
			{"9223372036854775807", UnixTime(9223372036854775807), false}, // max int64
			// Invalid string inputs
			{"invalid", UnixTime(0), true},
			{"123.45", UnixTime(0), true},
			{"", UnixTime(0), true},
			{"abc123", UnixTime(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixTime(testCase.inputValue)
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

	t.Run("NumericInput", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput UnixTime
			expectError    bool
		}{
			{int(0), UnixTime(0), false},
			{int8(127), UnixTime(127), false},
			{int16(-32768), UnixTime(-32768), false},
			{int32(2147483647), UnixTime(2147483647), false},
			{int64(1234567890), UnixTime(1234567890), false},
			{uint(123), UnixTime(123), false},
			{uint8(255), UnixTime(255), false},
			{uint16(65535), UnixTime(65535), false},
			{uint32(4294967295), UnixTime(4294967295), false},
			{uint64(1234567890), UnixTime(1234567890), false},
			{float32(123.45), UnixTime(123), false},
			{float64(987.65), UnixTime(987), false},
			// Invalid inputs
			{true, UnixTime(0), true},
			{[]string{"123"}, UnixTime(0), true},
			{nil, UnixTime(0), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewUnixTime(testCase.inputValue)
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

	t.Run("StringMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedOutput string
		}{
			{UnixTime(0), "0"},
			{UnixTime(1234567890), "1234567890"},
			{UnixTime(-1), "-1"},
			{UnixTime(9223372036854775807), "9223372036854775807"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("Int64Method", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedOutput int64
		}{
			{UnixTime(0), 0},
			{UnixTime(1234567890), 1234567890},
			{UnixTime(-1), -1},
			{UnixTime(9223372036854775807), 9223372036854775807},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.Int64()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadRfcDateMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedOutput string
		}{
			{UnixTime(0), "1970-01-01T00:00:00Z"},
			{UnixTime(1234567890), "2009-02-13T23:31:30Z"},
			{UnixTime(1609459200), "2021-01-01T00:00:00Z"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadRfcDate()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadDateOnlyMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedOutput string
		}{
			{UnixTime(0), "1970-01-01"},
			{UnixTime(1234567890), "2009-02-13"},
			{UnixTime(1609459200), "2021-01-01"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadDateOnly()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadTimeOnlyMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedOutput string
		}{
			{UnixTime(0), "00:00:00"},
			{UnixTime(1234567890), "23:31:30"},
			{UnixTime(1609459200), "00:00:00"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadTimeOnly()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("ReadDateTimeMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedOutput string
		}{
			{UnixTime(0), "01/01/1970 00:00:00"},
			{UnixTime(1234567890), "13/02/2009 23:31:30"},
			{UnixTime(1609459200), "01/01/2021 00:00:00"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadDateTime()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})

	t.Run("IsPastMethod", func(t *testing.T) {
		pastTime := UnixTime(1234567890)   // 2009-02-13
		futureTime := UnixTime(4102444800) // 2100-01-01

		if !pastTime.IsPast() {
			t.Errorf("UnexpectedOutputValue: expected true for past time [%v]", pastTime)
		}

		if futureTime.IsPast() {
			t.Errorf("UnexpectedOutputValue: expected false for future time [%v]", futureTime)
		}
	})

	t.Run("IsFutureMethod", func(t *testing.T) {
		pastTime := UnixTime(1234567890)   // 2009-02-13
		futureTime := UnixTime(4102444800) // 2100-01-01

		if pastTime.IsFuture() {
			t.Errorf("UnexpectedOutputValue: expected false for past time [%v]", pastTime)
		}

		if !futureTime.IsFuture() {
			t.Errorf("UnexpectedOutputValue: expected true for future time [%v]", futureTime)
		}
	})

	t.Run("ReadAsGoTimeMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue    UnixTime
			expectedYear  int
			expectedMonth int
			expectedDay   int
		}{
			{UnixTime(0), 1970, 1, 1},
			{UnixTime(1234567890), 2009, 2, 13},
			{UnixTime(1609459200), 2021, 1, 1},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadAsGoTime()
			if actualOutput.Year() != testCase.expectedYear ||
				int(actualOutput.Month()) != testCase.expectedMonth ||
				actualOutput.Day() != testCase.expectedDay {
				t.Errorf("UnexpectedOutputValue: expected %d-%02d-%02d, got %d-%02d-%02d [%v]",
					testCase.expectedYear, testCase.expectedMonth, testCase.expectedDay,
					actualOutput.Year(), int(actualOutput.Month()), actualOutput.Day(),
					testCase.inputValue)
			}
		}
	})

	t.Run("ReadStartOfDayMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedHour   int
			expectedMinute int
			expectedSecond int
		}{
			{UnixTime(1234567890), 0, 0, 0}, // Should be start of day
			{UnixTime(1609459200), 0, 0, 0}, // Should be start of day
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadStartOfDay()
			if actualOutput.Hour() != testCase.expectedHour ||
				actualOutput.Minute() != testCase.expectedMinute ||
				actualOutput.Second() != testCase.expectedSecond {
				t.Errorf("UnexpectedOutputValue: expected %02d:%02d:%02d, got %02d:%02d:%02d [%v]",
					testCase.expectedHour, testCase.expectedMinute, testCase.expectedSecond,
					actualOutput.Hour(), actualOutput.Minute(), actualOutput.Second(),
					testCase.inputValue)
			}
		}
	})

	t.Run("ReadEndOfDayMethod", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     UnixTime
			expectedHour   int
			expectedMinute int
			expectedSecond int
		}{
			{UnixTime(1234567890), 0, 0, 0}, // Should be start of next day
			{UnixTime(1609459200), 0, 0, 0}, // Should be start of next day
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.ReadEndOfDay()
			if actualOutput.Hour() != testCase.expectedHour ||
				actualOutput.Minute() != testCase.expectedMinute ||
				actualOutput.Second() != testCase.expectedSecond {
				t.Errorf("UnexpectedOutputValue: expected %02d:%02d:%02d, got %02d:%02d:%02d [%v]",
					testCase.expectedHour, testCase.expectedMinute, testCase.expectedSecond,
					actualOutput.Hour(), actualOutput.Minute(), actualOutput.Second(),
					testCase.inputValue)
			}
		}
	})

	t.Run("NewUnixTimeNowFactory", func(t *testing.T) {
		beforeTime := time.Now().Unix()
		unixTimeNow := NewUnixTimeNow()
		afterTime := time.Now().Unix()

		actualTime := unixTimeNow.Int64()
		if actualTime < beforeTime || actualTime > afterTime {
			t.Errorf("UnexpectedOutputValue: time %d should be between %d and %d", actualTime, beforeTime, afterTime)
		}
	})

	t.Run("NewUnixTimeBeforeNowFactory", func(t *testing.T) {
		duration := time.Hour
		beforeTime := time.Now().Add(-duration).Unix()
		unixTimeBefore := NewUnixTimeBeforeNow(duration)
		afterTime := time.Now().Add(-duration).Unix()

		actualTime := unixTimeBefore.Int64()
		if actualTime < beforeTime-1 || actualTime > afterTime+1 {
			t.Errorf("UnexpectedOutputValue: time %d should be approximately %d", actualTime, beforeTime)
		}
	})

	t.Run("NewUnixTimeAfterNowFactory", func(t *testing.T) {
		duration := time.Hour
		beforeTime := time.Now().Add(duration).Unix()
		unixTimeAfter := NewUnixTimeAfterNow(duration)
		afterTime := time.Now().Add(duration).Unix()

		actualTime := unixTimeAfter.Int64()
		if actualTime < beforeTime-1 || actualTime > afterTime+1 {
			t.Errorf("UnexpectedOutputValue: time %d should be approximately %d", actualTime, beforeTime)
		}
	})

	t.Run("NewUnixTimeWithGoTimeFactory", func(t *testing.T) {
		testCaseStructs := []struct {
			inputTime      time.Time
			expectedOutput UnixTime
		}{
			{time.Unix(0, 0), UnixTime(0)},
			{time.Unix(1234567890, 0), UnixTime(1234567890)},
			{time.Unix(1609459200, 0), UnixTime(1609459200)},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := NewUnixTimeWithGoTime(testCase.inputTime)
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputTime)
			}
		}
	})
}
