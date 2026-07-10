package entity

import (
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotHit struct {
	RequesterIpAddress tkValueObject.IpAddress         `json:"requesterIpAddress"`
	HoneypotPath       tkValueObject.UrlPath           `json:"honeypotPath"`
	HitClass           tkValueObject.HoneypotPathClass `json:"hitClass"`
	HitCount           uint64                          `json:"hitCount"`
	FirstHitAt         tkValueObject.UnixTime          `json:"firstHitAt"`
	CreatedAt          tkValueObject.UnixTime          `json:"createdAt"`
}

func NewHoneypotHit(
	requesterIpAddress tkValueObject.IpAddress,
	honeypotPath tkValueObject.UrlPath,
	hitClass tkValueObject.HoneypotPathClass,
	hitCount uint64,
	firstHitAt tkValueObject.UnixTime,
	createdAt tkValueObject.UnixTime,
) HoneypotHit {
	return HoneypotHit{
		RequesterIpAddress: requesterIpAddress,
		HoneypotPath:       honeypotPath,
		HitClass:           hitClass,
		HitCount:           hitCount,
		FirstHitAt:         firstHitAt,
		CreatedAt:          createdAt,
	}
}
