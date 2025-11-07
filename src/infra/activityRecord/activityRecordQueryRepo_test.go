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
		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		}

		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("ReadFailedOnEmptyDatabase: %v", err)
		}
		if len(response.ActivityRecords) != 0 {
			t.Errorf("ExpectedEmptyResponse: got %d records", len(response.ActivityRecords))
		}
	})

	t.Run("ReadWithRecordIdFilter", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordLevelVo := tkValueObject.ActivityRecordLevelInfo

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("TEST_CODE")
		if err != nil {
			t.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		resVo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource1")
		if err != nil {
			t.Fatalf("CreateResourceVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevelVo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{resVo},
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

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
			RecordId:   &recordIdVo,
		}

		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("ReadWithRecordIdFilterFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(response.ActivityRecords))
		}
		if response.ActivityRecords[0].RecordId.Uint64() != testRecord.RecordId.Uint64() {
			t.Errorf("RecordIdMismatch: expected %d, got %d", testRecord.RecordId.Uint64(), response.ActivityRecords[0].RecordId.Uint64())
		}
	})

	t.Run("ReadWithRecordLevelFilter", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordLevel1Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode1Vo, err := tkValueObject.NewActivityRecordCode("CODE1")
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
			t.Fatalf("CreateTestActivityRecord1Failed: %v", err)
		}

		recordLevel2Vo := tkValueObject.ActivityRecordLevelError

		recordCode2Vo, err := tkValueObject.NewActivityRecordCode("CODE2")
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
			t.Fatalf("CreateTestActivityRecord2Failed: %v", err)
		}

		recordLevelVo, err := tkValueObject.NewActivityRecordLevel("INFO")
		if err != nil {
			t.Fatalf("CreateRecordLevelVoFailed: %v", err)
		}

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination:  tkDto.PaginationUnpaginated,
			RecordLevel: &recordLevelVo,
		}

		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("ReadWithRecordLevelFilterFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(response.ActivityRecords))
		}
		if response.ActivityRecords[0].RecordLevel.String() != "INFO" {
			t.Errorf("RecordLevelMismatch: expected INFO, got %s", response.ActivityRecords[0].RecordLevel.String())
		}
	})

	t.Run("ReadWithAffectedResourcesFilter", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordLevel3Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode3Vo, err := tkValueObject.NewActivityRecordCode("CODE1")
		if err != nil {
			t.Fatalf("CreateRecordCode3VoFailed: %v", err)
		}

		res1Vo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res1")
		if err != nil {
			t.Fatalf("CreateRes1VoFailed: %v", err)
		}

		createDto3 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel3Vo,
			RecordCode:        recordCode3Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{res1Vo},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto3)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord3Failed: %v", err)
		}

		recordLevel4Vo, err := tkValueObject.NewActivityRecordLevel("INFO")
		if err != nil {
			t.Fatalf("CreateRecordLevel4VoFailed: %v", err)
		}

		recordCode4Vo, err := tkValueObject.NewActivityRecordCode("CODE2")
		if err != nil {
			t.Fatalf("CreateRecordCode4VoFailed: %v", err)
		}

		res2Vo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res2")
		if err != nil {
			t.Fatalf("CreateRes2VoFailed: %v", err)
		}

		createDto4 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel4Vo,
			RecordCode:        recordCode4Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{res2Vo},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto4)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord4Failed: %v", err)
		}

		resVo, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res1")
		if err != nil {
			t.Fatalf("CreateResourceVoFailed: %v", err)
		}

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{resVo},
		}

		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("ReadWithAffectedResourcesFilterFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecord: got %d", len(response.ActivityRecords))
		}
		if response.ActivityRecords[0].RecordCode.String() != "CODE1" {
			t.Errorf("RecordCodeMismatch: expected CODE1, got %s", response.ActivityRecords[0].RecordCode.String())
		}
	})

	t.Run("ReadWithTimeFilters", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordLevel5Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode5Vo, err := tkValueObject.NewActivityRecordCode("TIME_TEST")
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

		testRecord, err := createTestActivityRecord(dbSvc, createDto5)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord5Failed: %v", err)
		}

		// Wait a bit to ensure time difference
		time.Sleep(10 * time.Millisecond)

		now := time.Now()
		beforeTimeVo, err := tkValueObject.NewUnixTime(now.Unix())
		if err != nil {
			t.Fatalf("CreateBeforeTimeVoFailed: %v", err)
		}

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination:      tkDto.PaginationUnpaginated,
			CreatedBeforeAt: &beforeTimeVo,
		}

		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("ReadWithCreatedBeforeAtFilterFailed: %v", err)
		}
		if len(response.ActivityRecords) == 0 {
			t.Errorf("ExpectedRecordsBeforeTime: got none")
		}

		// Check that the test record is included
		found := false
		for _, record := range response.ActivityRecords {
			if record.RecordId.Uint64() == testRecord.RecordId.Uint64() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("TestRecordNotFoundInBeforeTimeFilter")
		}
	})

	t.Run("ReadWithPagination", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordLevel6Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode6Vo, err := tkValueObject.NewActivityRecordCode("PAGE1")
		if err != nil {
			t.Fatalf("CreateRecordCode6VoFailed: %v", err)
		}

		createDto6 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel6Vo,
			RecordCode:        recordCode6Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto6)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord6Failed: %v", err)
		}

		recordLevel7Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode7Vo, err := tkValueObject.NewActivityRecordCode("PAGE2")
		if err != nil {
			t.Fatalf("CreateRecordCode7VoFailed: %v", err)
		}

		createDto7 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel7Vo,
			RecordCode:        recordCode7Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		_, err = createTestActivityRecord(dbSvc, createDto7)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord7Failed: %v", err)
		}

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.Pagination{PageNumber: 0, ItemsPerPage: 1},
		}

		response, err := queryRepo.Read(requestDto)
		if err != nil {
			t.Errorf("ReadWithPaginationFailed: %v", err)
		}
		if len(response.ActivityRecords) != 1 {
			t.Errorf("ExpectedOneRecordWithPagination: got %d", len(response.ActivityRecords))
		}
	})
}

func TestActivityRecordQueryRepoReadFirst(t *testing.T) {
	dbSvc := SetupTestTrailDatabaseService(t)
	queryRepo := NewActivityRecordQueryRepo(dbSvc)

	t.Run("ReadFirstFromEmptyDatabase", func(t *testing.T) {

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		}

		_, err := queryRepo.ReadFirst(requestDto)
		if err == nil {
			t.Errorf("ReadFirstSucceededWhenShouldFailOnEmptyDatabase")
		}
		expectedError := "ActivityRecordNotFound"
		if err.Error() != expectedError {
			t.Errorf("UnexpectedErrorMessage: expected '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("ReadFirstSuccess", func(t *testing.T) {
		recordLevel8Vo := tkValueObject.ActivityRecordLevelInfo

		recordCode8Vo, err := tkValueObject.NewActivityRecordCode("FIRST_TEST")
		if err != nil {
			t.Fatalf("CreateRecordCode8VoFailed: %v", err)
		}

		createDto8 := tkDto.CreateActivityRecord{
			RecordLevel:       recordLevel8Vo,
			RecordCode:        recordCode8Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
			RecordDetails:     nil,
			OperatorAccountId: nil,
			OperatorIpAddress: nil,
		}

		testRecord, err := createTestActivityRecord(dbSvc, createDto8)
		if err != nil {
			t.Fatalf("CreateTestActivityRecord8Failed: %v", err)
		}

		requestDto := tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		}

		record, err := queryRepo.ReadFirst(requestDto)
		if err != nil {
			t.Errorf("ReadFirstFailed: %v", err)
		}
		if record.RecordId.Uint64() != testRecord.RecordId.Uint64() {
			t.Errorf("ReadFirstReturnedWrongRecord: expected ID %d, got %d", testRecord.RecordId.Uint64(), record.RecordId.Uint64())
		}
	})
}
