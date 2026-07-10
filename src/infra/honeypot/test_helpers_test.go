package tkInfraHoneypot

import (
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
	tkInfraDbModel "github.com/goinfinite/tk/src/infra/db/model"
	"gorm.io/gorm"
)

func setupTestTransientDatabaseService(
	t *testing.T,
) *tkInfraDb.TransientDatabaseService {
	t.Helper()
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	ormSvc, err := gorm.Open(
		sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"),
		&gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		},
	)
	if err != nil {
		t.Fatalf("SetupTransientDatabaseServiceFailed: %v", err)
	}
	err = ormSvc.AutoMigrate(&tkInfraDbModel.HoneypotHitModel{})
	if err != nil {
		t.Fatalf("SetupTransientDatabaseMigrationFailed: %v", err)
	}
	return &tkInfraDb.TransientDatabaseService{Handler: ormSvc}
}

func createTestHoneypotHit(
	t *testing.T,
	dbSvc *tkInfraDb.TransientDatabaseService,
	ipAddress string,
	hitClass string,
) tkDto.CreateHoneypotHit {
	t.Helper()
	ipVo, err := tkValueObject.NewIpAddress(ipAddress)
	if err != nil {
		t.Fatalf("CreateIpAddressVoFailed: %v", err)
	}
	pathVo, err := tkValueObject.NewUrlPath("/test/path")
	if err != nil {
		t.Fatalf("CreateUrlPathVoFailed: %v", err)
	}
	classVo, err := tkValueObject.NewHoneypotPathClass(hitClass)
	if err != nil {
		t.Fatalf("CreateHitClassVoFailed: %v", err)
	}
	return tkDto.CreateHoneypotHit{
		RequesterIpAddress: ipVo,
		HoneypotPath:       pathVo,
		HitClass:           classVo,
		HitCount:           1,
	}
}

func seedHoneypotHit(
	t *testing.T,
	dbSvc *tkInfraDb.TransientDatabaseService,
	ipAddress string,
	path string,
	hitClass string,
	hitCount uint64,
) {
	t.Helper()
	now := time.Now().UTC()
	execResult := dbSvc.Handler.Exec(
		`INSERT INTO honeypot_hits
			(requester_ip_address, honeypot_path, hit_class,
			 hit_count, first_hit_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		ipAddress, path, hitClass, hitCount, now, now,
	)
	if execResult.Error != nil {
		t.Fatalf("SeedHoneypotHitFailed: %v", execResult.Error)
	}
}
