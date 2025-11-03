package tkValueObject

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

// @see https://www.regular-expressions.info/unicodecategory.html
// UnixFileNameRegex combined with characters allowed in absolute paths.
var (
	unixAbsoluteFilePathStrictRegex = regexp.MustCompile(`^[\/\p{L}\p{N}\p{Pc}\p{Pd}\.][\p{L}\p{N}\p{Pc}\p{Pd}\p{Zs}\(\)\[\]\+\.\/]*$`)
	unixAbsoluteFilePathUnsafeRegex = regexp.MustCompile(`^[\/\p{L}\p{N}\p{Pc}\p{Pd}\.][\p{L}\p{N}\p{Pc}\p{Pd}\p{Zs}\p{S}\p{P}\(\)\[\]\+\.\/]*$`)
)

type UnixAbsoluteFilePath string

func NewUnixAbsoluteFilePath(value any, allowUnsafeChars bool) (
	filePath UnixAbsoluteFilePath, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return filePath, errors.New("UnixAbsoluteFilePathValueMustBeString")
	}

	if len(stringValue) == 0 {
		return filePath, errors.New("UnixAbsoluteFilePathValueMustNotBeEmpty")
	}

	if len(stringValue) > 4096 {
		return filePath, errors.New("UnixAbsoluteFilePathTooBig")
	}

	if !strings.Contains(stringValue, "/") {
		stringValue = "/" + stringValue
	}

	switch allowUnsafeChars {
	case true:
		if !unixAbsoluteFilePathUnsafeRegex.MatchString(stringValue) {
			return filePath, errors.New("InvalidUnixAbsoluteFilePath")
		}
	case false:
		if !unixAbsoluteFilePathStrictRegex.MatchString(stringValue) {
			return filePath, errors.New("InvalidUnixAbsoluteFilePath")
		}
	}

	if unixTypicalRelativeFilePathRegex.MatchString(stringValue) {
		return filePath, errors.New("RelativePathNotAllowed")
	}

	return UnixAbsoluteFilePath(stringValue), nil
}

func (vo UnixAbsoluteFilePath) ReadWithoutExtension(allowUnsafeChars bool) UnixAbsoluteFilePath {
	fileExt, err := vo.ReadCompoundFileExtension()
	if err != nil {
		return vo
	}

	extStr := "." + fileExt.String()
	rawFilePathWithoutExt := strings.TrimSuffix(string(vo), extStr)
	filePathWithoutExt, _ := NewUnixAbsoluteFilePath(rawFilePathWithoutExt, allowUnsafeChars)
	return filePathWithoutExt
}

func (vo UnixAbsoluteFilePath) ReadFileName(allowUnsafeChars bool) UnixFileName {
	unixFileBase := filepath.Base(string(vo))
	unixFileName, _ := NewUnixFileName(unixFileBase, allowUnsafeChars)
	return unixFileName
}

func (vo UnixAbsoluteFilePath) ReadFileExtension() (UnixFileExtension, error) {
	unixFileExtensionStr := filepath.Ext(string(vo))
	return NewUnixFileExtension(unixFileExtensionStr)
}

func (vo UnixAbsoluteFilePath) ReadCompoundFileExtension() (UnixFileExtension, error) {
	fileNameParts := strings.Split(vo.ReadFileName(true).String(), ".")
	if len(fileNameParts) < 3 {
		return vo.ReadFileExtension()
	}
	extensionsOnly := fileNameParts[1:]
	return NewUnixFileExtension(strings.Join(extensionsOnly, "."))
}

func (vo UnixAbsoluteFilePath) ReadFileNameWithoutExtension(allowUnsafeChars bool) UnixFileName {
	fileBase := filepath.Base(string(vo))
	fileExt, err := vo.ReadCompoundFileExtension()
	if err != nil {
		return vo.ReadFileName(allowUnsafeChars)
	}
	rawFileBaseWithoutExt := strings.TrimSuffix(fileBase, "."+fileExt.String())
	fileNameWithoutExt, _ := NewUnixFileName(rawFileBaseWithoutExt, allowUnsafeChars)
	return fileNameWithoutExt
}

func (vo UnixAbsoluteFilePath) ReadFileDir() UnixAbsoluteFilePath {
	unixFileDirPath, _ := NewUnixAbsoluteFilePath(filepath.Dir(string(vo)), true)
	return unixFileDirPath
}

func (vo UnixAbsoluteFilePath) String() string {
	return string(vo)
}
