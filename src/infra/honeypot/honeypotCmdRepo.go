package tkInfraHoneypot

import (
	"encoding/json"
	"log/slog"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
)

const honeypotHitKeyPrefix = "honeypot:hit:"

type HoneypotCmdRepo struct {
	transientDbSvc *tkInfraDb.TransientDatabaseService
}

func NewHoneypotCmdRepo(
	transientDbSvc *tkInfraDb.TransientDatabaseService,
) *HoneypotCmdRepo {
	return &HoneypotCmdRepo{transientDbSvc: transientDbSvc}
}

func (repo *HoneypotCmdRepo) buildHitKey(
	requesterIp tkValueObject.IpAddress,
) string {
	return honeypotHitKeyPrefix + requesterIp.String()
}

func (repo *HoneypotCmdRepo) IncrementHit(
	requesterIp tkValueObject.IpAddress,
	interceptPath string,
) {
	if repo.transientDbSvc == nil {
		return
	}

	hitKey := repo.buildHitKey(requesterIp)
	rawValue, readErr := repo.transientDbSvc.Read(hitKey)

	var hitData tkDto.HoneypotHitData
	if readErr == nil {
		parseErr := json.Unmarshal(
			[]byte(rawValue), &hitData,
		)
		if parseErr != nil {
			slog.Debug("HoneypotHitDataParseFailed",
				slog.String("err", parseErr.Error()))
		}
	}

	if hitData.Count == 0 {
		hitData.FirstHitAt = time.Now().UTC().Format(
			time.RFC3339,
		)
		hitData.Endpoints = make(map[string]int)
	}

	hitData.Count++
	hitData.Endpoints[interceptPath]++

	jsonBytes, marshalErr := json.Marshal(hitData)
	if marshalErr != nil {
		slog.Debug("HoneypotHitDataMarshalFailed",
			slog.String("err", marshalErr.Error()))
		return
	}

	setErr := repo.transientDbSvc.Set(
		hitKey, string(jsonBytes),
	)
	if setErr != nil {
		slog.Debug("HoneypotHitCountSetFailed",
			slog.String("err", setErr.Error()))
	}
}

func (repo *HoneypotCmdRepo) CleanExpiredEntries(
	banDuration time.Duration,
) {
	if repo.transientDbSvc == nil {
		return
	}

	cutoff := time.Now().Add(-banDuration)
	repo.transientDbSvc.Handler.Where(
		"created_at < ?", cutoff,
	).Delete(&tkInfraDb.KeyValueModel{})
}

func (repo *HoneypotCmdRepo) EnforceMaxEntries(
	maxEntries int,
) {
	if repo.transientDbSvc == nil {
		return
	}

	var totalCount int64
	repo.transientDbSvc.Handler.Model(
		&tkInfraDb.KeyValueModel{},
	).Count(&totalCount)

	if int(totalCount) <= maxEntries {
		return
	}

	excessCount := int(totalCount) - maxEntries
	keysToDelete := make([]string, 0, excessCount)
	repo.transientDbSvc.Handler.Model(
		&tkInfraDb.KeyValueModel{},
	).Order("created_at ASC").Limit(
		excessCount,
	).Pluck("key", &keysToDelete)

	if len(keysToDelete) > 0 {
		repo.transientDbSvc.Handler.Where(
			"key IN ?", keysToDelete,
		).Delete(&tkInfraDb.KeyValueModel{})
	}
}
