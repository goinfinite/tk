package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type ReadHoneypotBanDecisionRequest struct {
	RequesterIpAddress tkValueObject.IpAddress `json:"requesterIpAddress"`
}

type ReadHoneypotBanDecisionResponse struct {
	IsBanned         bool                                `json:"isBanned"`
	HitCount         uint64                              `json:"hitCount"`
	SuggestedAction  tkValueObject.HoneypotSuggestedAction `json:"suggestedAction"`
}
