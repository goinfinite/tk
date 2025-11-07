package tkInfraActivityRecord

import (
	"encoding/json"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
)

type ActivityRecordCmdRepo struct {
	trailDbSvc *tkInfraDb.TrailDatabaseService
}

func NewActivityRecordCmdRepo(
	trailDbSvc *tkInfraDb.TrailDatabaseService,
) *ActivityRecordCmdRepo {
	return &ActivityRecordCmdRepo{trailDbSvc: trailDbSvc}
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
	deleteModel := tkInfraDbModel.ActivityRecord{}
	if deleteDto.RecordId != nil {
		deleteModel.ID = deleteDto.RecordId.Uint64()
	}

	if deleteDto.RecordLevel != nil {
		deleteModel.RecordLevel = deleteDto.RecordLevel.String()
	}

	if deleteDto.RecordCode != nil {
		deleteModel.RecordCode = deleteDto.RecordCode.String()
	}

	affectedResources := []tkInfraDbModel.ActivityRecordAffectedResource{}
	for _, affectedResourceSri := range deleteDto.AffectedResources {
		affectedResourceModel := tkInfraDbModel.ActivityRecordAffectedResource{
			SystemResourceIdentifier: affectedResourceSri.String(),
		}
		affectedResources = append(affectedResources, affectedResourceModel)
	}
	deleteModel.AffectedResources = affectedResources

	if deleteDto.OperatorAccountId != nil {
		operatorAccountId := deleteDto.OperatorAccountId.Uint64()
		deleteModel.OperatorAccountId = &operatorAccountId
	}

	if deleteDto.OperatorIpAddress != nil {
		operatorIpAddressStr := deleteDto.OperatorIpAddress.String()
		deleteModel.OperatorIpAddress = &operatorIpAddressStr
	}

	dbQuery := repo.trailDbSvc.Handler.Model(&deleteModel).Where(&deleteModel)

	if deleteDto.CreatedBeforeAt != nil {
		dbQuery.Where("created_at < ?", deleteDto.CreatedBeforeAt.ReadAsGoTime())
	}
	if deleteDto.CreatedAfterAt != nil {
		dbQuery.Where("created_at > ?", deleteDto.CreatedAfterAt.ReadAsGoTime())
	}

	return dbQuery.Delete(&tkInfraDbModel.ActivityRecord{}).Error
}
