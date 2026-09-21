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

var validScheduledTaskStatuses = []ScheduledTaskStatus{
	ScheduledTaskStatusPending, ScheduledTaskStatusRunning,
	ScheduledTaskStatusCompleted, ScheduledTaskStatusFailed,
	ScheduledTaskStatusCancelled, ScheduledTaskStatusTimeout,
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

	status := ScheduledTaskStatus(stringValue)
	if !slices.Contains(validScheduledTaskStatuses, status) {
		return scheduledTaskStatus, errors.New("InvalidScheduledTaskStatus")
	}

	return status, nil
}

func (vo ScheduledTaskStatus) String() string {
	return string(vo)
}
