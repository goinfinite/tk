package tkUseCase

import (
	"errors"
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

func DeleteActivityRecord(
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	deleteDto tkDto.DeleteActivityRecord,
) error {
	err := activityRecordCmdRepo.Delete(deleteDto)
	if err != nil {
		logAttributes := []any{
			slog.String("error", err.Error()),
		}
		logAttributes = append(
			logAttributes,
			activityRecordLogAttributesBuilder(activityRecordLogAttributesInput{
				RecordId:          deleteDto.RecordId,
				RecordLevel:       deleteDto.RecordLevel,
				RecordCode:        deleteDto.RecordCode,
				AffectedResources: deleteDto.AffectedResources,
				OperatorSri:       deleteDto.OperatorSri,
				OperatorIpAddress: deleteDto.OperatorIpAddress,
				CreatedBeforeAt:   deleteDto.CreatedBeforeAt,
				CreatedAfterAt:    deleteDto.CreatedAfterAt,
			})...,
		)
		slog.Error("DeleteActivityRecordError", logAttributes...)
		return errors.New("DeleteActivityRecordInfraError")
	}

	return nil
}
