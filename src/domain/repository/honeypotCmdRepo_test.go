package tkRepository

import (
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type mockHoneypotCmdRepo struct{}

func (mock *mockHoneypotCmdRepo) Create(
	createDto tkDto.CreateHoneypotHit,
) error {
	return nil
}

func (mock *mockHoneypotCmdRepo) DeleteExpired(
	banDuration tkValueObject.HoneypotBanDuration,
) error {
	return nil
}

func (mock *mockHoneypotCmdRepo) EnforceMaxEntries(
	maxEntries tkValueObject.HoneypotMaxEntries,
) error {
	return nil
}

func TestHoneypotCmdRepoInterface(t *testing.T) {
	var repo HoneypotCmdRepo = &mockHoneypotCmdRepo{}

	testCaseStructs := []struct {
		testName string
	}{
		{"CreateMethodExists"},
		{"DeleteExpiredMethodExists"},
		{"EnforceMaxEntriesMethodExists"},
	}

	for _, testCase := range testCaseStructs {
		t.Run(testCase.testName, func(t *testing.T) {
			if repo == nil {
				t.Fatalf("RepoIsNil")
			}
		})
	}
}
