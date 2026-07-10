package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
)

func RunHoneypotMaintenance(
	honeypotCmdRepo tkRepository.HoneypotCmdRepo,
	requestDto tkDto.RunHoneypotMaintenanceRequest,
) error {
	err := honeypotCmdRepo.DeleteExpired(requestDto.BanDuration)
	if err != nil {
		slog.Error(
			"RunHoneypotMaintenanceDeleteExpiredError",
			slog.String("err", err.Error()),
		)
		return err
	}

	err = honeypotCmdRepo.EnforceMaxEntries(requestDto.MaxEntries)
	if err != nil {
		slog.Error(
			"RunHoneypotMaintenanceEnforceMaxEntriesError",
			slog.String("err", err.Error()),
		)
		return err
	}

	return nil
}
