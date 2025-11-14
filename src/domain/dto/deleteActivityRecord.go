package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type DeleteActivityRecord struct {
	RecordId          *tkValueObject.ActivityRecordId          `json:"recordId"`
	RecordLevel       *tkValueObject.ActivityRecordLevel       `json:"recordLevel"`
	RecordCode        *tkValueObject.ActivityRecordCode        `json:"recordCode"`
	AffectedResources []tkValueObject.SystemResourceIdentifier `json:"affectedResources"`
	OperatorSri       *tkValueObject.SystemResourceIdentifier  `json:"operatorSri"`
	OperatorIpAddress *tkValueObject.IpAddress                 `json:"operatorIpAddress"`
	CreatedBeforeAt   *tkValueObject.UnixTime                  `json:"createdBeforeAt"`
	CreatedAfterAt    *tkValueObject.UnixTime                  `json:"createdAfterAt"`
}

func NewDeleteActivityRecord(
	recordId *tkValueObject.ActivityRecordId,
	recordLevel *tkValueObject.ActivityRecordLevel,
	recordCode *tkValueObject.ActivityRecordCode,
	affectedResources []tkValueObject.SystemResourceIdentifier,
	operatorSri *tkValueObject.SystemResourceIdentifier,
	operatorIpAddress *tkValueObject.IpAddress,
	createdBeforeAt, createdAfterAt *tkValueObject.UnixTime,
) DeleteActivityRecord {
	return DeleteActivityRecord{
		RecordId:          recordId,
		RecordLevel:       recordLevel,
		RecordCode:        recordCode,
		AffectedResources: affectedResources,
		OperatorSri:       operatorSri,
		OperatorIpAddress: operatorIpAddress,
		CreatedBeforeAt:   createdBeforeAt,
		CreatedAfterAt:    createdAfterAt,
	}
}
