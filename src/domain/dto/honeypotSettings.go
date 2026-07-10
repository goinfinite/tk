package tkDto

import tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"

type HoneypotSettings struct {
	AggressivenessMode tkValueObject.HoneypotAggressivenessMode `json:"aggressivenessMode"`
	ActivePathCount    tkValueObject.HoneypotActivePathCount    `json:"activePathCount"`
	MaxEntries         tkValueObject.HoneypotMaxEntries         `json:"maxEntries"`
	MaxStreamSize      tkValueObject.HoneypotMaxStreamSize      `json:"maxStreamSize"`
	StatsInterval      tkValueObject.HoneypotStatsInterval      `json:"statsInterval"`
	BanDuration        tkValueObject.HoneypotBanDuration        `json:"banDuration"`
	RandomSeed         int64                                    `json:"randomSeed"`
}
