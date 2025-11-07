package tkDto

import (
	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type ReadActivityRecordsRequest struct {
	Pagination        Pagination                               `json:"pagination"`
	RecordId          *tkValueObject.ActivityRecordId          `json:"recordId"`
	RecordLevel       *tkValueObject.ActivityRecordLevel       `json:"recordLevel"`
	RecordCode        *tkValueObject.ActivityRecordCode        `json:"recordCode"`
	AffectedResources []tkValueObject.SystemResourceIdentifier `json:"affectedResources"`
	RecordDetails     *string                                  `json:"recordDetails"`
	OperatorAccountId *tkValueObject.AccountId                 `json:"operatorAccountId"`
	OperatorIpAddress *tkValueObject.IpAddress                 `json:"operatorIpAddress"`
	CreatedBeforeAt   *tkValueObject.UnixTime                  `json:"createdBeforeAt"`
	CreatedAfterAt    *tkValueObject.UnixTime                  `json:"createdAfterAt"`
}

type ReadActivityRecordsResponse struct {
	Pagination      Pagination                `json:"pagination"`
	ActivityRecords []tkEntity.ActivityRecord `json:"activityRecords"`
}
