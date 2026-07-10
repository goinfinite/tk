package tkRepository

import (
	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotCmdRepo interface {
	Create(createDto tkDto.CreateHoneypotHit) error
	DeleteExpired(banDuration tkValueObject.HoneypotBanDuration) error
	EnforceMaxEntries(maxEntries tkValueObject.HoneypotMaxEntries) error
}
