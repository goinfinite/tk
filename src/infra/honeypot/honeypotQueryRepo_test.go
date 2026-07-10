package tkInfraHoneypot

import (
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

// Note: Setup/teardown are intentionally inline — test independence
// requires each file to own its preconditions, even if it duplicates code.

func TestHoneypotQueryRepoReadBanDecision(t *testing.T) {
	t.Run("Good_IpWithHitsReturnsCount", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotQueryRepo(dbSvc)

		seedHoneypotHit(
			t, dbSvc, "192.168.1.1", "/test", "staticVulnerability", 3,
		)

		ipVo, err := tkValueObject.NewIpAddress("192.168.1.1")
		if err != nil {
			t.Fatalf("CreateIpAddressVoFailed: %v", err)
		}

		response, err := repo.ReadBanDecision(
			tkDto.ReadHoneypotBanDecisionRequest{
				RequesterIpAddress: ipVo,
			},
		)
		if err != nil {
			t.Fatalf("ReadBanDecisionFailed: %v", err)
		}
		if response.HitCount != 3 {
			t.Errorf("ExpectedHitCount3: got %d", response.HitCount)
		}
	})

	t.Run("Bad_UnknownIpReturnsZero", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotQueryRepo(dbSvc)

		ipVo, err := tkValueObject.NewIpAddress("10.99.99.99")
		if err != nil {
			t.Fatalf("CreateIpAddressVoFailed: %v", err)
		}

		response, err := repo.ReadBanDecision(
			tkDto.ReadHoneypotBanDecisionRequest{
				RequesterIpAddress: ipVo,
			},
		)
		if err != nil {
			t.Fatalf("ReadBanDecisionFailed: %v", err)
		}
		if response.HitCount != 0 {
			t.Errorf("ExpectedHitCount0: got %d", response.HitCount)
		}
	})

	t.Run("Ugly_HighHitCountReturnsCorrectValue", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotQueryRepo(dbSvc)

		seedHoneypotHit(
			t, dbSvc, "172.16.0.1", "/trap", "aiTrap", 9999,
		)

		ipVo, err := tkValueObject.NewIpAddress("172.16.0.1")
		if err != nil {
			t.Fatalf("CreateIpAddressVoFailed: %v", err)
		}

		response, err := repo.ReadBanDecision(
			tkDto.ReadHoneypotBanDecisionRequest{
				RequesterIpAddress: ipVo,
			},
		)
		if err != nil {
			t.Fatalf("ReadBanDecisionFailed: %v", err)
		}
		if response.HitCount != 9999 {
			t.Errorf("ExpectedHitCount9999: got %d", response.HitCount)
		}
	})
}

func TestHoneypotQueryRepoReadStatsReport(t *testing.T) {
	t.Run("Good_MultipleIpsAndClasses", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotQueryRepo(dbSvc)

		seedHoneypotHit(
			t, dbSvc, "192.168.1.1", "/a", "staticVulnerability", 5,
		)
		seedHoneypotHit(
			t, dbSvc, "192.168.1.2", "/b", "bandwidthExhaust", 3,
		)
		seedHoneypotHit(
			t, dbSvc, "192.168.1.3", "/c", "staticVulnerability", 2,
		)

		response, err := repo.ReadStatsReport(
			tkDto.ReadHoneypotStatsReportRequest{},
		)
		if err != nil {
			t.Fatalf("ReadStatsReportFailed: %v", err)
		}
		if response.TotalHits != 10 {
			t.Errorf("ExpectedTotalHits10: got %d", response.TotalHits)
		}
		if response.UniqueIps != 3 {
			t.Errorf("ExpectedUniqueIps3: got %d", response.UniqueIps)
		}
		if len(response.HitsByClass) != 2 {
			t.Errorf(
				"Expected2Classes: got %d", len(response.HitsByClass),
			)
		}
	})

	t.Run("Bad_EmptyDatabaseReturnsZeros", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotQueryRepo(dbSvc)

		response, err := repo.ReadStatsReport(
			tkDto.ReadHoneypotStatsReportRequest{},
		)
		if err != nil {
			t.Fatalf("ReadStatsReportFailed: %v", err)
		}
		if response.TotalHits != 0 {
			t.Errorf("ExpectedTotalHits0: got %d", response.TotalHits)
		}
		if response.UniqueIps != 0 {
			t.Errorf("ExpectedUniqueIps0: got %d", response.UniqueIps)
		}
		if len(response.HitsByClass) != 0 {
			t.Errorf(
				"ExpectedEmptyHitsByClass: got %d",
				len(response.HitsByClass),
			)
		}
	})

	t.Run("Ugly_AllHitsFromSameIp", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotQueryRepo(dbSvc)

		seedHoneypotHit(
			t, dbSvc, "10.0.0.1", "/x", "staticVulnerability", 50,
		)

		response, err := repo.ReadStatsReport(
			tkDto.ReadHoneypotStatsReportRequest{},
		)
		if err != nil {
			t.Fatalf("ReadStatsReportFailed: %v", err)
		}
		if response.TotalHits != 50 {
			t.Errorf("ExpectedTotalHits50: got %d", response.TotalHits)
		}
		if response.UniqueIps != 1 {
			t.Errorf("ExpectedUniqueIps1: got %d", response.UniqueIps)
		}
	})
}
