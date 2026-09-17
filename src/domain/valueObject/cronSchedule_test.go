package tkValueObject

import "testing"

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
			{"@daily", CronSchedule("@daily"), false},
			{"daily", CronSchedule("@daily"), false},
			{"@every 5m", CronSchedule("@every 5m"), false},
			{"@reboot", CronSchedule("@reboot"), false},
			// Invalid schedules
			{"", CronSchedule(""), true},
			{"not a cron", CronSchedule(""), true},
			{"0 0 * *", CronSchedule(""), true},
			{"every 5m", CronSchedule(""), true},
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
