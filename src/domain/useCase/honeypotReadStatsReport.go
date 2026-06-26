package tkUseCase

import (
	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func ReadHoneypotStatsReport(
	queryRepo tkRepository.HoneypotQueryRepo,
	banDuration tkValueObject.HoneypotBanDuration,
	aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
) (tkDto.HoneypotStatsReport, error) {
	if queryRepo == nil {
		return tkDto.HoneypotStatsReport{}, nil
	}

	return queryRepo.ReadReport(
		banDuration.Duration(), aggressivenessMode,
	)
}
