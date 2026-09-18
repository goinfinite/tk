package tkValueObject

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var cronPredefinedScheduleRegex = regexp.MustCompile(
	`^@?(?:annually|yearly|monthly|weekly|daily|hourly|reboot)$|` +
		`^@every (?:\d+(?:ns|us|µs|ms|s|m|h))+$`,
)

type CronSchedule string

func isValidCronStep(stepValue string) bool {
	stepNumber, err := strconv.ParseUint(stepValue, 10, 64)
	return err == nil && stepNumber >= 1
}

func isValidCronNumber(numberValue string, minValue, maxValue int) bool {
	number, err := strconv.ParseUint(numberValue, 10, 64)
	return err == nil && number >= uint64(minValue) && number <= uint64(maxValue)
}

func isValidCronFieldItem(fieldItem string, minValue, maxValue int) bool {
	if fieldItem == "*" {
		return true
	}

	baseValue, stepValue, hasStep := strings.Cut(fieldItem, "/")
	if hasStep && !isValidCronStep(stepValue) {
		return false
	}
	if baseValue == "*" {
		return true
	}

	startValue, endValue, hasRange := strings.Cut(baseValue, "-")
	if !isValidCronNumber(startValue, minValue, maxValue) {
		return false
	}

	return !hasRange || isValidCronNumber(endValue, minValue, maxValue)
}

func isValidCronField(fieldValue string, minValue, maxValue int) bool {
	for fieldItem := range strings.SplitSeq(fieldValue, ",") {
		if !isValidCronFieldItem(fieldItem, minValue, maxValue) {
			return false
		}
	}

	return true
}

// A schedule holds five space-separated fields: minute, hour, day, month, weekday.
// A field is a comma-separated list of items (isValidCronFieldItem).
// An item is a number (N), a range (N-M), or a step (*/S, N/S, N-M/S).
// A number (isValidCronNumber) is an unsigned value inside the field range.
// A step (isValidCronStep) is an interval of one or more.
// The grammar follows the standard crontab(5) format.
func NewCronSchedule(value any) (cronSchedule CronSchedule, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return cronSchedule, errors.New("CronScheduleMustBeString")
	}

	if len(stringValue) > 256 {
		return cronSchedule, errors.New("CronScheduleTooBig")
	}

	if cronPredefinedScheduleRegex.MatchString(stringValue) {
		if !strings.HasPrefix(stringValue, "@") {
			stringValue = "@" + stringValue
		}
		return CronSchedule(stringValue), nil
	}

	fieldValues := strings.Split(stringValue, " ")
	if len(fieldValues) != 5 {
		return cronSchedule, errors.New("InvalidCronSchedule")
	}

	minuteField := fieldValues[0]
	hourField := fieldValues[1]
	dayField := fieldValues[2]
	monthField := fieldValues[3]
	weekdayField := fieldValues[4]

	fieldSpecs := []struct {
		value    string
		minValue int
		maxValue int
	}{
		{minuteField, 0, 59},
		{hourField, 0, 23},
		{dayField, 1, 31},
		{monthField, 1, 12},
		{weekdayField, 0, 7},
	}

	for _, fieldSpec := range fieldSpecs {
		if !isValidCronField(fieldSpec.value, fieldSpec.minValue, fieldSpec.maxValue) {
			return cronSchedule, errors.New("InvalidCronSchedule")
		}
	}

	return CronSchedule(stringValue), nil
}

func (vo CronSchedule) String() string {
	return string(vo)
}
