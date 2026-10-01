package tkInfra

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	tkEntity "github.com/goinfinite/tk/src/domain/entity"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

const (
	tempFileNameSuffix       = ".tk-tmp"
	tempFileNameEntropyChars = 8

	// A swapped-in FIFO would block the open before the identity check.
	targetFileReadOpenFlags = unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_CLOEXEC |
		unix.O_NONBLOCK
	targetFileAppendOpenFlags = unix.O_WRONLY | unix.O_APPEND | unix.O_NOFOLLOW |
		unix.O_CLOEXEC | unix.O_NONBLOCK

	dirReadOpenFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_NOFOLLOW |
		unix.O_CLOEXEC
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
	FileClerkDefaultMaxFiles           uint64      = 1000

	FileClerkSymlinkPolicyStrictRefuse FileClerkSymlinkPolicy = "strict-refuse"
	FileClerkSymlinkPolicyResolve      FileClerkSymlinkPolicy = "resolve"

	// FileClerkDirChainPolicySharedWriteAllowed is the default policy.
	FileClerkDirChainPolicySharedWriteAllowed FileClerkDirChainPolicy = "shared-write-allowed"

	// FileClerkDirChainPolicySharedWriteRefused rejects a component that
	// is writable by group or others. A sticky component is allowed. The
	// rejection fails with ErrDirectoryWritableByOthers.
	FileClerkDirChainPolicySharedWriteRefused FileClerkDirChainPolicy = "shared-write-refused"

	FileClerkOverwritePolicyStrictRefuse FileClerkOverwritePolicy = "strict-refuse"
	FileClerkOverwritePolicyReplace      FileClerkOverwritePolicy = "replace"

	FileClerkOwnerSourceExistingFile        FileClerkOwnerSource = "existing-file"
	FileClerkOwnerSourceContainingDirectory FileClerkOwnerSource = "containing-directory"
	FileClerkOwnerSourceRunningProcess      FileClerkOwnerSource = "running-process"

	ErrSourceFileMissing            = errors.New("SourceFileNotFound")
	ErrTargetFileExists             = errors.New("TargetFileAlreadyExists")
	ErrFileMissing                  = errors.New("FileNotFound")
	ErrDirMissing                   = errors.New("DirNotFound")
	ErrGlobPatternInvalid           = errors.New("GlobPatternInvalid")
	ErrStartingPathMissing          = errors.New("StartingPathOrPatternMustBeSet")
	ErrStartingPathConflict         = errors.New("StartingPathConflictsWithStartingPathPattern")
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

// TouchFile behaves like touch(1). A dangling symlink fails with
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

// MoveFile never replaces an existing target. Cross-device moves
// fall back to copy and delete. An interrupted run leaves a partial
// target next to an untouched source.
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

// OverwriteFile atomically replaces the underlying file of targetPath
// with the content of sourcePath. Symlink targets are written
// through. The original file and every other reference stay in place.
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

// FileContentRegexFindings holds one regex match and its capture
// groups. LineNumRange is [start, end] inclusive. A single-line match
// has start == end.
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

	// DO NOT REPLACE with FindAllStringSubmatchIndex. Extracting match
	// text and capture groups from raw index pairs needs manual slice
	// arithmetic and makes the code unreadable. The gain is about a
	// millisecond on a 10 MiB file.
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

// FileContentRegexSearch finds every regex match in a file. Each
// result has capture groups and a 1-based inclusive line range. Use
// the (?m) flag for per-line anchors. Files at or above
// RegexLargeFileThresholdBytes stream line by line, so a multi-line
// pattern matches within one line only.
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
	// Chmod last. chown clears the setuid and setgid bits.
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

type fileStatSnapshot struct {
	Exists       bool
	IsSymlink    bool
	IsDirectory  bool
	OwnerUserId  tkValueObject.UnixUserId
	OwnerGroupId tkValueObject.UnixGroupId
	Permissions  os.FileMode
	SizeBytes    int64
	DeviceId     uint64
	InodeId      uint64
	ModifiedAt   time.Time
}

func (clerk FileClerk) fileSnapshotReader(
	dirHandle int,
	entryName string,
) (snapshot fileStatSnapshot, err error) {
	entryStat := unix.Stat_t{}
	statErr := unix.Fstatat(
		dirHandle, entryName, &entryStat, unix.AT_SYMLINK_NOFOLLOW,
	)
	if statErr != nil {
		if errors.Is(statErr, unix.ENOENT) {
			return snapshot, nil
		}
		return snapshot, statErr
	}

	ownerUserId, userIdErr := tkValueObject.NewUnixUserId(entryStat.Uid)
	if userIdErr != nil {
		return snapshot, userIdErr
	}
	ownerGroupId, groupIdErr := tkValueObject.NewUnixGroupId(entryStat.Gid)
	if groupIdErr != nil {
		return snapshot, groupIdErr
	}

	entryFormat := entryStat.Mode & unix.S_IFMT
	snapshot = fileStatSnapshot{
		Exists:       true,
		IsSymlink:    entryFormat == unix.S_IFLNK,
		IsDirectory:  entryFormat == unix.S_IFDIR,
		OwnerUserId:  ownerUserId,
		OwnerGroupId: ownerGroupId,
		Permissions:  clerk.unixFileModeConverter(entryStat.Mode),
		SizeBytes:    entryStat.Size,
		DeviceId:     entryStat.Dev,
		InodeId:      entryStat.Ino,
		ModifiedAt: time.Unix(
			int64(entryStat.Mtim.Sec), int64(entryStat.Mtim.Nsec),
		),
	}

	return snapshot, nil
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

// UpdateFilePermissions never chmods through a symlink. Unlike
// chmod(2), it requires read permission on the target. Root is
// exempt.
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
// shouldKeepSourceFilePtr requests otherwise. Zip extraction
// overwrites existing files without warning. Only decompress archives
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

func (clerk FileClerk) IsSymlinkTo(sourcePath, targetPath string) bool {
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

func (FileClerk) resolveFilePathBySymlinkPolicy(
	filePath tkValueObject.UnixAbsoluteFilePath,
	symlinkPolicy FileClerkSymlinkPolicy,
) (resolvedPath tkValueObject.UnixAbsoluteFilePath, err error) {
	if symlinkPolicy != FileClerkSymlinkPolicyResolve {
		return filePath, nil
	}

	rawFilePath := filePath.String()
	fileInfo, lstatErr := os.Lstat(rawFilePath)
	entryIsSymlink := lstatErr == nil && fileInfo.Mode()&os.ModeSymlink != 0
	if entryIsSymlink {
		resolvedFilePath, evalErr := filepath.EvalSymlinks(rawFilePath)
		if os.IsNotExist(evalErr) {
			return resolvedPath, ErrFileMissing
		}
		if evalErr != nil {
			return resolvedPath, evalErr
		}
		return tkValueObject.NewUnixAbsoluteFilePath(resolvedFilePath, true)
	}

	dirPath, fileName := filepath.Split(rawFilePath)
	resolvedDirPath, evalErr := filepath.EvalSymlinks(dirPath)
	if os.IsNotExist(evalErr) {
		return resolvedPath, ErrFileMissing
	}
	if evalErr != nil {
		return resolvedPath, evalErr
	}
	return tkValueObject.NewUnixAbsoluteFilePath(
		filepath.Join(resolvedDirPath, fileName), true,
	)
}

type fileWriteTarget struct {
	DirHandle    int
	FileName     tkValueObject.UnixFileName
	StatSnapshot fileStatSnapshot
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

	targetFilePath, resolveErr := clerk.resolveFilePathBySymlinkPolicy(
		filePath, symlinkPolicy,
	)
	if resolveErr != nil {
		return target, resolveErr
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

	targetSnapshot, snapshotErr := clerk.fileSnapshotReader(
		dirHandle, targetFileName.String(),
	)
	if snapshotErr != nil {
		_ = unix.Close(dirHandle)
		return target, snapshotErr
	}

	return fileWriteTarget{
		DirHandle:    dirHandle,
		FileName:     targetFileName,
		StatSnapshot: targetSnapshot,
	}, nil
}

func (FileClerk) atomicReplacementSettingsResolver(
	target fileWriteTarget,
) atomicFileWriteSettings {
	return atomicFileWriteSettings{
		DirHandle:       target.DirHandle,
		TargetFileName:  target.FileName,
		Permissions:     target.StatSnapshot.Permissions,
		ShouldOverwrite: true,
		OwnerUserId:     &target.StatSnapshot.OwnerUserId,
		OwnerGroupId:    &target.StatSnapshot.OwnerGroupId,
	}
}

func (FileClerk) regularFileSnapshotValidator(
	inspectedSnapshot fileStatSnapshot,
) error {
	if !inspectedSnapshot.Exists {
		return ErrFileMissing
	}
	if inspectedSnapshot.IsSymlink {
		return ErrTargetIsSymlink
	}
	if inspectedSnapshot.IsDirectory {
		return ErrTargetIsDirectory
	}

	return nil
}

func (FileClerk) targetFileSwapVerifier(
	openedHandle int,
	inspectedSnapshot fileStatSnapshot,
) error {
	openedStat := unix.Stat_t{}
	statErr := unix.Fstat(openedHandle, &openedStat)
	if statErr != nil {
		return statErr
	}

	identityChanged := openedStat.Dev != inspectedSnapshot.DeviceId ||
		openedStat.Ino != inspectedSnapshot.InodeId
	if identityChanged {
		return ErrTargetFileChanged
	}

	return nil
}

func (clerk FileClerk) inspectedTargetFileOpener(
	dirHandle int,
	targetFileName tkValueObject.UnixFileName,
	inspectedSnapshot fileStatSnapshot,
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

	verifyErr := clerk.targetFileSwapVerifier(fileHandle, inspectedSnapshot)
	if verifyErr != nil {
		_ = unix.Close(fileHandle)
		return 0, verifyErr
	}

	return fileHandle, nil
}

type FileRegexReplaceSettings struct {
	FilePath tkValueObject.UnixAbsoluteFilePath

	DirChainPolicy *FileClerkDirChainPolicy
	SymlinkPolicy  *FileClerkSymlinkPolicy

	// When empty, the running process account is trusted. Root is
	// always trusted.
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
) (replacementCount, writtenBytesTotal int, err error) {
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

// FileContentRegexReplace atomically substitutes regex matches. The
// target keeps its owner, group, and mode, including special bits. A
// swap during the operation fails with ErrTargetFileChanged. Symlinks
// are refused unless the policy resolves them. A zero-byte result
// fails with ErrReplacementWouldTruncateFile. Use
// TruncateFileContent to empty a file. Files at or above the
// large-file threshold stream line by line, so a multi-line pattern
// needs a smaller file.
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

	targetSnapshot := target.StatSnapshot
	targetSnapshotErr := clerk.regularFileSnapshotValidator(targetSnapshot)
	if targetSnapshotErr != nil {
		return 0, targetSnapshotErr
	}
	if targetSnapshot.SizeBytes == 0 {
		return 0, ErrFileEmpty
	}

	fileHandle, openErr := clerk.inspectedTargetFileOpener(
		target.DirHandle, target.FileName, targetSnapshot, targetFileReadOpenFlags,
	)
	if openErr != nil {
		return 0, openErr
	}
	fileHandler := os.NewFile(uintptr(fileHandle), target.FileName.String())
	defer func() { _ = fileHandler.Close() }()

	if targetSnapshot.SizeBytes >= RegexLargeFileThresholdBytes {
		slog.Warn(
			"FileContentRegexReplaceStreamingFallback",
			slog.String("filePath", settings.FilePath.String()),
			slog.Int64("fileSizeBytes", targetSnapshot.SizeBytes),
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

	// When empty, the running process account is trusted. Root is
	// always trusted.
	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId
}

// AppendFileContent appends through an O_APPEND write, so concurrent
// writers never lose data. The target keeps its owner, group, and
// mode. A swap during the append fails with ErrTargetFileChanged. A
// missing target fails with ErrFileMissing. Create it with
// UpsertFile. Symlinks are refused unless the policy resolves them.
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

	targetSnapshot := target.StatSnapshot
	targetSnapshotErr := clerk.regularFileSnapshotValidator(targetSnapshot)
	if targetSnapshotErr != nil {
		return targetSnapshotErr
	}

	fileHandle, openErr := clerk.inspectedTargetFileOpener(
		target.DirHandle, target.FileName, targetSnapshot, targetFileAppendOpenFlags,
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

	// When empty, the running process account is trusted. Root is
	// always trusted.
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
	targetSnapshot fileStatSnapshot,
	containingDirStat unix.Stat_t,
) (ownership fileUpsertOwnership, err error) {
	switch ownerSource {
	case FileClerkOwnerSourceExistingFile:
		if targetSnapshot.Exists {
			ownership = fileUpsertOwnership{
				UserId:  targetSnapshot.OwnerUserId,
				GroupId: targetSnapshot.OwnerGroupId,
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
	targetSnapshot fileStatSnapshot,
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
		ownerSource, targetSnapshot, containingDirStat,
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
	targetSnapshot fileStatSnapshot,
) os.FileMode {
	if statedPermissionsPtr != nil {
		return *statedPermissionsPtr
	}
	if targetSnapshot.Exists {
		return targetSnapshot.Permissions
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

	targetSnapshot := target.StatSnapshot
	shouldRefuseSymlink := targetSnapshot.IsSymlink &&
		symlinkPolicy == FileClerkSymlinkPolicyStrictRefuse
	if shouldRefuseSymlink {
		return ErrTargetIsSymlink
	}
	if targetSnapshot.IsDirectory {
		return ErrTargetIsDirectory
	}

	shouldOverwrite := overwritePolicy == FileClerkOverwritePolicyReplace
	if targetSnapshot.Exists && !shouldOverwrite {
		return ErrTargetFileExists
	}

	permissions := clerk.fileUpsertPermissionsResolver(
		settings.Permissions, targetSnapshot,
	)

	containingDirStat := unix.Stat_t{}
	containingDirStatErr := unix.Fstat(target.DirHandle, &containingDirStat)
	if containingDirStatErr != nil {
		return containingDirStatErr
	}

	ownership, ownerErr := clerk.fileUpsertOwnerResolver(
		ownerSource, settings.OwnerUsername, settings.OwnerUserId,
		settings.OwnerGroupId, targetSnapshot, containingDirStat,
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

type ownerNameCache struct {
	usernames       map[tkValueObject.UnixUserId]tkValueObject.UnixUsername
	groupNames      map[tkValueObject.UnixGroupId]tkValueObject.UnixGroupName
	usernameErrors  map[tkValueObject.UnixUserId]error
	groupNameErrors map[tkValueObject.UnixGroupId]error
}

type dirReadSettings struct {
	SymlinkPolicy  FileClerkSymlinkPolicy
	DirChainPolicy FileClerkDirChainPolicy
	MaxFiles       uint64
	OwnerNameCache *ownerNameCache

	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId
}

func (clerk FileClerk) dirReadSettingsResolver(
	symlinkPolicyPtr *FileClerkSymlinkPolicy,
	dirChainPolicyPtr *FileClerkDirChainPolicy,
	maxFilesPtr *uint64,
	trustedDirOwnerUsernames []tkValueObject.UnixUsername,
	trustedDirOwnerUserIds []tkValueObject.UnixUserId,
) (readSettings dirReadSettings, err error) {
	symlinkPolicy, symlinkPolicyErr := clerk.symlinkPolicyNormalizer(
		symlinkPolicyPtr,
	)
	if symlinkPolicyErr != nil {
		return readSettings, symlinkPolicyErr
	}
	dirChainPolicy, dirChainPolicyErr := clerk.dirChainPolicyNormalizer(
		dirChainPolicyPtr,
	)
	if dirChainPolicyErr != nil {
		return readSettings, dirChainPolicyErr
	}

	readSettings = dirReadSettings{
		SymlinkPolicy:  *symlinkPolicy,
		DirChainPolicy: *dirChainPolicy,
		MaxFiles:       FileClerkDefaultMaxFiles,
		OwnerNameCache: &ownerNameCache{
			usernames:       map[tkValueObject.UnixUserId]tkValueObject.UnixUsername{},
			groupNames:      map[tkValueObject.UnixGroupId]tkValueObject.UnixGroupName{},
			usernameErrors:  map[tkValueObject.UnixUserId]error{},
			groupNameErrors: map[tkValueObject.UnixGroupId]error{},
		},

		TrustedDirOwnerUsernames: trustedDirOwnerUsernames,
		TrustedDirOwnerUserIds:   trustedDirOwnerUserIds,
	}
	if maxFilesPtr != nil {
		readSettings.MaxFiles = *maxFilesPtr
	}

	return readSettings, nil
}

func (clerk FileClerk) openReadableDirHandler(
	dirPath tkValueObject.UnixAbsoluteFilePath,
	readSettings dirReadSettings,
) (dirFile *os.File, err error) {
	if dirPath == "" {
		return nil, ErrDirMissing
	}

	targetPath, resolveErr := clerk.resolveFilePathBySymlinkPolicy(
		dirPath, readSettings.SymlinkPolicy,
	)
	if resolveErr != nil {
		if errors.Is(resolveErr, ErrFileMissing) {
			return nil, ErrDirMissing
		}
		return nil, resolveErr
	}

	trustedDirOwnerIds, trustErr := clerk.trustedDirOwnerIdSetResolver(
		readSettings.TrustedDirOwnerUsernames,
		readSettings.TrustedDirOwnerUserIds,
	)
	if trustErr != nil {
		return nil, trustErr
	}

	verifiedHandle, chainErr := clerk.openRedirectProofDirChain(
		targetPath, trustedDirOwnerIds, readSettings.DirChainPolicy,
	)
	if chainErr != nil {
		if errors.Is(chainErr, unix.ENOENT) {
			return nil, ErrDirMissing
		}
		return nil, chainErr
	}
	defer func() { _ = unix.Close(verifiedHandle) }()

	readableHandle, openErr := unix.Openat(
		verifiedHandle, ".", dirReadOpenFlags, 0,
	)
	if openErr != nil {
		if errors.Is(openErr, unix.ENOENT) {
			return nil, ErrDirMissing
		}
		return nil, openErr
	}

	return os.NewFile(uintptr(readableHandle), targetPath.String()), nil
}

func (FileClerk) cappedDirEntriesReader(
	dirFile *os.File,
	maxFiles uint64,
) (rawEntries []os.DirEntry, err error) {
	const (
		readAllDirEntriesRequestCount        = -1
		maxDirEntryReadRequestCount   uint64 = math.MaxInt32
	)

	if maxFiles == 0 {
		return dirFile.ReadDir(readAllDirEntriesRequestCount)
	}

	readRequestCount := int(min(maxFiles, maxDirEntryReadRequestCount))
	chunkEntries, readErr := dirFile.ReadDir(readRequestCount)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return nil, readErr
	}

	return chunkEntries, nil
}

func (clerk FileClerk) inspectedSubDirOpener(
	dirHandle int,
	entryName string,
	inspectedSnapshot fileStatSnapshot,
) (subDirHandle int, err error) {
	subDirHandle, openErr := unix.Openat(
		dirHandle, entryName, dirReadOpenFlags, 0,
	)
	if openErr != nil {
		entryNoLongerDirectory := errors.Is(openErr, unix.ELOOP) ||
			errors.Is(openErr, unix.ENOTDIR)
		if entryNoLongerDirectory {
			return 0, fmt.Errorf("%w: %w", ErrTargetFileChanged, openErr)
		}
		return 0, openErr
	}

	verifyErr := clerk.targetFileSwapVerifier(subDirHandle, inspectedSnapshot)
	if verifyErr != nil {
		_ = unix.Close(subDirHandle)
		return 0, verifyErr
	}

	return subDirHandle, nil
}

func (FileClerk) usernameByUserIdResolver(
	userId tkValueObject.UnixUserId,
) (username tkValueObject.UnixUsername, err error) {
	account, lookupErr := user.LookupId(userId.String())
	if lookupErr != nil {
		return username, fmt.Errorf(
			"UsernameLookupFailed: %s: %w", userId.String(), lookupErr,
		)
	}

	return tkValueObject.NewUnixUsername(account.Username)
}

func (FileClerk) groupNameByGroupIdResolver(
	groupId tkValueObject.UnixGroupId,
) (groupName tkValueObject.UnixGroupName, err error) {
	group, lookupErr := user.LookupGroupId(groupId.String())
	if lookupErr != nil {
		return groupName, fmt.Errorf(
			"GroupNameLookupFailed: %s: %w", groupId.String(), lookupErr,
		)
	}

	return tkValueObject.NewUnixGroupName(group.Name)
}

func (FileClerk) unixFilePermissionsFormatter(
	fileMode os.FileMode,
) (tkValueObject.UnixFilePermissions, error) {
	permissionBits := fileMode.Perm()
	if fileMode&os.ModeSetuid != 0 {
		permissionBits |= unix.S_ISUID
	}
	if fileMode&os.ModeSetgid != 0 {
		permissionBits |= unix.S_ISGID
	}
	if fileMode&os.ModeSticky != 0 {
		permissionBits |= unix.S_ISVTX
	}

	return tkValueObject.NewUnixFilePermissions(
		fmt.Sprintf("%03o", permissionBits),
	)
}

func (clerk FileClerk) cachedUsernameResolver(
	ownerNames *ownerNameCache,
	ownerUserId tkValueObject.UnixUserId,
) (username tkValueObject.UnixUsername, err error) {
	username, isCached := ownerNames.usernames[ownerUserId]
	if isCached {
		return username, nil
	}
	cachedLookupErr, hasFailedBefore := ownerNames.usernameErrors[ownerUserId]
	if hasFailedBefore {
		return username, cachedLookupErr
	}

	username, err = clerk.usernameByUserIdResolver(ownerUserId)
	if err != nil {
		ownerNames.usernameErrors[ownerUserId] = err
		return username, err
	}
	ownerNames.usernames[ownerUserId] = username

	return username, nil
}

func (clerk FileClerk) cachedGroupNameResolver(
	ownerNames *ownerNameCache,
	ownerGroupId tkValueObject.UnixGroupId,
) (groupName tkValueObject.UnixGroupName, err error) {
	groupName, isCached := ownerNames.groupNames[ownerGroupId]
	if isCached {
		return groupName, nil
	}
	cachedLookupErr, hasFailedBefore := ownerNames.groupNameErrors[ownerGroupId]
	if hasFailedBefore {
		return groupName, cachedLookupErr
	}

	groupName, err = clerk.groupNameByGroupIdResolver(ownerGroupId)
	if err != nil {
		ownerNames.groupNameErrors[ownerGroupId] = err
		return groupName, err
	}
	ownerNames.groupNames[ownerGroupId] = groupName

	return groupName, nil
}

func (clerk FileClerk) unixFileEntityFactory(
	entryName, entryPath string,
	entrySnapshot fileStatSnapshot,
	ownerNames *ownerNameCache,
) (unixFile tkEntity.UnixFile, err error) {
	fileName, nameErr := tkValueObject.NewUnixFileName(entryName, true)
	if nameErr != nil {
		return unixFile, nameErr
	}
	filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(entryPath, true)
	if pathErr != nil {
		return unixFile, pathErr
	}

	var fileExtensionPtr *tkValueObject.UnixFileExtension
	fileMimeType := tkValueObject.MimeTypeGeneric
	fileExtension, extensionErr := filePath.ReadFileExtension()
	if extensionErr == nil && fileExtension != "" {
		fileExtensionPtr = &fileExtension
		fileMimeType = fileExtension.ReadMimeType()
	}
	if entrySnapshot.IsDirectory {
		fileMimeType = tkValueObject.MimeTypeDirectory
		fileExtensionPtr = nil
	}

	fileSize, sizeErr := tkValueObject.NewByte(entrySnapshot.SizeBytes)
	if sizeErr != nil {
		return unixFile, sizeErr
	}

	filePermissions, permissionsErr := clerk.unixFilePermissionsFormatter(
		entrySnapshot.Permissions,
	)
	if permissionsErr != nil {
		return unixFile, permissionsErr
	}

	fileOwner, ownerErr := clerk.cachedUsernameResolver(
		ownerNames, entrySnapshot.OwnerUserId,
	)
	if ownerErr != nil {
		return unixFile, ownerErr
	}
	fileGroup, groupErr := clerk.cachedGroupNameResolver(
		ownerNames, entrySnapshot.OwnerGroupId,
	)
	if groupErr != nil {
		return unixFile, groupErr
	}

	fileUpdatedAt := tkValueObject.NewUnixTimeWithGoTime(entrySnapshot.ModifiedAt)

	return tkEntity.NewUnixFile(
		fileName, filePath, fileMimeType,
		filePermissions,
		fileSize, fileExtensionPtr, entrySnapshot.OwnerUserId, fileOwner,
		entrySnapshot.OwnerGroupId, fileGroup, fileUpdatedAt,
		entrySnapshot.IsSymlink,
	), nil
}

func (clerk FileClerk) collectDirTreeEntities(
	dirFile *os.File,
	remainingDepth uint16,
	readSettings dirReadSettings,
	namePattern *regexp.Regexp,
	collectedFiles []tkEntity.UnixFile,
) (walkedFiles []tkEntity.UnixFile, err error) {
	dirPath := dirFile.Name()
	rawEntries, readErr := clerk.cappedDirEntriesReader(
		dirFile, readSettings.MaxFiles,
	)
	if readErr != nil {
		return nil, readErr
	}

	dirHandle := int(dirFile.Fd())
	for _, rawEntry := range rawEntries {
		if readSettings.MaxFiles > 0 &&
			uint64(len(collectedFiles)) >= readSettings.MaxFiles {
			return collectedFiles, nil
		}

		entryName := rawEntry.Name()
		entryPath := filepath.Join(dirPath, entryName)
		entrySnapshot, snapshotErr := clerk.fileSnapshotReader(
			dirHandle, entryName,
		)
		if snapshotErr != nil {
			slog.Warn(
				"DirEntryStatFailed",
				slog.String("entryPath", entryPath),
				slog.String("reason", snapshotErr.Error()),
			)
			continue
		}
		if !entrySnapshot.Exists {
			slog.Debug(
				"DirEntryVanishedMidWalk",
				slog.String("dirPath", dirPath),
				slog.String("entryName", entryName),
			)
			continue
		}

		if namePattern == nil || namePattern.MatchString(entryName) {
			unixFile, entityErr := clerk.unixFileEntityFactory(
				entryName, entryPath, entrySnapshot,
				readSettings.OwnerNameCache,
			)
			if entityErr != nil {
				slog.Debug(
					"UnixFileEntitySkipped",
					slog.String("entryPath", entryPath),
					slog.String("reason", entityErr.Error()),
				)
			}
			if entityErr == nil {
				collectedFiles = append(collectedFiles, unixFile)
			}
		}

		if !entrySnapshot.IsDirectory || remainingDepth == 0 {
			continue
		}

		subHandle, openErr := clerk.inspectedSubDirOpener(
			dirHandle, entryName, entrySnapshot,
		)
		if openErr != nil {
			if errors.Is(openErr, unix.ENOENT) {
				slog.Debug(
					"DirEntryVanishedMidWalk",
					slog.String("dirPath", dirPath),
					slog.String("entryName", entryName),
				)
				continue
			}
			if errors.Is(openErr, ErrTargetFileChanged) {
				return nil, fmt.Errorf(
					"DirEntryChanged: %s: %w", entryPath, openErr,
				)
			}
			slog.Warn(
				"DirOpenFailed",
				slog.String("entryPath", entryPath),
				slog.String("reason", openErr.Error()),
			)
			continue
		}
		subDirFile := os.NewFile(uintptr(subHandle), entryPath)
		subWalkedFiles, walkErr := clerk.collectDirTreeEntities(
			subDirFile, remainingDepth-1, readSettings, namePattern,
			collectedFiles,
		)
		closeErr := subDirFile.Close()
		if walkErr != nil || closeErr != nil {
			return nil, errors.Join(walkErr, closeErr)
		}
		collectedFiles = subWalkedFiles
	}

	return collectedFiles, nil
}

func (clerk FileClerk) dirTreeEntityReader(
	rootDirPath tkValueObject.UnixAbsoluteFilePath,
	maxDepth uint16,
	readSettings dirReadSettings,
	namePatternPtr *regexp.Regexp,
) (unixFiles []tkEntity.UnixFile, err error) {
	dirFile, dirErr := clerk.openReadableDirHandler(rootDirPath, readSettings)
	if dirErr != nil {
		return nil, dirErr
	}
	defer func() { _ = dirFile.Close() }()

	return clerk.collectDirTreeEntities(
		dirFile, maxDepth, readSettings, namePatternPtr,
		[]tkEntity.UnixFile{},
	)
}

type FileListDirSettings struct {
	DirPath  tkValueObject.UnixAbsoluteFilePath
	MaxDepth uint16

	DirChainPolicy *FileClerkDirChainPolicy
	SymlinkPolicy  *FileClerkSymlinkPolicy

	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId

	MaxFiles *uint64
}

// ListDir reads a directory tree into UnixFile entities sorted by
// entry name. The path breaks ties. MaxDepth 0 reads only the root's
// entries. MaxDepth N reads N levels of subdirectories. MaxFiles 0
// reads without a cap. A small MaxFiles truncates silently. An
// unreadable entry is skipped with a warning. A subdirectory replaced
// mid-walk fails with ErrTargetFileChanged.
func (clerk FileClerk) ListDir(
	settings FileListDirSettings,
) (unixFiles []tkEntity.UnixFile, err error) {
	readSettings, readSettingsErr := clerk.dirReadSettingsResolver(
		settings.SymlinkPolicy, settings.DirChainPolicy, settings.MaxFiles,
		settings.TrustedDirOwnerUsernames, settings.TrustedDirOwnerUserIds,
	)
	if readSettingsErr != nil {
		return nil, readSettingsErr
	}

	unixFiles, readErr := clerk.dirTreeEntityReader(
		settings.DirPath, settings.MaxDepth, readSettings, nil,
	)
	if readErr != nil {
		return nil, readErr
	}

	slices.SortFunc(unixFiles, func(first, second tkEntity.UnixFile) int {
		nameComparison := strings.Compare(
			first.Name.String(), second.Name.String(),
		)
		if nameComparison != 0 {
			return nameComparison
		}
		return strings.Compare(first.Path.String(), second.Path.String())
	})

	return unixFiles, nil
}

type FileFindSettings struct {
	StartingPath        tkValueObject.UnixAbsoluteFilePath
	StartingPathPattern tkValueObject.UnixAbsoluteGlobPath
	MaxDepth            uint16

	DirChainPolicy *FileClerkDirChainPolicy
	SymlinkPolicy  *FileClerkSymlinkPolicy

	TrustedDirOwnerUsernames []tkValueObject.UnixUsername
	TrustedDirOwnerUserIds   []tkValueObject.UnixUserId

	MaxFiles    *uint64
	NamePattern *regexp.Regexp
}

func (FileClerk) startingDirPathsResolver(
	startingPath tkValueObject.UnixAbsoluteFilePath,
	startingPathPattern tkValueObject.UnixAbsoluteGlobPath,
) (startingDirPaths []tkValueObject.UnixAbsoluteFilePath, err error) {
	startingPathIsSet := startingPath != ""
	patternIsSet := startingPathPattern != ""
	if startingPathIsSet && patternIsSet {
		return nil, ErrStartingPathConflict
	}
	if !startingPathIsSet && !patternIsSet {
		return nil, ErrStartingPathMissing
	}

	patternHasWildcard := startingPathPattern.HasWildcardChars()
	var candidatePaths []string
	if patternIsSet {
		patternStr := startingPathPattern.String()
		matchedPaths, globErr := filepath.Glob(patternStr)
		if globErr != nil {
			return nil, fmt.Errorf("%w: %s", ErrGlobPatternInvalid, patternStr)
		}

		if len(matchedPaths) == 0 {
			if patternHasWildcard {
				return []tkValueObject.UnixAbsoluteFilePath{}, nil
			}
			return nil, ErrDirMissing
		}
		candidatePaths = matchedPaths
	}
	if startingPathIsSet {
		_, statErr := os.Stat(startingPath.String())
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return nil, ErrDirMissing
			}
			return nil, statErr
		}
		candidatePaths = []string{startingPath.String()}
	}

	for _, candidatePath := range candidatePaths {
		candidateInfo, statErr := os.Stat(candidatePath)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				slog.Debug(
					"StartingDirVanished",
					slog.String("candidatePath", candidatePath),
				)
				continue
			}
			if !patternHasWildcard {
				return nil, statErr
			}
			slog.Warn(
				"StartingDirStatFailed",
				slog.String("candidatePath", candidatePath),
				slog.String("reason", statErr.Error()),
			)
			continue
		}
		if !candidateInfo.IsDir() {
			continue
		}

		startingDirPath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
			candidatePath, true,
		)
		if pathErr != nil {
			slog.Debug(
				"StartingDirPathRejected",
				slog.String("candidatePath", candidatePath),
				slog.String("reason", pathErr.Error()),
			)
			continue
		}
		startingDirPaths = append(startingDirPaths, startingDirPath)
	}
	slices.SortFunc(
		startingDirPaths,
		func(first, second tkValueObject.UnixAbsoluteFilePath) int {
			return strings.Compare(first.String(), second.String())
		},
	)

	return startingDirPaths, nil
}

// Find walks every directory that StartingPath or
// StartingPathPattern matches. It returns UnixFile entities sorted by
// path. The starting directory is not in the results. Exactly one of
// the two starting points must be set. A missing starting path fails
// with ErrDirMissing. A wildcard that matches nothing returns an
// empty result. Non-directories are skipped. A wildcard match that
// cannot be read is skipped with a warning. A directly named path
// fails with its read error. NamePattern matches the entry name. A
// nil NamePattern matches every entry. MaxFiles 0 reads without a
// cap. The cap counts across all matched starting directories. An
// unreadable entry is skipped with a warning. A subdirectory replaced
// mid-walk fails with ErrTargetFileChanged.
func (clerk FileClerk) Find(
	settings FileFindSettings,
) (unixFiles []tkEntity.UnixFile, err error) {
	readSettings, readSettingsErr := clerk.dirReadSettingsResolver(
		settings.SymlinkPolicy, settings.DirChainPolicy, settings.MaxFiles,
		settings.TrustedDirOwnerUsernames, settings.TrustedDirOwnerUserIds,
	)
	if readSettingsErr != nil {
		return nil, readSettingsErr
	}

	startingDirPaths, startingDirPathsErr := clerk.startingDirPathsResolver(
		settings.StartingPath, settings.StartingPathPattern,
	)
	if startingDirPathsErr != nil {
		return nil, startingDirPathsErr
	}

	patternHasWildcard := settings.StartingPathPattern.HasWildcardChars()
	unixFiles = []tkEntity.UnixFile{}
	remainingMaxFiles := readSettings.MaxFiles
	for _, startingDirPath := range startingDirPaths {
		maxFilesIsCapped := readSettings.MaxFiles > 0
		if maxFilesIsCapped && remainingMaxFiles == 0 {
			break
		}

		startingDirReadSettings := readSettings
		startingDirReadSettings.MaxFiles = remainingMaxFiles
		startingDirUnixFiles, readErr := clerk.dirTreeEntityReader(
			startingDirPath, settings.MaxDepth, startingDirReadSettings,
			settings.NamePattern,
		)
		if readErr != nil {
			readErrorIsFatal := !patternHasWildcard ||
				errors.Is(readErr, ErrTargetFileChanged) ||
				errors.Is(readErr, ErrDirMissing)
			if readErrorIsFatal {
				return nil, readErr
			}
			slog.Warn(
				"StartingDirReadFailed",
				slog.String("startingDirPath", startingDirPath.String()),
				slog.String("reason", readErr.Error()),
			)
			continue
		}
		unixFiles = append(unixFiles, startingDirUnixFiles...)
		if maxFilesIsCapped {
			remainingMaxFiles -= uint64(len(startingDirUnixFiles))
		}
	}

	slices.SortFunc(unixFiles, func(first, second tkEntity.UnixFile) int {
		return strings.Compare(first.Path.String(), second.Path.String())
	})

	return unixFiles, nil
}
