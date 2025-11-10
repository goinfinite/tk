package tkInfraActivityRecord

import (
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestActivityRecordQueryRepoRead(t *testing.T) {
	t.Run("EmptyDatabase", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			t.Errorf("ReadFailedOnEmptyDatabase: %v", err)
		}
		if len(responseDto.ActivityRecords) != 0 {
			t.Errorf("ExpectedEmptyResponse: got %d records", len(responseDto.ActivityRecords))
		}
	})

	t.Run("ReadWithRecordIdFilter", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("TEST_CODE")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource1")
		if err != nil {
			t.Fatalf("CreateResourceVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri},
		}

		testRecord, err := createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordIdVo, err := tkValueObject.NewActivityRecordId(testRecord.RecordId.Uint64())
		if err != nil {
			t.Fatalf("CreateRecordIdVoFailed: %v", err)
		}
		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
			RecordId:   &recordIdVo,
		})
		if err != nil {
			t.Errorf("ReadWithRecordIdFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordId.Uint64() != testRecord.RecordId.Uint64() {
			t.Errorf(
				"RecordIdMismatch: expected %d, got %d",
				testRecord.RecordId.Uint64(), responseDto.ActivityRecords[0].RecordId.Uint64(),
			)
		}
	})

	t.Run("ReadWithRecordLevelFilter", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode1Vo, err := tkValueObject.NewActivityRecordCode("CODE1")
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
			t.Fatalf("CreateTestActivityRecord1Failed: %v", err)
		}

		recordCode2Vo, err := tkValueObject.NewActivityRecordCode("CODE2")
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
			t.Fatalf("CreateTestActivityRecord2Failed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:  tkDto.PaginationUnpaginated,
			RecordLevel: &tkValueObject.ActivityRecordLevelInfo,
		})
		if err != nil {
			t.Errorf("ReadWithRecordLevelFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordLevel.String() != "INFO" {
			t.Errorf(
				"RecordLevelMismatch: expected INFO, got %s",
				responseDto.ActivityRecords[0].RecordLevel.String(),
			)
		}
	})

	t.Run("ReadWithAffectedResourcesFilter", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode3Vo, err := tkValueObject.NewActivityRecordCode("CODE1")
		if err != nil {
			t.Fatalf("CreateRecordCode3VoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res1")
		if err != nil {
			t.Fatalf("CreateRes1VoFailed: %v", err)
		}

		createDto3 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode3Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1},
		}

		_, err = createTestActivityRecord(dbSvc, createDto3)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord3Failed: %v", err)
		}

		recordCode4Vo, err := tkValueObject.NewActivityRecordCode("CODE2")
		if err != nil {
			t.Fatalf("CreateRecordCode4VoFailed: %v", err)
		}

		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res2")
		if err != nil {
			t.Fatalf("CreateRes2VoFailed: %v", err)
		}

		createDto4 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode4Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri2},
		}

		_, err = createTestActivityRecord(dbSvc, createDto4)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord4Failed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1},
		})
		if err != nil {
			t.Errorf("ReadWithAffectedResourcesFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordCode.String() != "CODE1" {
			t.Errorf(
				"RecordCodeMismatch: expected CODE1, got %s",
				responseDto.ActivityRecords[0].RecordCode.String(),
			)
		}
	})

	t.Run("ReadWithMultipleMatchingAffectedResources", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("MULTI_RES_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/multi1")
		if err != nil {
			t.Fatalf("CreateRes1VoFailed: %v", err)
		}
		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/multi2")
		if err != nil {
			t.Fatalf("CreateRes2VoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			t.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		})
		if err != nil {
			t.Errorf("ReadWithMultipleMatchingAffectedResourcesFailed: %v", err)
		}

		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}

		if responseDto.Pagination.ItemsTotal == nil {
			t.Fatalf("Expected ItemsTotal to be non-nil")
		}
		if *responseDto.Pagination.ItemsTotal != 1 {
			t.Errorf("ExpectedTotalItemsToBeOne: got %d", *responseDto.Pagination.ItemsTotal)
		}
	})

	t.Run("ReadWithTimeFilters", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode5Vo, err := tkValueObject.NewActivityRecordCode("TIME_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCode5VoFailed: %v", err)
		}

		createDto5 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode5Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		testRecord, err := createTestActivityRecord(dbSvc, createDto5)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord5Failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		now := time.Now()
		beforeTimeVo, err := tkValueObject.NewUnixTime(now.Unix())
		if err != nil {
			t.Fatalf("CreateBeforeTimeVoFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:      tkDto.PaginationUnpaginated,
			CreatedBeforeAt: &beforeTimeVo,
		})
		if err != nil {
			t.Errorf("ReadWithCreatedBeforeAtFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) == 0 {
			t.Errorf("ExpectedRecordsBeforeTime: got none")
		}

		recordExists := false
		for _, record := range responseDto.ActivityRecords {
			if record.RecordId.Uint64() == testRecord.RecordId.Uint64() {
				recordExists = true
				break
			}
		}
		if !recordExists {
			t.Errorf("TestRecordNotFoundInBeforeTimeFilter")
		}
	})

	t.Run("ReadWithPagination", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode6Vo, err := tkValueObject.NewActivityRecordCode("PAGE1")
		if err != nil {
			t.Fatalf("CreateRecordCode6VoFailed: %v", err)
		}

		createDto6 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode6Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto6)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord6Failed: %v", err)
		}

		recordCode7Vo, err := tkValueObject.NewActivityRecordCode("PAGE2")
		if err != nil {
			t.Fatalf("CreateRecordCode7VoFailed: %v", err)
		}

		createDto7 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode7Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto7)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord7Failed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.Pagination{PageNumber: 0, ItemsPerPage: 1},
		})
		if err != nil {
			t.Errorf("ReadWithPaginationFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecordWithPagination: got %d", len(responseDto.ActivityRecords))
		}
	})
}

func TestActivityRecordQueryRepoReadFirst(t *testing.T) {
	dbSvc := SetupTestTrailDatabaseService(t)
	queryRepo := NewActivityRecordQueryRepo(dbSvc)

	t.Run("ReadFirstFromEmptyDatabase", func(t *testing.T) {
		_, err := queryRepo.ReadFirst(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err == nil {
			t.Errorf("ReadFirstSucceededWhenShouldFailOnEmptyDatabase")
		}
		expectedError := "ActivityRecordNotFound"
		if err.Error() != expectedError {
			t.Errorf("UnexpectedErrorMessage: expected '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("ReadFirstSuccess", func(t *testing.T) {
		recordCode8Vo, err := tkValueObject.NewActivityRecordCode("FIRST_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCode8VoFailed: %v", err)
		}

		createDto8 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode8Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		createdRecordEntity, err := createTestActivityRecord(dbSvc, createDto8)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord8Failed: %v", err)
		}

		readRecordEntity, err := queryRepo.ReadFirst(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			t.Errorf("ReadFirstFailed: %v", err)
		}
		if readRecordEntity.RecordId.Uint64() != createdRecordEntity.RecordId.Uint64() {
			t.Errorf(
				"ReadFirstReturnedWrongRecord: expected ID %d, got %d",
				createdRecordEntity.RecordId.Uint64(), readRecordEntity.RecordId.Uint64(),
			)
		}
	})
}
