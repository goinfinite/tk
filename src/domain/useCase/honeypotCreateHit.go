package tkUseCase

import (
	"math/rand"

	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func CreateHoneypotHit(
	cmdRepo tkRepository.HoneypotCmdRepo,
	requesterIp tkValueObject.IpAddress,
	interceptPath string,
	maxEntries int,
) {
	if cmdRepo == nil {
		return
	}

	cmdRepo.IncrementHit(requesterIp, interceptPath)

	if rand.Float64() < 0.02 {
		cmdRepo.EnforceMaxEntries(maxEntries)
	}
}
