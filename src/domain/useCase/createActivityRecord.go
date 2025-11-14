package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

// CreateActivityRecord persists an activity record as a non-blocking side effect.
// Errors are logged but not returned to avoid failing the caller's primary operation.
func CreateActivityRecord(
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	createDto tkDto.CreateActivityRecord,
) {
	err := activityRecordCmdRepo.Create(createDto)
	if err != nil {
		slog.Error(
			"CreateActivityRecordInfraError",
			slog.String("err", err.Error()),
			slog.String("recordCode", createDto.RecordCode.String()),
		)
	}
}
