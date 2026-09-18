package tkValueObject

import (
	"strings"
	"testing"
)

func TestNewCronSchedule(t *testing.T) {
	t.Run("ValidCronSchedule", func(t *testing.T) {
		testCaseStructs := []struct {
			inputValue     any
			expectedOutput CronSchedule
			expectError    bool
		}{
			{"0 0 * * *", CronSchedule("0 0 * * *"), false},
			{"*/5 * * * *", CronSchedule("*/5 * * * *"), false},
			{"0 9-17 * * 1-5", CronSchedule("0 9-17 * * 1-5"), false},
			{"1,2,3 * * * *", CronSchedule("1,2,3 * * * *"), false},
			{"0 0 1 1 0", CronSchedule("0 0 1 1 0"), false},
			{"0 0 31 12 7", CronSchedule("0 0 31 12 7"), false},
			{"5/2 * * * *", CronSchedule("5/2 * * * *"), false},
			{"1-5/2 * * * *", CronSchedule("1-5/2 * * * *"), false},
			{"1,2-3 * * * *", CronSchedule("1,2-3 * * * *"), false},
			{"*/5,10 * * * *", CronSchedule("*/5,10 * * * *"), false},
			{"59 * * * *", CronSchedule("59 * * * *"), false},
			{"0 23 * * *", CronSchedule("0 23 * * *"), false},
			{"@every 5m30s", CronSchedule("@every 5m30s"), false},
			{"@daily", CronSchedule("@daily"), false},
			{"daily", CronSchedule("@daily"), false},
			{"@every 5m", CronSchedule("@every 5m"), false},
			{"@reboot", CronSchedule("@reboot"), false},
			// Invalid schedules
			{"", CronSchedule(""), true},
			{"not a cron", CronSchedule(""), true},
			{"0 0 * *", CronSchedule(""), true},
			{"0  0 * * *", CronSchedule(""), true},
			{"every 5m", CronSchedule(""), true},
			{"60 * * * *", CronSchedule(""), true},
			{"0 24 * * *", CronSchedule(""), true},
			{"0 0 32 * *", CronSchedule(""), true},
			{"0 0 0 * *", CronSchedule(""), true},
			{"0 0 * 13 *", CronSchedule(""), true},
			{"0 0 * 0 *", CronSchedule(""), true},
			{"0 0 * * 8", CronSchedule(""), true},
			{"*/0 * * * *", CronSchedule(""), true},
			{"1-5/0 * * * *", CronSchedule(""), true},
			{"+5 * * * *", CronSchedule(""), true},
			{"*/+5 * * * *", CronSchedule(""), true},
			{"1-+5 * * * *", CronSchedule(""), true},
			{"1-5/+2 * * * *", CronSchedule(""), true},
			{"+1-5 * * * *", CronSchedule(""), true},
			{"1,,2 * * * *", CronSchedule(""), true},
			{"1, * * * *", CronSchedule(""), true},
			{",1 * * * *", CronSchedule(""), true},
			{strings.Repeat("1,", 130) + "1 0 * * *", CronSchedule(""), true},
			{123, CronSchedule(""), true},
			{[]string{"0 0 * * *"}, CronSchedule(""), true},
		}

		for _, testCase := range testCaseStructs {
			actualOutput, conversionErr := NewCronSchedule(testCase.inputValue)
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
			inputValue     CronSchedule
			expectedOutput string
		}{
			{CronSchedule("0 0 * * *"), "0 0 * * *"},
			{CronSchedule("@daily"), "@daily"},
		}

		for _, testCase := range testCaseStructs {
			actualOutput := testCase.inputValue.String()
			if actualOutput != testCase.expectedOutput {
				t.Errorf("UnexpectedOutputValue: '%v' vs '%v' [%v]", actualOutput, testCase.expectedOutput, testCase.inputValue)
			}
		}
	})
}
