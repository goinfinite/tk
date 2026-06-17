package tkInfraDb

import (
	"errors"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	errTransientDbConnectionError string = "TransientDbConnectionError"
	errTransientDbMigrationError  string = "TransientDbMigrationError"
)

type KeyValueModel struct {
	Key       string    `gorm:"primaryKey"`
	Value     string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (KeyValueModel) TableName() string {
	return "key_values"
}

type TransientDatabaseService struct {
	Handler *gorm.DB
}

func NewTransientDatabaseService() (*TransientDatabaseService, error) {
	ormSvc, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		},
	)
	if err != nil {
		return nil, errors.New(errTransientDbConnectionError)
	}

	dbSvc := &TransientDatabaseService{Handler: ormSvc}

	err = ormSvc.AutoMigrate(&KeyValueModel{})
	if err != nil {
		return nil, errors.New(errTransientDbMigrationError)
	}

	return dbSvc, nil
}

func (service *TransientDatabaseService) Has(key string) bool {
	var model KeyValueModel
	err := service.Handler.Where("key = ?", key).First(&model).Error
	return err == nil
}

func (service *TransientDatabaseService) Read(key string) (string, error) {
	var model KeyValueModel
	err := service.Handler.Where("key = ?", key).First(&model).Error
	if err != nil {
		return "", err
	}

	return model.Value, nil
}

func (service *TransientDatabaseService) ReadAll() ([]KeyValueModel, error) {
	var entries []KeyValueModel
	err := service.Handler.Find(&entries).Error
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (service *TransientDatabaseService) Set(key, value string) error {
	existingEntry := KeyValueModel{Key: key}
	result := service.Handler.First(&existingEntry)

	if result.Error != nil {
		newEntry := KeyValueModel{Key: key, Value: value}
		return service.Handler.Create(&newEntry).Error
	}

	return service.Handler.Model(&existingEntry).Update("value", value).Error
}

func (service *TransientDatabaseService) Count() int64 {
	var count int64
	service.Handler.Model(&KeyValueModel{}).Count(&count)

	return count
}
