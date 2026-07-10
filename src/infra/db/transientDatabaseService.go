package tkInfraDb

import (
	"errors"
	"time"

	"github.com/glebarez/sqlite"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
	"gorm.io/gorm"
)

const (
	errTransientDatabaseConnectionError string = "TransientDatabaseConnectionError"
	errTransientDatabaseMigrationError  string = "TransientDatabaseMigrationError"
)

type TransientDatabaseService struct {
	Handler *gorm.DB
}

func NewTransientDatabaseService() (*TransientDatabaseService, error) {
	ormSvc, err := gorm.Open(
		sqlite.Open(
			"file::memory:" + DatabaseStandardConnectionParams,
		),
		&gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		},
	)
	if err != nil {
		return nil, errors.New(errTransientDatabaseConnectionError)
	}

	dbSvc := &TransientDatabaseService{Handler: ormSvc}
	return dbSvc, dbSvc.dbMigrate()
}

func (service *TransientDatabaseService) dbMigrate() error {
	err := service.Handler.AutoMigrate(&tkInfraDbModel.HoneypotHitModel{})
	if err != nil {
		return errors.New(
			errTransientDatabaseMigrationError + ": " + err.Error(),
		)
	}

	return nil
}
