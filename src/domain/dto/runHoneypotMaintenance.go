package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type RunHoneypotMaintenanceRequest struct {
	MaxEntries   tkValueObject.HoneypotMaxEntries    `json:"maxEntries"`
	BanDuration  tkValueObject.HoneypotBanDuration   `json:"banDuration"`
	StatsInterval tkValueObject.HoneypotStatsInterval `json:"statsInterval"`
}
