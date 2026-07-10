package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type ReadHoneypotStatsReportRequest struct{}

type ReadHoneypotStatsReportResponse struct {
	TotalHits  uint64                                      `json:"totalHits"`
	UniqueIps  uint64                                      `json:"uniqueIps"`
	HitsByClass map[tkValueObject.HoneypotPathClass]uint64 `json:"hitsByClass"`
	BannedIps  uint64                                      `json:"bannedIps"`
}
