package entity

import (
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type ActivityRecord struct {
	RecordId          tkValueObject.ActivityRecordId           `json:"recordId"`
	RecordLevel       tkValueObject.ActivityRecordLevel        `json:"recordLevel"`
	RecordCode        tkValueObject.ActivityRecordCode         `json:"recordCode"`
	AffectedResources []tkValueObject.SystemResourceIdentifier `json:"affectedResources"`
	RecordDetails     any                                      `json:"recordDetails"`
	OperatorSri       *tkValueObject.SystemResourceIdentifier  `json:"operatorSri"`
	OperatorIpAddress *tkValueObject.IpAddress                 `json:"operatorIpAddress"`
	CreatedAt         tkValueObject.UnixTime                   `json:"createdAt"`
}

func NewActivityRecord(
	recordId tkValueObject.ActivityRecordId,
	recordLevel tkValueObject.ActivityRecordLevel,
	recordCode tkValueObject.ActivityRecordCode,
	affectedResources []tkValueObject.SystemResourceIdentifier,
	recordDetails any,
	operatorSri *tkValueObject.SystemResourceIdentifier,
	operatorIpAddress *tkValueObject.IpAddress,
	createdAt tkValueObject.UnixTime,
) ActivityRecord {
	return ActivityRecord{
		RecordId:          recordId,
		RecordLevel:       recordLevel,
		RecordCode:        recordCode,
		AffectedResources: affectedResources,
		RecordDetails:     recordDetails,
		OperatorSri:       operatorSri,
		OperatorIpAddress: operatorIpAddress,
		CreatedAt:         createdAt,
	}
}
