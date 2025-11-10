package tkInfraActivityRecord

import (
	"encoding/json"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
)

type ActivityRecordCmdRepo struct {
	trailDbSvc *tkInfraDb.TrailDatabaseService
	queryRepo  *ActivityRecordQueryRepo
}

func NewActivityRecordCmdRepo(
	trailDbSvc *tkInfraDb.TrailDatabaseService,
	queryRepo *ActivityRecordQueryRepo,
) *ActivityRecordCmdRepo {
	return &ActivityRecordCmdRepo{
		trailDbSvc: trailDbSvc,
		queryRepo:  queryRepo,
	}
}

func (repo *ActivityRecordCmdRepo) Create(createDto tkDto.CreateActivityRecord) error {
	affectedResources := []tkInfraDbModel.ActivityRecordAffectedResource{}
	for _, affectedResourceSri := range createDto.AffectedResources {
		affectedResourceModel := tkInfraDbModel.ActivityRecordAffectedResource{
			SystemResourceIdentifier: affectedResourceSri.String(),
		}
		affectedResources = append(affectedResources, affectedResourceModel)
	}

	var recordDetails *string
	if createDto.RecordDetails != nil {
		recordDetailsBytes, err := json.Marshal(createDto.RecordDetails)
		if err != nil {
			return err
		}
		recordDetailsStr := string(recordDetailsBytes)
		recordDetails = &recordDetailsStr
	}

	var operatorAccountIdPtr *uint64
	if createDto.OperatorAccountId != nil {
		operatorAccountId := createDto.OperatorAccountId.Uint64()
		operatorAccountIdPtr = &operatorAccountId
	}

	var operatorIpAddressPtr *string
	if createDto.OperatorIpAddress != nil {
		operatorIpAddress := createDto.OperatorIpAddress.String()
		operatorIpAddressPtr = &operatorIpAddress
	}

	activityRecordModel := tkInfraDbModel.NewActivityRecord(
		0, createDto.RecordLevel.String(), createDto.RecordCode.String(),
		affectedResources, recordDetails, operatorAccountIdPtr, operatorIpAddressPtr,
	)

	return repo.trailDbSvc.Handler.Create(&activityRecordModel).Error
}

func (repo *ActivityRecordCmdRepo) Delete(deleteDto tkDto.DeleteActivityRecord) error {
	readResponseDto, err := repo.queryRepo.Read(tkDto.ReadActivityRecordsRequest{
		Pagination:        tkDto.PaginationUnpaginated,
		RecordId:          deleteDto.RecordId,
		RecordLevel:       deleteDto.RecordLevel,
		RecordCode:        deleteDto.RecordCode,
		AffectedResources: deleteDto.AffectedResources,
		OperatorAccountId: deleteDto.OperatorAccountId,
		OperatorIpAddress: deleteDto.OperatorIpAddress,
		CreatedBeforeAt:   deleteDto.CreatedBeforeAt,
		CreatedAfterAt:    deleteDto.CreatedAfterAt,
	})
	if err != nil {
		return err
	}

	if len(readResponseDto.ActivityRecords) == 0 {
		return nil
	}

	recordIds := []uint64{}
	for _, record := range readResponseDto.ActivityRecords {
		recordIds = append(recordIds, record.RecordId.Uint64())
	}

	return repo.trailDbSvc.Handler.Delete(&tkInfraDbModel.ActivityRecord{}, recordIds).Error
}
