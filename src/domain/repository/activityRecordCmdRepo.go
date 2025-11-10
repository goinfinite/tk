package tkRepository

import tkDto "github.com/goinfinite/tk/src/domain/dto"

type ActivityRecordCmdRepo interface {
	Create(createDto tkDto.CreateActivityRecord) error
	Delete(deleteDto tkDto.DeleteActivityRecord) error
}
