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
	unixTypicalRelativeFilePathRegex = regexp.MustCompile(`^\/?\.\/|^\/?\~\/|\/?\.\.\/`)
)

type UnixRelativeFilePath string

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

	extStr := "." + fileExt.String()
	rawFilePathWithoutExt := strings.TrimSuffix(string(vo), extStr)
	filePathWithoutExt, _ := NewUnixRelativeFilePath(rawFilePathWithoutExt)
	return filePathWithoutExt
}

func (vo UnixRelativeFilePath) ReadFileName() UnixFileName {
	unixFileBase := filepath.Base(string(vo))
	unixFileName, _ := NewUnixFileName(unixFileBase, true)
	return unixFileName
}

func (vo UnixRelativeFilePath) ReadFileExtension() (UnixFileExtension, error) {
	unixFileExtensionStr := filepath.Ext(string(vo))
	return NewUnixFileExtension(unixFileExtensionStr)
}

func (vo UnixRelativeFilePath) ReadCompoundFileExtension() (UnixFileExtension, error) {
	fileNameParts := strings.Split(vo.ReadFileName().String(), ".")
	if len(fileNameParts) < 3 {
		return vo.ReadFileExtension()
	}
	extensionsOnly := fileNameParts[1:]
	return NewUnixFileExtension(strings.Join(extensionsOnly, "."))
}

func (vo UnixRelativeFilePath) ReadFileNameWithoutExtension() UnixFileName {
	fileBase := filepath.Base(string(vo))
	fileExt, err := vo.ReadCompoundFileExtension()
	if err != nil {
		return vo.ReadFileName()
	}
	fileBaseWithoutExtStr := strings.TrimSuffix(fileBase, "."+fileExt.String())
	fileNameWithoutExt, _ := NewUnixFileName(fileBaseWithoutExtStr, true)
	return fileNameWithoutExt
}

func (vo UnixRelativeFilePath) ReadFileDir() UnixRelativeFilePath {
	unixFileDirPath, _ := NewUnixRelativeFilePath(filepath.Dir(string(vo)))
	return unixFileDirPath
}

func (vo UnixRelativeFilePath) String() string {
	return string(vo)
}
