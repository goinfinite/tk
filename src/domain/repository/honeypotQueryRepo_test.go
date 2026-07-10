package tkRepository

import (
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
)

type mockHoneypotQueryRepo struct{}

func (mock *mockHoneypotQueryRepo) ReadBanDecision(
	requestDto tkDto.ReadHoneypotBanDecisionRequest,
) (tkDto.ReadHoneypotBanDecisionResponse, error) {
	return tkDto.ReadHoneypotBanDecisionResponse{}, nil
}

func (mock *mockHoneypotQueryRepo) ReadStatsReport(
	requestDto tkDto.ReadHoneypotStatsReportRequest,
) (tkDto.ReadHoneypotStatsReportResponse, error) {
	return tkDto.ReadHoneypotStatsReportResponse{}, nil
}

func TestHoneypotQueryRepoInterface(t *testing.T) {
	var repo HoneypotQueryRepo = &mockHoneypotQueryRepo{}

	testCaseStructs := []struct {
		testName string
	}{
		{"ReadBanDecisionMethodExists"},
		{"ReadStatsReportMethodExists"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			if repo == nil {
				t.Fatalf("RepoIsNil")
			}
		})
	}
}
