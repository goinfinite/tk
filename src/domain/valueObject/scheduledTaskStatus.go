package tkValueObject

import (
	"errors"
	"slices"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type ScheduledTaskStatus string

const (
	ScheduledTaskStatusPending   ScheduledTaskStatus = "pending"
	ScheduledTaskStatusRunning   ScheduledTaskStatus = "running"
	ScheduledTaskStatusCompleted ScheduledTaskStatus = "completed"
	ScheduledTaskStatusFailed    ScheduledTaskStatus = "failed"
	ScheduledTaskStatusCancelled ScheduledTaskStatus = "cancelled"
	ScheduledTaskStatusTimeout   ScheduledTaskStatus = "timeout"
)

var ValidScheduledTaskStatuses = []string{
	ScheduledTaskStatusPending.String(), ScheduledTaskStatusRunning.String(),
	ScheduledTaskStatusCompleted.String(), ScheduledTaskStatusFailed.String(),
	ScheduledTaskStatusCancelled.String(), ScheduledTaskStatusTimeout.String(),
}

func NewScheduledTaskStatus(value any) (
	scheduledTaskStatus ScheduledTaskStatus, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return scheduledTaskStatus, errors.New("ScheduledTaskStatusMustBeString")
	}

	stringValue = strings.TrimSpace(stringValue)
	stringValue = strings.ToLower(stringValue)

	if !slices.Contains(ValidScheduledTaskStatuses, stringValue) {
		return scheduledTaskStatus, errors.New("InvalidScheduledTaskStatus")
	}

	return ScheduledTaskStatus(stringValue), nil
}

func (vo ScheduledTaskStatus) String() string {
	return string(vo)
}
