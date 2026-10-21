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

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("TIME_TEST")
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

		time.Sleep(10 * time.Millisecond)

		testCases := []struct {
			name            string
			timeFilterValue *tkValueObject.UnixTime
			isBeforeFilter  bool
		}{
			{
				name: "CreatedBeforeAt_Future",
				timeFilterValue: func() *tkValueObject.UnixTime {
					beforeTimeVoFuture := tkValueObject.NewUnixTimeAfterNow(1 * time.Minute)
					return &beforeTimeVoFuture
				}(),
				isBeforeFilter: true,
			},
			{
				name: "CreatedBeforeAt_DistantFuture",
				timeFilterValue: func() *tkValueObject.UnixTime {
					beforeTimeVoDistant := tkValueObject.NewUnixTimeAfterNow(1 * time.Hour)
					return &beforeTimeVoDistant
				}(),
				isBeforeFilter: true,
			},
			{
				name: "CreatedAfterAt_Recent",
				timeFilterValue: func() *tkValueObject.UnixTime {
					afterTimeVoRecent := tkValueObject.NewUnixTimeBeforeNow(1 * time.Minute)
					return &afterTimeVoRecent
				}(),
				isBeforeFilter: false,
			},
			{
				name: "CreatedAfterAt_Medium",
				timeFilterValue: func() *tkValueObject.UnixTime {
					afterTimeVoMedium := tkValueObject.NewUnixTimeBeforeNow(1 * time.Hour)
					return &afterTimeVoMedium
				}(),
				isBeforeFilter: false,
			},
			{
				name: "CreatedAfterAt_Old",
				timeFilterValue: func() *tkValueObject.UnixTime {
					afterTimeVoOld := tkValueObject.NewUnixTimeBeforeNow(24 * time.Hour)
					return &afterTimeVoOld
				}(),
				isBeforeFilter: false,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				requestDto := tkDto.ReadActivityRecordsRequest{
					RecordLevel: &tkValueObject.ActivityRecordLevelInfo,
					RecordCode:  &recordCodeVo,
				}

				if testCase.isBeforeFilter {
					requestDto.CreatedBeforeAt = testCase.timeFilterValue
				}
				if !testCase.isBeforeFilter {
					requestDto.CreatedAfterAt = testCase.timeFilterValue
				}

				recordEntity, err := queryRepo.ReadFirst(requestDto)
				if err != nil {
					t.Errorf("ReadWith%sFilterFailed: %v", testCase.name, err)
				}
				if recordEntity.RecordId.Uint64() != testRecord.RecordId.Uint64() {
					t.Errorf("TestRecordNotFoundIn%sFilter", testCase.name)
				}
			})
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

func TestActivityRecordQueryRepoCount(t *testing.T) {
	t.Run("EmptyDatabase", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		count, err := queryRepo.Count(tkDto.ReadActivityRecordsRequest{})
		if err != nil {
			t.Errorf("CountFailedOnEmptyDatabase: %v", err)
		}
		if count != 0 {
			t.Errorf("ExpectedZeroCount: got %d", count)
		}
	})

	t.Run("CountWithFilters", func(t *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(t)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		codeAVo, err := tkValueObject.NewActivityRecordCode("COUNT_CODE_A")
		if err != nil {
			t.Fatalf("CreateCodeAVoFailed: %v", err)
		}
		codeBVo, err := tkValueObject.NewActivityRecordCode("COUNT_CODE_B")
		if err != nil {
			t.Fatalf("CreateCodeBVoFailed: %v", err)
		}

		sri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/count1")
		if err != nil {
			t.Fatalf("CreateRes1VoFailed: %v", err)
		}
		sri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/count2")
		if err != nil {
			t.Fatalf("CreateRes2VoFailed: %v", err)
		}

		accountId1Vo, err := tkValueObject.NewAccountId(1)
		if err != nil {
			t.Fatalf("CreateAccountId1VoFailed: %v", err)
		}
		operatorSri1 := tkValueObject.NewSriAccount(accountId1Vo)
		accountId2Vo, err := tkValueObject.NewAccountId(2)
		if err != nil {
			t.Fatalf("CreateAccountId2VoFailed: %v", err)
		}
		operatorSri2 := tkValueObject.NewSriAccount(accountId2Vo)

		ip1, err := tkValueObject.NewIpAddress("192.168.1.1")
		if err != nil {
			t.Fatalf("CreateIp1VoFailed: %v", err)
		}
		ip2, err := tkValueObject.NewIpAddress("192.168.1.2")
		if err != nil {
			t.Fatalf("CreateIp2VoFailed: %v", err)
		}

		_, err = createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        codeAVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{sri1},
			OperatorSri:       &operatorSri1,
			OperatorIpAddress: &ip1,
		})
		if err != nil {
			t.Fatalf("CreateRecordAFailed: %v", err)
		}

		recordB, err := createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelError,
			RecordCode:        codeBVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{sri2},
			OperatorSri:       &operatorSri2,
			OperatorIpAddress: &ip2,
		})
		if err != nil {
			t.Fatalf("CreateRecordBFailed: %v", err)
		}

		_, err = createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        codeAVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		})
		if err != nil {
			t.Fatalf("CreateRecordCFailed: %v", err)
		}

		recordIdVo, err := tkValueObject.NewActivityRecordId(recordB.RecordId.Uint64())
		if err != nil {
			t.Fatalf("CreateRecordIdVoFailed: %v", err)
		}

		futureTimeVo := tkValueObject.NewUnixTimeAfterNow(1 * time.Hour)
		pastTimeVo := tkValueObject.NewUnixTimeBeforeNow(1 * time.Hour)
		infoLevel := tkValueObject.ActivityRecordLevelInfo
		errorLevel := tkValueObject.ActivityRecordLevelError

		testCases := []struct {
			name          string
			requestDto    tkDto.ReadActivityRecordsRequest
			expectedCount uint64
		}{
			{
				name:          "NoFilters",
				requestDto:    tkDto.ReadActivityRecordsRequest{},
				expectedCount: 3,
			},
			{
				name:          "RecordIdFilter",
				requestDto:    tkDto.ReadActivityRecordsRequest{RecordId: &recordIdVo},
				expectedCount: 1,
			},
			{
				name:          "RecordCodeFilter",
				requestDto:    tkDto.ReadActivityRecordsRequest{RecordCode: &codeAVo},
				expectedCount: 2,
			},
			{
				name:          "RecordLevelFilter",
				requestDto:    tkDto.ReadActivityRecordsRequest{RecordLevel: &infoLevel},
				expectedCount: 2,
			},
			{
				name: "AffectedResourcesFilter",
				requestDto: tkDto.ReadActivityRecordsRequest{
					AffectedResources: []tkValueObject.SystemResourceIdentifier{sri1},
				},
				expectedCount: 1,
			},
			{
				name:          "OperatorSriFilter",
				requestDto:    tkDto.ReadActivityRecordsRequest{OperatorSri: &operatorSri1},
				expectedCount: 1,
			},
			{
				name: "OperatorIpAddressFilter",
				requestDto: tkDto.ReadActivityRecordsRequest{
					OperatorIpAddress: &ip2,
				},
				expectedCount: 1,
			},
			{
				name: "CreatedBeforeAtFuture",
				requestDto: tkDto.ReadActivityRecordsRequest{
					CreatedBeforeAt: &futureTimeVo,
				},
				expectedCount: 3,
			},
			{
				name: "CreatedBeforeAtPast",
				requestDto: tkDto.ReadActivityRecordsRequest{
					CreatedBeforeAt: &pastTimeVo,
				},
				expectedCount: 0,
			},
			{
				name: "CreatedAfterAtPast",
				requestDto: tkDto.ReadActivityRecordsRequest{
					CreatedAfterAt: &pastTimeVo,
				},
				expectedCount: 3,
			},
			{
				name: "CreatedAfterAtFuture",
				requestDto: tkDto.ReadActivityRecordsRequest{
					CreatedAfterAt: &futureTimeVo,
				},
				expectedCount: 0,
			},
			{
				name: "CombinedFilters",
				requestDto: tkDto.ReadActivityRecordsRequest{
					RecordCode:  &codeAVo,
					RecordLevel: &errorLevel,
				},
				expectedCount: 0,
			},
			{
				name: "PaginationIgnored",
				requestDto: tkDto.ReadActivityRecordsRequest{
					Pagination: tkDto.Pagination{PageNumber: 1, ItemsPerPage: 1},
				},
				expectedCount: 3,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				count, err := queryRepo.Count(testCase.requestDto)
				if err != nil {
					t.Errorf("CountWith%sFilterFailed: %v", testCase.name, err)
				}
				if count != testCase.expectedCount {
					t.Errorf(
						"UnexpectedCountWith%sFilter: expected %d, got %d",
						testCase.name, testCase.expectedCount, count,
					)
				}
			})
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
