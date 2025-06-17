package tkInfra

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func (clerk FileClerk) RenameFile(sourcePath, targetPath string) error {
	return clerk.MoveFile(sourcePath, targetPath)
}

func (clerk FileClerk) DeleteFile(filePath string) error {
	if !clerk.IsFile(filePath) {
		return nil
	}

	return os.Remove(filePath)
}

func (clerk FileClerk) ReadFileContent(
	filePath string,
	maxContentSizeBytesPtr *int64,
) (string, error) {
	if !clerk.IsFile(filePath) {
		return "", errors.New("FileNotFound")
	}

	maxContentSizeBytes := int64(1 * 1073741824) // 1GiB
	if maxContentSizeBytesPtr != nil {
		maxContentSizeBytes = *maxContentSizeBytesPtr
	}

	fileHandler, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fileHandler.Close()

	limitedReader := io.LimitedReader{R: fileHandler, N: maxContentSizeBytes}
	fileContentBytes, err := io.ReadAll(&limitedReader)
	if err != nil {
		return "", err
	}

	return string(fileContentBytes), nil
}

func (clerk FileClerk) UpdateFileContent(
	filePath, newContent string,
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
	_, err = bufferWriter.WriteString(newContent)
	if err != nil {
		return err
	}

	return bufferWriter.Flush()
}

func (clerk FileClerk) DeleteFileContent(filePath string) error {
	return clerk.UpdateFileContent(filePath, "", true)
}

func (clerk FileClerk) TruncateFileContent(filePath string) error {
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

func (clerk FileClerk) CompressFile(
	sourcePath string,
	compressionFormatPtr *string,
) (compressedFilePath string, err error) {
	compressionCmd := "tar"
	compressionArgs := []string{"--create", "--file"}
	compressionSuffix := ".tar"
	if compressionFormatPtr != nil {
		switch *compressionFormatPtr {
		case "tar", "tarball":
		case "br", "brotli":
			compressionSuffix = ".br"
			compressionCmd = "brotli"
			compressionArgs = []string{"--quality=4", "--rm"}
		case "gz", "gzip":
			compressionSuffix = ".gz"
			compressionCmd = "gzip"
			compressionArgs = []string{"-6"}
		case "zip":
			compressionSuffix = ".zip"
			compressionCmd = "zip"
			compressionArgs = []string{"-6", "--quiet", "--move", "--test"}
		case "xz":
			compressionSuffix = ".xz"
			compressionCmd = "xz"
			compressionArgs = []string{"-1", "--memlimit=10%"}
		default:
			return compressedFilePath, errors.New("UnsupportedCompressionFormat")
		}
	}

	if !clerk.IsFile(sourcePath) {
		if !clerk.IsDir(sourcePath) {
			return compressedFilePath, errors.New("SourceFileNotFound")
		}
		if compressionSuffix != ".tar" {
			return compressedFilePath, errors.New("DirectoryCompressionMustUseTarFormat")
		}
	}

	targetPath := sourcePath + compressionSuffix
	if clerk.IsFile(targetPath) {
		return compressedFilePath, errors.New("TargetFileAlreadyExists")
	}
	switch compressionSuffix {
	case ".tar", ".zip":
		compressionArgs = append(compressionArgs, targetPath)
	}

	compressionArgs = append(compressionArgs, sourcePath)
	_, err = NewShell(
		ShellSettings{Command: compressionCmd, Args: compressionArgs},
	).Run()
	if err != nil {
		return compressedFilePath, err
	}

	if !clerk.IsFile(targetPath) {
		return compressedFilePath, errors.New("CompressedFileNotFound")
	}

	return targetPath, nil
}

func (clerk FileClerk) DecompressFile(
	sourcePath string,
	targetPathPtr *string,
	shouldKeepSourceFilePtr *bool,
) (decompressedFilePath string, err error) {
	if !clerk.IsFile(sourcePath) {
		return decompressedFilePath, errors.New("SourceFileNotFound")
	}

	decompressionCmd := "tar"
	decompressionArgs := []string{"--extract", "--file", sourcePath}
	if targetPathPtr != nil {
		decompressionArgs = append(decompressionArgs, "--directory", *targetPathPtr)
	}

	sourcePathExtStr := filepath.Ext(sourcePath)
	sourcePathExtNoDotStr := strings.TrimPrefix(sourcePathExtStr, ".")
	switch sourcePathExtNoDotStr {
	case "tar", "tarball":
	case "br", "brotli":
		decompressionCmd = "brotli"
		decompressionArgs = []string{"--decompress", "--keep", sourcePath}
		if targetPathPtr != nil {
			decompressionArgs = append(decompressionArgs, "--output", *targetPathPtr)
		}
	case "gz", "gzip":
		decompressionCmd = "gzip"
		decompressionArgs = []string{"--decompress", "--keep", "--quiet", sourcePath}
		if targetPathPtr != nil {
			decompressionArgs = append(decompressionArgs, "--stdout")
		}
	case "zip":
		decompressionCmd = "unzip"
		decompressionArgs = []string{"-q", "-o", sourcePath}
		if targetPathPtr != nil {
			decompressionArgs = append(decompressionArgs, "-d", *targetPathPtr)
		}
	case "xz":
		decompressionCmd = "xz"
		decompressionArgs = []string{"--decompress", "--keep", "--memlimit=10%", sourcePath}
		if targetPathPtr != nil {
			decompressionArgs = append(decompressionArgs, "--stdout")
		}
	default:
		return decompressedFilePath, errors.New("UnsupportedCompressionFormat")
	}

	shell := NewShell(
		ShellSettings{Command: decompressionCmd, Args: decompressionArgs},
	)
	switch sourcePathExtNoDotStr {
	case "gz", "gzip", "xz":
		if targetPathPtr != nil {
			shell.runtimeSettings.StdoutFilePath = *targetPathPtr
		}
	}

	_, err = shell.Run()
	if err != nil {
		return decompressedFilePath, err
	}

	shouldKeepSourceFile := false
	if shouldKeepSourceFilePtr != nil {
		shouldKeepSourceFile = *shouldKeepSourceFilePtr
	}
	if !shouldKeepSourceFile && clerk.FileExists(sourcePath) {
		err = os.Remove(sourcePath)
		if err != nil {
			return decompressedFilePath, err
		}
	}

	sourcePathNoExt := sourcePath[:len(sourcePath)-len(sourcePathExtStr)]
	targetPath := sourcePathNoExt
	if targetPathPtr != nil {
		targetPath = *targetPathPtr
	}

	return targetPath, nil
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

func (clerk FileClerk) CompressDir(
	sourcePath string,
	compressionFormatPtr *string,
) (compressedFilePath string, err error) {
	if !clerk.IsDir(sourcePath) {
		return compressedFilePath, errors.New("SourceDirNotFound")
	}

	tarCompressedFilePath, err := clerk.CompressFile(sourcePath, nil)
	if err != nil {
		return compressedFilePath, err
	}

	compressionFormat := "br"
	if compressionFormatPtr != nil {
		if *compressionFormatPtr == "tar" {
			return tarCompressedFilePath, nil
		}
		compressionFormat = *compressionFormatPtr
	}

	compressedFilePath, err = clerk.CompressFile(
		tarCompressedFilePath, &compressionFormat,
	)
	if err != nil {
		return compressedFilePath, err
	}

	err = os.Remove(tarCompressedFilePath)
	return compressedFilePath, err
}

func (clerk FileClerk) DecompressDir(
	sourcePath string,
	targetPathPtr *string,
	shouldKeepSourceFilePtr *bool,
) (decompressedDirPath string, err error) {
	sourcePathExt := filepath.Ext(decompressedDirPath)
	if sourcePathExt != ".tar" {
		sourcePath, err = clerk.DecompressFile(sourcePath, nil, shouldKeepSourceFilePtr)
		if err != nil {
			return decompressedDirPath, err
		}
		sourcePathExt := filepath.Ext(sourcePath)
		if sourcePathExt != ".tar" {
			return decompressedDirPath, errors.New("UnsupportedCompressionFormat")
		}
	}

	return clerk.DecompressFile(sourcePath, targetPathPtr, shouldKeepSourceFilePtr)
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
