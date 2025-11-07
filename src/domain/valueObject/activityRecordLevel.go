package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var (
	ActivityRecordLevelDebug    ActivityRecordLevel = "DEBUG"
	ActivityRecordLevelInfo     ActivityRecordLevel = "INFO"
	ActivityRecordLevelWarning  ActivityRecordLevel = "WARNING"
	ActivityRecordLevelError    ActivityRecordLevel = "ERROR"
	ActivityRecordLevelSecurity ActivityRecordLevel = "SECURITY"
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
	case ActivityRecordLevelDebug, ActivityRecordLevelInfo, ActivityRecordLevelWarning,
		ActivityRecordLevelError, ActivityRecordLevelSecurity:
		return stringValueVo, nil
	case "SEC":
		return ActivityRecordLevelSecurity, nil
	case "WARN":
		return ActivityRecordLevelWarning, nil
	default:
		return stringValueVo, errors.New("InvalidActivityRecordLevel")
	}
}

func (vo ActivityRecordLevel) String() string {
	return string(vo)
}
