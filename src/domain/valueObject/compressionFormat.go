package tkValueObject

import (
	"errors"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

type CompressionFormat string

const (
	CompressionFormatTarball CompressionFormat = "tar"
	CompressionFormatGzip    CompressionFormat = "gzip"
	CompressionFormatZip     CompressionFormat = "zip"
	CompressionFormatXz      CompressionFormat = "xz"
	CompressionFormatBrotli  CompressionFormat = "br"
)

var ValidCompressionFormats = []string{
	"tar", "gzip", "zip", "xz", "br",
}

func NewCompressionFormat(value any) (CompressionFormat, error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return "", errors.New("CompressionFormatMustBeString")
	}

	stringValue = strings.TrimPrefix(stringValue, ".")
	stringValue = strings.ToLower(stringValue)

	switch stringValue {
	case "gz":
		stringValue = "gzip"
	case "tarball":
		stringValue = "tar"
	case "brotli":
		stringValue = "br"
	case "":
		return "", errors.New("CompressionFormatCannotBeEmpty")
	}

	stringValueVo := CompressionFormat(stringValue)
	switch stringValueVo {
	case CompressionFormatTarball, CompressionFormatGzip, CompressionFormatZip,
		CompressionFormatXz, CompressionFormatBrotli:
		return stringValueVo, nil
	default:
		return "", errors.New("InvalidCompressionFormat")
	}
}

func (vo CompressionFormat) String() string {
	return string(vo)
}
