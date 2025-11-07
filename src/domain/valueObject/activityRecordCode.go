package tkValueObject

import (
	"errors"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var activityRecordRegex = regexp.MustCompile(`^[a-zA-Z]\w{2,127}$`)

type ActivityRecordCode string

func NewActivityRecordCode(value any) (recordCode ActivityRecordCode, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return recordCode, errors.New("ActivityRecordCodeMustBeString")
	}

	if !activityRecordRegex.MatchString(stringValue) {
		return recordCode, errors.New("InvalidActivityRecordCode")
	}

	return ActivityRecordCode(stringValue), nil
}

func (vo ActivityRecordCode) String() string {
	return string(vo)
}
