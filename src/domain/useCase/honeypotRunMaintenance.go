package tkUseCase

import (
	"encoding/json"
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func RunHoneypotMaintenance(
	cmdRepo tkRepository.HoneypotCmdRepo,
	queryRepo tkRepository.HoneypotQueryRepo,
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	request tkDto.RunHoneypotMaintenanceRequest,
) {
	if cmdRepo != nil {
		cmdRepo.CleanExpiredEntries(
			request.BanDuration.Duration(),
		)
		cmdRepo.EnforceMaxEntries(
			request.MaxEntries.Int(),
		)
	}

	if queryRepo == nil || activityRecordCmdRepo == nil {
		return
	}

	if queryRepo.Count() == 0 {
		return
	}

	statsReport, reportErr := ReadHoneypotStatsReport(
		queryRepo,
		request.BanDuration,
		request.AggressivenessMode,
	)
	if reportErr != nil {
		slog.Error("HoneypotStatsReadReportFailed",
			slog.String("err", reportErr.Error()))
		return
	}

	reportJson, marshalErr := json.Marshal(statsReport)
	if marshalErr != nil {
		slog.Error("HoneypotStatsMarshalFailed",
			slog.String("err", marshalErr.Error()))
		return
	}

	createRequest := tkDto.CreateActivityRecord{
		RecordLevel: request.StatsRecordLevel,
		RecordCode:  request.StatsRecordCode,
		AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		RecordDetails: map[string]string{
			"statsReport": string(reportJson),
		},
	}

	createErr := activityRecordCmdRepo.Create(createRequest)
	if createErr != nil {
		slog.Error("HoneypotStatsReportCreationFailed",
			slog.String("err", createErr.Error()))
	}
}
