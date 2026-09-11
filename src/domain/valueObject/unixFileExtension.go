package tkValueObject

import (
	"errors"
	"mime"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var unixFileExtensionRegex = regexp.MustCompile(`^([\w\-]{1,15}\.)?[\w\-]{1,15}$`)

var ErrFileExtensionInvalid = errors.New("FileExtensionInvalid")

type UnixFileExtension string

func NewUnixFileExtension(value any) (
	unixFileExtension UnixFileExtension, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unixFileExtension, errors.New("UnixFileExtensionMustBeString")
	}
	stringValue = strings.TrimPrefix(stringValue, ".")

	if !unixFileExtensionRegex.MatchString(stringValue) {
		return unixFileExtension, ErrFileExtensionInvalid
	}

	return UnixFileExtension(stringValue), nil
}

func (vo UnixFileExtension) ReadMimeType() MimeType {
	mimeTypeStr := "generic"

	fileExtWithLeadingDot := "." + string(vo)
	mimeTypeWithCharset := mime.TypeByExtension(fileExtWithLeadingDot)
	if len(mimeTypeWithCharset) > 0 {
		mimeTypeOnly, _, _ := strings.Cut(mimeTypeWithCharset, ";")
		mimeTypeStr = mimeTypeOnly
	}

	mimeType, _ := NewMimeType(mimeTypeStr)
	return mimeType
}

func (vo UnixFileExtension) String() string {
	return string(vo)
}
