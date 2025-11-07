package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

func CreateActivityRecord(
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	createDto tkDto.CreateActivityRecord,
) {
	err := activityRecordCmdRepo.Create(createDto)
	if err != nil {
		slog.Error(
			"CreateActivityRecordInfraError",
			slog.String("err", err.Error()),
			slog.Any("createDto", createDto),
		)
	}
}
