package tkUseCase

import (
	"errors"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type mockCmdRepoHitCreate struct {
	createError error
}

func (mock *mockCmdRepoHitCreate) Create(
	_ tkDto.CreateHoneypotHit,
) error {
	return mock.createError
}

func (mock *mockCmdRepoHitCreate) DeleteExpired(
	_ tkValueObject.HoneypotBanDuration,
) error {
	return nil
}

func (mock *mockCmdRepoHitCreate) EnforceMaxEntries(
	_ tkValueObject.HoneypotMaxEntries,
) error {
	return nil
}

func TestCreateHoneypotHit(t *testing.T) {
	ipAddress, _ := tkValueObject.NewIpAddress("10.0.0.1")
	urlPath, _ := tkValueObject.NewUrlPath("/admin/config")
	hitClass, _ := tkValueObject.NewHoneypotPathClass("adminPanel")

	testCaseStructs := []struct {
		testName    string
		createDto   tkDto.CreateHoneypotHit
		repoError   error
	}{
		{
			testName: "ValidHit_NoError",
			createDto: tkDto.CreateHoneypotHit{
				RequesterIpAddress: ipAddress,
				HoneypotPath:       urlPath,
				HitClass:           hitClass,
				HitCount:           1,
			},
			repoError: nil,
		},
		{
			testName: "RepoError_ErrorLoggedNotReturned",
			createDto: tkDto.CreateHoneypotHit{
				RequesterIpAddress: ipAddress,
				HoneypotPath:       urlPath,
				HitClass:           hitClass,
				HitCount:           2,
			},
			repoError: errors.New("DatabaseWriteFail"),
		},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			mockRepo := &mockCmdRepoHitCreate{
				createError: testCase.repoError,
			}

			CreateHoneypotHit(mockRepo, testCase.createDto)
		})
	}
}
