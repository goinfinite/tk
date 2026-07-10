package tkInfraHoneypot

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
)

type HoneypotQueryRepo struct {
	transientDbSvc *tkInfraDb.TransientDatabaseService
}

func NewHoneypotQueryRepo(
	transientDbSvc *tkInfraDb.TransientDatabaseService,
) *HoneypotQueryRepo {
	return &HoneypotQueryRepo{transientDbSvc: transientDbSvc}
}

func (repo *HoneypotQueryRepo) ReadBanDecision(
	requestDto tkDto.ReadHoneypotBanDecisionRequest,
) (tkDto.ReadHoneypotBanDecisionResponse, error) {
	var hitModel tkInfraDbModel.HoneypotHitModel
	err := repo.transientDbSvc.Handler.
		Where("requester_ip_address = ?", requestDto.RequesterIpAddress.String()).
		First(&hitModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tkDto.ReadHoneypotBanDecisionResponse{
				HitCount: 0,
			}, nil
		}

		slog.Debug(
			"ReadHoneypotBanDecisionDbError",
			"error", err,
		)
		return tkDto.ReadHoneypotBanDecisionResponse{}, err
	}

	return tkDto.ReadHoneypotBanDecisionResponse{
		HitCount: hitModel.HitCount,
	}, nil
}

func (repo *HoneypotQueryRepo) ReadStatsReport(
	requestDto tkDto.ReadHoneypotStatsReportRequest,
) (tkDto.ReadHoneypotStatsReportResponse, error) {
	var totalHits int64
	err := repo.transientDbSvc.Handler.
		Model(&tkInfraDbModel.HoneypotHitModel{}).
		Select("COALESCE(SUM(hit_count), 0)").
		Scan(&totalHits).Error
	if err != nil {
		return tkDto.ReadHoneypotStatsReportResponse{}, err
	}

	var uniqueIps int64
	err = repo.transientDbSvc.Handler.
		Model(&tkInfraDbModel.HoneypotHitModel{}).
		Count(&uniqueIps).Error
	if err != nil {
		return tkDto.ReadHoneypotStatsReportResponse{}, err
	}

	type classHitRow struct {
		HitClass string
		Total    uint64
	}
	var classHits []classHitRow
	err = repo.transientDbSvc.Handler.
		Model(&tkInfraDbModel.HoneypotHitModel{}).
		Select("hit_class, SUM(hit_count) as total").
		Group("hit_class").
		Scan(&classHits).Error
	if err != nil {
		return tkDto.ReadHoneypotStatsReportResponse{}, err
	}

	hitsByClass := map[tkValueObject.HoneypotPathClass]uint64{}
	for _, row := range classHits {
		pathClass, classErr := tkValueObject.NewHoneypotPathClass(row.HitClass)
		if classErr != nil {
			slog.Debug(
				"ReadHoneypotStatsReportInvalidClass",
				"hitClass", row.HitClass,
				"error", classErr,
			)
			continue
		}
		hitsByClass[pathClass] = row.Total
	}

	return tkDto.ReadHoneypotStatsReportResponse{
		TotalHits:   uint64(totalHits),
		UniqueIps:   uint64(uniqueIps),
		HitsByClass: hitsByClass,
		BannedIps:   0,
	}, nil
}
