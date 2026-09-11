package tkValueObject

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

// There is no allowUnsafeChars param because relative paths are always unsafe anyway.
var (
	unixRelativeFilePathRegex        = regexp.MustCompile(`^[\p{L}\p{N}\p{Pc}\p{Pd}\.\~][\p{L}\p{N}\p{Pc}\p{Pd}\p{Zs}\p{S}\p{P}\(\)\[\]\+\.\/]*$`)
	unixTypicalRelativeFilePathRegex = regexp.MustCompile(`^\/?\.\/|^\/?\~\/|\/?\.\.\/|^\/?~[^\/]+\/`)
)

type UnixRelativeFilePath string

// Besides validating the input, NewUnixRelativeFilePath will also normalize the path.
// If the path is not relative, it will be converted to a relative path. For example,
// "/my/file.go" will be converted to "./my/file.go". Be careful with relative paths.
// Consider using UnixAbsoluteFilePath VO whenever possible.
func NewUnixRelativeFilePath(value any) (filePath UnixRelativeFilePath, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return filePath, errors.New("UnixRelativeFilePathValueMustBeString")
	}
	if len(stringValue) == 0 {
		return filePath, errors.New("UnixRelativeFilePathValueMustNotBeEmpty")
	}

	if len(stringValue) > 4096 {
		return filePath, errors.New("UnixRelativeFilePathTooBig")
	}

	stringValue = strings.TrimPrefix(stringValue, "/")
	if !unixTypicalRelativeFilePathRegex.MatchString(stringValue) {
		switch stringValue {
		case "", ".":
			stringValue = "./"
		case "..":
			stringValue = "../"
		default:
			stringValue = "./" + stringValue
		}

		if !unixTypicalRelativeFilePathRegex.MatchString(stringValue) {
			return filePath, errors.New("PathMustBeRelative")
		}
	}

	if !unixRelativeFilePathRegex.MatchString(stringValue) {
		return filePath, errors.New("InvalidUnixRelativeFilePath")
	}

	return UnixRelativeFilePath(stringValue), nil
}

func (vo UnixRelativeFilePath) ReadWithoutExtension() UnixRelativeFilePath {
	fileExt, err := vo.ReadCompoundFileExtension()
	if err != nil {
		return vo
	}
	if fileExt == "" {
		return vo
	}

	extStr := "." + fileExt.String()
	rawFilePathWithoutExt := strings.TrimSuffix(string(vo), extStr)
	filePathWithoutExt, _ := NewUnixRelativeFilePath(rawFilePathWithoutExt)
	return filePathWithoutExt
}

func (vo UnixRelativeFilePath) ReadFileName() (UnixFileName, error) {
	_, rawFileName := filepath.Split(string(vo))
	return NewUnixFileName(rawFileName, true)
}

func (vo UnixRelativeFilePath) ReadFileExtension() (UnixFileExtension, error) {
	fileName, fileNameErr := vo.ReadFileName()
	if fileNameErr != nil {
		return "", fileNameErr
	}

	fileNameStr := fileName.String()
	fileNameWithoutLeadingDots := strings.TrimLeft(fileNameStr, ".")
	isExtensionlessDotfile := strings.HasPrefix(fileNameStr, ".") &&
		fileNameWithoutLeadingDots != "" &&
		!strings.Contains(fileNameWithoutLeadingDots, ".")
	if isExtensionlessDotfile {
		return "", nil
	}

	rawFileExtension := filepath.Ext(fileNameStr)
	isTrailingDotName := rawFileExtension == "."
	if rawFileExtension == "" || isTrailingDotName {
		return "", nil
	}

	return NewUnixFileExtension(rawFileExtension)
}

func (vo UnixRelativeFilePath) ReadCompoundFileExtension() (UnixFileExtension, error) {
	fileName, fileNameErr := vo.ReadFileName()
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

func (vo UnixRelativeFilePath) ReadFileNameWithoutExtension() (UnixFileName, error) {
	fileName, fileNameErr := vo.ReadFileName()
	if fileNameErr != nil {
		return "", fileNameErr
	}

	fileExt, extErr := vo.ReadCompoundFileExtension()
	if extErr != nil || fileExt == "" {
		return fileName, nil
	}

	fileBaseWithoutExtStr := strings.TrimSuffix(
		fileName.String(), "."+fileExt.String(),
	)
	return NewUnixFileName(fileBaseWithoutExtStr, true)
}

func (vo UnixRelativeFilePath) ReadFileDir() UnixRelativeFilePath {
	unixFileDirPath, _ := NewUnixRelativeFilePath(filepath.Dir(string(vo)))
	return unixFileDirPath
}

func (vo UnixRelativeFilePath) String() string {
	return string(vo)
}
