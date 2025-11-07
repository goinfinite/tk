package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

const (
	ErrActivityRecordNotFound string = "ActivityRecordNotFound"
)

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
		slog.Error("ReadActivityRecordsInfraError", slog.String("err", err.Error()))
	}

	return responseDto, err
}
