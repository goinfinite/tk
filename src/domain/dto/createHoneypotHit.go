package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type CreateHoneypotHit struct {
	RequesterIpAddress tkValueObject.IpAddress          `json:"requesterIpAddress"`
	HoneypotPath       tkValueObject.UrlPath            `json:"honeypotPath"`
	HitClass           tkValueObject.HoneypotPathClass  `json:"hitClass"`
	HitCount           uint64                           `json:"hitCount"`
}
