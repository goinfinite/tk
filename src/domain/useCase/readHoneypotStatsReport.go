package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

func ReadHoneypotStatsReport(
	honeypotQueryRepo tkRepository.HoneypotQueryRepo,
	requestDto tkDto.ReadHoneypotStatsReportRequest,
) (tkDto.ReadHoneypotStatsReportResponse, error) {
	responseDto, err := honeypotQueryRepo.ReadStatsReport(requestDto)
	if err != nil {
		slog.Error(
			"ReadHoneypotStatsReportInfraError",
			slog.String("err", err.Error()),
		)
		return responseDto, err
	}

	return responseDto, nil
}
