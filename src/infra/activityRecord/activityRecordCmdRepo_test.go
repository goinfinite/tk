package tkInfraActivityRecord

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
)

func TestActivityRecordCmdRepoCreate(t *testing.T) {
	dbSvc := SetupTestTrailDatabaseService(t)

	t.Run("CreateBasicRecord", func(t *testing.T) {
		recordCodeVo, err := tkValueObject.NewActivityRecordCode("CREATE_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Errorf("CreateBasicRecordFailed: %v", err)
		}
	})

	t.Run("CreateRecordWithAllFields", func(t *testing.T) {
		recordCodeVo, err := tkValueObject.NewActivityRecordCode("FULL_CREATE_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/test-resource")
		if err != nil {
			t.Fatalf("CreateResourceVoFailed: %v", err)
		}

		operatorAccountIdVo, err := tkValueObject.NewAccountId(123)
		if err != nil {
			t.Fatalf("CreateOperatorAccountIdVoFailed: %v", err)
		}

		operatorIpAddressVo, err := tkValueObject.NewIpAddress("192.168.1.1")
		if err != nil {
			t.Fatalf("CreateOperatorIpAddressVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelError,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri},
			RecordDetails:     map[string]any{"key": "value"},
			OperatorAccountId: &operatorAccountIdVo,
			OperatorIpAddress: &operatorIpAddressVo,
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Errorf("CreateRecordWithAllFieldsFailed: %v", err)
		}
	})

	t.Run("CreateRecordWithMultipleAffectedResources", func(t *testing.T) {
		recordCodeVo, err := tkValueObject.NewActivityRecordCode("MULTI_RES_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource1")
		if err != nil {
			t.Fatalf("CreateResource1VoFailed: %v", err)
		}
		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource2")
		if err != nil {
			t.Fatalf("CreateResource2VoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelWarning,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Errorf("CreateRecordWithMultipleAffectedResourcesFailed: %v", err)
		}
	})
}

func TestActivityRecordCmdRepoDelete(t *testing.T) {
	t.Run("DeleteWithRecordId", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_ID_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		testRecord, err := createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordIdVo, err := tkValueObject.NewActivityRecordId(testRecord.RecordId.Uint64())
		if err != nil {
			t.Fatalf("CreateRecordIdVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			&recordIdVo, nil, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			t.Errorf("DeleteWithRecordIdFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
			RecordId:   &recordIdVo,
		})
		if err != nil {
			t.Errorf("VerifyDeleteFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 0 {
			t.Errorf("RecordNotDeleted: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
	})

	t.Run("DeleteWithRecordLevel", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCode1Vo, err := tkValueObject.NewActivityRecordCode("DELETE_LEVEL_1")
		if err != nil {
			t.Fatalf("CreateRecordCode1VoFailed: %v", err)
		}

		createDto1 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode1Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto1)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordCode2Vo, err := tkValueObject.NewActivityRecordCode("DELETE_LEVEL_2")
		if err != nil {
			t.Fatalf("CreateRecordCode2VoFailed: %v", err)
		}

		createDto2 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelError,
			RecordCode:        recordCode2Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto2)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, &tkValueObject.ActivityRecordLevelInfo, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			t.Errorf("DeleteWithRecordLevelFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			t.Errorf("VerifyDeleteLevelFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecordRemaining: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordLevel != tkValueObject.ActivityRecordLevelError {
			t.Errorf(
				"WrongRecordRemaining: expected ERROR level, got %s",
				responseDto.ActivityRecords[0].RecordLevel.String(),
			)
		}
	})

	t.Run("DeleteWithRecordCode", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCode3Vo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_1")
		if err != nil {
			t.Fatalf("CreateRecordCode3VoFailed: %v", err)
		}

		createDto3 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode3Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto3)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordCode4Vo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_2")
		if err != nil {
			t.Fatalf("CreateRecordCode4VoFailed: %v", err)
		}

		createDto4 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode4Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto4)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_1")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, nil, &recordCodeVo, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			t.Errorf("DeleteWithRecordCodeFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			t.Errorf("VerifyDeleteCodeFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecordRemaining: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordCode.String() != "DELETE_CODE_2" {
			t.Errorf(
				"WrongRecordRemaining: expected DELETE_CODE_2, got %s",
				responseDto.ActivityRecords[0].RecordCode.String(),
			)
		}
	})

	t.Run("DeleteWithTimeFilters", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCode5Vo, err := tkValueObject.NewActivityRecordCode("DELETE_TIME_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCode5VoFailed: %v", err)
		}

		createDto5 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode5Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto5)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		// Use a future time to ensure nothing is deleted
		futureTime, err := tkValueObject.NewUnixTime(time.Now().Add(time.Hour).Unix())
		if err != nil {
			t.Fatalf("CreateFutureTimeVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, nil, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, &futureTime,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			t.Errorf("DeleteWithTimeFiltersFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			t.Errorf("VerifyDeleteTimeFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("RecordShouldStillExist: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
	})

	t.Run("DeleteWithMultipleMatchingAffectedResources", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_MULTI_RES_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/del-multi1")
		if err != nil {
			t.Fatalf("CreateRes1VoFailed: %v", err)
		}
		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/del-multi2")
		if err != nil {
			t.Fatalf("CreateRes2VoFailed: %v", err)
		}

		_, err = createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		})
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, nil, nil, []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			t.Errorf("DeleteWithMultipleMatchingAffectedResourcesFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		})
		if err != nil {
			t.Errorf("VerifyDeleteMultipleResourcesFailed: %v", err)
		}

		if len(responseDto.ActivityRecords) != 0 {
			t.Errorf("ExpectedZeroRecords: got %d", len(responseDto.ActivityRecords))
		}
	})
}

func SetupTestTrailDatabaseService(t *testing.T) *tkInfraDb.TrailDatabaseService {
	t.Helper()
	tempDir := t.TempDir()
	rawDatabaseFilePath := filepath.Join(tempDir, strings.ReplaceAll(t.Name(), "/", "_")+".db")
	rawDatabaseFilePath, err := filepath.Abs(rawDatabaseFilePath)
	if err != nil {
		t.Fatalf("ResolveRawDatabaseFilePathFailed: %v", err)
	}

	databaseFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawDatabaseFilePath, false)
	if err != nil {
		t.Fatalf("ResolveDatabaseFilePathFailed: %v", err)
	}

	originalValue := os.Getenv(tkInfraDb.TrailDatabaseFilePathEnvVarName)
	t.Cleanup(func() { os.Setenv(tkInfraDb.TrailDatabaseFilePathEnvVarName, originalValue) })
	os.Setenv(tkInfraDb.TrailDatabaseFilePathEnvVarName, databaseFilePath.String())

	dbSvc, err := tkInfraDb.NewTrailDatabaseService([]any{})
	if err != nil {
		t.Fatalf("SetupTrailDatabaseServiceFailed: %v", err)
	}

	return dbSvc
}

func createTestActivityRecord(
	trailDbSvc *tkInfraDb.TrailDatabaseService,
	createDto tkDto.CreateActivityRecord,
) (recordEntity tkEntity.ActivityRecord, err error) {
	queryRepo := NewActivityRecordQueryRepo(trailDbSvc)
	cmdRepo := NewActivityRecordCmdRepo(trailDbSvc)
	err = cmdRepo.Create(createDto)
	if err != nil {
		return recordEntity, err
	}

	requestDto := tkDto.ReadActivityRecordsRequest{
		Pagination: tkDto.PaginationUnpaginated,
		RecordCode: &createDto.RecordCode,
	}
	responseDto, err := queryRepo.Read(requestDto)
	if err != nil {
		return recordEntity, err
	}
	if len(responseDto.ActivityRecords) < 1 {
		return recordEntity, errors.New(
			"ExpectedAtLeastOneRecordAfterCreate: got " + strconv.Itoa(len(responseDto.ActivityRecords)),
		)
	}

	return responseDto.ActivityRecords[0], nil
}
