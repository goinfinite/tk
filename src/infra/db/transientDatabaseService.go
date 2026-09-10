package tkInfraDb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/glebarez/sqlite"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	errTransientDatabaseConnectionError string = "TransientDatabaseConnectionError"
	errTransientDatabaseMigrationError  string = "TransientDatabaseMigrationError"
)

var ErrKeyNotFound = errors.New("KeyNotFound")

type TransientDatabaseService struct {
	Handler *gorm.DB
}

func NewTransientDatabaseService() (*TransientDatabaseService, error) {
	ormSvc, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{NowFunc: func() time.Time { return time.Now().UTC() }},
	)
	if err != nil {
		return nil, errors.New(errTransientDatabaseConnectionError)
	}

	err = ormSvc.AutoMigrate(&tkInfraDbModel.KeyValue{})
	if err != nil {
		return nil, errors.New(errTransientDatabaseMigrationError + ": " + err.Error())
	}

	return &TransientDatabaseService{Handler: ormSvc}, nil
}

func (service *TransientDatabaseService) Has(key string) bool {
	var count int64
	result := service.Handler.Model(&tkInfraDbModel.KeyValue{}).
		Where("key = ?", key).Count(&count)
	if result.Error != nil {
		slog.Error(
			"TransientDatabaseKeyCountFailed",
			slog.String("key", key),
			slog.String("err", result.Error.Error()),
		)
		return false
	}

	return count > 0
}

func (service *TransientDatabaseService) Read(key string) (string, error) {
	var keyValue tkInfraDbModel.KeyValue
	result := service.Handler.Model(&tkInfraDbModel.KeyValue{}).
		Where("key = ?", key).Find(&keyValue)
	if result.Error != nil {
		return "", result.Error
	}

	if result.RowsAffected == 0 {
		return "", ErrKeyNotFound
	}

	return keyValue.Value, nil
}

func (service *TransientDatabaseService) Set(key string, value string) error {
	keyValue := tkInfraDbModel.KeyValue{Key: key, Value: value}

	result := service.Handler.Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&keyValue)
	return result.Error
}
