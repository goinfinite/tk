package tkUseCase

import (
	"errors"
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

var ErrActivityRecordNotFound = errors.New("ActivityRecordNotFound")

var ActivityRecordsDefaultPagination = tkDto.Pagination{
	PageNumber:   0,
	ItemsPerPage: 10,
}

func ReadActivityRecords(
	activityRecordQueryRepo tkRepository.ActivityRecordQueryRepo,
	requestDto tkDto.ReadActivityRecordsRequest,
) (responseDto tkDto.ReadActivityRecordsResponse, err error) {
	responseDto, err = activityRecordQueryRepo.Read(requestDto)
	if err != nil {
		logAttributes := []any{
			slog.String("error", err.Error()),
		}
		logAttributes = append(
			logAttributes,
			activityRecordLogAttributesBuilder(activityRecordLogAttributesInput{
				Pagination:        &requestDto.Pagination,
				RecordId:          requestDto.RecordId,
				RecordLevel:       requestDto.RecordLevel,
				RecordCode:        requestDto.RecordCode,
				AffectedResources: requestDto.AffectedResources,
				OperatorSri:       requestDto.OperatorSri,
				OperatorIpAddress: requestDto.OperatorIpAddress,
				CreatedBeforeAt:   requestDto.CreatedBeforeAt,
				CreatedAfterAt:    requestDto.CreatedAfterAt,
			})...,
		)
		slog.Error("ReadActivityRecordsInfraError", logAttributes...)
		if errors.Is(err, ErrActivityRecordNotFound) {
			return responseDto, ErrActivityRecordNotFound
		}
		return responseDto, errors.New("ReadActivityRecordsInfraError")
	}

	return responseDto, nil
}
