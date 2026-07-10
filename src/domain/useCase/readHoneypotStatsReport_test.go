package tkUseCase

import (
	"errors"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type mockQueryRepoStatsReport struct {
	statsToReturn tkDto.ReadHoneypotStatsReportResponse
	repoError     error
}

func (mock *mockQueryRepoStatsReport) ReadBanDecision(
	_ tkDto.ReadHoneypotBanDecisionRequest,
) (tkDto.ReadHoneypotBanDecisionResponse, error) {
	return tkDto.ReadHoneypotBanDecisionResponse{}, nil
}

func (mock *mockQueryRepoStatsReport) ReadStatsReport(
	_ tkDto.ReadHoneypotStatsReportRequest,
) (tkDto.ReadHoneypotStatsReportResponse, error) {
	if mock.repoError != nil {
		return tkDto.ReadHoneypotStatsReportResponse{}, mock.repoError
	}
	return mock.statsToReturn, nil
}

func TestReadHoneypotStatsReport(t *testing.T) {
	testCaseStructs := []struct {
		testName       string
		statsToReturn  tkDto.ReadHoneypotStatsReportResponse
		repoError      error
		expectError    bool
	}{
		{
			testName: "ValidStats_ReturnsAggregated",
			statsToReturn: tkDto.ReadHoneypotStatsReportResponse{
				TotalHits: 10,
				UniqueIps: 3,
				HitsByClass: map[tkValueObject.HoneypotPathClass]uint64{
					tkValueObject.HoneypotPathClass("adminPanel"): 6,
					tkValueObject.HoneypotPathClass("apiEndpoint"): 4,
				},
				BannedIps: 2,
			},
			repoError:   nil,
			expectError: false,
		},
		{
			testName: "EmptyStats_ReturnsZeros",
			statsToReturn: tkDto.ReadHoneypotStatsReportResponse{
				TotalHits:  0,
				UniqueIps:  0,
				HitsByClass: map[tkValueObject.HoneypotPathClass]uint64{},
				BannedIps:  0,
			},
			repoError:   nil,
			expectError: false,
		},
		{
			testName:       "RepoError_PropagatesError",
			statsToReturn:  tkDto.ReadHoneypotStatsReportResponse{},
			repoError:      errors.New("DatabaseReadFail"),
			expectError:    true,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			mockRepo := &mockQueryRepoStatsReport{
				statsToReturn: testCase.statsToReturn,
				repoError:     testCase.repoError,
			}
			requestDto := tkDto.ReadHoneypotStatsReportRequest{}

			response, err := ReadHoneypotStatsReport(mockRepo, requestDto)

			if testCase.expectError && err == nil {
				t.Fatalf("MissingExpectedError")
			}
			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if !testCase.expectError {
				if response.TotalHits != testCase.statsToReturn.TotalHits {
					t.Errorf(
						"UnexpectedTotalHits: '%v' vs '%v'",
						response.TotalHits,
						testCase.statsToReturn.TotalHits,
					)
				}
				if response.UniqueIps != testCase.statsToReturn.UniqueIps {
					t.Errorf(
						"UnexpectedUniqueIps: '%v' vs '%v'",
						response.UniqueIps,
						testCase.statsToReturn.UniqueIps,
					)
				}
				if response.BannedIps != testCase.statsToReturn.BannedIps {
					t.Errorf(
						"UnexpectedBannedIps: '%v' vs '%v'",
						response.BannedIps,
						testCase.statsToReturn.BannedIps,
					)
				}
			}
		})
	}
}
