package tkRepository

import (
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotQueryRepo interface {
	ReadHitRecord(
		requesterIp tkValueObject.IpAddress,
	) (tkDto.HoneypotHitData, error)
	Count() int64
	ReadReport(
		banDuration time.Duration,
		aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
	) (tkDto.HoneypotStatsReport, error)
}
