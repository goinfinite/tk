package tkInfra

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
)

type FileClerk struct{}

func (FileClerk) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (clerk FileClerk) IsFile(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir() && !clerk.IsSymlink(filePath)
}

func (clerk FileClerk) IsDir(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return fileInfo.IsDir() && !clerk.IsSymlink(filePath)
}

func (clerk FileClerk) CreateFile(filePath string) error {
	fileHandler, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	return nil
}

func (clerk FileClerk) CopyFile(sourcePath, targetPath string) error {
	if !clerk.IsFile(sourcePath) {
		return errors.New("SourceFileNotFound")
	}

	if clerk.IsFile(targetPath) {
		return errors.New("TargetFileAlreadyExists")
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	bufferReader := bufio.NewReader(sourceFile)
	bufferWriter := bufio.NewWriter(targetFile)

	_, err = bufferWriter.ReadFrom(bufferReader)
	if err != nil {
		return err
	}

	return bufferWriter.Flush()
}

func (clerk FileClerk) MoveFile(sourcePath, targetPath string) error {
	if !clerk.IsFile(sourcePath) {
		return errors.New("SourceFileNotFound")
	}

	if clerk.IsFile(targetPath) {
		return errors.New("TargetFileAlreadyExists")
	}

	return os.Rename(sourcePath, targetPath)
}

func (clerk FileClerk) DeleteFile(filePath string) error {
	if !clerk.IsFile(filePath) {
		return nil
	}

	return os.Remove(filePath)
}

func (clerk FileClerk) CreateDir(dirPath string) error {
	if clerk.IsDir(dirPath) {
		return nil
	}

	return os.MkdirAll(dirPath, 0755)
}

func (clerk FileClerk) CopyDir(sourcePath, targetPath string) error {
	if !clerk.IsDir(sourcePath) {
		return errors.New("SourceDirNotFound")
	}

	if clerk.IsDir(targetPath) {
		return errors.New("TargetDirAlreadyExists")
	}

	return filepath.Walk(
		sourcePath,
		func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relativePath, err := filepath.Rel(sourcePath, filePath)
			if err != nil {
				return err
			}

			targetFilePath := filepath.Join(targetPath, relativePath)
			if fileInfo.IsDir() {
				return os.MkdirAll(targetFilePath, fileInfo.Mode())
			}

			return clerk.CopyFile(filePath, targetFilePath)
		})
}

func (clerk FileClerk) MoveDir(sourcePath, targetPath string) error {
	if !clerk.IsDir(sourcePath) {
		return errors.New("SourceDirNotFound")
	}

	if clerk.IsDir(targetPath) {
		return errors.New("TargetDirAlreadyExists")
	}

	err := clerk.CopyDir(sourcePath, targetPath)
	if err != nil {
		return err
	}

	return os.RemoveAll(sourcePath)
}

func (clerk FileClerk) DeleteDir(dirPath string) error {
	if !clerk.IsDir(dirPath) {
		return nil
	}

	return os.RemoveAll(dirPath)
}

func (clerk FileClerk) ReadFileContent(filePath string) (string, error) {
	if !clerk.IsFile(filePath) {
		return "", errors.New("FileNotFound")
	}

	fileHandler, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fileHandler.Close()

	bufferReader := bufio.NewReader(fileHandler)
	content, err := bufferReader.ReadString(0)
	if err != nil {
		return "", err
	}

	return content, nil
}

func (clerk FileClerk) UpdateFileContent(
	filePath, content string,
	shouldOverwrite bool,
) error {
	fileFlags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if shouldOverwrite {
		fileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	fileHandler, err := os.OpenFile(filePath, fileFlags, 0644)
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	bufferWriter := bufio.NewWriter(fileHandler)
	_, err = bufferWriter.WriteString(content)
	if err != nil {
		return err
	}

	return bufferWriter.Flush()
}

func (clerk FileClerk) DeleteFileContent(filePath string) error {
	return clerk.UpdateFileContent(filePath, "", true)
}

func (FileClerk) IsSymlink(sourcePath string) bool {
	linkInfo, err := os.Lstat(sourcePath)
	if err != nil {
		return false
	}

	isSymlink := linkInfo.Mode()&os.ModeSymlink == os.ModeSymlink
	return isSymlink
}

func (clerk FileClerk) UpdateFileOwnership(
	filePath string,
	userId, groupId int,
) error {
	return os.Lchown(filePath, userId, groupId)
}

func (clerk FileClerk) UpdateFilePermissions(
	filePath string,
	permissionsPtr *int,
) error {
	defaultFilePermission := int(0644)
	if clerk.IsDir(filePath) {
		defaultFilePermission = 0755
	}

	if permissionsPtr == nil {
		permissionsPtr = &defaultFilePermission
	}

	return os.Chmod(filePath, os.FileMode(*permissionsPtr))
}

func (clerk FileClerk) IsSymlinkTo(sourcePath string, targetPath string) bool {
	isSymlink := clerk.IsSymlink(sourcePath)
	if !isSymlink {
		return false
	}

	linkTarget, err := os.Readlink(sourcePath)
	if err != nil {
		return false
	}

	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	absLinkTarget, err := filepath.Abs(linkTarget)
	if err != nil {
		return false
	}

	return absLinkTarget == absTargetPath
}

func (clerk FileClerk) CreateSymlink(
	sourcePath, targetPath string,
	shouldOverwrite bool,
) error {
	if !shouldOverwrite && !clerk.FileExists(sourcePath) {
		return errors.New("FileNotFound")
	}

	if shouldOverwrite {
		err := os.Remove(targetPath)
		if err != nil {
			return err
		}
	}

	return os.Symlink(sourcePath, targetPath)
}

func (FileClerk) RemoveSymlink(symlinkPath string) error {
	return os.Remove(symlinkPath)
}
