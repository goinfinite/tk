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
	tempFileNameSuffix       = ".tk-tmp"
	tempFileNameEntropyChars = 8

	// A swapped-in FIFO would block the open before the swap verifier runs.
	targetFileReadOpenFlags = unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_CLOEXEC |
		unix.O_NONBLOCK
	targetFileAppendOpenFlags = unix.O_WRONLY | unix.O_APPEND | unix.O_NOFOLLOW |
		unix.O_CLOEXEC | unix.O_NONBLOCK
)

type FileClerkSymlinkPolicy string

// FileClerkDirChainPolicy selects how the directory-chain walk treats a
// component writable by group or others.
type FileClerkDirChainPolicy string

type FileClerkOverwritePolicy string

type FileClerkOwnerSource string

var (
	RegexLargeFileThresholdBytes       int64       = 10 * 1024 * 1024
	ReadFileContentDefaultMaxSizeBytes int64       = 500 * 1024 * 1024
	FileClerkDefaultNewFileMode        os.FileMode = 0o600

	FileClerkSymlinkPolicyStrictRefuse FileClerkSymlinkPolicy = "strict-refuse"
	FileClerkSymlinkPolicyResolve      FileClerkSymlinkPolicy = "resolve"

	// FileClerkDirChainPolicySharedWriteAllowed is the default policy.
	FileClerkDirChainPolicySharedWriteAllowed FileClerkDirChainPolicy = "shared-write-allowed"

	// FileClerkDirChainPolicySharedWriteRefused rejects a component writable
	// by group or others, unless it is sticky (ErrDirectoryWritableByOthers).
	FileClerkDirChainPolicySharedWriteRefused FileClerkDirChainPolicy = "shared-write-refused"

	FileClerkOverwritePolicyStrictRefuse FileClerkOverwritePolicy = "strict-refuse"
	FileClerkOverwritePolicyReplace      FileClerkOverwritePolicy = "replace"

	FileClerkOwnerSourceExistingFile        FileClerkOwnerSource = "existing-file"
	FileClerkOwnerSourceContainingDirectory FileClerkOwnerSource = "containing-directory"
	FileClerkOwnerSourceRunningProcess      FileClerkOwnerSource = "running-process"

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
	ErrTargetNotRegularFile         = errors.New("TargetNotRegularFile")
	ErrDirPathTraversalInvalid      = errors.New("DirPathParentTraversalNotAllowed")
	ErrSymlinkedPathInvalid         = errors.New("PathComponentIsSymlink")
	ErrDirectoryOwnerInvalid        = errors.New("DirectoryNotOwnedByExpectedOwner")
	ErrDirectoryWritableByOthers    = errors.New("DirectoryWritableByOthers")
	ErrFileNameInvalid              = errors.New("FilePathFinalComponentIsNotAFileName")
	ErrTempFileNameTooLong          = errors.New("TempFileNameExceedsNameMax")
	ErrDirChainPolicyInvalid        = errors.New("DirChainPolicyInvalid")
	ErrSymlinkPolicyInvalid         = errors.New("SymlinkPolicyInvalid")
	ErrOverwritePolicyInvalid       = errors.New("OverwritePolicyInvalid")
	ErrOwnerSourceInvalid           = errors.New("OwnerSourceInvalid")
	ErrOwnerSourceConflict          = errors.New("OwnerSourceConflictsWithStatedOwner")
	ErrFileOwnerChangeFailed        = errors.New("FileOwnerChangeFailed")
	ErrTargetFileChanged            = errors.New("TargetFileChanged")
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

// TouchFile behaves like touch(1), except a dangling symlink fails with
// ErrTargetIsSymlink instead of creating the file behind the link.
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

func (FileClerk) readFileContentFromReader(
	fileReader io.Reader,
	maxContentSizeBytesPtr *int64,
) (fileContent string, err error) {
	maxContentSizeBytes := ReadFileContentDefaultMaxSizeBytes
	if maxContentSizeBytesPtr != nil {
		maxContentSizeBytes = *maxContentSizeBytesPtr
	}
	if maxContentSizeBytes < 0 {
		maxContentSizeBytes = 0
	}

	limitedReader := io.LimitedReader{R: fileReader, N: maxContentSizeBytes}
	fileContentBytes, err := io.ReadAll(&limitedReader)
	if err != nil {
		return fileContent, err
	}

	sizeCapReached := limitedReader.N == 0
	if !sizeCapReached {
		return string(fileContentBytes), nil
	}

	trailingByte := make([]byte, 1)
	trailingCount, trailingErr := fileReader.Read(trailingByte)
	if trailingCount > 0 {
		return fileContent, ErrFileTooLarge
	}
	if trailingErr != nil && !errors.Is(trailingErr, io.EOF) {
		return fileContent, trailingErr
	}

	return string(fileContentBytes), nil
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

	return clerk.readFileContentFromReader(fileHandler, maxContentSizeBytesPtr)
}

// FileContentRegexFindings holds one regex match and its capture groups.
// LineNumRange is [start, end] inclusive; a single-line match has start == end.
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

// FileContentRegexSearch finds every regex match in a file with its 1-based
// inclusive line range and capture groups. Files under
// RegexLargeFileThresholdBytes are matched in one pass, so per-line anchors
// need the (?m) flag and multi-line matches span every line they touch.
// Larger files stream line-by-line, so multi-line patterns match within a
// single line only.
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
		ownerChangeNotPermitted := errors.Is(chownErr, unix.EPERM)
		if ownerChangeNotPermitted {
			chownErr = fmt.Errorf("%w: %w", ErrFileOwnerChangeFailed, chownErr)
		}
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

func (FileClerk) unixFileModeConverter(rawMode uint32) os.FileMode {
	fileMode := os.FileMode(rawMode).Perm()
	if rawMode&unix.S_ISUID != 0 {
		fileMode |= os.ModeSetuid
	}
	if rawMode&unix.S_ISGID != 0 {
		fileMode |= os.ModeSetgid
	}
	if rawMode&unix.S_ISVTX != 0 {
		fileMode |= os.ModeSticky
	}

	return fileMode
}

type targetFileState struct {
	Exists       bool
	IsSymlink    bool
	IsDirectory  bool
	OwnerUserId  tkValueObject.UnixUserId
	OwnerGroupId tkValueObject.UnixGroupId
	Permissions  os.FileMode
	SizeBytes    int64
	DeviceId     uint64
	InodeId      uint64
}

func (clerk FileClerk) targetFileStateReader(
	dirHandle int,
	targetFileName tkValueObject.UnixFileName,
) (fileState targetFileState, err error) {
	entryStat := unix.Stat_t{}
	statErr := unix.Fstatat(
		dirHandle, targetFileName.String(), &entryStat, unix.AT_SYMLINK_NOFOLLOW,
	)
	if statErr != nil {
		if errors.Is(statErr, unix.ENOENT) {
			return fileState, nil
		}
		return fileState, statErr
	}

	ownerUserId, userIdErr := tkValueObject.NewUnixUserId(entryStat.Uid)
	if userIdErr != nil {
		return fileState, userIdErr
	}
	ownerGroupId, groupIdErr := tkValueObject.NewUnixGroupId(entryStat.Gid)
	if groupIdErr != nil {
		return fileState, groupIdErr
	}

	entryFormat := entryStat.Mode & unix.S_IFMT
	fileState = targetFileState{
		Exists:       true,
		IsSymlink:    entryFormat == unix.S_IFLNK,
		IsDirectory:  entryFormat == unix.S_IFDIR,
		OwnerUserId:  ownerUserId,
		OwnerGroupId: ownerGroupId,
		Permissions:  clerk.unixFileModeConverter(entryStat.Mode),
		SizeBytes:    entryStat.Size,
		DeviceId:     entryStat.Dev,
		InodeId:      entryStat.Ino,
	}

	return fileState, nil
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

// DecompressFile expands sourcePath and removes the archive unless
// shouldKeepSourceFilePtr requests otherwise. Zip extraction overwrites
// existing files at the destination without warning: only decompress archives
// you trust.
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

func (FileClerk) ownerUserIdResolver(
	ownerUsernamePtr *tkValueObject.UnixUsername,
	ownerUserIdPtr *tkValueObject.UnixUserId,
) (tkValueObject.UnixUserId, error) {
	switch {
	case ownerUsernamePtr != nil:
		account, lookupErr := user.Lookup(ownerUsernamePtr.String())
		if lookupErr != nil {
			return 0, fmt.Errorf(
				"OwnerLookupFailed: %s: %w", ownerUsernamePtr.String(), lookupErr,
			)
		}

		resolvedUserId, userIdErr := tkValueObject.NewUnixUserId(account.Uid)
		if userIdErr != nil {
			return 0, fmt.Errorf(
				"OwnerIdParseFailed: %s: %w", account.Username, userIdErr,
			)
		}

		return resolvedUserId, nil
	case ownerUserIdPtr != nil:
		return *ownerUserIdPtr, nil
	default:
		processUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid())
		if userIdErr != nil {
			return 0, userIdErr
		}

		return processUserId, nil
	}
}

func (FileClerk) ownerGroupIdResolver(
	ownerUserId tkValueObject.UnixUserId,
) (tkValueObject.UnixGroupId, error) {
	account, lookupErr := user.LookupId(ownerUserId.String())
	if lookupErr != nil {
		return 0, fmt.Errorf(
			"OwnerLookupFailed: %s: %w", ownerUserId.String(), lookupErr,
		)
	}

	resolvedGroupId, groupIdErr := tkValueObject.NewUnixGroupId(account.Gid)
	if groupIdErr != nil {
		return 0, fmt.Errorf(
			"OwnerIdParseFailed: %s: %w", account.Username, groupIdErr,
		)
	}

	return resolvedGroupId, nil
}

type trustedDirOwnerIdSet map[uint64]struct{}

func (clerk FileClerk) trustedDirOwnerIdSetResolver(
	trustedDirOwnerUsernames []tkValueObject.UnixUsername,
	trustedDirOwnerUserIds []tkValueObject.UnixUserId,
) (trustedDirOwnerIds trustedDirOwnerIdSet, err error) {
	trustedDirOwnerIds = trustedDirOwnerIdSet{}

	for _, trustedDirOwnerUsername := range trustedDirOwnerUsernames {
		resolvedUserId, resolveErr := clerk.ownerUserIdResolver(
			&trustedDirOwnerUsername, nil,
		)
		if resolveErr != nil {
			return nil, resolveErr
		}
		trustedDirOwnerIds[resolvedUserId.Uint64()] = struct{}{}
	}

	for _, trustedDirOwnerUserId := range trustedDirOwnerUserIds {
		trustedDirOwnerIds[trustedDirOwnerUserId.Uint64()] = struct{}{}
	}

	if len(trustedDirOwnerIds) == 0 {
		processUserId, resolveErr := clerk.ownerUserIdResolver(nil, nil)
		if resolveErr != nil {
			return nil, resolveErr
		}
		trustedDirOwnerIds[processUserId.Uint64()] = struct{}{}
	}

	return trustedDirOwnerIds, nil
}

func (FileClerk) openRedirectProofDirChain(
	dirPath tkValueObject.UnixAbsoluteFilePath,
	trustedDirOwnerIds trustedDirOwnerIdSet,
	dirChainPolicy FileClerkDirChainPolicy,
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

		_, componentOwnedByTrustedOwner := trustedDirOwnerIds[uint64(componentStat.Uid)]
		componentOwnedByRoot := componentStat.Uid == 0
		componentOwnedByTrustedAccount := componentOwnedByTrustedOwner ||
			componentOwnedByRoot
		if !componentOwnedByTrustedAccount {
			return 0, fmt.Errorf(
				"%w: %s", ErrDirectoryOwnerInvalid, pathComponent,
			)
		}

		componentHasSharedWrite :=
			componentStat.Mode&(unix.S_IWGRP|unix.S_IWOTH) != 0
		componentIsSticky := componentStat.Mode&unix.S_ISVTX != 0
		policyRefusesSharedWrite :=
			dirChainPolicy == FileClerkDirChainPolicySharedWriteRefused
		componentIsSharedWriteRefused := policyRefusesSharedWrite &&
			componentHasSharedWrite && !componentIsSticky
		if componentIsSharedWriteRefused {
			return 0, fmt.Errorf(
				"%w: %s", ErrDirectoryWritableByOthers, pathComponent,
			)
		}
	}

	return walkHandle, nil
}

func (FileClerk) symlinkPolicyNormalizer(
	symlinkPolicyPtr *FileClerkSymlinkPolicy,
) (*FileClerkSymlinkPolicy, error) {
	if symlinkPolicyPtr == nil {
		return &FileClerkSymlinkPolicyStrictRefuse, nil
	}
	switch *symlinkPolicyPtr {
	case FileClerkSymlinkPolicyStrictRefuse, FileClerkSymlinkPolicyResolve:
		return symlinkPolicyPtr, nil
	default:
		return nil, ErrSymlinkPolicyInvalid
	}
}

func (FileClerk) dirChainPolicyNormalizer(
	dirChainPolicyPtr *FileClerkDirChainPolicy,
) (*FileClerkDirChainPolicy, error) {
	if dirChainPolicyPtr == nil {
		return &FileClerkDirChainPolicySharedWriteAllowed, nil
	}
	switch *dirChainPolicyPtr {
	case FileClerkDirChainPolicySharedWriteAllowed, FileClerkDirChainPolicySharedWriteRefused:
		return dirChainPolicyPtr, nil
	default:
		return nil, ErrDirChainPolicyInvalid
	}
}

func (FileClerk) resolveSymlinkedFilePath(filePath string) (string, error) {
	fileInfo, lstatErr := os.Lstat(filePath)
	entryIsSymlink := lstatErr == nil && fileInfo.Mode()&os.ModeSymlink != 0
	if entryIsSymlink {
		resolvedFilePath, evalErr := filepath.EvalSymlinks(filePath)
		if os.IsNotExist(evalErr) {
			return "", ErrFileMissing
		}
		return resolvedFilePath, evalErr
	}

	dirPath, fileName := filepath.Split(filePath)
	resolvedDirPath, evalErr := filepath.EvalSymlinks(dirPath)
	if os.IsNotExist(evalErr) {
		return "", ErrFileMissing
	}
	if evalErr != nil {
		return "", evalErr
	}
	return filepath.Join(resolvedDirPath, fileName), nil
}

type fileWriteTarget struct {
	DirHandle   int
	FileName    tkValueObject.UnixFileName
	TargetState targetFileState
}

func (clerk FileClerk) fileWriteTargetResolver(
	filePath tkValueObject.UnixAbsoluteFilePath,
	symlinkPolicy FileClerkSymlinkPolicy,
	dirChainPolicy FileClerkDirChainPolicy,
	trustedDirOwnerUsernames []tkValueObject.UnixUsername,
	trustedDirOwnerUserIds []tkValueObject.UnixUserId,
) (target fileWriteTarget, err error) {
	hasTrailingSeparator := strings.HasSuffix(filePath.String(), "/")
	if hasTrailingSeparator {
		return target, ErrFileNameInvalid
	}

	targetFilePath := filePath
	if symlinkPolicy == FileClerkSymlinkPolicyResolve {
		resolvedFilePathStr, resolveErr := clerk.resolveSymlinkedFilePath(
			targetFilePath.String(),
		)
		if resolveErr != nil {
			return target, resolveErr
		}

		resolvedFilePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
			resolvedFilePathStr, true,
		)
		if pathErr != nil {
			return target, pathErr
		}
		targetFilePath = resolvedFilePath
	}

	targetFileName, fileNameErr := targetFilePath.ReadFileName(true)
	if fileNameErr != nil {
		return target, ErrFileNameInvalid
	}

	trustedDirOwnerIds, trustErr := clerk.trustedDirOwnerIdSetResolver(
		trustedDirOwnerUsernames, trustedDirOwnerUserIds,
	)
	if trustErr != nil {
		return target, trustErr
	}

	dirPath := targetFilePath.ReadFileDir()
	dirHandle, dirChainErr := clerk.openRedirectProofDirChain(
		dirPath, trustedDirOwnerIds, dirChainPolicy,
	)
	if dirChainErr != nil {
		return target, dirChainErr
	}

	targetState, stateErr := clerk.targetFileStateReader(dirHandle, targetFileName)
	if stateErr != nil {
		_ = unix.Close(dirHandle)
		return target, stateErr
	}

	return fileWriteTarget{
		DirHandle:   dirHandle,
		FileName:    targetFileName,
		TargetState: targetState,
	}, nil
}

func (FileClerk) atomicReplacementSettingsResolver(
	target fileWriteTarget,
) atomicFileWriteSettings {
	return atomicFileWriteSettings{
		DirHandle:       target.DirHandle,
		TargetFileName:  target.FileName,
		Permissions:     target.TargetState.Permissions,
		ShouldOverwrite: true,
		OwnerUserId:     &target.TargetState.OwnerUserId,
		OwnerGroupId:    &target.TargetState.OwnerGroupId,
	}
}

func (FileClerk) targetFileStateValidator(
	targetState targetFileState,
) error {
	if !targetState.Exists {
		return ErrFileMissing
	}
	if targetState.IsSymlink {
		return ErrTargetIsSymlink
	}
	if targetState.IsDirectory {
		return ErrTargetIsDirectory
	}

	return nil
}

func (FileClerk) targetFileSwapVerifier(
	fileHandle int,
	targetState targetFileState,
) error {
	openedFileStat := unix.Stat_t{}
	statErr := unix.Fstat(fileHandle, &openedFileStat)
	if statErr != nil {
		return statErr
	}

	identityChanged := openedFileStat.Dev != targetState.DeviceId ||
		openedFileStat.Ino != targetState.InodeId
	if identityChanged {
		return ErrTargetFileChanged
	}

	return nil
}

func (clerk FileClerk) inspectedTargetFileOpener(
	dirHandle int,
	targetFileName tkValueObject.UnixFileName,
	targetState targetFileState,
	openFlags int,
) (fileHandle int, err error) {
	fileHandle, openErr := unix.Openat(
		dirHandle, targetFileName.String(), openFlags, 0,
	)
	if openErr != nil {
		switch {
		case errors.Is(openErr, unix.ELOOP):
			return 0, ErrTargetIsSymlink
		case errors.Is(openErr, unix.ENOENT):
			return 0, ErrFileMissing
		case errors.Is(openErr, unix.EISDIR):
			return 0, ErrTargetIsDirectory
		case errors.Is(openErr, unix.ENXIO):
			return 0, ErrTargetNotRegularFile
		default:
			return 0, openErr
		}
	}

	swapErr := clerk.targetFileSwapVerifier(fileHandle, targetState)
	if swapErr != nil {
		_ = unix.Close(fileHandle)
		return 0, swapErr
	}

	return fileHandle, nil
}

type FileRegexReplaceSettings struct {
	FilePath tkValueObject.UnixAbsoluteFilePath

	DirChainPolicy *FileClerkDirChainPolicy
	SymlinkPolicy  *FileClerkSymlinkPolicy

	// When empty, the running process account is trusted; root is always trusted.
	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId
}

func (clerk FileClerk) regexReplaceWholeFile(
	fileHandler *os.File,
	target fileWriteTarget,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	fileContent, readErr := clerk.readFileContentFromReader(fileHandler, nil)
	if readErr != nil {
		return 0, readErr
	}

	foundMatches := regexPattern.FindAllString(fileContent, -1)
	replacementCount = len(foundMatches)
	replacedContent := regexPattern.ReplaceAllString(fileContent, replacement)

	if len(replacedContent) == 0 {
		return 0, ErrReplacementWouldTruncateFile
	}

	writeErr := clerk.writeFileAtomically(
		clerk.atomicReplacementSettingsResolver(target),
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
	fileHandler *os.File,
	target fileWriteTarget,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	bufferReader := bufio.NewReader(fileHandler)
	writeErr := clerk.writeFileAtomically(
		clerk.atomicReplacementSettingsResolver(target),
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

// FileContentRegexReplace atomically substitutes regex matches in a file. It
// preserves the target's owner, group, and mode, including special bits. The
// parent chain is held and the opened inode is verified against the inspected
// target, so a swap between the two fails with ErrTargetFileChanged. Symlinks
// are refused unless the policy resolves them. A zero-byte result fails; use
// TruncateFileContent to empty a file. Files at or above the large-file
// threshold stream line-by-line, so multi-line patterns need smaller files.
func (clerk FileClerk) FileContentRegexReplace(
	settings FileRegexReplaceSettings,
	regexPattern *regexp.Regexp,
	replacement string,
) (replacementCount int, err error) {
	if regexPattern == nil {
		return 0, ErrRegexPatternMissing
	}

	symlinkPolicy, symlinkPolicyErr := clerk.symlinkPolicyNormalizer(
		settings.SymlinkPolicy,
	)
	if symlinkPolicyErr != nil {
		return 0, symlinkPolicyErr
	}
	dirChainPolicy, dirChainPolicyErr := clerk.dirChainPolicyNormalizer(
		settings.DirChainPolicy,
	)
	if dirChainPolicyErr != nil {
		return 0, dirChainPolicyErr
	}

	target, targetErr := clerk.fileWriteTargetResolver(
		settings.FilePath, *symlinkPolicy, *dirChainPolicy,
		settings.TrustedDirOwnerUsernames, settings.TrustedDirOwnerUserIds,
	)
	if targetErr != nil {
		return 0, targetErr
	}
	defer func() { _ = unix.Close(target.DirHandle) }()

	targetState := target.TargetState
	targetStateErr := clerk.targetFileStateValidator(targetState)
	if targetStateErr != nil {
		return 0, targetStateErr
	}
	if targetState.SizeBytes == 0 {
		return 0, ErrFileEmpty
	}

	fileHandle, openErr := clerk.inspectedTargetFileOpener(
		target.DirHandle, target.FileName, targetState, targetFileReadOpenFlags,
	)
	if openErr != nil {
		return 0, openErr
	}
	fileHandler := os.NewFile(uintptr(fileHandle), target.FileName.String())
	defer func() { _ = fileHandler.Close() }()

	if targetState.SizeBytes >= RegexLargeFileThresholdBytes {
		slog.Warn(
			"FileContentRegexReplaceStreamingFallback",
			slog.String("filePath", settings.FilePath.String()),
			slog.Int64("fileSizeBytes", targetState.SizeBytes),
			slog.Int64("thresholdBytes", RegexLargeFileThresholdBytes),
			slog.String("reason", "FileSizeExceedsThreshold"),
		)
		return clerk.regexReplaceStreaming(
			fileHandler, target, regexPattern, replacement,
		)
	}
	return clerk.regexReplaceWholeFile(
		fileHandler, target, regexPattern, replacement,
	)
}

type FileAppendSettings struct {
	FilePath tkValueObject.UnixAbsoluteFilePath

	DirChainPolicy *FileClerkDirChainPolicy
	SymlinkPolicy  *FileClerkSymlinkPolicy

	// When empty, the running process account is trusted; root is always trusted.
	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId
}

// AppendFileContent verifies the opened inode against the inspected target and
// appends through an O_APPEND write, so concurrent writers never lose data and
// the target's owner, group, and mode stay untouched. A missing target fails
// with ErrFileMissing; create it with UpsertFile. Symlinks are refused unless
// the policy resolves them.
func (clerk FileClerk) AppendFileContent(
	settings FileAppendSettings,
	content string,
) error {
	symlinkPolicy, symlinkPolicyErr := clerk.symlinkPolicyNormalizer(
		settings.SymlinkPolicy,
	)
	if symlinkPolicyErr != nil {
		return symlinkPolicyErr
	}
	dirChainPolicy, dirChainPolicyErr := clerk.dirChainPolicyNormalizer(
		settings.DirChainPolicy,
	)
	if dirChainPolicyErr != nil {
		return dirChainPolicyErr
	}

	target, targetErr := clerk.fileWriteTargetResolver(
		settings.FilePath, *symlinkPolicy, *dirChainPolicy,
		settings.TrustedDirOwnerUsernames, settings.TrustedDirOwnerUserIds,
	)
	if targetErr != nil {
		return targetErr
	}
	defer func() { _ = unix.Close(target.DirHandle) }()

	targetState := target.TargetState
	targetStateErr := clerk.targetFileStateValidator(targetState)
	if targetStateErr != nil {
		return targetStateErr
	}

	fileHandle, openErr := clerk.inspectedTargetFileOpener(
		target.DirHandle, target.FileName, targetState, targetFileAppendOpenFlags,
	)
	if openErr != nil {
		return openErr
	}
	targetFile := os.NewFile(uintptr(fileHandle), target.FileName.String())

	_, writeErr := targetFile.WriteString(content)
	closeErr := targetFile.Close()
	return errors.Join(writeErr, closeErr)
}

type FileUpsertSettings struct {
	FilePath tkValueObject.UnixAbsoluteFilePath

	DirChainPolicy  *FileClerkDirChainPolicy
	OverwritePolicy *FileClerkOverwritePolicy
	SymlinkPolicy   *FileClerkSymlinkPolicy

	// When empty, the running process account is trusted; root is always trusted.
	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId

	Permissions *os.FileMode

	OwnerSource   *FileClerkOwnerSource
	OwnerUsername *tkValueObject.UnixUsername
	OwnerUserId   *tkValueObject.UnixUserId
	OwnerGroupId  *tkValueObject.UnixGroupId
}

type fileUpsertOwnership struct {
	UserId  tkValueObject.UnixUserId
	GroupId tkValueObject.UnixGroupId
}

func (clerk FileClerk) runningProcessOwnershipResolver() (
	ownership fileUpsertOwnership, err error,
) {
	processUserId, userIdErr := clerk.ownerUserIdResolver(nil, nil)
	if userIdErr != nil {
		return ownership, userIdErr
	}

	processGroupId, groupIdErr := tkValueObject.NewUnixGroupId(os.Getegid())
	if groupIdErr != nil {
		return ownership, groupIdErr
	}

	ownership = fileUpsertOwnership{
		UserId:  processUserId,
		GroupId: processGroupId,
	}

	return ownership, nil
}

func (clerk FileClerk) fileUpsertSourceOwnershipResolver(
	ownerSource FileClerkOwnerSource,
	targetState targetFileState,
	containingDirStat unix.Stat_t,
) (ownership fileUpsertOwnership, err error) {
	switch ownerSource {
	case FileClerkOwnerSourceExistingFile:
		if targetState.Exists {
			ownership = fileUpsertOwnership{
				UserId:  targetState.OwnerUserId,
				GroupId: targetState.OwnerGroupId,
			}
			return ownership, nil
		}
		return clerk.runningProcessOwnershipResolver()
	case FileClerkOwnerSourceContainingDirectory:
		containingDirUserId, userIdErr := tkValueObject.NewUnixUserId(
			containingDirStat.Uid,
		)
		if userIdErr != nil {
			return ownership, userIdErr
		}
		containingDirGroupId, groupIdErr := tkValueObject.NewUnixGroupId(
			containingDirStat.Gid,
		)
		if groupIdErr != nil {
			return ownership, groupIdErr
		}
		ownership = fileUpsertOwnership{
			UserId:  containingDirUserId,
			GroupId: containingDirGroupId,
		}
		return ownership, nil
	case FileClerkOwnerSourceRunningProcess:
		return clerk.runningProcessOwnershipResolver()
	default:
		return ownership, ErrOwnerSourceInvalid
	}
}

func (clerk FileClerk) fileUpsertOwnerResolver(
	ownerSource FileClerkOwnerSource,
	ownerUsername *tkValueObject.UnixUsername,
	ownerUserId *tkValueObject.UnixUserId,
	ownerGroupId *tkValueObject.UnixGroupId,
	targetState targetFileState,
	containingDirStat unix.Stat_t,
) (ownership fileUpsertOwnership, err error) {
	ownerAccountStated := ownerUsername != nil || ownerUserId != nil
	if ownerAccountStated {
		ownerSourceContradictsAccount :=
			ownerSource == FileClerkOwnerSourceContainingDirectory ||
				ownerSource == FileClerkOwnerSourceRunningProcess
		if ownerSourceContradictsAccount {
			return ownership, ErrOwnerSourceConflict
		}

		statedUserId, userIdErr := clerk.ownerUserIdResolver(
			ownerUsername, ownerUserId,
		)
		if userIdErr != nil {
			return ownership, userIdErr
		}
		if ownerGroupId != nil {
			ownership = fileUpsertOwnership{
				UserId:  statedUserId,
				GroupId: *ownerGroupId,
			}
			return ownership, nil
		}

		statedGroupId, groupIdErr := clerk.ownerGroupIdResolver(statedUserId)
		if groupIdErr != nil {
			return ownership, groupIdErr
		}
		ownership = fileUpsertOwnership{
			UserId:  statedUserId,
			GroupId: statedGroupId,
		}
		return ownership, nil
	}

	ownership, err = clerk.fileUpsertSourceOwnershipResolver(
		ownerSource, targetState, containingDirStat,
	)
	if err != nil {
		return ownership, err
	}
	if ownerGroupId != nil {
		ownership.GroupId = *ownerGroupId
	}

	return ownership, nil
}

func (FileClerk) fileUpsertPermissionsResolver(
	statedPermissionsPtr *os.FileMode,
	targetState targetFileState,
) os.FileMode {
	if statedPermissionsPtr != nil {
		return *statedPermissionsPtr
	}
	if targetState.Exists {
		return targetState.Permissions
	}

	return FileClerkDefaultNewFileMode
}

func (clerk FileClerk) fileUpsertSettingsNormalizer(
	settings FileUpsertSettings,
) (FileUpsertSettings, error) {
	dirChainPolicy, dirChainPolicyErr := clerk.dirChainPolicyNormalizer(
		settings.DirChainPolicy,
	)
	if dirChainPolicyErr != nil {
		return settings, dirChainPolicyErr
	}
	settings.DirChainPolicy = dirChainPolicy

	symlinkPolicy, symlinkPolicyErr := clerk.symlinkPolicyNormalizer(
		settings.SymlinkPolicy,
	)
	if symlinkPolicyErr != nil {
		return settings, symlinkPolicyErr
	}
	settings.SymlinkPolicy = symlinkPolicy

	if settings.OverwritePolicy == nil {
		settings.OverwritePolicy = &FileClerkOverwritePolicyStrictRefuse
	}
	switch *settings.OverwritePolicy {
	case FileClerkOverwritePolicyStrictRefuse, FileClerkOverwritePolicyReplace:
	default:
		return settings, ErrOverwritePolicyInvalid
	}

	if settings.OwnerSource == nil {
		settings.OwnerSource = &FileClerkOwnerSourceExistingFile
	}
	switch *settings.OwnerSource {
	case FileClerkOwnerSourceExistingFile, FileClerkOwnerSourceContainingDirectory,
		FileClerkOwnerSourceRunningProcess:
	default:
		return settings, ErrOwnerSourceInvalid
	}

	return settings, nil
}

func (clerk FileClerk) UpsertFile(
	settings FileUpsertSettings,
	fileContent []byte,
) error {
	settings, normalizeErr := clerk.fileUpsertSettingsNormalizer(settings)
	if normalizeErr != nil {
		return normalizeErr
	}
	symlinkPolicy := *settings.SymlinkPolicy
	overwritePolicy := *settings.OverwritePolicy
	ownerSource := *settings.OwnerSource

	target, targetErr := clerk.fileWriteTargetResolver(
		settings.FilePath, symlinkPolicy, *settings.DirChainPolicy,
		settings.TrustedDirOwnerUsernames, settings.TrustedDirOwnerUserIds,
	)
	if targetErr != nil {
		return targetErr
	}
	defer func() { _ = unix.Close(target.DirHandle) }()

	targetState := target.TargetState
	shouldRefuseSymlink := targetState.IsSymlink &&
		symlinkPolicy == FileClerkSymlinkPolicyStrictRefuse
	if shouldRefuseSymlink {
		return ErrTargetIsSymlink
	}
	if targetState.IsDirectory {
		return ErrTargetIsDirectory
	}

	shouldOverwrite := overwritePolicy == FileClerkOverwritePolicyReplace
	if targetState.Exists && !shouldOverwrite {
		return ErrTargetFileExists
	}

	permissions := clerk.fileUpsertPermissionsResolver(
		settings.Permissions, targetState,
	)

	containingDirStat := unix.Stat_t{}
	containingDirStatErr := unix.Fstat(target.DirHandle, &containingDirStat)
	if containingDirStatErr != nil {
		return containingDirStatErr
	}

	ownership, ownerErr := clerk.fileUpsertOwnerResolver(
		ownerSource, settings.OwnerUsername, settings.OwnerUserId,
		settings.OwnerGroupId, targetState, containingDirStat,
	)
	if ownerErr != nil {
		return ownerErr
	}

	writeSettings := atomicFileWriteSettings{
		DirHandle:       target.DirHandle,
		TargetFileName:  target.FileName,
		Permissions:     permissions,
		ShouldOverwrite: shouldOverwrite,
		OwnerUserId:     &ownership.UserId,
		OwnerGroupId:    &ownership.GroupId,
	}
	return clerk.writeFileAtomically(
		writeSettings,
		func(writer io.Writer) error {
			_, writeErr := writer.Write(fileContent)
			return writeErr
		},
	)
}
