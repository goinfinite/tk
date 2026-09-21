package tkValueObject

import (
	"errors"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

const maxUnixCommandOutputBytes = 4096

type UnixCommandOutput string

func NewUnixCommandOutput(value any) (
	unixCommandOutput UnixCommandOutput, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unixCommandOutput, errors.New("UnixCommandOutputMustBeString")
	}

	stringValue = tkVoUtil.SafeTruncateString(
		stringValue, maxUnixCommandOutputBytes,
	)

	return UnixCommandOutput(stringValue), nil
}

func (vo UnixCommandOutput) String() string {
	return string(vo)
}
