package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	ActivityRecordLevelDebug ActivityRecordLevel = "DEBUG"
	ActivityRecordLevelInfo  ActivityRecordLevel = "INFO"
	ActivityRecordLevelWarn  ActivityRecordLevel = "WARN"
	ActivityRecordLevelError ActivityRecordLevel = "ERROR"
	ActivityRecordLevelSec   ActivityRecordLevel = "SEC"
)

type ActivityRecordLevel string

func NewActivityRecordLevel(value any) (recordLevel ActivityRecordLevel, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return recordLevel, errors.New("ActivityRecordLevelMustBeString")
	}
	stringValue = strings.ToUpper(stringValue)

	stringValueVo := ActivityRecordLevel(stringValue)
	switch stringValueVo {
	case ActivityRecordLevelDebug, ActivityRecordLevelInfo, ActivityRecordLevelWarn,
		ActivityRecordLevelError, ActivityRecordLevelSec:
		return stringValueVo, nil
	case "SECURITY":
		return ActivityRecordLevelSec, nil
	case "WARNING":
		return ActivityRecordLevelWarn, nil
	default:
		return stringValueVo, errors.New("InvalidActivityRecordLevel")
	}
}

func (vo ActivityRecordLevel) String() string {
	return string(vo)
}
