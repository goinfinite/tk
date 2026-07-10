package tkUseCase

import (
	"errors"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type mockQueryRepoBanDecision struct {
	hitsToReturn uint64
	repoError    error
}

func (mock *mockQueryRepoBanDecision) ReadBanDecision(
	requestDto tkDto.ReadHoneypotBanDecisionRequest,
) (tkDto.ReadHoneypotBanDecisionResponse, error) {
	if mock.repoError != nil {
		return tkDto.ReadHoneypotBanDecisionResponse{}, mock.repoError
	}
	return tkDto.ReadHoneypotBanDecisionResponse{
		HitCount: mock.hitsToReturn,
	}, nil
}

func (mock *mockQueryRepoBanDecision) ReadStatsReport(
	_ tkDto.ReadHoneypotStatsReportRequest,
) (tkDto.ReadHoneypotStatsReportResponse, error) {
	return tkDto.ReadHoneypotStatsReportResponse{}, nil
}

func TestReadHoneypotBanDecision(t *testing.T) {
	ipAddress, _ := tkValueObject.NewIpAddress("192.168.1.1")
	requestDto := tkDto.ReadHoneypotBanDecisionRequest{
		RequesterIpAddress: ipAddress,
	}

	testCaseStructs := []struct {
		testName            string
		hitCount            uint64
		repoError           error
		aggressivenessMode  tkValueObject.HoneypotAggressivenessMode
		expectedIsBanned    bool
		expectedAction      tkValueObject.HoneypotSuggestedAction
		expectError         bool
	}{
		{
			testName:           "BalancedMode_HitCountAboveThreshold_IsBanned",
			hitCount:           5,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			expectedIsBanned:   true,
			expectedAction:     tkValueObject.HoneypotSuggestedActionBan,
			expectError:        false,
		},
		{
			testName:           "BalancedMode_HitCountBelowThreshold_NotBanned",
			hitCount:           1,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			expectedIsBanned:   false,
			expectedAction:     tkValueObject.HoneypotSuggestedActionServeMixed,
			expectError:        false,
		},
		{
			testName:           "ImmediateMode_AnyHitCount_IsBanned",
			hitCount:           1,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeImmediate,
			expectedIsBanned:   true,
			expectedAction:     tkValueObject.HoneypotSuggestedActionBan,
			expectError:        false,
		},
		{
			testName:           "ObserveMode_AnyHitCount_NotBanned",
			hitCount:           100,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeObserve,
			expectedIsBanned:   false,
			expectedAction:     tkValueObject.HoneypotSuggestedActionPassthrough,
			expectError:        false,
		},
		{
			testName:           "TolerantMode_HitCountBelowThreshold_NotBanned",
			hitCount:           5,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeTolerant,
			expectedIsBanned:   false,
			expectedAction:     tkValueObject.HoneypotSuggestedActionServeStream,
			expectError:        false,
		},
		{
			testName:           "TolerantMode_HitCountAtThreshold_IsBanned",
			hitCount:           10,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeTolerant,
			expectedIsBanned:   true,
			expectedAction:     tkValueObject.HoneypotSuggestedActionBan,
			expectError:        false,
		},
		{
			testName:           "RepoError_PropagatesError",
			hitCount:           0,
			repoError:          errors.New("DatabaseConnectionFail"),
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeBalanced,
			expectedIsBanned:   false,
			expectedAction:     tkValueObject.HoneypotSuggestedAction(""),
			expectError:        true,
		},
		{
			testName:           "ZeroHits_NotBanned",
			hitCount:           0,
			repoError:          nil,
			aggressivenessMode: tkValueObject.HoneypotAggressivenessModeImmediate,
			expectedIsBanned:   false,
			expectedAction:     tkValueObject.HoneypotSuggestedActionServePayload,
			expectError:        false,
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			mockRepo := &mockQueryRepoBanDecision{
				hitsToReturn: testCase.hitCount,
				repoError:    testCase.repoError,
			}
			settings := tkDto.HoneypotSettings{
				AggressivenessMode: testCase.aggressivenessMode,
			}

			response, err := ReadHoneypotBanDecision(
				mockRepo, requestDto, settings,
			)

			if testCase.expectError && err == nil {
				t.Fatalf("MissingExpectedError")
			}
			if !testCase.expectError && err != nil {
				t.Fatalf("UnexpectedError: '%s'", err.Error())
			}
			if !testCase.expectError {
				if response.IsBanned != testCase.expectedIsBanned {
					t.Errorf(
						"UnexpectedIsBanned: '%v' vs '%v'",
						response.IsBanned,
						testCase.expectedIsBanned,
					)
				}
				if response.SuggestedAction != testCase.expectedAction {
					t.Errorf(
						"UnexpectedSuggestedAction: '%v' vs '%v'",
						response.SuggestedAction,
						testCase.expectedAction,
					)
				}
			}
		})
	}
}
