package tkUseCase

import (
	"errors"
	"time"

	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func ReadHoneypotBanDecision(
	queryRepo tkRepository.HoneypotQueryRepo,
	requesterIp tkValueObject.IpAddress,
	banDuration time.Duration,
	aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
) (int, error) {
	if queryRepo == nil {
		return 0, errors.New("NilQueryRepo")
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

	if time.Since(firstHitAt) > banDuration {
		return 0, nil
	}

	return aggressivenessMode.ResolveTier(
		existentHitData.Count,
	), nil
}
