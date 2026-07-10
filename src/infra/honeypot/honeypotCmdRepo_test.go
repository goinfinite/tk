package tkInfraHoneypot

import (
	"testing"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

// Note: Setup/teardown are intentionally inline — test independence
// requires each file to own its preconditions, even if it duplicates code.

func TestHoneypotCmdRepoCreate(t *testing.T) {
	t.Run("Good_NewHitPersisted", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		createDto := createTestHoneypotHit(
			t, dbSvc, "192.168.1.1", "staticVulnerability",
		)

		err := repo.Create(createDto)
		if err != nil {
			t.Fatalf("CreateFailed: %v", err)
		}

		var hitCount int64
		err = dbSvc.Handler.Raw(
			"SELECT hit_count FROM honeypot_hits WHERE requester_ip_address = ?",
			"192.168.1.1",
		).Scan(&hitCount).Error
		if err != nil {
			t.Fatalf("VerifyHitCountFailed: %v", err)
		}
		if hitCount != 1 {
			t.Errorf("ExpectedHitCount1: got %d", hitCount)
		}
	})

	t.Run("Good_AllFieldsPopulated", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		ipVo, err := tkValueObject.NewIpAddress("10.0.0.1")
		if err != nil {
			t.Fatalf("CreateIpAddressVoFailed: %v", err)
		}
		pathVo, err := tkValueObject.NewUrlPath("/admin/config")
		if err != nil {
			t.Fatalf("CreateUrlPathVoFailed: %v", err)
		}
		classVo, err := tkValueObject.NewHoneypotPathClass("bandwidthExhaust")
		if err != nil {
			t.Fatalf("CreateHitClassVoFailed: %v", err)
		}

		createDto := tkDto.CreateHoneypotHit{
			RequesterIpAddress: ipVo,
			HoneypotPath:       pathVo,
			HitClass:           classVo,
			HitCount:           3,
		}

		beforeCreate := time.Now().UTC().Add(-1 * time.Second)
		err = repo.Create(createDto)
		if err != nil {
			t.Fatalf("CreateFailed: %v", err)
		}

		type storedHitRow struct {
			HoneypotPath string
			HitClass     string
			HitCount     uint64
			FirstHitAt   time.Time
			CreatedAt    time.Time
		}
		var storedRow storedHitRow
		err = dbSvc.Handler.Raw(
			`SELECT honeypot_path, hit_class, hit_count,
				first_hit_at, created_at
			FROM honeypot_hits WHERE requester_ip_address = ?`,
			"10.0.0.1",
		).Scan(&storedRow).Error
		if err != nil {
			t.Fatalf("ReadStoredFieldsFailed: %v", err)
		}
		if storedRow.HoneypotPath != "/admin/config" {
			t.Errorf(
				"PathMismatch: expected /admin/config, got %s",
				storedRow.HoneypotPath,
			)
		}
		if storedRow.HitClass != "bandwidthExhaust" {
			t.Errorf(
				"ClassMismatch: expected bandwidthExhaust, got %s",
				storedRow.HitClass,
			)
		}
		if storedRow.HitCount != 3 {
			t.Errorf("CountMismatch: expected 3, got %d", storedRow.HitCount)
		}
		if storedRow.FirstHitAt.Before(beforeCreate) {
			t.Errorf("FirstHitAtTooOld: %v", storedRow.FirstHitAt)
		}
		if storedRow.CreatedAt.Before(beforeCreate) {
			t.Errorf("CreatedAtTooOld: %v", storedRow.CreatedAt)
		}
	})

	t.Run("Ugly_DuplicateIpUpsertsIncrementCount", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		createDto := createTestHoneypotHit(
			t, dbSvc, "172.16.0.1", "staticVulnerability",
		)

		err := repo.Create(createDto)
		if err != nil {
			t.Fatalf("FirstCreateFailed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		var firstHitAtBefore time.Time
		err = dbSvc.Handler.Raw(
			"SELECT first_hit_at FROM honeypot_hits WHERE requester_ip_address = ?",
			"172.16.0.1",
		).Scan(&firstHitAtBefore).Error
		if err != nil {
			t.Fatalf("ReadFirstHitAtFailed: %v", err)
		}

		err = repo.Create(createDto)
		if err != nil {
			t.Fatalf("SecondCreateFailed: %v", err)
		}

		type upsertResultRow struct {
			HitCount   uint64
			FirstHitAt time.Time
			CreatedAt  time.Time
		}
		var upsertRow upsertResultRow
		err = dbSvc.Handler.Raw(
			`SELECT hit_count, first_hit_at, created_at
			FROM honeypot_hits WHERE requester_ip_address = ?`,
			"172.16.0.1",
		).Scan(&upsertRow).Error
		if err != nil {
			t.Fatalf("ReadAfterUpsertFailed: %v", err)
		}

		if upsertRow.HitCount != 2 {
			t.Errorf("ExpectedHitCount2: got %d", upsertRow.HitCount)
		}
		if !upsertRow.FirstHitAt.Equal(firstHitAtBefore) {
			t.Errorf(
				"FirstHitAtChanged: expected %v, got %v",
				firstHitAtBefore, upsertRow.FirstHitAt,
			)
		}
	})

	t.Run("Ugly_UpsertIsSingleAtomicStatement", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		createDto := createTestHoneypotHit(
			t, dbSvc, "10.10.10.10", "aiTrap",
		)

		err := repo.Create(createDto)
		if err != nil {
			t.Fatalf("FirstCreateFailed: %v", err)
		}

		err = repo.Create(createDto)
		if err != nil {
			t.Fatalf("SecondCreateFailed: %v", err)
		}

		var entryCount int64
		err = dbSvc.Handler.Raw(
			"SELECT COUNT(*) FROM honeypot_hits WHERE requester_ip_address = ?",
			"10.10.10.10",
		).Scan(&entryCount).Error
		if err != nil {
			t.Fatalf("CountEntriesFailed: %v", err)
		}
		if entryCount != 1 {
			t.Errorf("ExpectedExactly1Entry: got %d", entryCount)
		}
	})
}

func TestHoneypotCmdRepoDeleteExpired(t *testing.T) {
	t.Run("Good_ExpiredEntriesDeleted", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		oldModel := createTestHoneypotHit(
			t, dbSvc, "192.168.1.10", "staticVulnerability",
		)
		err := repo.Create(oldModel)
		if err != nil {
			t.Fatalf("CreateOldEntryFailed: %v", err)
		}

		oldTime := time.Now().UTC().Add(-2 * time.Hour)
		dbSvc.Handler.Exec(
			`UPDATE honeypot_hits SET created_at = ?, first_hit_at = ?
			WHERE requester_ip_address = ?`,
			oldTime, oldTime, "192.168.1.10",
		)

		recentModel := createTestHoneypotHit(
			t, dbSvc, "192.168.1.11", "staticVulnerability",
		)
		err = repo.Create(recentModel)
		if err != nil {
			t.Fatalf("CreateRecentEntryFailed: %v", err)
		}

		banDuration, err := tkValueObject.NewHoneypotBanDuration("1h")
		if err != nil {
			t.Fatalf("CreateBanDurationFailed: %v", err)
		}

		err = repo.DeleteExpired(banDuration)
		if err != nil {
			t.Fatalf("DeleteExpiredFailed: %v", err)
		}

		var remainingCount int64
		err = dbSvc.Handler.Raw(
			"SELECT COUNT(*) FROM honeypot_hits",
		).Scan(&remainingCount).Error
		if err != nil {
			t.Fatalf("CountRemainingFailed: %v", err)
		}
		if remainingCount != 1 {
			t.Errorf("Expected1Remaining: got %d", remainingCount)
		}

		var remainingIp string
		err = dbSvc.Handler.Raw(
			"SELECT requester_ip_address FROM honeypot_hits",
		).Scan(&remainingIp).Error
		if err != nil {
			t.Fatalf("ReadRemainingIpFailed: %v", err)
		}
		if remainingIp != "192.168.1.11" {
			t.Errorf(
				"WrongEntryRemaining: expected 192.168.1.11, got %s",
				remainingIp,
			)
		}
	})

	t.Run("Bad_ZeroBanDurationVoRejects", func(t *testing.T) {
		_, err := tkValueObject.NewHoneypotBanDuration("0s")
		if err == nil {
			t.Errorf("ExpectedZeroDurationRejection")
		}
	})

	t.Run("Ugly_OldCreatedAtRecentFirstHitAtKept", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		createDto := createTestHoneypotHit(
			t, dbSvc, "10.0.0.50", "bandwidthExhaust",
		)
		err := repo.Create(createDto)
		if err != nil {
			t.Fatalf("CreateEntryFailed: %v", err)
		}

		oldCreatedAt := time.Now().UTC().Add(-2 * time.Hour)
		recentFirstHitAt := time.Now().UTC().Add(-10 * time.Minute)
		dbSvc.Handler.Exec(
			`UPDATE honeypot_hits SET created_at = ?, first_hit_at = ?
			WHERE requester_ip_address = ?`,
			oldCreatedAt, recentFirstHitAt, "10.0.0.50",
		)

		banDuration, err := tkValueObject.NewHoneypotBanDuration("1h")
		if err != nil {
			t.Fatalf("CreateBanDurationFailed: %v", err)
		}

		err = repo.DeleteExpired(banDuration)
		if err != nil {
			t.Fatalf("DeleteExpiredFailed: %v", err)
		}

		var remainingCount int64
		err = dbSvc.Handler.Raw(
			"SELECT COUNT(*) FROM honeypot_hits WHERE requester_ip_address = ?",
			"10.0.0.50",
		).Scan(&remainingCount).Error
		if err != nil {
			t.Fatalf("CountRemainingFailed: %v", err)
		}
		if remainingCount != 1 {
			t.Errorf("EntryShouldBeKept: got %d remaining", remainingCount)
		}
	})
}

func TestHoneypotCmdRepoEnforceMaxEntries(t *testing.T) {
	t.Run("Good_OldestEntriesDeleted", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		ips := []string{
			"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4",
			"192.168.1.5", "192.168.1.6", "192.168.1.7", "192.168.1.8",
			"192.168.1.9", "192.168.1.10",
		}
		for _, ip := range ips {
			createDto := createTestHoneypotHit(
				t, dbSvc, ip, "staticVulnerability",
			)
			err := repo.Create(createDto)
			if err != nil {
				t.Fatalf("CreateEntryFailed(%s): %v", ip, err)
			}
			time.Sleep(5 * time.Millisecond)
		}

		maxEntries, err := tkValueObject.NewHoneypotMaxEntries(5)
		if err != nil {
			t.Fatalf("CreateMaxEntriesFailed: %v", err)
		}

		err = repo.EnforceMaxEntries(maxEntries)
		if err != nil {
			t.Fatalf("EnforceMaxEntriesFailed: %v", err)
		}

		var remainingCount int64
		err = dbSvc.Handler.Raw(
			"SELECT COUNT(*) FROM honeypot_hits",
		).Scan(&remainingCount).Error
		if err != nil {
			t.Fatalf("CountRemainingFailed: %v", err)
		}
		if remainingCount != 5 {
			t.Errorf("Expected5Remaining: got %d", remainingCount)
		}
	})

	t.Run("Bad_ZeroMaxEntriesVoRejects", func(t *testing.T) {
		_, err := tkValueObject.NewHoneypotMaxEntries(0)
		if err == nil {
			t.Errorf("ExpectedZeroMaxEntriesRejection")
		}
	})

	t.Run("Ugly_FewerEntriesThanMaxNoOp", func(t *testing.T) {
		dbSvc := setupTestTransientDatabaseService(t)
		repo := NewHoneypotCmdRepo(dbSvc)

		createDto := createTestHoneypotHit(
			t, dbSvc, "10.0.0.1", "staticVulnerability",
		)
		err := repo.Create(createDto)
		if err != nil {
			t.Fatalf("CreateEntryFailed: %v", err)
		}

		maxEntries, err := tkValueObject.NewHoneypotMaxEntries(10)
		if err != nil {
			t.Fatalf("CreateMaxEntriesFailed: %v", err)
		}

		err = repo.EnforceMaxEntries(maxEntries)
		if err != nil {
			t.Fatalf("EnforceMaxEntriesFailed: %v", err)
		}

		var remainingCount int64
		err = dbSvc.Handler.Raw(
			"SELECT COUNT(*) FROM honeypot_hits",
		).Scan(&remainingCount).Error
		if err != nil {
			t.Fatalf("CountRemainingFailed: %v", err)
		}
		if remainingCount != 1 {
			t.Errorf("Expected1Remaining: got %d", remainingCount)
		}
	})
}
