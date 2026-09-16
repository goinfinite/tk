package tkInfraDbModel

import "time"

type KeyValue struct {
	Key       string `gorm:"primaryKey"`
	Value     string `gorm:"not null"`
	ExpiresAt *time.Time
}

func (KeyValue) TableName() string {
	return "key_values"
}
