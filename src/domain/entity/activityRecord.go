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
	OperatorAccountId *tkValueObject.AccountId                 `json:"operatorAccountId"`
	OperatorIpAddress *tkValueObject.IpAddress                 `json:"operatorIpAddress"`
	CreatedAt         tkValueObject.UnixTime                   `json:"createdAt"`
}

func NewActivityRecord(
	recordId tkValueObject.ActivityRecordId,
	recordLevel tkValueObject.ActivityRecordLevel,
	recordCode tkValueObject.ActivityRecordCode,
	affectedResources []tkValueObject.SystemResourceIdentifier,
	recordDetails any,
	operatorAccountId *tkValueObject.AccountId,
	operatorIpAddress *tkValueObject.IpAddress,
	createdAt tkValueObject.UnixTime,
) (ActivityRecord, error) {
	return ActivityRecord{
		RecordId:          recordId,
		RecordLevel:       recordLevel,
		RecordCode:        recordCode,
		AffectedResources: affectedResources,
		RecordDetails:     recordDetails,
		OperatorAccountId: operatorAccountId,
		OperatorIpAddress: operatorIpAddress,
		CreatedAt:         createdAt,
	}, nil
}
