package tkRepository

import (
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotCmdRepo interface {
	IncrementHit(
		requesterIp tkValueObject.IpAddress,
		interceptPath string,
	)
	CleanExpiredEntries(banDuration time.Duration)
	EnforceMaxEntries(maxEntries int)
}
