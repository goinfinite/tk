package tkRepository

import (
	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkEntity "github.com/goinfinite/tk/src/domain/entity"
)

type ActivityRecordQueryRepo interface {
	Read(tkDto.ReadActivityRecordsRequest) (tkDto.ReadActivityRecordsResponse, error)
	ReadFirst(tkDto.ReadActivityRecordsRequest) (tkEntity.ActivityRecord, error)
}
