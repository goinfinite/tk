package tkInfraDbModel

import (
	"time"

	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type ActivityRecord struct {
	ID                uint64 `gorm:"primaryKey"`
	RecordLevel       string `gorm:"not null"`
	RecordCode        string `gorm:"not null;index:idx_activity_records_code_ip_created_at,priority:1"`
	AffectedResources []ActivityRecordAffectedResource
	RecordDetails     *string
	OperatorSri       *string   `gorm:"index:idx_activity_records_operator_sri_created_at,priority:1"`
	OperatorIpAddress *string   `gorm:"index:idx_activity_records_code_ip_created_at,priority:2"`
	CreatedAt         time.Time `gorm:"not null;index:idx_activity_records_created_at;index:idx_activity_records_operator_sri_created_at,priority:2;index:idx_activity_records_code_ip_created_at,priority:3"`
}

func (ActivityRecord) TableName() string {
	return "activity_records"
}

func NewActivityRecord(
	recordId uint64,
	recordLevel, recordCode string,
	affectedResources []ActivityRecordAffectedResource,
	recordDetails, operatorSri, operatorIpAddress *string,
) ActivityRecord {
	model := ActivityRecord{
		RecordLevel:       recordLevel,
		RecordCode:        recordCode,
		AffectedResources: affectedResources,
		RecordDetails:     recordDetails,
		OperatorSri:       operatorSri,
		OperatorIpAddress: operatorIpAddress,
	}

	if recordId != 0 {
		model.ID = recordId
	}

	return model
}

func (model ActivityRecord) ToEntity() (recordEntity tkEntity.ActivityRecord, err error) {
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

	var operatorSriPtr *tkValueObject.SystemResourceIdentifier
	if model.OperatorSri != nil {
		operatorSri, err := tkValueObject.NewSystemResourceIdentifier(*model.OperatorSri)
		if err != nil {
			return recordEntity, err
		}
		operatorSriPtr = &operatorSri
	}

	var operatorIpAddressPtr *tkValueObject.IpAddress
	if model.OperatorIpAddress != nil {
		operatorIpAddress, err := tkValueObject.NewIpAddress(*model.OperatorIpAddress)
		if err != nil {
			return recordEntity, err
		}
		operatorIpAddressPtr = &operatorIpAddress
	}

	return tkEntity.NewActivityRecord(
		recordId, recordLevel, recordCode, affectedResources, recordDetails,
		operatorSriPtr, operatorIpAddressPtr,
		tkValueObject.NewUnixTimeWithGoTime(model.CreatedAt),
	), nil
}
