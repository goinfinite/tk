package tkUseCase

import (
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func ReadHoneypotStatsReport(
	queryRepo tkRepository.HoneypotQueryRepo,
	banDuration time.Duration,
	aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
) (tkDto.HoneypotStatsReport, error) {
	if queryRepo == nil {
		return tkDto.HoneypotStatsReport{}, nil
	}

	return queryRepo.ReadReport(
		banDuration, aggressivenessMode,
	)
}
