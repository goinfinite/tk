package tkInfraDb

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
	"gorm.io/gorm"
)

const (
	DatabaseStandardConnectionParams string = "?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"
	TrailDatabaseFilePathEnvVarName  string = "TRAIL_DATABASE_FILE_PATH"
	errTrailDatabaseFilePathNotSet   string = "TrailDatabaseFilePathNotSet"
	errTrailDatabaseFilePathNotValid string = "TrailDatabaseFilePathNotValid"
	errTrailDatabaseConnectionError  string = "TrailDatabaseConnectionError"
	errTrailDatabaseMigrationError   string = "TrailDatabaseMigrationError"
)

type TrailDatabaseService struct {
	Handler *gorm.DB
}

func NewTrailDatabaseService(extraModelsPtrs []any) (*TrailDatabaseService, error) {
	rawDatabaseFilePath := os.Getenv(TrailDatabaseFilePathEnvVarName)
	if rawDatabaseFilePath == "" {
		return nil, errors.New(errTrailDatabaseFilePathNotSet)
	}
	rawDatabaseFilePath, err := filepath.Abs(rawDatabaseFilePath)
	if err != nil {
		return nil, errors.New(errTrailDatabaseFilePathNotValid)
	}

	databaseFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawDatabaseFilePath, false)
	if err != nil {
		return nil, errors.New(errTrailDatabaseFilePathNotValid)
	}

	ormSvc, err := gorm.Open(
		sqlite.Open("file:"+databaseFilePath.String()+DatabaseStandardConnectionParams),
		&gorm.Config{},
	)
	if err != nil {
		return nil, errors.New(errTrailDatabaseConnectionError)
	}

	dbSvc := &TrailDatabaseService{Handler: ormSvc}
	return dbSvc, dbSvc.dbMigrate(extraModelsPtrs)
}

func (service *TrailDatabaseService) dbMigrate(extraModelsPtrs []any) error {
	standardModelsPtrs := []any{
		&tkInfraDbModel.ActivityRecord{},
		&tkInfraDbModel.ActivityRecordAffectedResource{},
	}
	allModelsPtrs := append(standardModelsPtrs, extraModelsPtrs...)
	err := service.Handler.AutoMigrate(allModelsPtrs...)
	if err != nil {
		return errors.New(errTrailDatabaseMigrationError + ": " + err.Error())
	}

	return nil
}
