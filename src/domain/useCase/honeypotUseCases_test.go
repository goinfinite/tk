package tkUseCase

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type mockHoneypotCmdRepo struct {
	incrementHitFunc func(
		ip tkValueObject.IpAddress, path string,
	)
	cleanExpiredFunc func(banDuration time.Duration)
	enforceMaxFunc   func(maxEntries int)
}

func (mock mockHoneypotCmdRepo) IncrementHit(
	ip tkValueObject.IpAddress, path string,
) {
	if mock.incrementHitFunc != nil {
		mock.incrementHitFunc(ip, path)
	}
}

func (mock mockHoneypotCmdRepo) CleanExpiredEntries(
	banDuration time.Duration,
) {
	if mock.cleanExpiredFunc != nil {
		mock.cleanExpiredFunc(banDuration)
	}
}

func (mock mockHoneypotCmdRepo) EnforceMaxEntries(
	maxEntries int,
) {
	if mock.enforceMaxFunc != nil {
		mock.enforceMaxFunc(maxEntries)
	}
}

type mockHoneypotQueryRepo struct {
	readHitRecordFunc func(
		ip tkValueObject.IpAddress,
	) (tkDto.HoneypotHitData, error)
	countFunc func() int64
	readReportFunc func(
		banDuration time.Duration,
		mode tkValueObject.HoneypotAggressivenessMode,
	) (tkDto.HoneypotStatsReport, error)
}

func (mock mockHoneypotQueryRepo) ReadHitRecord(
	ip tkValueObject.IpAddress,
) (tkDto.HoneypotHitData, error) {
	if mock.readHitRecordFunc != nil {
		return mock.readHitRecordFunc(ip)
	}
	return tkDto.HoneypotHitData{}, nil
}

func (mock mockHoneypotQueryRepo) Count() int64 {
	if mock.countFunc != nil {
		return mock.countFunc()
	}
	return 0
}

func (mock mockHoneypotQueryRepo) ReadReport(
	banDuration time.Duration,
	mode tkValueObject.HoneypotAggressivenessMode,
) (tkDto.HoneypotStatsReport, error) {
	if mock.readReportFunc != nil {
		return mock.readReportFunc(banDuration, mode)
	}
	return tkDto.HoneypotStatsReport{}, nil
}

type mockActivityRecordCmdRepoForHoneypot struct {
	createFunc func(
		dto tkDto.CreateActivityRecord,
	) error
}

func (mock mockActivityRecordCmdRepoForHoneypot) Create(
	dto tkDto.CreateActivityRecord,
) error {
	if mock.createFunc != nil {
		return mock.createFunc(dto)
	}
	return nil
}

func (mock mockActivityRecordCmdRepoForHoneypot) Delete(
	dto tkDto.DeleteActivityRecord,
) error {
	return nil
}

func newTestIpAddress(value string) tkValueObject.IpAddress {
	ipAddr, _ := tkValueObject.NewIpAddress(value)
	return ipAddr
}

func mustNewHoneypotBanDuration(
	rawValue any,
) tkValueObject.HoneypotBanDuration {
	banDuration, _ := tkValueObject.NewHoneypotBanDuration(rawValue)
	return banDuration
}

func mustNewHoneypotMaxEntries(
	rawValue any,
) tkValueObject.HoneypotMaxEntries {
	maxEntries, _ := tkValueObject.NewHoneypotMaxEntries(rawValue)
	return maxEntries
}

func TestReadHoneypotBanDecision(t *testing.T) {
	testCaseStructs := []struct {
		name        string
		queryRepo   tkRepository.HoneypotQueryRepo
		requesterIp tkValueObject.IpAddress
		banDuration tkValueObject.HoneypotBanDuration
		mode        tkValueObject.HoneypotAggressivenessMode
		expectedTier int
		expectError bool
	}{
		{
			"BalancedCountOneReturnsTierOne",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      1,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			1, false,
		},
		{
			"BalancedCountTwoReturnsTierTwo",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      2,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			2, false,
		},
		{
			"BalancedCountThreeReturnsTierThree",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      3,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			3, false,
		},
		{
			"ImmediateCountOneReturnsTierThree",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      1,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeImmediate,
			3, false,
		},
		{
			"TolerantCountTwoReturnsTierOne",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      2,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeTolerant,
			1, false,
		},
		{
			"ObserveModeAlwaysReturnsTierOne",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      50,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeObserve,
			1, false,
		},
		{
			"ExpiredHitReturnsTierZero",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count: 5,
						FirstHitAt: time.Now().UTC().Add(
							-25 * time.Hour,
						).Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			0, false,
		},
		{
			"ZeroCountReturnsTierZero",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      0,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			0, false,
		},
		{
			"ImmediateCountZeroReturnsTierZero",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      0,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeImmediate,
			0, false,
		},
		{
			"TolerantCountFiveReturnsTierTwo",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      5,
						FirstHitAt: time.Now().UTC().Format(time.RFC3339),
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeTolerant,
			2, false,
		},
		{
			"NilQueryRepoReturnsSentinelError",
			nil,
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			0, true,
		},
		{
			"ReadErrorReturnsError",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{},
						errors.New("RecordNotFound")
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			0, true,
		},
		{
			"MalformedTimestampReturnsError",
			mockHoneypotQueryRepo{
				readHitRecordFunc: func(
					ip tkValueObject.IpAddress,
				) (tkDto.HoneypotHitData, error) {
					return tkDto.HoneypotHitData{
						Count:      1,
						FirstHitAt: "not-a-timestamp",
					}, nil
				},
			},
			newTestIpAddress("1.2.3.4"),
			mustNewHoneypotBanDuration(24 * time.Hour),
			tkValueObject.HoneypotAggressivenessModeBalanced,
			0, true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			resolvedTier, resolveErr := ReadHoneypotBanDecision(
				testCase.queryRepo,
				testCase.requesterIp,
				testCase.banDuration,
				testCase.mode,
			)

			if testCase.expectError {
				if resolveErr == nil {
					t.Errorf("MissingExpectedError")
				}
				if testCase.queryRepo == nil &&
					!errors.Is(resolveErr, ErrNilHoneypotQueryRepo) {
					t.Errorf("ExpectedSentinelError: got=%v",
						resolveErr)
				}
				return
			}

			if resolveErr != nil {
				t.Fatalf("UnexpectedError: %v", resolveErr)
			}

			if resolvedTier != testCase.expectedTier {
				t.Errorf("TierMismatch: got=%d, want=%d",
					resolvedTier, testCase.expectedTier)
			}
		})
	}
}

func TestCreateHoneypotHit(t *testing.T) {
	testCaseStructs := []struct {
		name               string
		cmdRepo            tkRepository.HoneypotCmdRepo
		requesterIp        tkValueObject.IpAddress
		interceptPath      string
		maxEntries         tkValueObject.HoneypotMaxEntries
		expectIncrement    bool
	}{
		{
			"NilCmdRepoIsNoop",
			nil,
			newTestIpAddress("1.2.3.4"),
			"/.env", mustNewHoneypotMaxEntries(5000),
			false,
		},
		{
			"IncrementsHitCount",
			mockHoneypotCmdRepo{
				incrementHitFunc: func(
					ip tkValueObject.IpAddress,
					path string,
				) {},
			},
			newTestIpAddress("1.2.3.4"),
			"/.env", mustNewHoneypotMaxEntries(5000),
			true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			incrementCalled := false
			existentCmdRepo := testCase.cmdRepo
			if testCase.expectIncrement {
				existentCmdRepo = mockHoneypotCmdRepo{
					incrementHitFunc: func(
						ip tkValueObject.IpAddress,
						path string,
					) {
						incrementCalled = true
					},
				}
			}

			CreateHoneypotHit(
				existentCmdRepo,
				testCase.requesterIp,
				testCase.interceptPath,
				testCase.maxEntries,
			)

			if testCase.expectIncrement && !incrementCalled {
				t.Errorf("IncrementHitNotCalled")
			}
			if !testCase.expectIncrement && incrementCalled {
				t.Errorf("IncrementHitShouldNotBeCalled")
			}
		})
	}
}

func TestReadHoneypotStatsReport(t *testing.T) {
	existentReport := tkDto.HoneypotStatsReport{
		BannedIpCount: 5,
		TopOffenders: []tkDto.HoneypotStatsOffender{
			{
				IpAddress: "1.2.3.4",
				HitCount:  10,
				Tier:      3,
			},
		},
		TopEndpoints: []tkDto.HoneypotStatsEndpoint{
			{Path: "/.env", HitCount: 10},
		},
	}

	testCaseStructs := []struct {
		name          string
		queryRepo     tkRepository.HoneypotQueryRepo
		expectedCount int
		expectError   bool
	}{
		{
			"NilQueryRepoReturnsEmptyReport",
			nil,
			0, false,
		},
		{
			"ReturnsReportFromRepo",
			mockHoneypotQueryRepo{
				readReportFunc: func(
					banDuration time.Duration,
					mode tkValueObject.HoneypotAggressivenessMode,
				) (tkDto.HoneypotStatsReport, error) {
					return existentReport, nil
				},
			},
			5, false,
		},
		{
			"ReturnsErrorFromRepo",
			mockHoneypotQueryRepo{
				readReportFunc: func(
					banDuration time.Duration,
					mode tkValueObject.HoneypotAggressivenessMode,
				) (tkDto.HoneypotStatsReport, error) {
					return tkDto.HoneypotStatsReport{},
						errors.New("ReportReadFailed")
				},
			},
			0, true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			report, reportErr := ReadHoneypotStatsReport(
				testCase.queryRepo,
				mustNewHoneypotBanDuration(24*time.Hour),
				tkValueObject.HoneypotAggressivenessModeBalanced,
			)

			if testCase.expectError {
				if reportErr == nil {
					t.Errorf("MissingExpectedError")
				}
				return
			}

			if reportErr != nil {
				t.Fatalf("UnexpectedError: %v", reportErr)
			}

			if report.BannedIpCount != testCase.expectedCount {
				t.Errorf("BannedIpCountMismatch: got=%d, want=%d",
					report.BannedIpCount,
					testCase.expectedCount)
			}
		})
	}
}

func TestRunHoneypotMaintenance(t *testing.T) {
	testRecordCode, _ := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)
	testRecordLevel := tkValueObject.ActivityRecordLevelSecurity
	balancedMode := tkValueObject.HoneypotAggressivenessModeBalanced

	testCaseStructs := []struct {
		name               string
		cmdRepo            tkRepository.HoneypotCmdRepo
		queryRepo          tkRepository.HoneypotQueryRepo
		activityRecordRepo tkRepository.ActivityRecordCmdRepo
		expectCleanup      bool
		expectEnforce      bool
		expectStatsRecord  bool
	}{
		{
			"AllReposWithEntriesRunsFullCycle",
			mockHoneypotCmdRepo{
				cleanExpiredFunc: func(
					banDuration time.Duration,
				) {},
				enforceMaxFunc: func(maxEntries int) {},
			},
			mockHoneypotQueryRepo{
				countFunc: func() int64 { return 5 },
				readReportFunc: func(
					banDuration time.Duration,
					mode tkValueObject.HoneypotAggressivenessMode,
				) (tkDto.HoneypotStatsReport, error) {
					return tkDto.HoneypotStatsReport{
						BannedIpCount: 2,
					}, nil
				},
			},
			mockActivityRecordCmdRepoForHoneypot{
				createFunc: func(
					dto tkDto.CreateActivityRecord,
				) error {
					return nil
				},
			},
			true, true, true,
		},
		{
			"NilCmdRepoSkipsCleanup",
			nil,
			mockHoneypotQueryRepo{
				countFunc: func() int64 { return 5 },
				readReportFunc: func(
					banDuration time.Duration,
					mode tkValueObject.HoneypotAggressivenessMode,
				) (tkDto.HoneypotStatsReport, error) {
					return tkDto.HoneypotStatsReport{
						BannedIpCount: 2,
					}, nil
				},
			},
			mockActivityRecordCmdRepoForHoneypot{
				createFunc: func(
					dto tkDto.CreateActivityRecord,
				) error {
					return nil
				},
			},
			false, false, true,
		},
		{
			"NilQueryRepoSkipsStats",
			mockHoneypotCmdRepo{
				cleanExpiredFunc: func(
					banDuration time.Duration,
				) {},
				enforceMaxFunc: func(maxEntries int) {},
			},
			nil,
			mockActivityRecordCmdRepoForHoneypot{
				createFunc: func(
					dto tkDto.CreateActivityRecord,
				) error {
					return nil
				},
			},
			true, true, false,
		},
		{
			"NilActivityRecordRepoSkipsStats",
			mockHoneypotCmdRepo{
				cleanExpiredFunc: func(
					banDuration time.Duration,
				) {},
				enforceMaxFunc: func(maxEntries int) {},
			},
			mockHoneypotQueryRepo{
				countFunc: func() int64 { return 5 },
			},
			nil,
			true, true, false,
		},
		{
			"EmptyDbSkipsStats",
			mockHoneypotCmdRepo{
				cleanExpiredFunc: func(
					banDuration time.Duration,
				) {},
				enforceMaxFunc: func(maxEntries int) {},
			},
			mockHoneypotQueryRepo{
				countFunc: func() int64 { return 0 },
			},
			mockActivityRecordCmdRepoForHoneypot{
				createFunc: func(
					dto tkDto.CreateActivityRecord,
				) error {
					return nil
				},
			},
			true, true, false,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.name, func(t *testing.T) {
			cleanupCalled := false
			enforceCalled := false
			statsRecordCreated := false

			existentCmdRepo := testCase.cmdRepo
			if testCase.expectCleanup || testCase.expectEnforce {
				existentCmdRepo = mockHoneypotCmdRepo{
					cleanExpiredFunc: func(
						banDuration time.Duration,
					) {
						cleanupCalled = true
					},
					enforceMaxFunc: func(
						maxEntries int,
					) {
						enforceCalled = true
					},
				}
			}

			existentActivityRepo := testCase.activityRecordRepo
			if testCase.activityRecordRepo != nil {
				existentActivityRepo = mockActivityRecordCmdRepoForHoneypot{
					createFunc: func(
						dto tkDto.CreateActivityRecord,
					) error {
						if dto.RecordCode.String() == "HoneypotPeriodicReport" {
							statsRecordCreated = true
						}
						return nil
					},
				}
			}

			RunHoneypotMaintenance(
				existentCmdRepo,
				testCase.queryRepo,
				existentActivityRepo,
				tkDto.RunHoneypotMaintenanceRequest{
					BanDuration: mustNewHoneypotBanDuration(
						24 * time.Hour,
					),
					MaxEntries: mustNewHoneypotMaxEntries(5000),
					AggressivenessMode: balancedMode,
					StatsRecordCode:    testRecordCode,
					StatsRecordLevel:   testRecordLevel,
				},
			)

			if testCase.expectCleanup && !cleanupCalled {
				t.Errorf("CleanExpiredEntriesNotCalled")
			}
			if testCase.expectEnforce && !enforceCalled {
				t.Errorf("EnforceMaxEntriesNotCalled")
			}
			if testCase.expectStatsRecord && !statsRecordCreated {
				t.Errorf("StatsRecordNotCreated")
			}
			if !testCase.expectStatsRecord && statsRecordCreated {
				t.Errorf("StatsRecordShouldNotBeCreated")
			}
		})
	}
}

func TestRunHoneypotMaintenanceStatsReportContent(t *testing.T) {
	var capturedRecord *tkDto.CreateActivityRecord
	testRecordCode, _ := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)
	testRecordLevel := tkValueObject.ActivityRecordLevelSecurity

	queryRepo := mockHoneypotQueryRepo{
		countFunc: func() int64 { return 3 },
		readReportFunc: func(
			banDuration time.Duration,
			mode tkValueObject.HoneypotAggressivenessMode,
		) (tkDto.HoneypotStatsReport, error) {
			return tkDto.HoneypotStatsReport{
				BannedIpCount: 2,
				TopOffenders: []tkDto.HoneypotStatsOffender{
					{
						IpAddress: "1.1.1.1",
						HitCount:  5,
						Tier:      3,
					},
				},
				TopEndpoints: []tkDto.HoneypotStatsEndpoint{
					{Path: "/.env", HitCount: 5},
				},
			}, nil
		},
	}

	activityRepo := mockActivityRecordCmdRepoForHoneypot{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			capturedRecord = &dto
			return nil
		},
	}

	RunHoneypotMaintenance(
		mockHoneypotCmdRepo{},
		queryRepo, activityRepo,
		tkDto.RunHoneypotMaintenanceRequest{
			BanDuration: mustNewHoneypotBanDuration(
				24 * time.Hour,
			),
			MaxEntries: mustNewHoneypotMaxEntries(5000),
			AggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			StatsRecordCode:    testRecordCode,
			StatsRecordLevel:   testRecordLevel,
		},
	)

	if capturedRecord == nil {
		t.Fatalf("StatsRecordNotCreated")
	}

	if capturedRecord.RecordCode.String() != "HoneypotPeriodicReport" {
		t.Errorf("RecordCodeMismatch: got=%s",
			capturedRecord.RecordCode.String())
	}

	if capturedRecord.RecordLevel.String() != "SECURITY" {
		t.Errorf("RecordLevelMismatch: got=%s",
			capturedRecord.RecordLevel.String())
	}

	detailsMap, assertOk := capturedRecord.RecordDetails.(map[string]string)
	if !assertOk {
		t.Fatalf("RecordDetailsTypeMismatch")
	}

	var statsReport tkDto.HoneypotStatsReport
	json.Unmarshal(
		[]byte(detailsMap["statsReport"]),
		&statsReport,
	)

	if statsReport.BannedIpCount != 2 {
		t.Errorf("BannedIpCountMismatch: got=%d, want=2",
			statsReport.BannedIpCount)
	}
}

func TestRunHoneypotMaintenanceReportErrorLoggedNotPropagated(t *testing.T) {
	testRecordCode, _ := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)
	testRecordLevel := tkValueObject.ActivityRecordLevelSecurity

	queryRepo := mockHoneypotQueryRepo{
		countFunc: func() int64 { return 3 },
		readReportFunc: func(
			banDuration time.Duration,
			mode tkValueObject.HoneypotAggressivenessMode,
		) (tkDto.HoneypotStatsReport, error) {
			return tkDto.HoneypotStatsReport{},
				errors.New("ReportReadFailed")
		},
	}

	recordCreated := false
	activityRepo := mockActivityRecordCmdRepoForHoneypot{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			recordCreated = true
			return nil
		},
	}

	RunHoneypotMaintenance(
		mockHoneypotCmdRepo{},
		queryRepo, activityRepo,
		tkDto.RunHoneypotMaintenanceRequest{
			BanDuration: mustNewHoneypotBanDuration(
				24 * time.Hour,
			),
			MaxEntries: mustNewHoneypotMaxEntries(5000),
			AggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			StatsRecordCode:    testRecordCode,
			StatsRecordLevel:   testRecordLevel,
		},
	)

	if recordCreated {
		t.Errorf("RecordShouldNotBeCreatedOnReportError")
	}
}

func TestRunHoneypotMaintenanceCreateErrorLoggedNotPropagated(t *testing.T) {
	testRecordCode, _ := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)
	testRecordLevel := tkValueObject.ActivityRecordLevelSecurity

	queryRepo := mockHoneypotQueryRepo{
		countFunc: func() int64 { return 3 },
		readReportFunc: func(
			banDuration time.Duration,
			mode tkValueObject.HoneypotAggressivenessMode,
		) (tkDto.HoneypotStatsReport, error) {
			return tkDto.HoneypotStatsReport{
				BannedIpCount: 1,
			}, nil
		},
	}

	activityRepo := mockActivityRecordCmdRepoForHoneypot{
		createFunc: func(
			dto tkDto.CreateActivityRecord,
		) error {
			return errors.New("CreateFailed")
		},
	}

	RunHoneypotMaintenance(
		mockHoneypotCmdRepo{},
		queryRepo, activityRepo,
		tkDto.RunHoneypotMaintenanceRequest{
			BanDuration: mustNewHoneypotBanDuration(
				24 * time.Hour,
			),
			MaxEntries: mustNewHoneypotMaxEntries(5000),
			AggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			StatsRecordCode:    testRecordCode,
			StatsRecordLevel:   testRecordLevel,
		},
	)
}

func TestReadHoneypotBanDecisionTtlBoundary(t *testing.T) {
	queryRepo := mockHoneypotQueryRepo{
		readHitRecordFunc: func(
			ip tkValueObject.IpAddress,
		) (tkDto.HoneypotHitData, error) {
			return tkDto.HoneypotHitData{
				Count: 3,
				FirstHitAt: time.Now().UTC().Add(
					-23*time.Hour - 59*time.Minute,
				).Format(time.RFC3339),
			}, nil
		},
	}

	resolvedTier, resolveErr := ReadHoneypotBanDecision(
		queryRepo,
		newTestIpAddress("1.2.3.4"),
		mustNewHoneypotBanDuration(24*time.Hour),
		tkValueObject.HoneypotAggressivenessModeBalanced,
	)

	if resolveErr != nil {
		t.Fatalf("UnexpectedError: %v", resolveErr)
	}

	if resolvedTier != 3 {
		t.Errorf("TierMismatch: got=%d, want=3", resolvedTier)
	}
}

func TestCreateHoneypotHitProbabilisticEnforcement(t *testing.T) {
	enforceCallCount := 0
	cmdRepo := mockHoneypotCmdRepo{
		incrementHitFunc: func(
			ip tkValueObject.IpAddress,
			path string,
		) {},
		enforceMaxFunc: func(maxEntries int) {
			enforceCallCount++
		},
	}

	testIp := newTestIpAddress("1.2.3.4")
	for range 1000 {
		CreateHoneypotHit(
			cmdRepo, testIp, "/.env",
			mustNewHoneypotMaxEntries(5000),
		)
	}

	if enforceCallCount == 0 {
		t.Errorf("EnforcementShouldTriggerOccasionally")
	}

	if enforceCallCount > 100 {
		t.Errorf("EnforcementTriggeredTooOften: %d",
			enforceCallCount)
	}
}

func TestRunHoneypotMaintenanceAllNilRepos(t *testing.T) {
	testRecordCode, _ := tkValueObject.NewActivityRecordCode(
		"HoneypotPeriodicReport",
	)
	testRecordLevel := tkValueObject.ActivityRecordLevelSecurity

	RunHoneypotMaintenance(
		nil, nil, nil,
		tkDto.RunHoneypotMaintenanceRequest{
			BanDuration: mustNewHoneypotBanDuration(
				24 * time.Hour,
			),
			MaxEntries: mustNewHoneypotMaxEntries(5000),
			AggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			StatsRecordCode:    testRecordCode,
			StatsRecordLevel:   testRecordLevel,
		},
	)
}
