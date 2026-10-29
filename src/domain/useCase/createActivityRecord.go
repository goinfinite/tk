package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

// CreateActivityRecord persists an activity record as a non-blocking side effect.
// Errors are logged but not returned so the caller's primary operation succeeds.
func CreateActivityRecord(
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	createDto tkDto.CreateActivityRecord,
) {
	err := activityRecordCmdRepo.Create(createDto)
	if err != nil {
		logAttributes := []any{
			slog.String("error", err.Error()),
		}
		logAttributes = append(
			logAttributes,
			activityRecordLogAttributesBuilder(activityRecordLogAttributesInput{
				RecordLevel:       &createDto.RecordLevel,
				RecordCode:        &createDto.RecordCode,
				AffectedResources: createDto.AffectedResources,
				OperatorSri:       createDto.OperatorSri,
				OperatorIpAddress: createDto.OperatorIpAddress,
			})...,
		)
		slog.Error("CreateActivityRecordInfraError", logAttributes...)
	}
}
