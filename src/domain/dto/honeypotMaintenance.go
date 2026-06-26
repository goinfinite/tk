package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type RunHoneypotMaintenanceRequest struct {
	AggressivenessMode tkValueObject.HoneypotAggressivenessMode
	BanDuration        tkValueObject.HoneypotBanDuration
	MaxEntries         tkValueObject.HoneypotMaxEntries
	StatsRecordCode    tkValueObject.ActivityRecordCode
	StatsRecordLevel   tkValueObject.ActivityRecordLevel
}
