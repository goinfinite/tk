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

func SetupTestTrailDatabaseService(t *testing.T) *tkInfraDb.TrailDatabaseService {
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
	cmdRepo := NewActivityRecordCmdRepo(trailDbSvc)
	err = cmdRepo.Create(createDto)
	if err != nil {
		return recordEntity, err
	}

	queryRepo := NewActivityRecordQueryRepo(trailDbSvc)
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

func TestActivityRecordCmdRepoCreate(t *testing.T) {
	dbSvc := SetupTestTrailDatabaseService(t)

	t.Run("CreateBasicRecord", func(t *testing.T) {
		recordLevelVo := tkValueObject.ActivityRecordLevelInfo

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("CREATE_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevelVo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Errorf("CreateBasicRecordFailed: %v", err)
		}
	})

	t.Run("CreateRecordWithAllFields", func(t *testing.T) {
		recordLevelVo := tkValueObject.ActivityRecordLevelError

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("FULL_CREATE_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		resVo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/test-resource")
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
			RecordLevel:       recordLevelVo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{resVo},
			RecordDetails:     map[string]interface{}{"key": "value"},
			OperatorAccountId: &operatorAccountIdVo,
			OperatorIpAddress: &operatorIpAddressVo,
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Errorf("CreateRecordWithAllFieldsFailed: %v", err)
		}
	})

	t.Run("CreateRecordWithMultipleAffectedResources", func(t *testing.T) {
		recordLevelVo := tkValueObject.ActivityRecordLevelWarning

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("MULTI_RES_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		res1Vo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource1")
		if err != nil {
			t.Fatalf("CreateResource1VoFailed: %v", err)
		}
		res2Vo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource2")
		if err != nil {
			t.Fatalf("CreateResource2VoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevelVo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{res1Vo, res2Vo},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
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
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordLevelVo := tkValueObject.ActivityRecordLevelInfo

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("DELETE_ID_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevelVo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
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

		// Verify it's deleted
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
			RecordId:   &recordIdVo,
		}
		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("VerifyDeleteFailed: %v", err)
		}
		if len(response.ActivityRecords) != 0 {
			t.Errorf("RecordNotDeleted: still found %d records", len(response.ActivityRecords))
		}
	})

	t.Run("DeleteWithRecordLevel", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordLevel1Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode1Vo, err := tkValueObject.NewActivityRecordCode("DELETE_LEVEL_1")
		if err != nil {
			t.Fatalf("CreateRecordCode1VoFailed: %v", err)
		}

		createDto1 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel1Vo,
			RecordCode:        recordCode1Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto1)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordLevel2Vo := tkValueObject.ActivityRecordLevelError

		recordCode2Vo, err := tkValueObject.NewActivityRecordCode("DELETE_LEVEL_2")
		if err != nil {
			t.Fatalf("CreateRecordCode2VoFailed: %v", err)
		}

		createDto2 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel2Vo,
			RecordCode:        recordCode2Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto2)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordLevelVo, err := tkValueObject.NewActivityRecordLevel("INFO")
		if err != nil {
			t.Fatalf("CreateRecordLevelVoFailed: %v", err)
		}

		deleteDto := tkDto.NewDeleteActivityRecord(
			nil, &recordLevelVo, nil, []tkValueObject.SystemResourceIdentifier{},
			nil, nil, nil, nil,
		)

		err = cmdRepo.Delete(deleteDto)
		if err != nil {
			t.Errorf("DeleteWithRecordLevelFailed: %v", err)
		}

		// Verify only INFO records are deleted
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		}
		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("VerifyDeleteLevelFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecordRemaining: got %d", len(response.ActivityRecords))
		}
		if response.ActivityRecords[0].RecordLevel.String() != "ERROR" {
			t.Errorf("WrongRecordRemaining: expected ERROR level, got %s", response.ActivityRecords[0].RecordLevel.String())
		}
	})

	t.Run("DeleteWithRecordCode", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)
		// Create records with different codes
		recordLevel3Vo, err := tkValueObject.NewActivityRecordLevel("INFO")
		if err != nil {
			t.Fatalf("CreateRecordLevel3VoFailed: %v", err)
		}

		recordCode3Vo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_1")
		if err != nil {
			t.Fatalf("CreateRecordCode3VoFailed: %v", err)
		}

		createDto3 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel3Vo,
			RecordCode:        recordCode3Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto3)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordLevel4Vo, err := tkValueObject.NewActivityRecordLevel("INFO")
		if err != nil {
			t.Fatalf("CreateRecordLevel4VoFailed: %v", err)
		}

		recordCode4Vo, err := tkValueObject.NewActivityRecordCode("DELETE_CODE_2")
		if err != nil {
			t.Fatalf("CreateRecordCode4VoFailed: %v", err)
		}

		createDto4 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel4Vo,
			RecordCode:        recordCode4Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
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

		// Verify only records with DELETE_CODE_1 are deleted
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		}
		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("VerifyDeleteCodeFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecordRemaining: got %d", len(response.ActivityRecords))
		}
		if response.ActivityRecords[0].RecordCode.String() != "DELETE_CODE_2" {
			t.Errorf("WrongRecordRemaining: expected DELETE_CODE_2, got %s", response.ActivityRecords[0].RecordCode.String())
		}
	})

	t.Run("DeleteWithTimeFilters", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		cmdRepo := NewActivityRecordCmdRepo(dbSvc)

		recordLevel5Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode5Vo, err := tkValueObject.NewActivityRecordCode("DELETE_TIME_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCode5VoFailed: %v", err)
		}

		createDto5 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel5Vo,
			RecordCode:        recordCode5Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
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

		// Verify record still exists
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		}
		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("VerifyDeleteTimeFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("RecordShouldStillExist: got %d records", len(response.ActivityRecords))
		}
	})
}
