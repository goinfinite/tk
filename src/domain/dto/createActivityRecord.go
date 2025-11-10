package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type CreateActivityRecord struct {
	RecordLevel       tkValueObject.ActivityRecordLevel        `json:"recordLevel"`
	RecordCode        tkValueObject.ActivityRecordCode         `json:"recordCode"`
	AffectedResources []tkValueObject.SystemResourceIdentifier `json:"affectedResources"`
	RecordDetails     any                                      `json:"recordDetails"`
	OperatorAccountId *tkValueObject.AccountId                 `json:"operatorAccountId"`
	OperatorIpAddress *tkValueObject.IpAddress                 `json:"operatorIpAddress"`
}
