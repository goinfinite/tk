package tkInfra

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	RegexLargeFileThresholdBytes       int64 = 10 * 1024 * 1024
	ReadFileContentDefaultMaxSizeBytes int64 = 500 * 1024 * 1024
	tempFileNameSuffix                       = ".tk-tmp"
	tempFileNameEntropyChars                 = 8
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
	ErrFileTooLarge                 = errors.New("FileTooLarge")
	ErrTargetIsSymlink              = errors.New("TargetIsSymlink")
	ErrTargetNotDirectory           = errors.New("TargetNotDirectory")
	ErrDirPathTraversalInvalid      = errors.New("DirPathParentTraversalNotAllowed")
	ErrSymlinkedPathInvalid         = errors.New("PathComponentIsSymlink")
	ErrDirectoryOwnerInvalid        = errors.New("DirectoryNotOwnedByExpectedOwner")
	ErrFilePermissionsInvalid       = errors.New("FilePermissionsUnset")
	ErrFileNameInvalid              = errors.New("FilePathFinalComponentIsNotAFileName")
	ErrTempFileNameTooLong          = errors.New("TempFileNameExceedsNameMax")
)

type FileClerk struct{}

func (FileClerk) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func (FileClerk) IsSymlink(sourcePath string) bool {
	linkInfo, err := os.Lstat(sourcePath)
	if err != nil {
		return false
	}

	isSymlink := linkInfo.Mode()&os.ModeSymlink == os.ModeSymlink
	return isSymlink
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

// TouchFile refreshes timestamps or creates the file, like touch(1) — but a
// dangling symlink at the path fails with ErrTargetIsSymlink instead of
// creating the file behind the link.
func (FileClerk) TouchFile(filePath string) error {
	timestampNow := time.Now()

	chtimesErr := os.Chtimes(filePath, timestampNow, timestampNow)
	if chtimesErr == nil {
		return nil
	}
	if !os.IsNotExist(chtimesErr) {
		return chtimesErr
	}

	fileHandler, openErr := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0666,
	)
	if openErr == nil {
		return fileHandler.Close()
	}
	if !os.IsExist(openErr) {
		return openErr
	}

	_, statErr := os.Stat(filePath)
	fileAlreadyCreated := statErr == nil
	if fileAlreadyCreated {
		return nil
	}

	return ErrTargetIsSymlink
}

func (FileClerk) openNewFileHandler(
	filePath string,
	permissions os.FileMode,
) (*os.File, error) {
	fileHandler, openErr := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		permissions,
	)
	if openErr != nil {
		if os.IsExist(openErr) {
			return nil, ErrTargetFileExists
		}
		return nil, openErr
	}

	chmodErr := fileHandler.Chmod(permissions)
	if chmodErr != nil {
		closeErr := fileHandler.Close()
		removeErr := os.Remove(filePath)
		return nil, errors.Join(chmodErr, closeErr, removeErr)
	}

	return fileHandler, nil
}

func (clerk FileClerk) WriteNewFile(
	filePath, content string,
	permissions os.FileMode,
) error {
	fileHandler, createErr := clerk.openNewFileHandler(filePath, permissions)
	if createErr != nil {
		return createErr
	}

	_, writeErr := fileHandler.WriteString(content)
	closeErr := fileHandler.Close()
	if writeErr == nil {
		return closeErr
	}

	removeErr := os.Remove(filePath)
	return errors.Join(writeErr, closeErr, removeErr)
}

func (clerk FileClerk) CopyFile(sourcePath, targetPath string) error {
	if !clerk.IsFile(sourcePath) {
		return ErrSourceFileMissing
	}

	sourceFile, openErr := os.Open(sourcePath)
	if openErr != nil {
		return openErr
	}
	defer func() { _ = sourceFile.Close() }()

	sourceInfo, statErr := sourceFile.Stat()
	if statErr != nil {
		return statErr
	}

	targetFile, createErr := clerk.openNewFileHandler(
		targetPath, sourceInfo.Mode().Perm(),
	)
	if createErr != nil {
		return createErr
	}
	defer func() { _ = targetFile.Close() }()

	bufferReader := bufio.NewReader(sourceFile)
	bufferWriter := bufio.NewWriter(targetFile)

	_, copyErr := bufferWriter.ReadFrom(bufferReader)
	if copyErr == nil {
		copyErr = bufferWriter.Flush()
	}
	if copyErr != nil {
		closeErr := targetFile.Close()
		removeErr := os.Remove(targetPath)
		return errors.Join(copyErr, closeErr, removeErr)
	}

	return targetFile.Close()
}

func (FileClerk) isTargetSource(sourcePath, targetPath string) bool {
	sourceInfo, sourceErr := os.Stat(sourcePath)
	targetInfo, targetErr := os.Stat(targetPath)
	return sourceErr == nil && targetErr == nil && os.SameFile(sourceInfo, targetInfo)
}

// MoveFile never replaces an existing target. Cross-device moves fall back
// to copy+delete: the two files briefly coexist, and an interrupted run
// leaves a partial target next to an untouched source.
func (clerk FileClerk) MoveFile(sourcePath, targetPath string) error {
	if !clerk.IsFile(sourcePath) {
		return ErrSourceFileMissing
	}

	moveErr := unix.Renameat2(
		unix.AT_FDCWD,
		sourcePath,
		unix.AT_FDCWD,
		targetPath,
		unix.RENAME_NOREPLACE,
	)
	if moveErr == nil {
		return nil
	}
	moveCrossesDevices := errors.Is(moveErr, unix.EXDEV)
	if moveCrossesDevices {
		copyErr := clerk.CopyFile(sourcePath, targetPath)
		if copyErr != nil {
			return copyErr
		}

		return os.Remove(sourcePath)
	}
	if !errors.Is(moveErr, unix.EEXIST) {
		return moveErr
	}
	if clerk.isTargetSource(sourcePath, targetPath) {
		return nil
	}

	return ErrTargetFileExists
}

// OverwriteFile atomically replaces targetPath's underlying file with
// sourcePath's content. Symlink targets are written through, not replaced:
// the original file and every other reference to it stay in place.
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

	if clerk.IsDir(actualFilePath) {
		return ErrTargetIsDirectory
	}

	return os.Rename(sourcePath, actualFilePath)
}

func (FileClerk) DeleteFile(filePath string) error {
	fileInfo, lstatErr := os.Lstat(filePath)
	if lstatErr != nil {
		if os.IsNotExist(lstatErr) {
			return nil
		}
		return lstatErr
	}

	if fileInfo.IsDir() {
		return ErrTargetIsDirectory
	}

	return os.Remove(filePath)
}

func (clerk FileClerk) ReadFileContent(
	filePath string,
	maxContentSizeBytesPtr *int64,
) (fileContent string, err error) {
	fileHandler, openErr := os.Open(filePath)
	if openErr != nil {
		if os.IsNotExist(openErr) {
			return fileContent, ErrFileMissing
		}
		return fileContent, openErr
	}
	defer func() { _ = fileHandler.Close() }()

	maxContentSizeBytes := ReadFileContentDefaultMaxSizeBytes
	if maxContentSizeBytesPtr != nil {
		maxContentSizeBytes = *maxContentSizeBytesPtr
	}
	if maxContentSizeBytes < 0 {
		maxContentSizeBytes = 0
	}

	limitedReader := io.LimitedReader{R: fileHandler, N: maxContentSizeBytes}
	fileContentBytes, err := io.ReadAll(&limitedReader)
	if err != nil {
		return fileContent, err
	}

	sizeCapReached := limitedReader.N == 0
	if !sizeCapReached {
		return string(fileContentBytes), nil
	}

	trailingByte := make([]byte, 1)
	trailingCount, trailingErr := fileHandler.Read(trailingByte)
	if trailingCount > 0 {
		return fileContent, ErrFileTooLarge
	}
	if trailingErr != nil && !errors.Is(trailingErr, io.EOF) {
		return fileContent, trailingErr
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

	// DO NOT REPLACE: FindAllStringSubmatchIndex would halve the regex
	// scan, but extracting match text and capture groups from raw index
	// pairs requires manual slice arithmetic that makes the code
	// unreadable. The gain is negligible (~ms on 10MiB files).
	// Readability over performance.
	matchesWithGroups := regexPattern.FindAllStringSubmatch(fileContent, -1)
	matchByteRanges := regexPattern.FindAllStringIndex(fileContent, -1)
	if len(matchesWithGroups) != len(matchByteRanges) {
		return regexSearchFindings, fmt.Errorf(
			"RegexMatchCountMismatch: submatches %d != byteRanges %d",
			len(matchesWithGroups), len(matchByteRanges),
		)
	}
	regexSearchFindings = make([]FileContentRegexFindings, 0, len(matchesWithGroups))

	for matchIdx, matchedGroups := range matchesWithGroups {
		matchByteRange := matchByteRanges[matchIdx]
		matchStart := matchByteRange[0]
		matchEnd := matchByteRange[1]

		expectedMatchLength := matchEnd - matchStart
		actualMatchLength := len(matchedGroups[0])
		if actualMatchLength != expectedMatchLength {
			return regexSearchFindings, fmt.Errorf(
				"RegexMatchRangeLengthMismatch: matchIdx %d: %d != %d",
				matchIdx, actualMatchLength, expectedMatchLength,
			)
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
	defer func() { _ = fileHandler.Close() }()

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

func (clerk FileClerk) regexTargetFileInspector(
	filePath tkValueObject.UnixAbsoluteFilePath,
) (os.FileInfo, error) {
	filePathStr := filePath.String()
	fileInfo, statErr := os.Stat(filePathStr)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, ErrFileMissing
		}
		return nil, statErr
	}
	if fileInfo.IsDir() {
		return nil, ErrTargetIsDirectory
	}

	return fileInfo, nil
}

// FileContentRegexSearch finds every regex match in a file, returning each
// match's 1-based inclusive line range and capture groups. Files under
// RegexLargeFileThresholdBytes are matched in one pass — per-line anchors
// need the (?m) flag — and multi-line matches span every line they touch.
// Larger files stream line-by-line: multi-line patterns then match within a
// single line only, and LineNumRange is always [n, n].
func (clerk FileClerk) FileContentRegexSearch(
	filePath tkValueObject.UnixAbsoluteFilePath,
	regexPattern *regexp.Regexp,
) (regexSearchFindings []FileContentRegexFindings, err error) {
	if regexPattern == nil {
		return regexSearchFindings, ErrRegexPatternMissing
	}

	filePathStr := filePath.String()
	fileInfo, inspectErr := clerk.regexTargetFileInspector(filePath)
	if inspectErr != nil {
		return regexSearchFindings, inspectErr
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

func (FileClerk) TempFileNameFactory(
	targetFileName tkValueObject.UnixFileName,
) string {
	entropy := (&Synthesizer{}).randomStringFactory(
		tempFileNameEntropyChars,
		CharsetLowercaseLetters+CharsetNumbers,
	)
	return "." + targetFileName.String() + "." + entropy + tempFileNameSuffix
}

func (clerk FileClerk) TempFilePathFactory(
	targetFilePath tkValueObject.UnixAbsoluteFilePath,
) (string, error) {
	targetFileName, fileNameErr := targetFilePath.ReadFileName(true)
	if fileNameErr != nil {
		return "", fileNameErr
	}

	tempFileName := clerk.TempFileNameFactory(targetFileName)
	return filepath.Join(targetFilePath.ReadFileDir().String(), tempFileName), nil
}

type stagedFileContentWriter func(io.Writer) error

type atomicFileWriteSettings struct {
	DirHandle       int
	TargetFileName  tkValueObject.UnixFileName
	Permissions     os.FileMode
	ShouldOverwrite bool
	OwnerUserId     *tkValueObject.UnixUserId
	OwnerGroupId    *tkValueObject.UnixGroupId
}

func (clerk FileClerk) writeFileAtomically(
	settings atomicFileWriteSettings,
	writeContent stagedFileContentWriter,
) error {
	tempFileName := clerk.TempFileNameFactory(settings.TargetFileName)
	tempNameFitsKernelLimit := len(tempFileName) <= unix.NAME_MAX
	if !tempNameFitsKernelLimit {
		return ErrTempFileNameTooLong
	}

	const privateTempFileMode uint32 = 0o600
	tempFileHandle, createErr := unix.Openat(
		settings.DirHandle,
		tempFileName,
		unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC,
		privateTempFileMode,
	)
	if createErr != nil {
		return createErr
	}
	tempFile := os.NewFile(uintptr(tempFileHandle), tempFileName)

	bufferWriter := bufio.NewWriter(tempFile)
	writeErr := writeContent(bufferWriter)
	var flushErr error
	if writeErr == nil {
		flushErr = bufferWriter.Flush()
	}

	var chownErr error
	if settings.OwnerUserId != nil && settings.OwnerGroupId != nil {
		chownErr = tempFile.Chown(
			int(settings.OwnerUserId.Uint64()),
			int(settings.OwnerGroupId.Uint64()),
		)
	}
	// Chmod last: chown clears the setuid and setgid bits a mode may carry.
	chmodErr := tempFile.Chmod(settings.Permissions)
	closeErr := tempFile.Close()

	writeFailure := errors.Join(writeErr, flushErr, chownErr, chmodErr, closeErr)
	if writeFailure != nil {
		removeErr := unix.Unlinkat(settings.DirHandle, tempFileName, 0)
		return errors.Join(writeFailure, removeErr)
	}

	renameFlags := uint(unix.RENAME_NOREPLACE)
	if settings.ShouldOverwrite {
		renameFlags = 0
	}
	swapErr := unix.Renameat2(
		settings.DirHandle, tempFileName,
		settings.DirHandle, settings.TargetFileName.String(),
		renameFlags,
	)
	if errors.Is(swapErr, unix.EEXIST) {
		swapErr = ErrTargetFileExists
	}
	if swapErr != nil {
		removeErr := unix.Unlinkat(settings.DirHandle, tempFileName, 0)
		return errors.Join(swapErr, removeErr)
	}

	return nil
}

func (clerk FileClerk) atomicReplacementSettingsResolver(
	filePathStr string,
) (atomicFileWriteSettings, error) {
	actualFilePath := filePathStr
	if clerk.IsSymlink(filePathStr) {
		resolvedFilePath, evalErr := filepath.EvalSymlinks(filePathStr)
		if evalErr != nil {
			return atomicFileWriteSettings{}, evalErr
		}
		actualFilePath = resolvedFilePath
	}

	fileInfo, statErr := os.Stat(actualFilePath)
	if statErr != nil {
		return atomicFileWriteSettings{}, statErr
	}
	existingFilePermissions := fileInfo.Mode().Perm()

	dirPath, rawTargetFileName := filepath.Split(actualFilePath)
	targetFileName, fileNameErr := tkValueObject.NewUnixFileName(
		rawTargetFileName, true,
	)
	if fileNameErr != nil {
		return atomicFileWriteSettings{}, fileNameErr
	}

	dirHandle, openErr := unix.Open(dirPath, unix.O_PATH|unix.O_CLOEXEC, 0)
	if openErr != nil {
		return atomicFileWriteSettings{}, openErr
	}

	return atomicFileWriteSettings{
		DirHandle:       dirHandle,
		TargetFileName:  targetFileName,
		Permissions:     existingFilePermissions,
		ShouldOverwrite: true,
	}, nil
}

func (clerk FileClerk) regexReplaceWholeFile(
	filePathStr string,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	fileContent, readErr := clerk.ReadFileContent(filePathStr, nil)
	if readErr != nil {
		return 0, readErr
	}

	foundMatches := regexPattern.FindAllString(fileContent, -1)
	replacementCount = len(foundMatches)
	replacedContent := regexPattern.ReplaceAllString(fileContent, replacement)

	if len(replacedContent) == 0 {
		return 0, ErrReplacementWouldTruncateFile
	}

	settings, settingsErr := clerk.atomicReplacementSettingsResolver(filePathStr)
	if settingsErr != nil {
		return 0, settingsErr
	}
	defer func() { _ = unix.Close(settings.DirHandle) }()

	writeErr := clerk.writeFileAtomically(
		settings,
		func(writer io.Writer) error {
			_, writeErr := io.WriteString(writer, replacedContent)
			return writeErr
		},
	)
	if writeErr != nil {
		return 0, writeErr
	}

	return replacementCount, nil
}

func (clerk FileClerk) writeRegexReplacedLines(
	writer io.Writer,
	bufferReader *bufio.Reader,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, writtenBytesTotal int, err error) {
	for {
		rawLine, readErr := bufferReader.ReadString('\n')
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
		writtenCount, writeErr := writer.Write(
			[]byte(replacedLine + lineTerminator),
		)
		writtenBytesTotal += writtenCount
		if writeErr != nil {
			return replacementCount, writtenBytesTotal, writeErr
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return replacementCount, writtenBytesTotal, readErr
		}
	}

	if writtenBytesTotal == 0 {
		return replacementCount, writtenBytesTotal, ErrReplacementWouldTruncateFile
	}

	return replacementCount, writtenBytesTotal, nil
}

func (clerk FileClerk) regexReplaceStreaming(
	filePathStr string,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	fileHandler, osOpenErr := os.Open(filePathStr)
	if osOpenErr != nil {
		if os.IsNotExist(osOpenErr) {
			return 0, ErrFileMissing
		}
		return 0, osOpenErr
	}
	defer func() { _ = fileHandler.Close() }()

	settings, settingsErr := clerk.atomicReplacementSettingsResolver(filePathStr)
	if settingsErr != nil {
		return 0, settingsErr
	}
	defer func() { _ = unix.Close(settings.DirHandle) }()

	bufferReader := bufio.NewReader(fileHandler)
	writeErr := clerk.writeFileAtomically(
		settings,
		func(writer io.Writer) error {
			var replaceErr error
			replacementCount, _, replaceErr = clerk.writeRegexReplacedLines(
				writer, bufferReader, regexPattern, replacement,
			)
			return replaceErr
		},
	)
	if writeErr != nil {
		return 0, writeErr
	}

	return replacementCount, nil
}

// FileContentRegexReplace atomically substitutes regex matches in a file.
// Empty files and directories are rejected; symlinks are followed. A result
// with zero bytes fails as a call-site bug — use TruncateFileContent to
// empty a file on purpose; a file whose every line matches keeps its line
// terminators. Files at or above RegexLargeFileThresholdBytes are processed
// line-by-line, so multi-line patterns only match in smaller files.
func (clerk FileClerk) FileContentRegexReplace(
	filePath tkValueObject.UnixAbsoluteFilePath,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	if regexPattern == nil {
		return 0, ErrRegexPatternMissing
	}

	filePathStr := filePath.String()
	fileInfo, inspectErr := clerk.regexTargetFileInspector(filePath)
	if inspectErr != nil {
		return 0, inspectErr
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

func (FileClerk) TruncateFileContent(
	filePath tkValueObject.UnixAbsoluteFilePath,
) error {
	return os.Truncate(filePath.String(), 0)
}

func (clerk FileClerk) UpdateFileOwnership(
	filePath string,
	userId, groupId int,
) error {
	return os.Lchown(filePath, userId, groupId)
}

// UpdateFilePermissions never chmods through a symlink. Unlike chmod(2), it
// requires read permission on the target; root is exempt.
func (FileClerk) UpdateFilePermissions(
	filePath string,
	permissionsPtr *os.FileMode,
) error {
	fileHandler, openErr := os.OpenFile(filePath, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if openErr != nil {
		if errors.Is(openErr, syscall.ELOOP) {
			return ErrTargetIsSymlink
		}
		return openErr
	}
	defer func() { _ = fileHandler.Close() }()

	fileInfo, statErr := fileHandler.Stat()
	if statErr != nil {
		return statErr
	}

	defaultFilePermission := os.FileMode(0644)
	if fileInfo.IsDir() {
		defaultFilePermission = 0755
	}

	if permissionsPtr == nil {
		permissionsPtr = &defaultFilePermission
	}

	return fileHandler.Chmod(*permissionsPtr)
}

func (clerk FileClerk) CompressFile(
	sourcePath string,
	compressionFormatPtr *string,
	shouldKeepSourceFilePtr *bool,
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
			compressionArgs = []string{"--quality=4", "--keep"}
		case "gz", "gzip":
			compressionSuffix = ".gz"
			compressionCmd = "gzip"
			compressionArgs = []string{"-6", "--keep"}
		case "zip":
			compressionSuffix = ".zip"
			compressionCmd = "zip"
			compressionArgs = []string{"-6", "--quiet", "--test"}
		case "xz":
			compressionSuffix = ".xz"
			compressionCmd = "xz"
			compressionArgs = []string{"-1", "--keep", "--memlimit=10%"}
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

	shouldKeepSourceFile := false
	if shouldKeepSourceFilePtr != nil {
		shouldKeepSourceFile = *shouldKeepSourceFilePtr
	}
	if !shouldKeepSourceFile && clerk.IsFile(sourcePath) {
		removeErr := os.Remove(sourcePath)
		if removeErr != nil && !os.IsNotExist(removeErr) {
			return targetPath, removeErr
		}
	}

	return targetPath, nil
}

// DecompressFile expands sourcePath (optionally into targetPathPtr) and
// removes the archive unless shouldKeepSourceFilePtr requests otherwise.
// Zip extraction overwrites existing files at the destination without
// warning: only decompress archives you trust.
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

func (FileClerk) DeleteDir(dirPath string) error {
	dirInfo, lstatErr := os.Lstat(dirPath)
	if lstatErr != nil {
		if os.IsNotExist(lstatErr) {
			return nil
		}
		return lstatErr
	}

	if !dirInfo.IsDir() {
		return ErrTargetNotDirectory
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

	tarCompressedFilePath, err := clerk.CompressFile(sourcePath, nil, nil)
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

	return clerk.CompressFile(tarCompressedFilePath, &compressionFormat, nil)
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

func (FileClerk) ownerIdsResolver(
	ownerUsernamePtr *tkValueObject.UnixUsername,
	ownerUserIdPtr *tkValueObject.UnixUserId,
) (tkValueObject.UnixUserId, tkValueObject.UnixGroupId, error) {
	if ownerUsernamePtr == nil && ownerUserIdPtr == nil {
		processUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid())
		processGroupId, groupIdErr := tkValueObject.NewUnixGroupId(os.Getegid())
		if userIdErr != nil || groupIdErr != nil {
			return 0, 0, errors.Join(userIdErr, groupIdErr)
		}

		return processUserId, processGroupId, nil
	}

	accountIdentifier := ""
	var account *user.User
	var lookupErr error
	switch {
	case ownerUsernamePtr != nil:
		accountIdentifier = ownerUsernamePtr.String()
		account, lookupErr = user.Lookup(accountIdentifier)
	default:
		accountIdentifier = ownerUserIdPtr.String()
		account, lookupErr = user.LookupId(accountIdentifier)
	}
	if lookupErr != nil {
		return 0, 0, fmt.Errorf(
			"OwnerLookupFailed: %s: %w", accountIdentifier, lookupErr,
		)
	}

	resolvedUserId, userIdErr := tkValueObject.NewUnixUserId(account.Uid)
	resolvedGroupId, groupIdErr := tkValueObject.NewUnixGroupId(account.Gid)
	if userIdErr != nil || groupIdErr != nil {
		return 0, 0, fmt.Errorf(
			"OwnerIdParseFailed: %s: %w",
			account.Username, errors.Join(userIdErr, groupIdErr),
		)
	}

	return resolvedUserId, resolvedGroupId, nil
}

func (FileClerk) openRedirectProofDirChain(
	dirPath tkValueObject.UnixAbsoluteFilePath,
	ownerUserId tkValueObject.UnixUserId,
) (dirHandle int, err error) {
	walkFlags := unix.O_PATH | unix.O_NOFOLLOW | unix.O_CLOEXEC
	walkHandle, openErr := unix.Open("/", walkFlags, 0)
	if openErr != nil {
		return 0, fmt.Errorf("PathCheckFailed: /: %w", openErr)
	}
	defer func() {
		if err != nil {
			_ = unix.Close(walkHandle)
		}
	}()

	for pathComponent := range strings.SplitSeq(dirPath.String(), "/") {
		if pathComponent == "" || pathComponent == "." {
			continue
		}
		componentAscendsToParent := pathComponent == ".."
		if componentAscendsToParent {
			return 0, ErrDirPathTraversalInvalid
		}

		stepHandle, stepErr := unix.Openat(
			walkHandle, pathComponent, walkFlags, 0,
		)
		if stepErr != nil {
			return 0, fmt.Errorf(
				"PathCheckFailed: %s: %w", pathComponent, stepErr,
			)
		}
		_ = unix.Close(walkHandle)
		walkHandle = stepHandle

		componentStat := unix.Stat_t{}
		statErr := unix.Fstat(walkHandle, &componentStat)
		if statErr != nil {
			return 0, fmt.Errorf(
				"PathCheckFailed: %s: %w", pathComponent, statErr,
			)
		}

		componentFileFormat := componentStat.Mode & unix.S_IFMT
		componentIsSymlink := componentFileFormat == unix.S_IFLNK
		if componentIsSymlink {
			return 0, fmt.Errorf(
				"%w: %s", ErrSymlinkedPathInvalid, pathComponent,
			)
		}
		componentIsDir := componentFileFormat == unix.S_IFDIR
		if !componentIsDir {
			return 0, fmt.Errorf(
				"%w: %s", ErrTargetNotDirectory, pathComponent,
			)
		}

		ownedByRequestedUser := uint64(componentStat.Uid) == ownerUserId.Uint64()
		ownedByRoot := componentStat.Uid == 0
		ownedByTrustedAccount := ownedByRequestedUser || ownedByRoot
		if !ownedByTrustedAccount {
			return 0, fmt.Errorf(
				"%w: %s", ErrDirectoryOwnerInvalid, pathComponent,
			)
		}
	}

	return walkHandle, nil
}

func (clerk FileClerk) VerifyDirPathRedirectSafety(
	dirPath tkValueObject.UnixAbsoluteFilePath,
	ownerUsernamePtr *tkValueObject.UnixUsername,
	ownerUserIdPtr *tkValueObject.UnixUserId,
) error {
	ownerUserId, _, resolveErr := clerk.ownerIdsResolver(
		ownerUsernamePtr, ownerUserIdPtr,
	)
	if resolveErr != nil {
		return resolveErr
	}

	dirHandle, dirChainErr := clerk.openRedirectProofDirChain(dirPath, ownerUserId)
	if dirChainErr != nil {
		return dirChainErr
	}
	return unix.Close(dirHandle)
}

func (clerk FileClerk) AppendFileContent(
	filePath tkValueObject.UnixAbsoluteFilePath,
	content string,
) error {
	targetFileName, fileNameErr := filePath.ReadFileName(true)
	if fileNameErr != nil {
		return ErrFileNameInvalid
	}

	ownerUserId, _, resolveErr := clerk.ownerIdsResolver(nil, nil)
	if resolveErr != nil {
		return resolveErr
	}

	dirPath := filePath.ReadFileDir()
	dirHandle, dirChainErr := clerk.openRedirectProofDirChain(dirPath, ownerUserId)
	if dirChainErr != nil {
		return dirChainErr
	}
	defer func() { _ = unix.Close(dirHandle) }()

	fileHandle, openErr := unix.Openat(
		dirHandle,
		targetFileName.String(),
		unix.O_WRONLY|unix.O_CREAT|unix.O_APPEND|unix.O_NOFOLLOW|unix.O_CLOEXEC,
		0644,
	)
	if openErr != nil {
		if errors.Is(openErr, unix.ELOOP) {
			return ErrTargetIsSymlink
		}
		return openErr
	}
	targetFile := os.NewFile(uintptr(fileHandle), targetFileName.String())

	_, writeErr := targetFile.WriteString(content)
	closeErr := targetFile.Close()
	return errors.Join(writeErr, closeErr)
}

func (FileClerk) resolveSymlinkedFilePath(filePath string) (string, error) {
	fileInfo, lstatErr := os.Lstat(filePath)
	entryIsSymlink := lstatErr == nil && fileInfo.Mode()&os.ModeSymlink != 0
	if entryIsSymlink {
		return filepath.EvalSymlinks(filePath)
	}

	dirPath, fileName := filepath.Split(filePath)
	resolvedDirPath, evalErr := filepath.EvalSymlinks(dirPath)
	if evalErr != nil {
		return "", evalErr
	}
	return filepath.Join(resolvedDirPath, fileName), nil
}

type FileUpsertSettings struct {
	FilePath             tkValueObject.UnixAbsoluteFilePath
	Permissions          os.FileMode
	ShouldFollowSymlinks bool
	ShouldOverwrite      bool
	OwnerUsername        *tkValueObject.UnixUsername
	OwnerUserId          *tkValueObject.UnixUserId
}

func (clerk FileClerk) UpsertFile(
	settings FileUpsertSettings,
	fileContent []byte,
) error {
	if settings.Permissions == 0 {
		return ErrFilePermissionsInvalid
	}

	targetFilePath := settings.FilePath
	if settings.ShouldFollowSymlinks {
		resolvedFilePathStr, resolveErr := clerk.resolveSymlinkedFilePath(
			targetFilePath.String(),
		)
		if resolveErr != nil {
			return resolveErr
		}

		resolvedFilePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
			resolvedFilePathStr, true,
		)
		if pathErr != nil {
			return pathErr
		}
		targetFilePath = resolvedFilePath
	}

	targetFileName, fileNameErr := targetFilePath.ReadFileName(true)
	if fileNameErr != nil {
		return ErrFileNameInvalid
	}

	ownerUserId, ownerGroupId, resolveErr := clerk.ownerIdsResolver(
		settings.OwnerUsername, settings.OwnerUserId,
	)
	if resolveErr != nil {
		return resolveErr
	}

	dirPath := targetFilePath.ReadFileDir()
	dirHandle, dirChainErr := clerk.openRedirectProofDirChain(dirPath, ownerUserId)
	if dirChainErr != nil {
		return dirChainErr
	}
	defer func() { _ = unix.Close(dirHandle) }()

	writeSettings := atomicFileWriteSettings{
		DirHandle:       dirHandle,
		TargetFileName:  targetFileName,
		Permissions:     settings.Permissions,
		ShouldOverwrite: settings.ShouldOverwrite,
	}
	shouldChown := settings.OwnerUsername != nil || settings.OwnerUserId != nil
	if shouldChown {
		writeSettings.OwnerUserId = &ownerUserId
		writeSettings.OwnerGroupId = &ownerGroupId
	}

	writeErr := clerk.writeFileAtomically(
		writeSettings,
		func(writer io.Writer) error {
			_, writeErr := writer.Write(fileContent)
			return writeErr
		},
	)
	return writeErr
}
