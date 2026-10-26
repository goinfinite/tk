package tkInfraDbModel

import (
	"strings"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewActivityRecord(test *testing.T) {
	test.Run("RecordIdZero", func(subtest *testing.T) {
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
				recordDetails:     new("test details"),
				operatorSri:       new("sri://1:account/123"),
				operatorIpAddress: new("192.168.1.1"),
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
				subtest.Errorf("IdNotSetCorrectly: expected %d, got %d", testCase.expectedId, model.ID)
			}
			if model.RecordLevel != testCase.recordLevel {
				subtest.Errorf("RecordLevelNotSetCorrectly: expected %s, got %s", testCase.recordLevel, model.RecordLevel)
			}
			if model.RecordCode != testCase.recordCode {
				subtest.Errorf("RecordCodeNotSetCorrectly: expected %s, got %s", testCase.recordCode, model.RecordCode)
			}
			if len(model.AffectedResources) != len(testCase.affectedResources) {
				subtest.Errorf("AffectedResourcesNotSetCorrectly: expected %d, got %d", len(testCase.affectedResources), len(model.AffectedResources))
			}
			assertStringPtrEqual(subtest, model.RecordDetails, testCase.recordDetails)
			assertStringPtrEqual(subtest, model.OperatorSri, testCase.operatorSri)
			assertStringPtrEqual(subtest, model.OperatorIpAddress, testCase.operatorIpAddress)
		}
	})

	test.Run("RecordIdNonZero", func(subtest *testing.T) {
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
				recordDetails:     new("update details"),
				operatorSri:       new("sri://1:account/999"),
				operatorIpAddress: new("10.0.0.1"),
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
				subtest.Errorf("IdNotSetCorrectly: expected %d, got %d", testCase.expectedId, model.ID)
			}
			if model.RecordLevel != testCase.recordLevel {
				subtest.Errorf("RecordLevelNotSetCorrectly: expected %s, got %s", testCase.recordLevel, model.RecordLevel)
			}
			if model.RecordCode != testCase.recordCode {
				subtest.Errorf("RecordCodeNotSetCorrectly: expected %s, got %s", testCase.recordCode, model.RecordCode)
			}
			if len(model.AffectedResources) != len(testCase.affectedResources) {
				subtest.Errorf("AffectedResourcesNotSetCorrectly: expected %d, got %d", len(testCase.affectedResources), len(model.AffectedResources))
			}
			assertStringPtrEqual(subtest, model.RecordDetails, testCase.recordDetails)
			assertStringPtrEqual(subtest, model.OperatorSri, testCase.operatorSri)
			assertStringPtrEqual(subtest, model.OperatorIpAddress, testCase.operatorIpAddress)
		}
	})
}

func TestToEntity(test *testing.T) {
	test.Run("ValidData", func(subtest *testing.T) {
		model := ActivityRecord{
			ID:          42,
			RecordLevel: tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:  "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{
				{SystemResourceIdentifier: "sri://1:account/120"},
				{SystemResourceIdentifier: "sri://10:virtualHost/local.os"},
			},
			RecordDetails:     new("UserLoggedSuccessfully"),
			OperatorSri:       new("sri://1:account/123"),
			OperatorIpAddress: new("192.168.1.100"),
			CreatedAt:         time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		entity, err := model.ToEntity()
		if err != nil {
			subtest.Errorf("ToEntityFailed: %v", err)
			return
		}

		if entity.RecordId.Uint64() != 42 {
			subtest.Errorf("RecordIdMismatch: expected 42, got %d", entity.RecordId.Uint64())
		}
		if entity.RecordLevel.String() != "INFO" {
			subtest.Errorf("RecordLevelMismatch: expected INFO, got %s", entity.RecordLevel.String())
		}
		if entity.RecordCode.String() != "LoginSuccessful" {
			subtest.Errorf("RecordCodeMismatch: expected LoginSuccessful, got %s", entity.RecordCode.String())
		}
		if len(entity.AffectedResources) != 2 {
			subtest.Errorf("AffectedResourcesLengthMismatch: expected 2, got %d", len(entity.AffectedResources))
		}
		if entity.AffectedResources[0].String() != "sri://1:account/120" {
			subtest.Errorf("FirstAffectedResourceMismatch: expected sri://1:account/120, got %s", entity.AffectedResources[0].String())
		}
		if entity.AffectedResources[1].String() != "sri://10:virtualHost/local.os" {
			subtest.Errorf("SecondAffectedResourceMismatch: expected sri://10:virtualHost/local.os, got %s", entity.AffectedResources[1].String())
		}
		if entity.RecordDetails != "UserLoggedSuccessfully" {
			subtest.Errorf("RecordDetailsMismatch: expected 'UserLoggedSuccessfully', got %v", entity.RecordDetails)
		}
		if entity.OperatorSri == nil || entity.OperatorSri.String() != "sri://1:account/123" {
			subtest.Errorf("OperatorSriMismatch: expected sri://1:account/123, got %v", entity.OperatorSri)
		}
		if entity.OperatorIpAddress == nil || entity.OperatorIpAddress.String() != "192.168.1.100" {
			subtest.Errorf("OperatorIpAddressMismatch: expected 192.168.1.100, got %v", entity.OperatorIpAddress)
		}
		if entity.CreatedAt.Int64() != 1673778600 {
			subtest.Errorf("CreatedAtMismatch: expected 1673778600, got %d", entity.CreatedAt.Int64())
		}
	})

	test.Run("ValidDataWithNilPointers", func(subtest *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       tkValueObject.ActivityRecordLevelError.String(),
			RecordCode:        "LoginFailed",
			AffectedResources: []ActivityRecordAffectedResource{},
			CreatedAt:         time.Date(2023, 2, 20, 15, 45, 30, 0, time.UTC),
		}

		entity, err := model.ToEntity()
		if err != nil {
			subtest.Errorf("ToEntityFailed: %v", err)
			return
		}

		if entity.OperatorSri != nil {
			subtest.Errorf("OperatorSriShouldBeNil")
		}
		if entity.OperatorIpAddress != nil {
			subtest.Errorf("OperatorIpAddressShouldBeNil")
		}
		if entity.RecordDetails != nil {
			subtest.Errorf("RecordDetailsShouldBeNil: got %v", entity.RecordDetails)
		}
	})

	test.Run("InvalidRecordLevel", func(subtest *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       "InvalidRecordLevel",
			RecordCode:        "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{},
			CreatedAt:         time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			subtest.Errorf("ToEntityShouldFailWithInvalidRecordLevel")
		}
		if !strings.Contains(err.Error(), "ActivityRecordLevel") {
			subtest.Errorf("ErrorShouldContainActivityRecordLevel: got '%s'", err.Error())
		}
	})

	test.Run("InvalidRecordCode", func(subtest *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:        "a",
			AffectedResources: []ActivityRecordAffectedResource{},
			CreatedAt:         time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			subtest.Errorf("ToEntityShouldFailWithInvalidRecordCode")
		}
		if !strings.Contains(err.Error(), "InvalidActivityRecordCode") {
			subtest.Errorf("ErrorShouldContainInvalidActivityRecordCode: got '%s'", err.Error())
		}
	})

	test.Run("InvalidSystemResourceIdentifier", func(subtest *testing.T) {
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
			subtest.Errorf("ToEntityShouldFailWithInvalidSystemResourceIdentifier")
		}
		if !strings.Contains(err.Error(), "SystemResourceIdentifier") {
			subtest.Errorf("ErrorShouldContainSystemResourceIdentifier: got '%s'", err.Error())
		}
	})

	test.Run("InvalidOperatorIpAddress", func(subtest *testing.T) {
		model := ActivityRecord{
			ID:                1,
			RecordLevel:       tkValueObject.ActivityRecordLevelInfo.String(),
			RecordCode:        "LoginSuccessful",
			AffectedResources: []ActivityRecordAffectedResource{},
			OperatorIpAddress: new("invalid-ip"),
			CreatedAt:         time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			subtest.Errorf("ToEntityShouldFailWithInvalidOperatorIpAddress")
		}
		if !strings.Contains(err.Error(), "InvalidIpAddress") {
			subtest.Errorf("ErrorShouldContainInvalidIpAddress: got '%s'", err.Error())
		}
	})
}

func assertStringPtrEqual(test *testing.T, actual, expected *string) {
	test.Helper()
	if (actual == nil && expected != nil) || (actual != nil && expected == nil) {
		test.Errorf("StringPtrMismatch: actual=%v, expected=%v", actual, expected)
		return
	}
	if actual != nil && expected != nil && *actual != *expected {
		test.Errorf("StringPtrValueMismatch: actual=%s, expected=%s", *actual, *expected)
	}
}
