package tkInfraDbModel

type KeyValue struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

func (KeyValue) TableName() string {
	return "key_values"
}
