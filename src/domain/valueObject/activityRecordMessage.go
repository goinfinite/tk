package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

// ActivityRecordMessage is a string that represents the message of an activity record.
// Since it's supposed to be used by the system itself, it's not validated as much as
// other value objects on purpose. There is only a length limit of 2048 characters.
type ActivityRecordMessage string

func NewActivityRecordMessage(value any) (recordMessage ActivityRecordMessage, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return recordMessage, errors.New("ActivityRecordMessageMustBeString")
	}

	if len(stringValue) > 2048 {
		stringValue = stringValue[:2048]
	}

	return ActivityRecordMessage(stringValue), nil
}

func (vo ActivityRecordMessage) String() string {
	return string(vo)
}
