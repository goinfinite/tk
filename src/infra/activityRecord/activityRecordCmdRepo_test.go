package tkInfraActivityRecord

import (
	"errors"
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

func TestActivityRecordCmdRepoCreate(test *testing.T) {
	dbSvc := SetupTestTrailDatabaseService(test)

	test.Run("CreateBasicRecord", func(subtest *testing.T) {
		recordCodeVo, err := tkValueObject.NewActivityRecordCode("CREATE_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			subtest.Errorf("CreateBasicRecordFailed: %v", err)
		}
	})

	test.Run("CreateRecordWithAllFields", func(subtest *testing.T) {
		recordCodeVo, err := tkValueObject.NewActivityRecordCode("FULL_CREATE_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/test-resource")
		if err != nil {
			subtest.Fatalf("CreateResourceVoFailed: %v", err)
		}

		accountId, err := tkValueObject.NewAccountId(123)
		if err != nil {
			subtest.Fatalf("CreateAccountIdVoFailed: %v", err)
		}
		operatorSri := tkValueObject.NewSriAccount(accountId)

		operatorIpAddressVo, err := tkValueObject.NewIpAddress("192.168.1.1")
		if err != nil {
			subtest.Fatalf("CreateOperatorIpAddressVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelError,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri},
			RecordDetails:     map[string]any{"key": "value"},
			OperatorSri:       &operatorSri,
			OperatorIpAddress: &operatorIpAddressVo,
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			subtest.Errorf("CreateRecordWithAllFieldsFailed: %v", err)
		}
	})

	test.Run("CreateRecordWithMultipleAffectedResources", func(subtest *testing.T) {
		recordCodeVo, err := tkValueObject.NewActivityRecordCode("MULTI_RES_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource1")
		if err != nil {
			subtest.Fatalf("CreateResource1VoFailed: %v", err)
		}
		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource2")
		if err != nil {
			subtest.Fatalf("CreateResource2VoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelWarning,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			subtest.Errorf("CreateRecordWithMultipleAffectedResourcesFailed: %v", err)
		}
	})
}

func TestActivityRecordCmdRepoDelete(test *testing.T) {
	test.Run("DeleteWithRecordId", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_ID_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		testRecord, err := createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordIdVo, err := tkValueObject.NewActivityRecordId(testRecord.RecordId.Uint64())
		if err != nil {
			subtest.Fatalf("CreateRecordIdVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			&recordIdVo, nil, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			subtest.Errorf("DeleteWithRecordIdFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
			RecordId:   &recordIdVo,
		})
		if err != nil {
			subtest.Errorf("VerifyDeleteFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 0 {
			subtest.Errorf("RecordNotDeleted: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
	})

	test.Run("DeleteWithRecordLevel", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCode1Vo, err := tkValueObject.NewActivityRecordCode("DELETE_LEVEL_1")
		if err != nil {
			subtest.Fatalf("CreateRecordCode1VoFailed: %v", err)
		}

		createDto1 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode1Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto1)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordCode2Vo, err := tkValueObject.NewActivityRecordCode("DELETE_LEVEL_2")
		if err != nil {
			subtest.Fatalf("CreateRecordCode2VoFailed: %v", err)
		}

		createDto2 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelError,
			RecordCode:        recordCode2Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto2)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, &tkValueObject.ActivityRecordLevelInfo, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			subtest.Errorf("DeleteWithRecordLevelFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			subtest.Errorf("VerifyDeleteLevelFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecordRemaining: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordLevel != tkValueObject.ActivityRecordLevelError {
			subtest.Errorf(
				"WrongRecordRemaining: expected ERROR level, got %s",
				responseDto.ActivityRecords[0].RecordLevel.String(),
			)
		}
	})

	test.Run("DeleteWithRecordCode", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCode3Vo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_1")
		if err != nil {
			subtest.Fatalf("CreateRecordCode3VoFailed: %v", err)
		}

		createDto3 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode3Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto3)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordCode4Vo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_2")
		if err != nil {
			subtest.Fatalf("CreateRecordCode4VoFailed: %v", err)
		}

		createDto4 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode4Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto4)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_1")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, nil, &recordCodeVo, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			subtest.Errorf("DeleteWithRecordCodeFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			subtest.Errorf("VerifyDeleteCodeFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecordRemaining: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordCode.String() != "DELETE_CODE_2" {
			subtest.Errorf(
				"WrongRecordRemaining: expected DELETE_CODE_2, got %s",
				responseDto.ActivityRecords[0].RecordCode.String(),
			)
		}
	})

	test.Run("DeleteWithTimeFilters", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCode5Vo, err := tkValueObject.NewActivityRecordCode("DELETE_TIME_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCode5VoFailed: %v", err)
		}

		createDto5 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode5Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto5)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		// Use a future time to ensure nothing is deleted
		futureTime, err := tkValueObject.NewUnixTime(time.Now().Add(time.Hour).Unix())
		if err != nil {
			subtest.Fatalf("CreateFutureTimeVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, nil, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, &futureTime,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			subtest.Errorf("DeleteWithTimeFiltersFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			subtest.Errorf("VerifyDeleteTimeFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("RecordShouldStillExist: FoundRecordsCount: %d", len(responseDto.ActivityRecords))
		}
	})

	test.Run("DeleteWithMultipleMatchingAffectedResources", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_MULTI_RES_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/del-multi1")
		if err != nil {
			subtest.Fatalf("CreateRes1VoFailed: %v", err)
		}
		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/del-multi2")
		if err != nil {
			subtest.Fatalf("CreateRes2VoFailed: %v", err)
		}

		_, err = createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		})
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, nil, nil, []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			subtest.Errorf("DeleteWithMultipleMatchingAffectedResourcesFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		})
		if err != nil {
			subtest.Errorf("VerifyDeleteMultipleResourcesFailed: %v", err)
		}

		if len(responseDto.ActivityRecords) != 0 {
			subtest.Errorf("ExpectedZeroRecords: got %d", len(responseDto.ActivityRecords))
		}
	})
}

func SetupTestTrailDatabaseService(test *testing.T) *tkInfraDb.TrailDatabaseService {
	test.Helper()
	tempDir := test.TempDir()
	rawDatabaseFilePath := filepath.Join(tempDir, strings.ReplaceAll(test.Name(), "/", "_")+".db")
	rawDatabaseFilePath, err := filepath.Abs(rawDatabaseFilePath)
	if err != nil {
		test.Fatalf("ResolveRawDatabaseFilePathFailed: %v", err)
	}

	databaseFilePath, err := tkValueObject.NewUnixAbsoluteFilePath(rawDatabaseFilePath, false)
	if err != nil {
		test.Fatalf("ResolveDatabaseFilePathFailed: %v", err)
	}

	test.Setenv(tkInfraDb.TrailDatabaseFilePathEnvVarName, databaseFilePath.String())

	dbSvc, err := tkInfraDb.NewTrailDatabaseService([]any{})
	if err != nil {
		test.Fatalf("SetupTrailDatabaseServiceFailed: %v", err)
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
