package tkInfraActivityRecord

import (
	"errors"
	"strings"
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkUseCase "github.com/goinfinite/tk/src/domain/useCase"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestActivityRecordQueryRepoRead(test *testing.T) {
	test.Run("EmptyDatabase", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)
		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			subtest.Errorf("ReadFailedOnEmptyDatabase: %v", err)
		}
		if len(responseDto.ActivityRecords) != 0 {
			subtest.Errorf("ExpectedEmptyResponse: got %d records", len(responseDto.ActivityRecords))
		}
	})

	testName := "ReadWithUnknownSortFieldReturnsPaginationError"
	test.Run(testName, func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		sortBy := tkValueObject.PaginationSortBy("notAField")
		_, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.Pagination{
				PageNumber:   0,
				ItemsPerPage: 10,
				SortBy:       &sortBy,
			},
		})
		if err == nil {
			subtest.Fatalf("MissingExpectedError: unknown sort field")
		}
		if !strings.Contains(err.Error(), "PaginationQueryBuilderError") {
			subtest.Errorf("UnexpectedError: %v", err)
		}
	})

	test.Run("ReadWithRecordIdFilter", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("TEST_CODE")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/resource1")
		if err != nil {
			subtest.Fatalf("CreateResourceVoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri},
		}

		testRecord, err := createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		recordIdVo, err := tkValueObject.NewActivityRecordId(testRecord.RecordId.Uint64())
		if err != nil {
			subtest.Fatalf("CreateRecordIdVoFailed: %v", err)
		}
		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
			RecordId:   &recordIdVo,
		})
		if err != nil {
			subtest.Errorf("ReadWithRecordIdFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordId.Uint64() != testRecord.RecordId.Uint64() {
			subtest.Errorf(
				"RecordIdMismatch: expected %d, got %d",
				testRecord.RecordId.Uint64(), responseDto.ActivityRecords[0].RecordId.Uint64(),
			)
		}
	})

	test.Run("ReadWithRecordLevelFilter", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode1Vo, err := tkValueObject.NewActivityRecordCode("CODE1")
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
			subtest.Fatalf("CreateTestActivityRecord1Failed: %v", err)
		}

		recordCode2Vo, err := tkValueObject.NewActivityRecordCode("CODE2")
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
			subtest.Fatalf("CreateTestActivityRecord2Failed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:  tkDto.PaginationUnpaginated,
			RecordLevel: &tkValueObject.ActivityRecordLevelInfo,
		})
		if err != nil {
			subtest.Errorf("ReadWithRecordLevelFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordLevel.String() != "INFO" {
			subtest.Errorf(
				"RecordLevelMismatch: expected INFO, got %s",
				responseDto.ActivityRecords[0].RecordLevel.String(),
			)
		}
	})

	test.Run("ReadWithAffectedResourcesFilter", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode3Vo, err := tkValueObject.NewActivityRecordCode("CODE1")
		if err != nil {
			subtest.Fatalf("CreateRecordCode3VoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res1")
		if err != nil {
			subtest.Fatalf("CreateRes1VoFailed: %v", err)
		}

		createDto3 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode3Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1},
		}

		_, err = createTestActivityRecord(dbSvc, createDto3)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecord3Failed: %v", err)
		}

		recordCode4Vo, err := tkValueObject.NewActivityRecordCode("CODE2")
		if err != nil {
			subtest.Fatalf("CreateRecordCode4VoFailed: %v", err)
		}

		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/res2")
		if err != nil {
			subtest.Fatalf("CreateRes2VoFailed: %v", err)
		}

		createDto4 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode4Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri2},
		}

		_, err = createTestActivityRecord(dbSvc, createDto4)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecord4Failed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1},
		})
		if err != nil {
			subtest.Errorf("ReadWithAffectedResourcesFilterFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}
		if responseDto.ActivityRecords[0].RecordCode.String() != "CODE1" {
			subtest.Errorf(
				"RecordCodeMismatch: expected CODE1, got %s",
				responseDto.ActivityRecords[0].RecordCode.String(),
			)
		}
	})

	test.Run("ReadWithMultipleMatchingAffectedResources", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("MULTI_RES_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCodeVoFailed: %v", err)
		}

		testSri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/multi1")
		if err != nil {
			subtest.Fatalf("CreateRes1VoFailed: %v", err)
		}
		testSri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/multi2")
		if err != nil {
			subtest.Fatalf("CreateRes2VoFailed: %v", err)
		}

		createDto := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCodeVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		}

		_, err = createTestActivityRecord(dbSvc, createDto)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecordFailed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination:        tkDto.PaginationUnpaginated,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{testSri1, testSri2},
		})
		if err != nil {
			subtest.Errorf("ReadWithMultipleMatchingAffectedResourcesFailed: %v", err)
		}

		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecord: got %d", len(responseDto.ActivityRecords))
		}

		if responseDto.Pagination.ItemsTotal == nil {
			subtest.Fatalf("Expected ItemsTotal to be non-nil")
		}
		if *responseDto.Pagination.ItemsTotal != 1 {
			subtest.Errorf("ExpectedTotalItemsToBeOne: got %d", *responseDto.Pagination.ItemsTotal)
		}
	})

	test.Run("ReadWithTimeFilters", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCodeVo, err := tkValueObject.NewActivityRecordCode("TIME_TEST")
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
			subtest.Run(testCase.name, func(subtest *testing.T) {
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
					subtest.Errorf("ReadWith%sFilterFailed: %v", testCase.name, err)
				}
				if recordEntity.RecordId.Uint64() != testRecord.RecordId.Uint64() {
					subtest.Errorf("TestRecordNotFoundIn%sFilter", testCase.name)
				}
			})
		}
	})

	test.Run("ReadWithPagination", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		recordCode6Vo, err := tkValueObject.NewActivityRecordCode("PAGE1")
		if err != nil {
			subtest.Fatalf("CreateRecordCode6VoFailed: %v", err)
		}

		createDto6 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode6Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto6)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecord6Failed: %v", err)
		}

		recordCode7Vo, err := tkValueObject.NewActivityRecordCode("PAGE2")
		if err != nil {
			subtest.Fatalf("CreateRecordCode7VoFailed: %v", err)
		}

		createDto7 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode7Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		_, err = createTestActivityRecord(dbSvc, createDto7)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecord7Failed: %v", err)
		}

		responseDto, err := queryRepo.Read(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.Pagination{PageNumber: 0, ItemsPerPage: 1},
		})
		if err != nil {
			subtest.Errorf("ReadWithPaginationFailed: %v", err)
		}
		if len(responseDto.ActivityRecords) != 1 {
			subtest.Errorf("ExpectedOneRecordWithPagination: got %d", len(responseDto.ActivityRecords))
		}
	})
}

func TestActivityRecordQueryRepoCount(test *testing.T) {
	test.Run("EmptyDatabase", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		count, err := queryRepo.Count(tkDto.ReadActivityRecordsRequest{})
		if err != nil {
			subtest.Errorf("CountFailedOnEmptyDatabase: %v", err)
		}
		if count != 0 {
			subtest.Errorf("ExpectedZeroCount: got %d", count)
		}
	})

	test.Run("CountWithFilters", func(subtest *testing.T) {
		dbSvc := SetupTestTrailDatabaseService(subtest)
		queryRepo := NewActivityRecordQueryRepo(dbSvc)

		codeAVo, err := tkValueObject.NewActivityRecordCode("COUNT_CODE_A")
		if err != nil {
			subtest.Fatalf("CreateCodeAVoFailed: %v", err)
		}
		codeBVo, err := tkValueObject.NewActivityRecordCode("COUNT_CODE_B")
		if err != nil {
			subtest.Fatalf("CreateCodeBVoFailed: %v", err)
		}

		sri1, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/count1")
		if err != nil {
			subtest.Fatalf("CreateRes1VoFailed: %v", err)
		}
		sri2, err := tkValueObject.NewSystemResourceIdentifier("sri://0:test/count2")
		if err != nil {
			subtest.Fatalf("CreateRes2VoFailed: %v", err)
		}

		accountId1Vo, err := tkValueObject.NewAccountId(1)
		if err != nil {
			subtest.Fatalf("CreateAccountId1VoFailed: %v", err)
		}
		operatorSri1 := tkValueObject.NewSriAccount(accountId1Vo)
		accountId2Vo, err := tkValueObject.NewAccountId(2)
		if err != nil {
			subtest.Fatalf("CreateAccountId2VoFailed: %v", err)
		}
		operatorSri2 := tkValueObject.NewSriAccount(accountId2Vo)

		ip1, err := tkValueObject.NewIpAddress("192.168.1.1")
		if err != nil {
			subtest.Fatalf("CreateIp1VoFailed: %v", err)
		}
		ip2, err := tkValueObject.NewIpAddress("192.168.1.2")
		if err != nil {
			subtest.Fatalf("CreateIp2VoFailed: %v", err)
		}

		_, err = createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        codeAVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{sri1},
			OperatorSri:       &operatorSri1,
			OperatorIpAddress: &ip1,
		})
		if err != nil {
			subtest.Fatalf("CreateRecordAFailed: %v", err)
		}

		recordB, err := createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelError,
			RecordCode:        codeBVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{sri2},
			OperatorSri:       &operatorSri2,
			OperatorIpAddress: &ip2,
		})
		if err != nil {
			subtest.Fatalf("CreateRecordBFailed: %v", err)
		}

		_, err = createTestActivityRecord(dbSvc, tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        codeAVo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		})
		if err != nil {
			subtest.Fatalf("CreateRecordCFailed: %v", err)
		}

		recordIdVo, err := tkValueObject.NewActivityRecordId(recordB.RecordId.Uint64())
		if err != nil {
			subtest.Fatalf("CreateRecordIdVoFailed: %v", err)
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
			subtest.Run(testCase.name, func(subtest *testing.T) {
				count, err := queryRepo.Count(testCase.requestDto)
				if err != nil {
					subtest.Errorf("CountWith%sFilterFailed: %v", testCase.name, err)
				}
				if count != testCase.expectedCount {
					subtest.Errorf(
						"UnexpectedCountWith%sFilter: expected %d, got %d",
						testCase.name, testCase.expectedCount, count,
					)
				}
			})
		}
	})
}

func TestActivityRecordQueryRepoReadFirst(test *testing.T) {
	dbSvc := SetupTestTrailDatabaseService(test)
	queryRepo := NewActivityRecordQueryRepo(dbSvc)

	test.Run("ReadFirstFromEmptyDatabase", func(subtest *testing.T) {
		_, err := queryRepo.ReadFirst(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err == nil {
			subtest.Errorf("ReadFirstSucceededWhenShouldFailOnEmptyDatabase")
		}
		expectedError := "ActivityRecordNotFound"
		if err.Error() != expectedError {
			subtest.Errorf("UnexpectedErrorMessage: expected '%s', got '%s'", expectedError, err.Error())
		}
		if !errors.Is(err, tkUseCase.ErrActivityRecordNotFound) {
			subtest.Errorf("ExpectedActivityRecordNotFoundSentinel: %v", err)
		}
	})

	test.Run("ReadFirstSuccess", func(subtest *testing.T) {
		recordCode8Vo, err := tkValueObject.NewActivityRecordCode("FIRST_TEST")
		if err != nil {
			subtest.Fatalf("CreateRecordCode8VoFailed: %v", err)
		}

		createDto8 := tkDto.CreateActivityRecord{
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo,
			RecordCode:        recordCode8Vo,
			AffectedResources: []tkValueObject.SystemResourceIdentifier{},
		}

		createdRecordEntity, err := createTestActivityRecord(dbSvc, createDto8)
		if err != nil {
			subtest.Fatalf("CreateTestActivityRecord8Failed: %v", err)
		}

		readRecordEntity, err := queryRepo.ReadFirst(tkDto.ReadActivityRecordsRequest{
			Pagination: tkDto.PaginationUnpaginated,
		})
		if err != nil {
			subtest.Errorf("ReadFirstFailed: %v", err)
		}
		if readRecordEntity.RecordId.Uint64() != createdRecordEntity.RecordId.Uint64() {
			subtest.Errorf(
				"ReadFirstReturnedWrongRecord: expected ID %d, got %d",
				createdRecordEntity.RecordId.Uint64(), readRecordEntity.RecordId.Uint64(),
			)
		}
	})
}
