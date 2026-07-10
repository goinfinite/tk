package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

func CreateHoneypotHit(
	honeypotCmdRepo tkRepository.HoneypotCmdRepo,
	createDto tkDto.CreateHoneypotHit,
) {
	err := honeypotCmdRepo.Create(createDto)
	if err != nil {
		slog.Error(
			"CreateHoneypotHitInfraError",
			slog.String("err", err.Error()),
		)
	}
}
