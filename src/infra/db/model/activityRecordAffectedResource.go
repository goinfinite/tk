package tkInfraDbModel

type ActivityRecordAffectedResource struct {
	ID                       uint64 `gorm:"primaryKey"`
	SystemResourceIdentifier string `gorm:"not null;index:idx_activity_records_affected_resources_resource_record,priority:1"`
	ActivityRecordID         uint64 `gorm:"not null;index:idx_activity_records_affected_resources_record;index:idx_activity_records_affected_resources_resource_record,priority:2"`
}

func (ActivityRecordAffectedResource) TableName() string {
	return "activity_records_affected_resources"
}
