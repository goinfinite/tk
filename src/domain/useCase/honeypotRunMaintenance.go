package tkUseCase

import (
	"encoding/json"
	"log/slog"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func RunHoneypotMaintenance(
	cmdRepo tkRepository.HoneypotCmdRepo,
	queryRepo tkRepository.HoneypotQueryRepo,
	activityRecordCmdRepo tkRepository.ActivityRecordCmdRepo,
	banDuration time.Duration,
	maxEntries int,
	aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
	recordCode tkValueObject.ActivityRecordCode,
	recordLevel tkValueObject.ActivityRecordLevel,
) {
	if cmdRepo != nil {
		cmdRepo.CleanExpiredEntries(banDuration)
		cmdRepo.EnforceMaxEntries(maxEntries)
	}

	if queryRepo == nil || activityRecordCmdRepo == nil {
		return
	}

	if queryRepo.Count() == 0 {
		return
	}

	statsReport, reportErr := ReadHoneypotStatsReport(
		queryRepo, banDuration, aggressivenessMode,
	)
	if reportErr != nil {
		slog.Debug("HoneypotStatsReadReportFailed",
			slog.String("err", reportErr.Error()))
		return
	}

	reportJson, marshalErr := json.Marshal(statsReport)
	if marshalErr != nil {
		slog.Debug("HoneypotStatsMarshalFailed",
			slog.String("err", marshalErr.Error()))
		return
	}

	createRequest := tkDto.CreateActivityRecord{
		RecordLevel: recordLevel,
		RecordCode:  recordCode,
		AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		RecordDetails: map[string]string{
			"statsReport": string(reportJson),
		},
	}

	createErr := activityRecordCmdRepo.Create(createRequest)
	if createErr != nil {
		slog.Debug("HoneypotStatsReportCreationFailed",
			slog.String("err", createErr.Error()))
	}
}
