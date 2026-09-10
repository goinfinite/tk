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
// Blacklist approach: block control chars and shell-dangerous chars; allow all else.
var (
	unixAbsoluteFilePathStrictRegex = regexp.MustCompile(`^[\/\p{L}\p{N}\p{Pc}\p{Pd}\.][^\x00-\x1f\x7f;|&$` + "`" + `><{}!#*?@%\\]*$`)
	unixAbsoluteFilePathUnsafeRegex = regexp.MustCompile(`^[\/\p{L}\p{N}\p{Pc}\p{Pd}\.][^\x00-\x1f\x7f]*$`)
)

type UnixAbsoluteFilePath string

func NewUnixAbsoluteFilePath(value any, allowUnsafeChars bool) (
	filePath UnixAbsoluteFilePath, err error,
) {
	if existentFilePath, assertOk := value.(UnixAbsoluteFilePath); assertOk {
		return existentFilePath, nil
	}

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

	if !strings.HasPrefix(stringValue, "/") {
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

func (vo UnixAbsoluteFilePath) ReadFileName(
	allowUnsafeChars bool,
) (UnixFileName, error) {
	_, rawFileName := filepath.Split(string(vo))
	return NewUnixFileName(rawFileName, allowUnsafeChars)
}

func (vo UnixAbsoluteFilePath) ReadFileExtension() (UnixFileExtension, error) {
	fileName, fileNameErr := vo.ReadFileName(true)
	if fileNameErr != nil {
		return "", fileNameErr
	}

	fileNameWithoutLeadingDots := strings.TrimLeft(fileName.String(), ".")
	isExtensionlessDotfile := strings.HasPrefix(fileName.String(), ".") &&
		fileNameWithoutLeadingDots != "" &&
		!strings.Contains(fileNameWithoutLeadingDots, ".")
	if isExtensionlessDotfile {
		return NewUnixFileExtension("")
	}

	unixFileExtensionStr := filepath.Ext(string(vo))
	return NewUnixFileExtension(unixFileExtensionStr)
}

func (vo UnixAbsoluteFilePath) ReadCompoundFileExtension() (UnixFileExtension, error) {
	fileName, fileNameErr := vo.ReadFileName(true)
	if fileNameErr != nil {
		return "", fileNameErr
	}

	fileNameParts := strings.Split(
		strings.TrimPrefix(fileName.String(), "."), ".",
	)
	if len(fileNameParts) < 3 {
		return vo.ReadFileExtension()
	}
	extensionsOnly := fileNameParts[1:]
	return NewUnixFileExtension(strings.Join(extensionsOnly, "."))
}

func (vo UnixAbsoluteFilePath) ReadFileNameWithoutExtension(
	allowUnsafeChars bool,
) (UnixFileName, error) {
	fileName, fileNameErr := vo.ReadFileName(allowUnsafeChars)
	if fileNameErr != nil {
		return "", fileNameErr
	}

	fileExt, extErr := vo.ReadCompoundFileExtension()
	if extErr != nil {
		return fileName, nil
	}

	rawFileBaseWithoutExt := strings.TrimSuffix(
		fileName.String(), "."+fileExt.String(),
	)
	return NewUnixFileName(rawFileBaseWithoutExt, allowUnsafeChars)
}

func (vo UnixAbsoluteFilePath) ReadFileDir() UnixAbsoluteFilePath {
	unixFileDirPath, _ := NewUnixAbsoluteFilePath(filepath.Dir(string(vo)), true)
	return unixFileDirPath
}

func (vo UnixAbsoluteFilePath) String() string {
	return string(vo)
}
