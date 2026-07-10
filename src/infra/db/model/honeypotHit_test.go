package tkInfraDbModel

import (
	"strings"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestNewHoneypotHitModel(t *testing.T) {
	testCases := []struct {
		name                string
		requesterIpAddress  string
		honeypotPath        string
		hitClass            string
		hitCount            uint64
		firstHitAt          time.Time
		createdAt           time.Time
		expectedIp          string
		expectedPath        string
		expectedClass       string
		expectedCount       uint64
	}{
		{
			name:               "ValidParams",
			requesterIpAddress: "192.168.1.1",
			honeypotPath:       "/wp-admin",
			hitClass:           "staticVulnerability",
			hitCount:           5,
			firstHitAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			createdAt:          time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedIp:         "192.168.1.1",
			expectedPath:       "/wp-admin",
			expectedClass:      "staticVulnerability",
			expectedCount:      5,
		},
		{
			name:               "EmptyIpAddress",
			requesterIpAddress: "",
			honeypotPath:       "/wp-admin",
			hitClass:           "staticVulnerability",
			hitCount:           1,
			firstHitAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			createdAt:          time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedIp:         "",
			expectedPath:       "/wp-admin",
			expectedClass:      "staticVulnerability",
			expectedCount:      1,
		},
		{
			name:               "HitCountZero",
			requesterIpAddress: "10.0.0.1",
			honeypotPath:       "/login",
			hitClass:           "aiTrap",
			hitCount:           0,
			firstHitAt:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			createdAt:          time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedIp:         "10.0.0.1",
			expectedPath:       "/login",
			expectedClass:      "aiTrap",
			expectedCount:      0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			model := NewHoneypotHitModel(
				testCase.requesterIpAddress,
				testCase.honeypotPath,
				testCase.hitClass,
				testCase.hitCount,
				testCase.firstHitAt,
				testCase.createdAt,
			)

			if model.RequesterIpAddress != testCase.expectedIp {
				t.Errorf(
					"IpAddressMismatch: expected %s, got %s",
					testCase.expectedIp,
					model.RequesterIpAddress,
				)
			}
			if model.HoneypotPath != testCase.expectedPath {
				t.Errorf(
					"HoneypotPathMismatch: expected %s, got %s",
					testCase.expectedPath,
					model.HoneypotPath,
				)
			}
			if model.HitClass != testCase.expectedClass {
				t.Errorf(
					"HitClassMismatch: expected %s, got %s",
					testCase.expectedClass,
					model.HitClass,
				)
			}
			if model.HitCount != testCase.expectedCount {
				t.Errorf(
					"HitCountMismatch: expected %d, got %d",
					testCase.expectedCount,
					model.HitCount,
				)
			}
			if model.FirstHitAt != testCase.firstHitAt {
				t.Errorf(
					"FirstHitAtMismatch: expected %v, got %v",
					testCase.firstHitAt,
					model.FirstHitAt,
				)
			}
			if model.CreatedAt != testCase.createdAt {
				t.Errorf(
					"CreatedAtMismatch: expected %v, got %v",
					testCase.createdAt,
					model.CreatedAt,
				)
			}
		})
	}
}

func TestHoneypotHitModelToEntity(t *testing.T) {
	t.Run("ValidModel", func(t *testing.T) {
		firstHitAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

		model := HoneypotHitModel{
			RequesterIpAddress: "192.168.1.1",
			HoneypotPath:       "/wp-admin",
			HitClass:           "staticVulnerability",
			HitCount:           5,
			FirstHitAt:         firstHitAt,
			CreatedAt:          createdAt,
		}

		entity, err := model.ToEntity()
		if err != nil {
			t.Errorf("ToEntityFailed: %v", err)
			return
		}

		if entity.RequesterIpAddress.String() != "192.168.1.1" {
			t.Errorf(
				"RequesterIpMismatch: expected 192.168.1.1, got %s",
				entity.RequesterIpAddress.String(),
			)
		}
		if entity.HoneypotPath.String() != "/wp-admin" {
			t.Errorf(
				"HoneypotPathMismatch: expected /wp-admin, got %s",
				entity.HoneypotPath.String(),
			)
		}
		if entity.HitClass.String() != "staticVulnerability" {
			t.Errorf(
				"HitClassMismatch: expected staticVulnerability, got %s",
				entity.HitClass.String(),
			)
		}
		if entity.HitCount != 5 {
			t.Errorf("HitCountMismatch: expected 5, got %d", entity.HitCount)
		}
		expectedFirstHitAt := tkValueObject.NewUnixTimeWithGoTime(firstHitAt)
		if entity.FirstHitAt != expectedFirstHitAt {
			t.Errorf(
				"FirstHitAtMismatch: expected %d, got %d",
				expectedFirstHitAt,
				entity.FirstHitAt,
			)
		}
		expectedCreatedAt := tkValueObject.NewUnixTimeWithGoTime(createdAt)
		if entity.CreatedAt != expectedCreatedAt {
			t.Errorf(
				"CreatedAtMismatch: expected %d, got %d",
				expectedCreatedAt,
				entity.CreatedAt,
			)
		}
	})

	t.Run("InvalidIpAddress", func(t *testing.T) {
		model := HoneypotHitModel{
			RequesterIpAddress: "not-an-ip",
			HoneypotPath:       "/wp-admin",
			HitClass:           "staticVulnerability",
			HitCount:           1,
			FirstHitAt:         time.Now(),
			CreatedAt:          time.Now(),
		}

		_, err := model.ToEntity()
		if err == nil {
			t.Errorf("ToEntityShouldFailWithInvalidIpAddress")
		}
		if !strings.Contains(err.Error(), "InvalidIpAddress") {
			t.Errorf(
				"ErrorShouldContainInvalidIpAddress: got '%s'",
				err.Error(),
			)
		}
	})

	t.Run("ZeroTimestamps", func(t *testing.T) {
		model := HoneypotHitModel{
			RequesterIpAddress: "192.168.1.1",
			HoneypotPath:       "/wp-admin",
			HitClass:           "staticVulnerability",
			HitCount:           1,
			FirstHitAt:         time.Time{},
			CreatedAt:          time.Time{},
		}

		entity, err := model.ToEntity()
		if err != nil {
			t.Errorf("ToEntityFailed: %v", err)
			return
		}

		expectedZeroTime := tkValueObject.NewUnixTimeWithGoTime(time.Time{})
		if entity.FirstHitAt != expectedZeroTime {
			t.Errorf(
				"FirstHitAtShouldBeZero: got %d",
				entity.FirstHitAt,
			)
		}
		if entity.CreatedAt != expectedZeroTime {
			t.Errorf(
				"CreatedAtShouldBeZero: got %d",
				entity.CreatedAt,
			)
		}
	})
}
