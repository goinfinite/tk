package tkInfraDb

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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
		slog.Error(
			"TrailDatabaseFilePathResolutionError",
			slog.String("rawDatabaseFilePath", rawDatabaseFilePath),
			slog.String("error", err.Error()),
		)
		return nil, errors.New(errTrailDatabaseFilePathNotValid)
	}

	databaseFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(
		rawDatabaseFilePath, false,
	)
	if err != nil {
		slog.Error(
			"TrailDatabaseFilePathValidationError",
			slog.Int(
				"rawDatabaseFilePathLength", len(rawDatabaseFilePath),
			),
			slog.String("error", err.Error()),
		)
		return nil, errors.New(errTrailDatabaseFilePathNotValid)
	}

	ormSvc, err := gorm.Open(
		sqlite.Open("file:"+databaseFilePath.String()+DatabaseStandardConnectionParams),
		&gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
			Logger:  gormLogger.Default.LogMode(gormLogger.Silent),
		},
	)
	if err != nil {
		slog.Error(
			"TrailDatabaseConnectionError",
			slog.String("databaseFilePath", databaseFilePath.String()),
			slog.String("error", err.Error()),
		)
		return nil, errors.New(errTrailDatabaseConnectionError)
	}

	dbSvc := &TrailDatabaseService{Handler: ormSvc}
	err = dbSvc.dbMigrate(extraModelsPtrs)
	if err != nil {
		slog.Error(
			"TrailDatabaseMigrationError",
			slog.String("databaseFilePath", databaseFilePath.String()),
			slog.String("error", err.Error()),
		)
		return nil, errors.New(errTrailDatabaseMigrationError)
	}

	return dbSvc, nil
}

func (service *TrailDatabaseService) dbMigrate(extraModelsPtrs []any) error {
	standardModelsPtrs := []any{
		&tkInfraDbModel.ActivityRecord{},
		&tkInfraDbModel.ActivityRecordAffectedResource{},
	}
	allModelsPtrs := append(standardModelsPtrs, extraModelsPtrs...)
	return service.Handler.AutoMigrate(allModelsPtrs...)
}
