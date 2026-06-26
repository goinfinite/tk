package tkUseCase

import (
	"time"

	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func ReadHoneypotBanDecision(
	queryRepo tkRepository.HoneypotQueryRepo,
	requesterIp tkValueObject.IpAddress,
	banDuration tkValueObject.HoneypotBanDuration,
	aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
) (int, error) {
	if queryRepo == nil {
		return 0, ErrNilHoneypotQueryRepo
	}

	existentHitData, readErr := queryRepo.ReadHitRecord(
		requesterIp,
	)
	if readErr != nil {
		return 0, readErr
	}

	firstHitAt, timeParseErr := time.Parse(
		time.RFC3339, existentHitData.FirstHitAt,
	)
	if timeParseErr != nil {
		return 0, timeParseErr
	}

	if time.Since(firstHitAt) > banDuration.Duration() {
		return 0, nil
	}

	return aggressivenessMode.ResolveTier(
		existentHitData.Count,
	), nil
}
