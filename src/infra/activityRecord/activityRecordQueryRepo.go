package tkInfraActivityRecord

import (
	"errors"
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkUseCase "github.com/goinfinite/tk/src/domain/useCase"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
)

type ActivityRecordQueryRepo struct {
	trailDbSvc *tkInfraDb.TrailDatabaseService
}

func NewActivityRecordQueryRepo(
	trailDbSvc *tkInfraDb.TrailDatabaseService,
) *ActivityRecordQueryRepo {
	return &ActivityRecordQueryRepo{trailDbSvc: trailDbSvc}
}

func (repo *ActivityRecordQueryRepo) Read(
	requestDto tkDto.ReadActivityRecordsRequest,
) (responseDto tkDto.ReadActivityRecordsResponse, err error) {
	recordModel := tkInfraDbModel.ActivityRecord{}
	if requestDto.RecordId != nil {
		recordId := requestDto.RecordId.Uint64()
		recordModel.ID = recordId
	}

	if requestDto.RecordLevel != nil {
		recordLevelStr := requestDto.RecordLevel.String()
		recordModel.RecordLevel = recordLevelStr
	}

	if requestDto.RecordCode != nil {
		recordCodeStr := requestDto.RecordCode.String()
		recordModel.RecordCode = recordCodeStr
	}

	if requestDto.OperatorAccountId != nil {
		operatorAccountId := requestDto.OperatorAccountId.Uint64()
		recordModel.OperatorAccountId = &operatorAccountId
	}

	if requestDto.OperatorIpAddress != nil {
		operatorIpAddressStr := requestDto.OperatorIpAddress.String()
		recordModel.OperatorIpAddress = &operatorIpAddressStr
	}

	dbQuery := repo.trailDbSvc.Handler.Model(&recordModel).Where(&recordModel)

	if len(requestDto.AffectedResources) > 0 {
		systemResourceIdentifiersStrSlice := []string{}
		for _, sri := range requestDto.AffectedResources {
			systemResourceIdentifiersStrSlice = append(systemResourceIdentifiersStrSlice, sri.String())
		}

		affectedResourcesSubQuery := repo.trailDbSvc.Handler.
			Model(&tkInfraDbModel.ActivityRecordAffectedResource{}).
			Select("DISTINCT activity_record_id").
			Where("system_resource_identifier IN ?", systemResourceIdentifiersStrSlice)

		dbQuery = dbQuery.Where("activity_records.id IN (?)", affectedResourcesSubQuery)
	}

	if requestDto.CreatedBeforeAt != nil {
		dbQuery = dbQuery.Where("created_at < ?", requestDto.CreatedBeforeAt.ReadAsGoTime())
	}
	if requestDto.CreatedAfterAt != nil {
		dbQuery = dbQuery.Where("created_at > ?", requestDto.CreatedAfterAt.ReadAsGoTime())
	}

	paginatedDbQuery, responsePagination, err := tkInfraDb.PaginationQueryBuilder(
		dbQuery, requestDto.Pagination,
	)
	if err != nil {
		return responseDto, errors.New("PaginationQueryBuilderError: " + err.Error())
	}

	recordModels := []tkInfraDbModel.ActivityRecord{}
	err = paginatedDbQuery.Preload("AffectedResources").Find(&recordModels).Error
	if err != nil {
		return responseDto, err
	}

	for _, recordModel := range recordModels {
		activityRecordEntity, err := recordModel.ToEntity()
		if err != nil {
			slog.Debug(
				"ActivityRecordModelToEntityError",
				slog.Uint64("id", recordModel.ID),
				slog.String("err", err.Error()),
			)
			continue
		}
		responseDto.ActivityRecords = append(responseDto.ActivityRecords, activityRecordEntity)
	}

	return tkDto.ReadActivityRecordsResponse{
		ActivityRecords: responseDto.ActivityRecords,
		Pagination:      responsePagination,
	}, nil
}

func (repo *ActivityRecordQueryRepo) ReadFirst(
	requestDto tkDto.ReadActivityRecordsRequest,
) (activityRecord tkEntity.ActivityRecord, err error) {
	requestDto.Pagination = tkDto.PaginationSingleItem
	responseDto, err := repo.Read(requestDto)
	if err != nil {
		return activityRecord, err
	}

	if len(responseDto.ActivityRecords) == 0 {
		return activityRecord, errors.New(tkUseCase.ErrActivityRecordNotFound)
	}

	return responseDto.ActivityRecords[0], nil
}
