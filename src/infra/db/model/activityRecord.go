package tkInfraDbModel

import (
	"time"

	"github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type ActivityRecord struct {
	ID                uint64 `gorm:"primaryKey"`
	RecordLevel       string `gorm:"not null"`
	RecordCode        string `gorm:"not null"`
	AffectedResources []ActivityRecordAffectedResource
	RecordDetails     *string
	OperatorAccountId *uint64
	OperatorIpAddress *string
	CreatedAt         time.Time `gorm:"not null"`
}

func (ActivityRecord) TableName() string {
	return "activity_records"
}

func NewActivityRecord(
	recordId uint64,
	recordLevel, recordCode string,
	affectedResources []ActivityRecordAffectedResource,
	recordDetails *string,
	operatorAccountId *uint64,
	operatorIpAddress *string,
) ActivityRecord {
	model := ActivityRecord{
		RecordLevel:       recordLevel,
		RecordCode:        recordCode,
		AffectedResources: affectedResources,
		RecordDetails:     recordDetails,
		OperatorAccountId: operatorAccountId,
		OperatorIpAddress: operatorIpAddress,
	}

	if recordId != 0 {
		model.ID = recordId
	}

	return model
}

func (model ActivityRecord) ToEntity() (recordEntity entity.ActivityRecord, err error) {
	recordId, err := tkValueObject.NewActivityRecordId(model.ID)
	if err != nil {
		return recordEntity, err
	}

	recordLevel, err := tkValueObject.NewActivityRecordLevel(model.RecordLevel)
	if err != nil {
		return recordEntity, err
	}

	recordCode, err := tkValueObject.NewActivityRecordCode(model.RecordCode)
	if err != nil {
		return recordEntity, err
	}

	affectedResources := []tkValueObject.SystemResourceIdentifier{}
	for _, resource := range model.AffectedResources {
		sri, err := tkValueObject.NewSystemResourceIdentifier(resource.SystemResourceIdentifier)
		if err != nil {
			return recordEntity, err
		}
		affectedResources = append(affectedResources, sri)
	}

	var recordDetails any
	if model.RecordDetails != nil {
		recordDetails = *model.RecordDetails
	}

	var operatorAccountIdPtr *tkValueObject.AccountId
	if model.OperatorAccountId != nil {
		operatorAccountId, err := tkValueObject.NewAccountId(*model.OperatorAccountId)
		if err != nil {
			return recordEntity, err
		}
		operatorAccountIdPtr = &operatorAccountId
	}

	var operatorIpAddressPtr *tkValueObject.IpAddress
	if model.OperatorIpAddress != nil {
		operatorIpAddress, err := tkValueObject.NewIpAddress(*model.OperatorIpAddress)
		if err != nil {
			return recordEntity, err
		}
		operatorIpAddressPtr = &operatorIpAddress
	}

	return entity.NewActivityRecord(
		recordId, recordLevel, recordCode, affectedResources, recordDetails,
		operatorAccountIdPtr, operatorIpAddressPtr,
		tkValueObject.NewUnixTimeWithGoTime(model.CreatedAt),
	)
}
