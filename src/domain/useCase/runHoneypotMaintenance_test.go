package tkUseCase

import (
	"errors"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type mockCmdRepoMaintenance struct {
	deleteExpiredError    error
	enforceMaxEntriesError error
}

func (mock *mockCmdRepoMaintenance) Create(
	_ tkDto.CreateHoneypotHit,
) error {
	return nil
}

func (mock *mockCmdRepoMaintenance) DeleteExpired(
	_ tkValueObject.HoneypotBanDuration,
) error {
	return mock.deleteExpiredError
}

func (mock *mockCmdRepoMaintenance) EnforceMaxEntries(
	_ tkValueObject.HoneypotMaxEntries,
) error {
	return mock.enforceMaxEntriesError
}

func TestRunHoneypotMaintenance(t *testing.T) {
	banDuration, _ := tkValueObject.NewHoneypotBanDuration("24h")
	maxEntries, _ := tkValueObject.NewHoneypotMaxEntries(uint64(1000))
	statsInterval, _ := tkValueObject.NewHoneypotStatsInterval("1h")

	testCaseStructs := []struct {
		testName              string
		deleteExpiredError    error
		enforceMaxEntriesError error
		expectError           bool
	}{
		{
			testName:              "BothSucceed_NilError",
			deleteExpiredError:    nil,
			enforceMaxEntriesError: nil,
			expectError:           false,
		},
		{
			testName:              "DeleteExpiredFails_ReturnsError",
			deleteExpiredError:    errors.New("DatabaseDeleteFail"),
			enforceMaxEntriesError: nil,
			expectError:           true,
		},
		{
			testName:              "EnforceMaxEntriesFails_ReturnsError",
			deleteExpiredError:    nil,
			enforceMaxEntriesError: errors.New("DatabaseEnforceFail"),
			expectError:           true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			mockRepo := &mockCmdRepoMaintenance{
				deleteExpiredError:    testCase.deleteExpiredError,
				enforceMaxEntriesError: testCase.enforceMaxEntriesError,
			}
			requestDto := tkDto.RunHoneypotMaintenanceRequest{
				MaxEntries:    maxEntries,
				BanDuration:   banDuration,
				StatsInterval: statsInterval,
			}

			err := RunHoneypotMaintenance(mockRepo, requestDto)

			if testCase.expectError && err == nil {
				t.Fatalf("MissingExpectedError")
			}
			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
		})
	}
}
