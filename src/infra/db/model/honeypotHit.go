package tkInfraDbModel

import (
	"time"

	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotHitModel struct {
	RequesterIpAddress string    `gorm:"primaryKey;not null;uniqueIndex"`
	HoneypotPath       string    `gorm:"not null"`
	HitClass           string    `gorm:"not null"`
	HitCount           uint64    `gorm:"not null"`
	FirstHitAt         time.Time `gorm:"not null"`
	CreatedAt          time.Time `gorm:"not null"`
}

func (HoneypotHitModel) TableName() string {
	return "honeypot_hits"
}

func NewHoneypotHitModel(
	requesterIpAddress string,
	honeypotPath string,
	hitClass string,
	hitCount uint64,
	firstHitAt time.Time,
	createdAt time.Time,
) HoneypotHitModel {
	return HoneypotHitModel{
		RequesterIpAddress: requesterIpAddress,
		HoneypotPath:       honeypotPath,
		HitClass:           hitClass,
		HitCount:           hitCount,
		FirstHitAt:         firstHitAt,
		CreatedAt:          createdAt,
	}
}

func (model HoneypotHitModel) ToEntity() (tkEntity.HoneypotHit, error) {
	requesterIpAddress, err := tkValueObject.NewIpAddress(
		model.RequesterIpAddress,
	)
	if err != nil {
		return tkEntity.HoneypotHit{}, err
	}

	honeypotPath, err := tkValueObject.NewUrlPath(model.HoneypotPath)
	if err != nil {
		return tkEntity.HoneypotHit{}, err
	}

	hitClass, err := tkValueObject.NewHoneypotPathClass(model.HitClass)
	if err != nil {
		return tkEntity.HoneypotHit{}, err
	}

	return tkEntity.NewHoneypotHit(
		requesterIpAddress,
		honeypotPath,
		hitClass,
		model.HitCount,
		tkValueObject.NewUnixTimeWithGoTime(model.FirstHitAt),
		tkValueObject.NewUnixTimeWithGoTime(model.CreatedAt),
	), nil
}
