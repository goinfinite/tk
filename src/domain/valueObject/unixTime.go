package tkValueObject

import (
	"errors"
	"strconv"
	"time"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type UnixTime int64

func NewUnixTime(value any) (unixTime UnixTime, err error) {
	intValue, err := tkVoUtil.InterfaceToInt64(value)
	if err != nil {
		return unixTime, errors.New("UnixTimeMustBeInt64")
	}

	return UnixTime(intValue), nil
}

func NewUnixTimeNow() UnixTime {
	return UnixTime(time.Now().UTC().Unix())
}

func NewUnixTimeBeforeNow(duration time.Duration) UnixTime {
	return UnixTime(time.Now().Add(-duration).UTC().Unix())
}

func NewUnixTimeAfterNow(duration time.Duration) UnixTime {
	return UnixTime(time.Now().Add(duration).UTC().Unix())
}

func NewUnixTimeWithGoTime(goTime time.Time) UnixTime {
	return UnixTime(goTime.UTC().Unix())
}

func (vo UnixTime) Int64() int64 {
	return time.Unix(int64(vo), 0).UTC().Unix()
}

func (vo UnixTime) ReadRfcDate() string {
	return time.Unix(int64(vo), 0).UTC().Format(time.RFC3339)
}

func (vo UnixTime) ReadDateOnly() string {
	return time.Unix(int64(vo), 0).UTC().Format("2006-01-02")
}

func (vo UnixTime) ReadTimeOnly() string {
	return time.Unix(int64(vo), 0).UTC().Format("15:04:05")
}

func (ut UnixTime) ReadDateTime() string {
	return time.Unix(int64(ut), 0).UTC().Format("02/01/2006 15:04:05")
}

func (ut UnixTime) ReadStartOfDay() time.Time {
	return time.Unix(int64(ut), 0).UTC().Truncate(24 * time.Hour)
}

func (ut UnixTime) ReadEndOfDay() time.Time {
	return ut.ReadStartOfDay().Add(24 * time.Hour)
}

func (vo UnixTime) ReadAsGoTime() time.Time {
	return time.Unix(int64(vo), 0).UTC()
}

func (vo UnixTime) IsPast() bool {
	return vo.ReadAsGoTime().Before(time.Now().UTC())
}

func (vo UnixTime) IsFuture() bool {
	return vo.ReadAsGoTime().After(time.Now().UTC())
}

func (vo UnixTime) IsBetween(startDate, endDate UnixTime) bool {
	voGoTime := vo.ReadAsGoTime()
	startDateGoTime := startDate.ReadAsGoTime()
	endDateGoTime := endDate.ReadAsGoTime()
	if voGoTime.Equal(startDateGoTime) || voGoTime.Equal(endDateGoTime) {
		return true
	}

	if startDateGoTime.After(endDateGoTime) {
		startDateGoTime, endDateGoTime = endDateGoTime, startDateGoTime
	}

	isBeforeStartDate := voGoTime.Before(startDateGoTime)
	isAfterEndDate := voGoTime.After(endDateGoTime)
	if isBeforeStartDate || isAfterEndDate {
		return false
	}

	return true
}

func (vo UnixTime) String() string {
	return strconv.FormatInt(int64(vo), 10)
}
