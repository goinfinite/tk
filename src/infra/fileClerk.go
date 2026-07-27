package tkInfra

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	RegexLargeFileThresholdBytes int64 = 10 * 1024 * 1024
	regexTempFileSuffix                 = ".tmp"
)

var (
	ErrSourceFileMissing            = errors.New("SourceFileNotFound")
	ErrTargetFileExists             = errors.New("TargetFileAlreadyExists")
	ErrFileMissing                  = errors.New("FileNotFound")
	ErrFileEmpty                    = errors.New("FileEmpty")
	ErrReplacementWouldTruncateFile = errors.New("ReplacementWouldTruncateFile")
	ErrDirCompressionWrongFormat    = errors.New("DirectoryCompressionMustUseTarFormat")
	ErrCompressedFileMissing        = errors.New("CompressedFileNotFound")
	ErrSourceDirMissing             = errors.New("SourceDirNotFound")
	ErrTargetDirExists              = errors.New("TargetDirAlreadyExists")
	ErrSourcePathMissing            = errors.New("SourcePathNotFound")
	ErrSymlinkExists                = errors.New("SymlinkAlreadyExists")
	ErrTargetPathExists             = errors.New("TargetPathAlreadyExists")
	ErrRegexPatternMissing          = errors.New("RegexPatternCannotBeNil")
	ErrUnsupportedCompressionFormat = errors.New("UnsupportedCompressionFormat")
	ErrTargetIsDirectory            = errors.New("TargetIsDirectory")
	ErrSourceIsDirectory            = errors.New("SourceIsDirectory")
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

func (clerk FileClerk) CopyFile(sourcePath, targetPath string) (methodErr error) {
	if !clerk.IsFile(sourcePath) {
		return ErrSourceFileMissing
	}

	if clerk.IsFile(targetPath) {
		return ErrTargetFileExists
	}

	sourceFile, openErr := os.Open(sourcePath)
	if openErr != nil {
		return openErr
	}
	defer sourceFile.Close()

	targetFile, createErr := os.Create(targetPath)
	if createErr != nil {
		return createErr
	}
	defer func() {
		if methodErr == nil {
			methodErr = targetFile.Close()
		}
	}()

	bufferReader := bufio.NewReader(sourceFile)
	bufferWriter := bufio.NewWriter(targetFile)

	_, readFromErr := bufferWriter.ReadFrom(bufferReader)
	if readFromErr != nil {
		return readFromErr
	}

	return bufferWriter.Flush()
}

func (clerk FileClerk) MoveFile(sourcePath, targetPath string) error {
	if !clerk.IsFile(sourcePath) {
		return ErrSourceFileMissing
	}

	if clerk.IsFile(targetPath) {
		return ErrTargetFileExists
	}

	return os.Rename(sourcePath, targetPath)
}

func (clerk FileClerk) RenameFile(sourcePath, targetPath string) error {
	return clerk.MoveFile(sourcePath, targetPath)
}

func (clerk FileClerk) UpdateFileContent(
	filePath, newContent string,
	shouldOverwrite bool,
) (methodErr error) {
	fileFlags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if shouldOverwrite {
		fileFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	fileHandler, openErr := os.OpenFile(filePath, fileFlags, 0644)
	if openErr != nil {
		return openErr
	}
	defer func() {
		if methodErr == nil {
			methodErr = fileHandler.Close()
		}
	}()

	bufferWriter := bufio.NewWriter(fileHandler)
	_, writeErr := bufferWriter.WriteString(newContent)
	if writeErr != nil {
		return writeErr
	}

	return bufferWriter.Flush()
}

// OverwriteFile atomically replaces targetPath's underlying file with
// sourcePath's content. If targetPath is a symlink, the chain is resolved
// via filepath.EvalSymlinks so the rename modifies the underlying file's
// content rather than replacing the symlink path entry (which would leave
// the original target untouched and orphan any other references to it).
// EvSymlinks also handles relative symlink targets and walks multi-level
// chains, neither of which os.Readlink does on its own.
func (clerk FileClerk) OverwriteFile(sourcePath, targetPath string) error {
	sourceInfo, statErr := os.Stat(sourcePath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return ErrSourceFileMissing
		}
		return statErr
	}
	if sourceInfo.IsDir() {
		return ErrSourceIsDirectory
	}

	actualFilePath := targetPath
	if clerk.IsSymlink(targetPath) {
		resolvedPath, evalErr := filepath.EvalSymlinks(targetPath)
		if evalErr != nil {
			return evalErr
		}
		actualFilePath = resolvedPath
	}

	return os.Rename(sourcePath, actualFilePath)
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
		return "", ErrFileMissing
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

// FileContentRegexFindings holds one regex match plus its 1-based line
// range and capture groups. LineNumRange is [start, end] inclusive;
// a single-line match has start == end.
type FileContentRegexFindings struct {
	Match        string
	Groups       []string
	LineNumRange []int
}

func (clerk FileClerk) regexSearchWholeFile(
	filePathStr string,
	regexPattern *regexp.Regexp,
) (regexSearchFindings []FileContentRegexFindings, err error) {
	fileContent, readErr := clerk.ReadFileContent(filePathStr, nil)
	if readErr != nil {
		return regexSearchFindings, readErr
	}

	matchesWithGroups := regexPattern.FindAllStringSubmatch(fileContent, -1)
	matchByteRanges := regexPattern.FindAllStringIndex(fileContent, -1)
	if len(matchesWithGroups) != len(matchByteRanges) {
		slog.Error(
			"RegexMatchCountMismatch",
			slog.String("filePath", filePathStr),
			slog.Int("submatches", len(matchesWithGroups)),
			slog.Int("byteRanges", len(matchByteRanges)),
		)
		return regexSearchFindings, errors.New("RegexMatchCountMismatch")
	}
	regexSearchFindings = make([]FileContentRegexFindings, 0, len(matchesWithGroups))

	for matchIdx, matchedGroups := range matchesWithGroups {
		matchByteRange := matchByteRanges[matchIdx]
		matchStart := matchByteRange[0]
		matchEnd := matchByteRange[1]

		expectedMatchLength := matchEnd - matchStart
		actualMatchLength := len(matchedGroups[0])
		if actualMatchLength != expectedMatchLength {
			slog.Error(
				"RegexMatchRangeLengthMismatch",
				slog.String("filePath", filePathStr),
				slog.Int("matchIdx", matchIdx),
				slog.Int("matchTextLen", actualMatchLength),
				slog.Int("byteRangeSpan", expectedMatchLength),
			)
			return regexSearchFindings, errors.New("RegexMatchRangeLengthMismatch")
		}

		startLineNumber := strings.Count(fileContent[:matchStart], "\n") + 1
		endLineNumber := strings.Count(fileContent[:matchEnd], "\n") + 1

		regexSearchFindings = append(regexSearchFindings, FileContentRegexFindings{
			Match:        matchedGroups[0],
			Groups:       matchedGroups[1:],
			LineNumRange: []int{startLineNumber, endLineNumber},
		})
	}

	return regexSearchFindings, nil
}

func (clerk FileClerk) regexSearchStreaming(
	filePathStr string,
	regexPattern *regexp.Regexp,
) (regexSearchFindings []FileContentRegexFindings, err error) {
	fileHandler, osOpenErr := os.Open(filePathStr)
	if osOpenErr != nil {
		if os.IsNotExist(osOpenErr) {
			return regexSearchFindings, ErrFileMissing
		}
		return regexSearchFindings, osOpenErr
	}
	defer fileHandler.Close()

	fileScanner := bufio.NewScanner(fileHandler)
	for currentLineNumber := 1; fileScanner.Scan(); currentLineNumber++ {
		for _, lineMatchAndGroups := range regexPattern.FindAllStringSubmatch(
			fileScanner.Text(), -1,
		) {
			fullMatch := lineMatchAndGroups[0]
			captureGroups := lineMatchAndGroups[1:]

			regexSearchFindings = append(regexSearchFindings, FileContentRegexFindings{
				Match:        fullMatch,
				Groups:       captureGroups,
				LineNumRange: []int{currentLineNumber, currentLineNumber},
			})
		}
	}

	return regexSearchFindings, fileScanner.Err()
}

// FileContentRegexSearch finds every regex match in a file and returns
// each match's 1-based inclusive line range plus capture groups.
//
// Files below RegexLargeFileThresholdBytes are matched in a single
// regex pass; patterns with per-line anchors (^ / $) need the (?m) flag.
// Multi-line matches get a proper LineNumRange spanning every line they touch.
//
// Larger files fall back to bufio.Scanner streaming, which splits on
// newlines — multi-line patterns only match within a single line and
// LineNumRange is always [n, n].
func (clerk FileClerk) FileContentRegexSearch(
	filePath tkValueObject.UnixAbsoluteFilePath,
	regexPattern *regexp.Regexp,
) (regexSearchFindings []FileContentRegexFindings, err error) {
	if regexPattern == nil {
		return regexSearchFindings, ErrRegexPatternMissing
	}

	filePathStr := filePath.String()
	fileInfo, statErr := os.Stat(filePathStr)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return regexSearchFindings, ErrFileMissing
		}
		return regexSearchFindings, statErr
	}
	if fileInfo.IsDir() {
		return regexSearchFindings, ErrTargetIsDirectory
	}

	if fileInfo.Size() >= RegexLargeFileThresholdBytes {
		slog.Warn(
			"FileContentRegexSearchStreamingFallback",
			slog.String("filePath", filePathStr),
			slog.Int64("fileSizeBytes", fileInfo.Size()),
			slog.Int64("thresholdBytes", RegexLargeFileThresholdBytes),
			slog.String("reason", "FileSizeExceedsThreshold"),
		)
		return clerk.regexSearchStreaming(filePathStr, regexPattern)
	}
	return clerk.regexSearchWholeFile(filePathStr, regexPattern)
}

func (clerk FileClerk) regexReplaceWholeFile(
	filePathStr string,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, methodErr error) {
	fileContent, readErr := clerk.ReadFileContent(filePathStr, nil)
	if readErr != nil {
		return 0, readErr
	}

	unprocessedMatches := regexPattern.FindAllString(fileContent, -1)
	replacementCount = len(unprocessedMatches)
	replacedContent := regexPattern.ReplaceAllString(fileContent, replacement)

	if len(replacedContent) == 0 {
		return 0, ErrReplacementWouldTruncateFile
	}

	existingFilePermissions := os.FileMode(0644)
	existingFileInfo, statErr := os.Stat(filePathStr)
	if statErr == nil {
		existingFilePermissions = existingFileInfo.Mode().Perm()
	}

	tempFileHandle, createErr := os.CreateTemp(
		filepath.Dir(filePathStr),
		filepath.Base(filePathStr)+regexTempFileSuffix,
	)
	if createErr != nil {
		return 0, createErr
	}
	tempFilePathStr := tempFileHandle.Name()
	defer func() {
		_ = os.Remove(tempFilePathStr)
		if methodErr == nil {
			methodErr = tempFileHandle.Close()
		}
	}()

	_, writeErr := tempFileHandle.WriteString(replacedContent)
	if writeErr != nil {
		return 0, writeErr
	}

	if existingFilePermissions != 0600 {
		chmodErr := os.Chmod(tempFilePathStr, existingFilePermissions)
		if chmodErr != nil {
			return 0, chmodErr
		}
	}

	renameErr := clerk.OverwriteFile(tempFilePathStr, filePathStr)
	if renameErr != nil {
		return 0, renameErr
	}

	return replacementCount, nil
}

func (clerk FileClerk) regexReplaceStreaming(
	filePathStr string,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, methodErr error) {
	fileHandler, osOpenErr := os.Open(filePathStr)
	if osOpenErr != nil {
		if os.IsNotExist(osOpenErr) {
			return 0, ErrFileMissing
		}
		return 0, osOpenErr
	}
	defer fileHandler.Close()

	existingFilePermissions := os.FileMode(0644)
	existingFileInfo, statErr := os.Stat(filePathStr)
	if statErr == nil {
		existingFilePermissions = existingFileInfo.Mode().Perm()
	}

	tempFileHandle, createErr := os.CreateTemp(
		filepath.Dir(filePathStr),
		filepath.Base(filePathStr)+regexTempFileSuffix,
	)
	if createErr != nil {
		return 0, createErr
	}
	tempFilePathStr := tempFileHandle.Name()
	defer func() {
		_ = os.Remove(tempFilePathStr)
		if methodErr == nil {
			methodErr = tempFileHandle.Close()
		}
	}()

	bufferWriter := bufio.NewWriter(tempFileHandle)
	bufferedReader := bufio.NewReader(fileHandler)

	for {
		rawLine, readErr := bufferedReader.ReadString('\n')
		if readErr == io.EOF && len(rawLine) == 0 {
			break
		}

		lineTerminator := ""
		switch {
		case strings.HasSuffix(rawLine, "\r\n"):
			lineTerminator = "\r\n"
		case strings.HasSuffix(rawLine, "\n"):
			lineTerminator = "\n"
		}
		lineContent := rawLine[:len(rawLine)-len(lineTerminator)]

		replacedLine := regexPattern.ReplaceAllString(lineContent, replacement)
		replacementCount += len(regexPattern.FindAllString(lineContent, -1))
		_, writeErr := bufferWriter.WriteString(replacedLine + lineTerminator)
		if writeErr != nil {
			return 0, writeErr
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return 0, readErr
		}
	}

	flushErr := bufferWriter.Flush()
	if flushErr != nil {
		return 0, flushErr
	}

	if existingFilePermissions != 0600 {
		chmodErr := os.Chmod(tempFilePathStr, existingFilePermissions)
		if chmodErr != nil {
			return 0, chmodErr
		}
	}

	tempFileInfo, statErr := os.Stat(tempFilePathStr)
	if statErr != nil {
		return 0, statErr
	}
	if tempFileInfo.Size() == 0 {
		return 0, ErrReplacementWouldTruncateFile
	}

	renameErr := clerk.OverwriteFile(tempFilePathStr, filePathStr)
	if renameErr != nil {
		return 0, renameErr
	}

	return replacementCount, nil
}

// FileContentRegexReplace atomically substitutes regex matches in a file's
// content. Empty source files and directories are rejected; symlinks are
// followed. Replacing everything with "" is rejected as a call-site bug
// because it would silently truncate the source. Use TruncateFileContent to
// intentionally empty a file or strip only non-whitespace content.
//
// Files at or above RegexLargeFileThresholdBytes are processed line-by-line,
// so multi-line patterns only match in smaller files. Original line
// terminators are preserved unchanged.
func (clerk FileClerk) FileContentRegexReplace(
	filePath tkValueObject.UnixAbsoluteFilePath,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	if regexPattern == nil {
		return 0, ErrRegexPatternMissing
	}

	filePathStr := filePath.String()
	fileInfo, statErr := os.Stat(filePathStr)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return 0, ErrFileMissing
		}
		return 0, statErr
	}
	if fileInfo.IsDir() {
		return 0, ErrTargetIsDirectory
	}
	originalFileSize := fileInfo.Size()

	if originalFileSize == 0 {
		return 0, ErrFileEmpty
	}

	if originalFileSize >= RegexLargeFileThresholdBytes {
		slog.Warn(
			"FileContentRegexReplaceStreamingFallback",
			slog.String("filePath", filePathStr),
			slog.Int64("fileSizeBytes", originalFileSize),
			slog.Int64("thresholdBytes", RegexLargeFileThresholdBytes),
			slog.String("reason", "FileSizeExceedsThreshold"),
		)
		return clerk.regexReplaceStreaming(filePathStr, regexPattern, replacement)
	}
	return clerk.regexReplaceWholeFile(filePathStr, regexPattern, replacement)
}

func (clerk FileClerk) DeleteFileContent(filePath string) error {
	return clerk.TruncateFileContent(filePath)
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
			return compressedFilePath, ErrUnsupportedCompressionFormat
		}
	}

	if !clerk.IsFile(sourcePath) {
		if !clerk.IsDir(sourcePath) {
			return compressedFilePath, ErrSourceFileMissing
		}
		if compressionSuffix != ".tar" {
			return compressedFilePath, ErrDirCompressionWrongFormat
		}
	}

	targetPath := sourcePath + compressionSuffix
	if clerk.IsFile(targetPath) {
		return compressedFilePath, ErrTargetFileExists
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
		return compressedFilePath, ErrCompressedFileMissing
	}

	return targetPath, nil
}

func (clerk FileClerk) DecompressFile(
	sourcePath string,
	targetPathPtr *string,
	shouldKeepSourceFilePtr *bool,
) (decompressedFilePath string, err error) {
	if !clerk.IsFile(sourcePath) {
		return decompressedFilePath, ErrSourceFileMissing
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
		decompressionArgs = []string{
			"--decompress", "--keep", "--memlimit=10%", sourcePath,
		}
		if targetPathPtr != nil {
			decompressionArgs = append(decompressionArgs, "--stdout")
		}
	default:
		return decompressedFilePath, ErrUnsupportedCompressionFormat
	}

	shell := NewShell(
		ShellSettings{
			Command:          decompressionCmd,
			Args:             decompressionArgs,
			WorkingDirectory: filepath.Dir(sourcePath),
		},
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
		return ErrSourceDirMissing
	}

	if clerk.IsDir(targetPath) {
		return ErrTargetDirExists
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
		return ErrSourceDirMissing
	}

	if clerk.IsDir(targetPath) {
		return ErrTargetDirExists
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
		return compressedFilePath, ErrSourceDirMissing
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
	sourcePathExt := filepath.Ext(sourcePath)
	if sourcePathExt != ".tar" {
		sourcePath, err = clerk.DecompressFile(sourcePath, nil, shouldKeepSourceFilePtr)
		if err != nil {
			return decompressedDirPath, err
		}
		sourcePathExt := filepath.Ext(sourcePath)
		if sourcePathExt != ".tar" {
			return decompressedDirPath, ErrUnsupportedCompressionFormat
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
	if !clerk.FileExists(sourcePath) {
		return ErrSourcePathMissing
	}

	if !shouldOverwrite {
		if clerk.IsSymlink(targetPath) {
			return ErrSymlinkExists
		}
		if clerk.IsFile(targetPath) || clerk.IsDir(targetPath) {
			return ErrTargetPathExists
		}
	}

	if clerk.FileExists(targetPath) {
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
