package tkInfraHoneypot

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
)

type HoneypotCmdRepo struct {
	transientDbSvc *tkInfraDb.TransientDatabaseService
}

func NewHoneypotCmdRepo(
	transientDbSvc *tkInfraDb.TransientDatabaseService,
) *HoneypotCmdRepo {
	return &HoneypotCmdRepo{transientDbSvc: transientDbSvc}
}

func (repo *HoneypotCmdRepo) Create(
	createDto tkDto.CreateHoneypotHit,
) error {
	now := time.Now().UTC()
	hitModel := tkInfraDbModel.NewHoneypotHitModel(
		createDto.RequesterIpAddress.String(),
		createDto.HoneypotPath.String(),
		createDto.HitClass.String(),
		createDto.HitCount,
		now,
		now,
	)

	return repo.transientDbSvc.Handler.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "requester_ip_address"}},
		DoUpdates: clause.Assignments(map[string]any{
			"hit_count":  gorm.Expr("honeypot_hits.hit_count + ?", 1),
			"created_at": now,
		}),
	}).Create(&hitModel).Error
}

func (repo *HoneypotCmdRepo) DeleteExpired(
	banDuration tkValueObject.HoneypotBanDuration,
) error {
	cutoffTime := time.Now().UTC().Add(-banDuration.Duration())

	return repo.transientDbSvc.Handler.
		Where("created_at < ?", cutoffTime).
		Where("first_hit_at < ?", cutoffTime).
		Delete(&tkInfraDbModel.HoneypotHitModel{}).Error
}

func (repo *HoneypotCmdRepo) EnforceMaxEntries(
	maxEntries tkValueObject.HoneypotMaxEntries,
) error {
	var entryCount int64
	err := repo.transientDbSvc.Handler.
		Model(&tkInfraDbModel.HoneypotHitModel{}).
		Count(&entryCount).Error
	if err != nil {
		return err
	}

	maxEntriesCount := int64(maxEntries.Uint64())
	if entryCount <= maxEntriesCount {
		return nil
	}

	entriesToDelete := entryCount - maxEntriesCount

	subQuery := repo.transientDbSvc.Handler.
		Model(&tkInfraDbModel.HoneypotHitModel{}).
		Select("requester_ip_address").
		Order("created_at ASC").
		Limit(int(entriesToDelete))

	return repo.transientDbSvc.Handler.
		Where("requester_ip_address IN (?)", subQuery).
		Delete(&tkInfraDbModel.HoneypotHitModel{}).Error
}
