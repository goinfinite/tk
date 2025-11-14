package tkInfraDbModel

import (
	"strings"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewActivityRecord(t *testing.T) {
	t.Run("RecordIdZero", func(t *testing.T) {
		testCaseStructs := []struct {
			recordId          uint64
			recordLevel       string
			recordCode        string
			affectedResources []ActivityRecordAffectedResource
			recordDetails     *string
			operatorSri       *string
			operatorIpAddress *string
			expectedId        uint64
		}{
			{
				recordId:          0,
				recordLevel:       tkValueObject.ActivityRecordLevelInfo.String(),
				recordCode:        "TEST_CODE",
				affectedResources: []ActivityRecordAffectedResource{},
				expectedId:        0,
			},
			{
				recordId:    0,
				recordLevel: tkValueObject.ActivityRecordLevelError.String(),
				recordCode:  "ANOTHER_CODE",
				affectedResources: []ActivityRecordAffectedResource{
					{SystemResourceIdentifier: "test-resource"},
				},
				recordDetails:     stringPtr("test details"),
				operatorSri:       stringPtr("sri://1:account/123"),
				operatorIpAddress: stringPtr("192.168.1.1"),
				expectedId:        0,
			},
		}

		for _, testCase := range testCaseStructs {
			model := NewActivityRecord(
				testCase.recordId, testCase.recordLevel, testCase.recordCode,
				testCase.affectedResources, testCase.recordDetails,
				testCase.operatorSri, testCase.operatorIpAddress,
			)

			if model.ID != testCase.expectedId {
				t.Errorf("IdNotSetCorrectly: expected %d, got %d", testCase.expectedId, model.ID)
			}
			if model.RecordLevel != testCase.recordLevel {
				t.Errorf("RecordLevelNotSetCorrectly: expected %s, got %s", testCase.recordLevel, model.RecordLevel)
			}
			if model.RecordCode != testCase.recordCode {
				t.Errorf("RecordCodeNotSetCorrectly: expected %s, got %s", testCase.recordCode, model.RecordCode)
			}
			if len(model.AffectedResources) != len(testCase.affectedResources) {
				t.Errorf("AffectedResourcesNotSetCorrectly: expected %d, got %d", len(testCase.affectedResources), len(model.AffectedResources))
			}
			assertStringPtrEqual(t, model.RecordDetails, testCase.recordDetails)
			assertStringPtrEqual(t, model.OperatorSri, testCase.operatorSri)
			assertStringPtrEqual(t, model.OperatorIpAddress, testCase.operatorIpAddress)
		}
	})

	t.Run("RecordIdNonZero", func(t *testing.T) {
		testCaseStructs := []struct {
			recordId          uint64
			recordLevel       string
			recordCode        string
			affectedResources []ActivityRecordAffectedResource
			recordDetails     *string
			operatorSri       *string
			operatorIpAddress *string
			expectedId        uint64
		}{
			{
				recordId:    42,
				recordLevel: tkValueObject.ActivityRecordLevelWarning.String(),
				recordCode:  "UPDATE_CODE",
				affectedResources: []ActivityRecordAffectedResource{
					{SystemResourceIdentifier: "resource-1"},
					{SystemResourceIdentifier: "resource-2"},
				},
				recordDetails:     stringPtr("update details"),
				operatorSri:       stringPtr("sri://1:account/999"),
				operatorIpAddress: stringPtr("10.0.0.1"),
				expectedId:        42,
			},
		}

		for _, testCase := range testCaseStructs {
			model := NewActivityRecord(
				testCase.recordId, testCase.recordLevel, testCase.recordCode,
				testCase.affectedResources, testCase.recordDetails,
				testCase.operatorSri, testCase.operatorIpAddress,
			)

			if model.ID != testCase.expectedId {
				t.Errorf("IdNotSetCorrectly: expected %d, got %d", testCase.expectedId, model.ID)
			}
			if model.RecordLevel != testCase.recordLevel {
				t.Errorf("RecordLevelNotSetCorrectly: expected %s, got %s", testCase.recordLevel, model.RecordLevel)
			}
			if model.RecordCode != testCase.recordCode {
				t.Errorf("RecordCodeNotSetCorrectly: expected %s, got %s", testCase.recordCode, model.RecordCode)
			}
			if len(model.AffectedResources) != len(testCase.affectedResources) {
				t.Errorf("AffectedResourcesNotSetCorrectly: expected %d, got %d", len(testCase.affectedResources), len(model.AffectedResources))
			}
			assertStringPtrEqual(t, model.RecordDetails, testCase.recordDetails)
			assertStringPtrEqual(t, model.OperatorSri, testCase.operatorSri)
			assertStringPtrEqual(t, model.OperatorIpAddress, testCase.operatorIpAddress)
		}
	})
}

func TestToEntity(t *testing.T) {
	t.Run("ValidData", func(t *testing.T) {
		model := ActivityRecord{
			ID:          42,
			RecordLevel: tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:  "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{
				{SystemResourceIdentifier: "sri://1:account/120"},
				{SystemResourceIdentifier: "sri://10:virtualHost/local.os"},
			},
			RecordDetails:     stringPtr("UserLoggedSuccessfully"),
			OperatorSri:       stringPtr("sri://1:account/123"),
			OperatorIpAddress: stringPtr("192.168.1.100"),
			CreatedAt:         time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		entity, err := model.ToEntity()
		if err != nil {
			t.Errorf("ToEntityFailed: %v", err)
			return
		}

		if entity.RecordId.Uint64() != 42 {
			t.Errorf("RecordIdMismatch: expected 42, got %d", entity.RecordId.Uint64())
		}
		if entity.RecordLevel.String() != "INFO" {
			t.Errorf("RecordLevelMismatch: expected INFO, got %s", entity.RecordLevel.String())
		}
		if entity.RecordCode.String() != "LoginSuccessful" {
			t.Errorf("RecordCodeMismatch: expected LoginSuccessful, got %s", entity.RecordCode.String())
		}
		if len(entity.AffectedResources) != 2 {
			t.Errorf("AffectedResourcesLengthMismatch: expected 2, got %d", len(entity.AffectedResources))
		}
		if entity.AffectedResources[0].String() != "sri://1:account/120" {
			t.Errorf("FirstAffectedResourceMismatch: expected sri://1:account/120, got %s", entity.AffectedResources[0].String())
		}
		if entity.AffectedResources[1].String() != "sri://10:virtualHost/local.os" {
			t.Errorf("SecondAffectedResourceMismatch: expected sri://10:virtualHost/local.os, got %s", entity.AffectedResources[1].String())
		}
		if entity.RecordDetails != "UserLoggedSuccessfully" {
			t.Errorf("RecordDetailsMismatch: expected 'UserLoggedSuccessfully', got %v", entity.RecordDetails)
		}
		if entity.OperatorSri == nil || entity.OperatorSri.String() != "sri://1:account/123" {
			t.Errorf("OperatorSriMismatch: expected sri://1:account/123, got %v", entity.OperatorSri)
		}
		if entity.OperatorIpAddress == nil || entity.OperatorIpAddress.String() != "192.168.1.100" {
			t.Errorf("OperatorIpAddressMismatch: expected 192.168.1.100, got %v", entity.OperatorIpAddress)
		}
		if entity.CreatedAt.Int64() != 1673778600 {
			t.Errorf("CreatedAtMismatch: expected 1673778600, got %d", entity.CreatedAt.Int64())
		}
	})

	t.Run("ValidDataWithNilPointers", func(t *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       tkValueObject.ActivityRecordLevelError.String(),
			RecordCode:        "LoginFailed",
			AffectedResources: []ActivityRecordAffectedResource{},
			CreatedAt:         time.Date(2023, 2, 20, 15, 45, 30, 0, time.UTC),
		}

		entity, err := model.ToEntity()
		if err != nil {
			t.Errorf("ToEntityFailed: %v", err)
			return
		}

		if entity.OperatorSri != nil {
			t.Errorf("OperatorSriShouldBeNil")
		}
		if entity.OperatorIpAddress != nil {
			t.Errorf("OperatorIpAddressShouldBeNil")
		}
		if entity.RecordDetails != nil {
			t.Errorf("RecordDetailsShouldBeNil: got %v", entity.RecordDetails)
		}
	})

	t.Run("InvalidRecordLevel", func(t *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       "InvalidRecordLevel",
			RecordCode:        "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{},
			CreatedAt:         time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			t.Errorf("ToEntityShouldFailWithInvalidRecordLevel")
		}
		if !strings.Contains(err.Error(), "ActivityRecordLevel") {
			t.Errorf("ErrorShouldContainActivityRecordLevel: got '%s'", err.Error())
		}
	})

	t.Run("InvalidRecordCode", func(t *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:        "a",
			AffectedResources: []ActivityRecordAffectedResource{},
			CreatedAt:         time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			t.Errorf("ToEntityShouldFailWithInvalidRecordCode")
		}
		if !strings.Contains(err.Error(), "InvalidActivityRecordCode") {
			t.Errorf("ErrorShouldContainInvalidActivityRecordCode: got '%s'", err.Error())
		}
	})

	t.Run("InvalidSystemResourceIdentifier", func(t *testing.T) {
		model := ActivityRecord{
			ID:          1,
			RecordLevel: tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:  "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{
				{SystemResourceIdentifier: "invalid-sri"},
			},
			OperatorSri: nil,
			CreatedAt:   time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			t.Errorf("ToEntityShouldFailWithInvalidSystemResourceIdentifier")
		}
		if !strings.Contains(err.Error(), "SystemResourceIdentifier") {
			t.Errorf("ErrorShouldContainSystemResourceIdentifier: got '%s'", err.Error())
		}
	})

	t.Run("InvalidOperatorIpAddress", func(t *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:        "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{},
			OperatorIpAddress: stringPtr("invalid-ip"),
			CreatedAt:         time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			t.Errorf("ToEntityShouldFailWithInvalidOperatorIpAddress")
		}
		if !strings.Contains(err.Error(), "InvalidIpAddress") {
			t.Errorf("ErrorShouldContainInvalidIpAddress: got '%s'", err.Error())
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

func assertStringPtrEqual(t *testing.T, actual, expected *string) {
	t.Helper()
	if (actual == nil && expected != nil) || (actual != nil && expected == nil) {
		t.Errorf("StringPtrMismatch: actual=%v, expected=%v", actual, expected)
		return
	}
	if actual != nil && expected != nil && *actual != *expected {
		t.Errorf("StringPtrValueMismatch: actual=%s, expected=%s", *actual, *expected)
	}
}
