package tkInfra

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestFileExists(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("ExistingFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "existing.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.FileExists(testFile) {
			subtest.Errorf("FileExistsShouldReturnTrue: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("NonExistentFile", func(subtest *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		if clerk.FileExists(nonExistentFile) {
			subtest.Errorf("FileExistsShouldReturnFalse: %s", nonExistentFile)
		}
	})

	test.Run("ExistingDirectory", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		if !clerk.FileExists(testDir) {
			subtest.Errorf("FileExistsShouldReturnTrueForDir: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})
}

func TestIsFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("RegularFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			subtest.Errorf("IsFileShouldReturnTrue: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("Directory", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		if clerk.IsFile(testDir) {
			subtest.Errorf("IsFileShouldReturnFalseForDir: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})

	test.Run("NonExistentPath", func(subtest *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")
		if clerk.IsFile(nonExistentPath) {
			subtest.Errorf("IsFileShouldReturnFalseForNonExistent: %s", nonExistentPath)
		}
	})
}

func TestIsDir(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("RegularDirectory", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		if !clerk.IsDir(testDir) {
			subtest.Errorf("IsDirShouldReturnTrue: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})

	test.Run("RegularFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		if clerk.IsDir(testFile) {
			subtest.Errorf("IsDirShouldReturnFalseForFile: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("NonExistentPath", func(subtest *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")
		if clerk.IsDir(nonExistentPath) {
			subtest.Errorf("IsDirShouldReturnFalseForNonExistent: %s", nonExistentPath)
		}
	})
}

func TestTouchFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("CreateNewFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "newfile.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Errorf("TouchFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			subtest.Errorf("FileShouldExistAfterCreation: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("TouchFileInNonExistentDir", func(subtest *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent", "subdir")
		testFile := filepath.Join(nonExistentDir, "file.txt")
		err := clerk.TouchFile(testFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TouchFileInNonExistentDir")
		}
	})

	test.Run("TouchExistingFileRefreshesTimestamps", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "existing_touch.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		pastTimestamp := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		err = os.Chtimes(testFile, pastTimestamp, pastTimestamp)
		if err != nil {
			subtest.Fatalf("ChtimesFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			subtest.Errorf("TouchFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if !fileInfo.ModTime().After(pastTimestamp) {
			subtest.Errorf("TimestampsNotRefreshed: mtime %v", fileInfo.ModTime())
		}
	})

	test.Run("TouchExistingFileKeepsContent", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "kept_content.txt")

		err := clerk.WriteNewFile(testFile, "do not truncate me", 0644)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			subtest.Errorf("TouchFileFailed: %v", err)
		}

		content, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "do not truncate me" {
			subtest.Errorf("ContentWasTruncated: '%s'", content)
		}
	})

	test.Run("TouchDanglingSymlinkFails", func(subtest *testing.T) {
		targetPath := filepath.Join(tempDir, "dangling_touch_target")
		symlinkPath := filepath.Join(tempDir, "dangling_touch.txt")

		err := os.Symlink(targetPath, symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.TouchFile(symlinkPath)
		if err == nil || err.Error() != "TargetIsSymlink" {
			subtest.Errorf("MissingExpectedError: TargetIsSymlink, got %v", err)
		}
		if clerk.FileExists(targetPath) {
			subtest.Errorf("TouchFollowedDanglingSymlink")
		}
	})
}

func TestWriteNewFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("WritesContentWithExactPermissions", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "written.txt")

		err := clerk.WriteNewFile(testFile, "hello clerk", 0666)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0666 {
			subtest.Errorf(
				"PermissionsMismatch: umask must not shrink granted mode, got %04o",
				fileInfo.Mode().Perm(),
			)
		}

		content, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "hello clerk" {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", content, "hello clerk")
		}
	})

	test.Run("EmptyContentWritesEmptyFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "empty.txt")

		err := clerk.WriteNewFile(testFile, "", 0644)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Size() != 0 {
			subtest.Errorf("FileShouldBeEmpty: size %d", fileInfo.Size())
		}
	})

	test.Run("ExistingPathFails", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "occupied.txt")

		err := clerk.WriteNewFile(testFile, "first", 0644)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = clerk.WriteNewFile(testFile, "second", 0644)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}

		content, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "first" {
			subtest.Errorf("ExistingContentWasOverwritten: '%s'", content)
		}
	})

	test.Run("DanglingSymlinkPathFails", func(subtest *testing.T) {
		symlinkPath := filepath.Join(tempDir, "dangling_link.txt")

		err := os.Symlink(filepath.Join(tempDir, "nowhere"), symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.WriteNewFile(symlinkPath, "payload", 0644)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if clerk.FileExists(filepath.Join(tempDir, "nowhere")) {
			subtest.Errorf("WriteFollowedDanglingSymlink")
		}
	})
}

func TestCopyFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("CopyExistingFile", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		testContent := "test content for copy"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(testContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, targetFile)
		if err != nil {
			subtest.Errorf("CopyFileFailed: %v", err)
		}

		if !clerk.IsFile(targetFile) {
			subtest.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		targetContent, err := clerk.ReadFileContent(targetFile, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if targetContent != testContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", targetContent, testContent)
		}

		err = clerk.DeleteFile(sourceFile)
		if err != nil {
			subtest.Errorf("DeleteSourceFileFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	test.Run("CopyNonExistentFile", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.CopyFile(sourceFile, targetFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})

	test.Run("CopyToExistingTarget", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "existing_target.txt")

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, targetFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}

		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}

		err = clerk.DeleteFile(sourceFile)
		if err != nil {
			subtest.Errorf("DeleteSourceFileFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	test.Run("CopyPreservesSourceModeIncludingExecBit", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "executable.sh")
		targetFile := filepath.Join(tempDir, "executable_copy.sh")

		err := clerk.WriteNewFile(sourceFile, "#!/bin/sh\necho hi\n", 0755)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, targetFile)
		if err != nil {
			subtest.Fatalf("CopyFileFailed: %v", err)
		}

		targetInfo, statErr := os.Stat(targetFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if targetInfo.Mode().Perm() != 0755 {
			subtest.Errorf(
				"ModeMismatch: exec bit must survive copy, got %04o",
				targetInfo.Mode().Perm(),
			)
		}
	})

	test.Run("CopyToDanglingSymlinkFails", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "copy_source.txt")
		symlinkPath := filepath.Join(tempDir, "dangling_target.txt")

		err := clerk.WriteNewFile(sourceFile, "payload", 0644)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = os.Symlink(filepath.Join(tempDir, "nowhere_copy"), symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, symlinkPath)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}
		if clerk.FileExists(filepath.Join(tempDir, "nowhere_copy")) {
			subtest.Errorf("CopyFollowedDanglingSymlink")
		}
	})
}

func TestMoveFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("MoveExistingFile", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		testContent := "test content for move"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(testContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.MoveFile(sourceFile, targetFile)
		if err != nil {
			subtest.Errorf("MoveFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			subtest.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			subtest.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		targetContent, err := clerk.ReadFileContent(targetFile, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if targetContent != testContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", targetContent, testContent)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	test.Run("MoveNonExistentFile", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.MoveFile(sourceFile, targetFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})

	test.Run("MoveToExistingTargetFails", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "move_source.txt")
		targetFile := filepath.Join(tempDir, "move_existing_target.txt")

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.MoveFile(sourceFile, targetFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}

		if !clerk.IsFile(sourceFile) || !clerk.IsFile(targetFile) {
			subtest.Errorf("FailedMoveMustLeaveBothFilesIntact")
		}
	})

	test.Run("MoveToDanglingSymlinkFails", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "move_dangling_source.txt")
		symlinkPath := filepath.Join(tempDir, "move_dangling_target.txt")

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = os.Symlink(filepath.Join(tempDir, "move_dangling_nowhere"), symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.MoveFile(sourceFile, symlinkPath)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if !clerk.IsFile(sourceFile) {
			subtest.Errorf("SourceFileMustSurviveFailedMove: %s", sourceFile)
		}
	})

	test.Run("MoveFileOntoItselfSucceeds", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "move_onto_itself.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.MoveFile(testFile, testFile)
		if err != nil {
			subtest.Errorf("MoveFileOntoItselfFailed: %v", err)
		}
		if !clerk.IsFile(testFile) {
			subtest.Errorf("FileShouldStillExist: %s", testFile)
		}
	})

	test.Run("MoveFileCrossDeviceFallsBackToCopyAndDelete", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "xdev_source.txt")

		err := clerk.WriteNewFile(sourceFile, "cross device payload", 0644)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		otherDeviceBase := directoryOnDifferentDevice(subtest, tempDir)
		otherDeviceDir, mkErr := os.MkdirTemp(otherDeviceBase, "tk-xdev-")
		if mkErr != nil {
			subtest.Fatalf("MkdirTempFailed: %v", mkErr)
		}
		defer func() { _ = os.RemoveAll(otherDeviceDir) }()

		targetFile := filepath.Join(otherDeviceDir, "xdev_target.txt")

		err = clerk.MoveFile(sourceFile, targetFile)
		if err != nil {
			subtest.Fatalf("MoveFileCrossDeviceFailed: %v", err)
		}
		if clerk.FileExists(sourceFile) {
			subtest.Errorf("SourceShouldBeRemovedAfterCrossDeviceMove: %s", sourceFile)
		}

		content, readErr := clerk.ReadFileContent(targetFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "cross device payload" {
			subtest.Errorf("ContentMismatch: '%s'", content)
		}
	})
}

// directoryOnDifferentDevice finds a world-writable directory living on
// another device than referenceDir, so MoveFile must take its EXDEV path.
func directoryOnDifferentDevice(test *testing.T, referenceDir string) string {
	test.Helper()

	referenceInfo, statErr := os.Stat(referenceDir)
	if statErr != nil {
		test.Fatalf("StatFailed: %v", statErr)
	}
	referenceDevice := referenceInfo.Sys().(*syscall.Stat_t).Dev

	for _, candidate := range []string{"/var/tmp", "/dev/shm"} {
		candidateInfo, statErr := os.Stat(candidate)
		if statErr != nil {
			continue
		}
		if candidateInfo.Sys().(*syscall.Stat_t).Dev != referenceDevice {
			return candidate
		}
	}

	test.Skip("NoDirectoryOnDifferentDevice")
	return ""
}

func TestDeleteFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("DeleteExistingFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "todelete.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			subtest.Fatalf("FileShouldExistBeforeDeletion: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}

		if clerk.IsFile(testFile) {
			subtest.Errorf("FileShouldNotExistAfterDeletion: %s", testFile)
		}
	})

	test.Run("DeleteNonExistentFile", func(subtest *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		err := clerk.DeleteFile(nonExistentFile)
		if err != nil {
			subtest.Errorf("DeleteNonExistentFileFailed: %v", err)
		}
	})

	test.Run("DeleteDirectoryFails", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "undeletable_dir")

		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.DeleteFile(testDir)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetIsDirectory")
		}
		if err != nil && err.Error() != "TargetIsDirectory" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetIsDirectory", err.Error())
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})

	test.Run("DeleteSymlinkRemovesLinkOnly", func(subtest *testing.T) {
		targetFile := filepath.Join(tempDir, "symlink_target_delete.txt")
		symlinkPath := filepath.Join(tempDir, "symlink_delete.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.Symlink(targetFile, symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(symlinkPath)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}

		if clerk.IsSymlink(symlinkPath) {
			subtest.Errorf("SymlinkShouldBeGone: %s", symlinkPath)
		}
		if !clerk.IsFile(targetFile) {
			subtest.Errorf("SymlinkTargetMustSurvive: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestReadFileContent(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("ReadExistingFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "content.txt")
		expectedContent := "Hello, World!\nThis is test content."

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(expectedContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != expectedContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, expectedContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("ReadNonExistentFile", func(subtest *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := clerk.ReadFileContent(nonExistentFile, nil)
		if err == nil {
			subtest.Errorf("MissingExpectedError: FileNotFound")
		}

		if err != nil && err.Error() != "FileNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "FileNotFound", err.Error())
		}
	})

	test.Run("ReadExceedingMaxContentSizeFails", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "limited.txt")
		fullContent := "This is a longer content that should be truncated when reading with size limit."
		maxSize := int64(20)

		err := os.WriteFile(testFile, []byte(fullContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		_, err = clerk.ReadFileContent(testFile, &maxSize)
		if err == nil {
			subtest.Fatalf("MissingExpectedError: FileTooLarge")
		}
		if err.Error() != "FileTooLarge" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "FileTooLarge", err.Error())
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("ReadExactlyAtMaxContentSize", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "exact_limit.txt")
		fullContent := "exactly-twenty-chars"
		maxSize := int64(len(fullContent))

		err := os.WriteFile(testFile, []byte(fullContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, &maxSize)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}
		if actualContent != fullContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, fullContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("ReadWithNegativeMaxSizeFails", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "negative_limit.txt")
		negativeSize := int64(-1)

		err := os.WriteFile(testFile, []byte("content"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		_, err = clerk.ReadFileContent(testFile, &negativeSize)
		if err == nil || err.Error() != "FileTooLarge" {
			subtest.Errorf("WrongErrorForNegativeCap: %v", err)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("ReadWithDefaultMaxSize", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "default_size.txt")
		testContent := "Content with default size limit"

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != testContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, testContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("ReadSymlinkToFile", func(subtest *testing.T) {
		targetFile := filepath.Join(tempDir, "symlink_target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")
		expectedContent := "content behind the symlink"

		err := clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = os.WriteFile(targetFile, []byte(expectedContent), 0644)
		if err != nil {
			subtest.Fatalf("UpdateTargetFileContentFailed: %v", err)
		}

		err = os.Symlink(targetFile, symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(symlinkPath, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != expectedContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, expectedContent)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			subtest.Errorf("RemoveSymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	test.Run("ReadDanglingSymlink", func(subtest *testing.T) {
		deletedTarget := filepath.Join(tempDir, "deleted_target.txt")
		symlinkPath := filepath.Join(tempDir, "dangling.txt")

		err := clerk.TouchFile(deletedTarget)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.Symlink(deletedTarget, symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(deletedTarget)
		if err != nil {
			subtest.Fatalf("DeleteTargetFileFailed: %v", err)
		}

		_, err = clerk.ReadFileContent(symlinkPath, nil)
		if err == nil {
			subtest.Errorf("MissingExpectedError: FileNotFound")
		}

		if err != nil && err.Error() != "FileNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "FileNotFound", err.Error())
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			subtest.Errorf("RemoveSymlinkFailed: %v", err)
		}
	})
}

func TestAppendFileContent(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("AppendsToExistingFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "append.txt")
		initialContent := "Initial"
		appendContent := " Appended"

		err := os.WriteFile(testFile, []byte(initialContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, testFile),
		}, appendContent)
		if err != nil {
			subtest.Fatalf("AppendFileContentFailed: %v", err)
		}

		actualContent, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		expectedContent := initialContent + appendContent
		if actualContent != expectedContent {
			subtest.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, expectedContent)
		}
	})

	test.Run("MissingFileReturnsFileNotFound", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "missing.txt")

		err := clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, testFile),
		}, "content")
		if !errors.Is(err, ErrFileMissing) {
			subtest.Errorf("ExpectedErrFileMissing, got: %v", err)
		}
		if clerk.FileExists(testFile) {
			subtest.Errorf("MissingTargetWasCreated")
		}
	})

	test.Run("KeepsTargetMode", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "mode.txt")
		err := os.WriteFile(testFile, []byte("Initial"), 0640)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, testFile),
		}, " Appended")
		if err != nil {
			subtest.Fatalf("AppendFileContentFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0640 {
			subtest.Errorf("ModeChanged: expected 0640, got %o", fileInfo.Mode().Perm())
		}
	})

	test.Run("KeepsTargetOwner", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			subtest.Fatalf("GidParseFailed: %v", gidErr)
		}

		testFile := filepath.Join(tempDir, "owned.txt")
		err := os.WriteFile(testFile, []byte("Initial"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		err = os.Chown(testFile, nobodyUid, nobodyGid)
		if err != nil {
			subtest.Fatalf("ChownFailed: %v", err)
		}

		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, testFile),
		}, " Appended")
		if err != nil {
			subtest.Fatalf("AppendFileContentFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			subtest.Errorf(
				"OwnerChanged: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("RefusesSymlinkTargetByDefault", func(subtest *testing.T) {
		realFile := filepath.Join(tempDir, "symlinkReal.txt")
		err := os.WriteFile(realFile, []byte("real"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		linkFile := filepath.Join(tempDir, "symlinkLink.txt")
		err = os.Symlink(realFile, linkFile)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, linkFile),
		}, " appended")
		if !errors.Is(err, ErrTargetIsSymlink) {
			subtest.Errorf("ExpectedErrTargetIsSymlink, got: %v", err)
		}
	})

	test.Run("ResolvesSymlinkTargetWhenRequested", func(subtest *testing.T) {
		realFile := filepath.Join(tempDir, "resolveReal.txt")
		err := os.WriteFile(realFile, []byte("real"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		linkFile := filepath.Join(tempDir, "resolveLink.txt")
		err = os.Symlink(realFile, linkFile)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		symlinkPolicy := FileClerkSymlinkPolicyResolve
		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath:      absoluteFilePathForTest(subtest, linkFile),
			SymlinkPolicy: &symlinkPolicy,
		}, " appended")
		if err != nil {
			subtest.Fatalf("AppendFileContentFailed: %v", err)
		}

		actualContent, readErr := clerk.ReadFileContent(realFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if actualContent != "real appended" {
			subtest.Errorf("ContentMismatch: '%s'", actualContent)
		}
	})

	test.Run("DanglingSymlinkWithResolveReturnsFileNotFound", func(subtest *testing.T) {
		linkFile := filepath.Join(tempDir, "danglingLink.txt")
		err := os.Symlink(filepath.Join(tempDir, "missingTarget.txt"), linkFile)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		symlinkPolicy := FileClerkSymlinkPolicyResolve
		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath:      absoluteFilePathForTest(subtest, linkFile),
			SymlinkPolicy: &symlinkPolicy,
		}, "content")
		if !errors.Is(err, ErrFileMissing) {
			subtest.Errorf("ExpectedErrFileMissing, got: %v", err)
		}
	})

	test.Run("RefusesDirectoryTarget", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "appendDir")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, testDir),
		}, "content")
		if !errors.Is(err, ErrTargetIsDirectory) {
			subtest.Errorf("ExpectedErrTargetIsDirectory, got: %v", err)
		}
	})

	test.Run("RefusesFifoWithoutReader", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "append.fifo")
		fifoErr := unix.Mkfifo(testFile, 0o600)
		if fifoErr != nil {
			subtest.Fatalf("MkfifoFailed: %v", fifoErr)
		}
		targetPath := absoluteFilePathForTest(subtest, testFile)

		appendResultChan := make(chan error, 1)
		go func() {
			appendResultChan <- clerk.AppendFileContent(
				FileAppendSettings{FilePath: targetPath}, "content",
			)
		}()

		select {
		case appendErr := <-appendResultChan:
			if !errors.Is(appendErr, ErrTargetNotRegularFile) {
				subtest.Fatalf("ExpectedErrTargetNotRegularFile, got: %v", appendErr)
			}
		case <-time.After(2 * time.Second):
			subtest.Fatalf("AppendBlockedOnFifo")
		}
	})

	test.Run("RefusesWritableDirChainWhenRequired", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "writableAppendDir")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chmod(testDir, 0o777)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}
		testFile := filepath.Join(testDir, "append.txt")
		err = os.WriteFile(testFile, []byte("Initial"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused
		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath:       absoluteFilePathForTest(subtest, testFile),
			DirChainPolicy: &sharedWriteRefusedPolicy,
		}, " Appended")
		if !errors.Is(err, ErrDirectoryWritableByOthers) {
			subtest.Errorf("ExpectedErrDirectoryWritableByOthers, got: %v", err)
		}
	})

	test.Run("ReportsUnknownTrustedDirOwner", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "unknownOwner.txt")
		err := os.WriteFile(testFile, []byte("Initial"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		unknownUsername, usernameErr := tkValueObject.NewUnixUsername(
			"no-such-user-xyz-42",
		)
		if usernameErr != nil {
			subtest.Fatalf("UnknownUsernameInvalid: %v", usernameErr)
		}

		err = clerk.AppendFileContent(FileAppendSettings{
			FilePath: absoluteFilePathForTest(subtest, testFile),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				unknownUsername,
			},
		}, " Appended")
		if err == nil {
			subtest.Fatalf("MissingErrorForUnknownTrustedDirOwner")
		}
		if !strings.Contains(err.Error(), "OwnerLookupFailed") {
			subtest.Errorf("ErrorMissingOwnerLookupFailed: %v", err)
		}
	})
}

func TestTruncateFileContent(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("TruncateContentFromFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "clear.txt")
		initialContent := "Content to be cleared"

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(initialContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.TruncateFileContent(absoluteFilePathForTest(subtest, testFile))
		if err != nil {
			subtest.Errorf("TruncateFileContentFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != "" {
			subtest.Errorf("ContentShouldBeEmpty: '%s'", actualContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})
}

func TestCreateDir(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("CreateNewDirectory", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "newdir")

		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Errorf("UnexpectedError: '%s'", err.Error())
		}

		if !clerk.IsDir(testDir) {
			subtest.Errorf("DirectoryShouldExist: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})

	test.Run("CreateNestedDirectory", func(subtest *testing.T) {
		nestedDir := filepath.Join(tempDir, "parent", "child", "grandchild")

		err := clerk.CreateDir(nestedDir)
		if err != nil {
			subtest.Errorf("CreateDirFailed: %v", err)
		}

		if !clerk.IsDir(nestedDir) {
			subtest.Errorf("NestedDirectoryShouldExist: %s", nestedDir)
		}

		parentDir := filepath.Join(tempDir, "parent")
		err = clerk.DeleteDir(parentDir)
		if err != nil {
			subtest.Errorf("DeleteParentDirFailed: %v", err)
		}
	})

	test.Run("CreateExistingDirectory", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "existing")

		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.CreateDir(testDir)
		if err != nil {
			subtest.Errorf("CreateDirFailed: %v", err)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})
}

func TestCopyDir(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("CopyDirectoryWithFiles", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "sourcedir")
		targetDir := filepath.Join(tempDir, "targetdir")
		testFile := filepath.Join(sourceDir, "testfile.txt")
		testContent := "test content"

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			subtest.Fatalf("CreateSourceDirFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("CreateTestFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.CopyDir(sourceDir, targetDir)
		if err != nil {
			subtest.Errorf("CopyDirFailed: %v", err)
		}

		if !clerk.IsDir(targetDir) {
			subtest.Errorf("TargetDirectoryShouldExist: %s", targetDir)
		}

		copiedFile := filepath.Join(targetDir, "testfile.txt")
		if !clerk.IsFile(copiedFile) {
			subtest.Errorf("CopiedFileShouldExist: %s", copiedFile)
		}

		copiedContent, err := clerk.ReadFileContent(copiedFile, nil)
		if err != nil {
			subtest.Errorf("ReadCopiedFileContentFailed: %v", err)
		}

		if copiedContent != testContent {
			subtest.Errorf("CopiedContentMismatch: '%s' vs '%s'", copiedContent, testContent)
		}

		err = clerk.DeleteDir(sourceDir)
		if err != nil {
			subtest.Errorf("DeleteSourceDirFailed: %v", err)
		}

		err = clerk.DeleteDir(targetDir)
		if err != nil {
			subtest.Errorf("DeleteTargetDirFailed: %v", err)
		}
	})

	test.Run("CopyNonExistentDirectory", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "nonexistent")
		targetDir := filepath.Join(tempDir, "target")

		err := clerk.CopyDir(sourceDir, targetDir)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceDirNotFound")
		}

		if err != nil && err.Error() != "SourceDirNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceDirNotFound", err.Error())
		}
	})
}

func TestMoveDir(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("MoveDirectoryWithFiles", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "movesource")
		targetDir := filepath.Join(tempDir, "movetarget")
		testFile := filepath.Join(sourceDir, "movefile.txt")
		testContent := "content to move"

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			subtest.Fatalf("CreateSourceDirFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("CreateTestFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.MoveDir(sourceDir, targetDir)
		if err != nil {
			subtest.Errorf("MoveDirFailed: %v", err)
		}

		if clerk.IsDir(sourceDir) {
			subtest.Errorf("SourceDirectoryShouldNotExist: %s", sourceDir)
		}

		if !clerk.IsDir(targetDir) {
			subtest.Errorf("TargetDirectoryShouldExist: %s", targetDir)
		}

		movedFile := filepath.Join(targetDir, "movefile.txt")
		if !clerk.IsFile(movedFile) {
			subtest.Errorf("MovedFileShouldExist: %s", movedFile)
		}

		movedContent, err := clerk.ReadFileContent(movedFile, nil)
		if err != nil {
			subtest.Errorf("ReadMovedFileContentFailed: %v", err)
		}

		if movedContent != testContent {
			subtest.Errorf("MovedContentMismatch: '%s' vs '%s'", movedContent, testContent)
		}

		err = clerk.DeleteDir(targetDir)
		if err != nil {
			subtest.Errorf("DeleteTargetDirFailed: %v", err)
		}
	})

	test.Run("MoveNonExistentDirectory", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "nonexistent")
		targetDir := filepath.Join(tempDir, "target")

		err := clerk.MoveDir(sourceDir, targetDir)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceDirNotFound")
		}

		if err != nil && err.Error() != "SourceDirNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceDirNotFound", err.Error())
		}
	})
}

func TestDeleteDir(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("DeleteExistingDirectory", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "deleteme")
		testFile := filepath.Join(testDir, "file.txt")

		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.IsDir(testDir) {
			subtest.Fatalf("DirectoryShouldExistBeforeDeletion: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}

		if clerk.IsDir(testDir) {
			subtest.Errorf("DirectoryShouldNotExistAfterDeletion: %s", testDir)
		}
	})

	test.Run("DeleteNonExistentDirectory", func(subtest *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")

		err := clerk.DeleteDir(nonExistentDir)
		if err != nil {
			subtest.Errorf("DeleteNonExistentDirFailed: %v", err)
		}
	})

	test.Run("DeleteFileThroughDeleteDirFails", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "not_a_dir.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.DeleteDir(testFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetNotDirectory")
		}
		if err != nil && err.Error() != "TargetNotDirectory" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetNotDirectory", err.Error())
		}

		if !clerk.IsFile(testFile) {
			subtest.Errorf("FileMustSurviveRejectedDelete: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})
}

func TestIsSymlink(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("RegularFile", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		if clerk.IsSymlink(testFile) {
			subtest.Errorf("IsSymlinkShouldReturnFalseForRegularFile: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("NonExistentPath", func(subtest *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")

		if clerk.IsSymlink(nonExistentPath) {
			subtest.Errorf("IsSymlinkShouldReturnFalseForNonExistent: %s", nonExistentPath)
		}
	})
}

func TestCreateSymlink(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("CreateSymlinkToFile", func(subtest *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CreateSymlink(targetFile, symlinkPath, false)
		if err != nil {
			subtest.Errorf("CreateSymlinkFailed: %v", err)
		}

		if !clerk.IsSymlink(symlinkPath) {
			subtest.Errorf("SymlinkShouldExist: %s", symlinkPath)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			subtest.Errorf("RemoveSymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	test.Run("CreateSymlinkToNonExistentFile", func(subtest *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.CreateSymlink(nonExistentFile, symlinkPath, false)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourcePathNotFound")
		}

		if err != nil && err.Error() != "SourcePathNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourcePathNotFound", err.Error())
		}
	})
}

func TestIsSymlinkTo(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("ValidSymlinkTarget", func(subtest *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CreateSymlink(targetFile, symlinkPath, false)
		if err != nil {
			subtest.Fatalf("CreateSymlinkFailed: %v", err)
		}

		if !clerk.IsSymlinkTo(symlinkPath, targetFile) {
			subtest.Errorf("IsSymlinkToShouldReturnTrue: %s -> %s", symlinkPath, targetFile)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			subtest.Errorf("RemoveSymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	test.Run("RegularFileNotSymlink", func(subtest *testing.T) {
		regularFile := filepath.Join(tempDir, "regular.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.TouchFile(regularFile)
		if err != nil {
			subtest.Fatalf("CreateRegularFileFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		if clerk.IsSymlinkTo(regularFile, targetFile) {
			subtest.Errorf("IsSymlinkToShouldReturnFalseForRegularFile: %s", regularFile)
		}

		err = clerk.DeleteFile(regularFile)
		if err != nil {
			subtest.Errorf("DeleteRegularFileFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestRemoveSymlink(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("RemoveExistingSymlink", func(subtest *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CreateSymlink(targetFile, symlinkPath, false)
		if err != nil {
			subtest.Fatalf("CreateSymlinkFailed: %v", err)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			subtest.Errorf("RemoveSymlinkFailed: %v", err)
		}

		if clerk.IsSymlink(symlinkPath) {
			subtest.Errorf("SymlinkShouldNotExistAfterRemoval: %s", symlinkPath)
		}

		if !clerk.IsFile(targetFile) {
			subtest.Errorf("TargetFileShouldStillExist: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestUpdateFilePermissions(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("UpdateFilePermissionsWithCustomValue", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "permissions.txt")
		customPermissions := os.FileMode(0600)

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(testFile, &customPermissions)
		if err != nil {
			subtest.Errorf("UpdateFilePermissionsFailed: %v", err)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("UpdateFilePermissionsWithDefaultValue", func(subtest *testing.T) {
		testFile := filepath.Join(tempDir, "default_permissions.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			subtest.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(testFile, nil)
		if err != nil {
			subtest.Errorf("UpdateFilePermissionsFailed: %v", err)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("UpdateDirectoryPermissions", func(subtest *testing.T) {
		testDir := filepath.Join(tempDir, "permissions_dir")

		err := clerk.CreateDir(testDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(testDir, nil)
		if err != nil {
			subtest.Errorf("UpdateFilePermissionsFailed: %v", err)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			subtest.Errorf("DeleteDirFailed: %v", err)
		}
	})

	test.Run("UpdateSymlinkPermissionsFails", func(subtest *testing.T) {
		targetFile := filepath.Join(tempDir, "chmod_target.txt")
		symlinkPath := filepath.Join(tempDir, "chmod_link.txt")
		restrictivePermissions := os.FileMode(0600)

		err := os.WriteFile(targetFile, []byte("target"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = os.Symlink(targetFile, symlinkPath)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(symlinkPath, &restrictivePermissions)
		if err == nil {
			subtest.Errorf("MissingExpectedError: TargetIsSymlink")
		}
		if err != nil && err.Error() != "TargetIsSymlink" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetIsSymlink", err.Error())
		}

		targetInfo, statErr := os.Stat(targetFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if targetInfo.Mode().Perm() == 0600 {
			subtest.Errorf("ChmodLandedOnSymlinkTarget")
		}
	})
}

func TestCompressFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("CompressFileWithFormats", func(subtest *testing.T) {
		tarFormat := "tar"
		gzFormat := "gz"
		zipFormat := "zip"
		xzFormat := "xz"
		brFormat := "br"
		unsupportedFormat := "unsupported"

		testCaseStructs := []struct {
			format         *string
			expectedSuffix string
			shouldSucceed  bool
			expectedError  string
		}{
			{nil, ".tar", true, ""},
			{&tarFormat, ".tar", true, ""},
			{&gzFormat, ".gz", true, ""},
			{&zipFormat, ".zip", true, ""},
			{&xzFormat, ".xz", true, ""},
			{&brFormat, ".br", true, ""},
			{&unsupportedFormat, "", false, "UnsupportedCompressionFormat"},
		}

		for _, testCase := range testCaseStructs {
			formatName := "default"
			if testCase.format != nil {
				formatName = *testCase.format
			}

			testFile := filepath.Join(tempDir, "compress_"+formatName+".txt")
			testContent := "content to compress with " + formatName

			err := clerk.TouchFile(testFile)
			if err != nil {
				subtest.Fatalf("TouchFileFailed: %v", err)
			}

			err = os.WriteFile(testFile, []byte(testContent), 0644)
			if err != nil {
				subtest.Fatalf("WriteFileFailed: %v", err)
			}

			compressedFilePath, err := clerk.CompressFile(testFile, testCase.format, nil)
			if err != nil && testCase.shouldSucceed {
				subtest.Errorf("[%s] CompressFileFailed: %v", formatName, err)
			}

			if err != nil && !testCase.shouldSucceed {
				if testCase.expectedError != "" && err.Error() != testCase.expectedError {
					subtest.Errorf(
						"[%s] WrongErrorMessage: '%s' vs '%s'",
						formatName, testCase.expectedError, err.Error(),
					)
				}
			}

			if err == nil && !testCase.shouldSucceed {
				subtest.Errorf("[%s] UnexpectedCompressionSuccess: %s", formatName, testCase.expectedError)
			}

			if err == nil && !strings.HasSuffix(compressedFilePath, testCase.expectedSuffix) {
				subtest.Errorf(
					"[%s] WrongCompressionSuffix: '%s' vs '%s'",
					formatName, testCase.expectedSuffix, compressedFilePath,
				)
			}

			err = clerk.DeleteFile(testFile)
			if err != nil {
				subtest.Errorf("DeleteOriginalFileFailed: %v", err)
			}
		}
	})

	test.Run("CompressNonExistentFile", func(subtest *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := clerk.CompressFile(nonExistentFile, nil, nil)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})

	test.Run("SourceDeletionContractIsFormatIndependent", func(subtest *testing.T) {
		keepSourceTrue := true
		formats := []string{"tar", "gz", "zip", "xz", "br"}

		for _, format := range formats {
			formatName := format

			subtest.Run("Keep_"+formatName, func(subtest *testing.T) {
				testFile := filepath.Join(tempDir, "keep_"+formatName+".txt")
				err := clerk.WriteNewFile(testFile, "content for keep "+formatName, 0644)
				if err != nil {
					subtest.Fatalf("WriteNewFileFailed: %v", err)
				}

				compressedFile, err := clerk.CompressFile(testFile, &formatName, &keepSourceTrue)
				if err != nil {
					subtest.Fatalf("[%s] CompressFileFailed: %v", formatName, err)
				}
				if !clerk.IsFile(testFile) {
					subtest.Errorf("[%s] SourceMustBeKept: %s", formatName, testFile)
				}

				_ = clerk.DeleteFile(testFile)
				_ = clerk.DeleteFile(compressedFile)
			})

			subtest.Run("Remove_"+formatName, func(subtest *testing.T) {
				testFile := filepath.Join(tempDir, "remove_"+formatName+".txt")
				err := clerk.WriteNewFile(testFile, "content for remove "+formatName, 0644)
				if err != nil {
					subtest.Fatalf("WriteNewFileFailed: %v", err)
				}

				compressedFile, err := clerk.CompressFile(testFile, &formatName, nil)
				if err != nil {
					subtest.Fatalf("[%s] CompressFileFailed: %v", formatName, err)
				}
				if clerk.FileExists(testFile) {
					subtest.Errorf("[%s] SourceMustBeRemoved: %s", formatName, testFile)
				}

				_ = clerk.DeleteFile(compressedFile)
			})
		}
	})
}

func TestCompressDir(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("DefaultBrotliStageSucceedsAndCleansIntermediateTar", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "compressdir_br")
		nestedFile := filepath.Join(sourceDir, "nested.txt")

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.WriteNewFile(nestedFile, "nested content", 0644)
		if err != nil {
			subtest.Fatalf("WriteNewFileFailed: %v", err)
		}

		compressedFilePath, err := clerk.CompressDir(sourceDir, nil)
		if err != nil {
			subtest.Fatalf("CompressDirFailed: %v", err)
		}

		if !strings.HasSuffix(compressedFilePath, ".tar.br") {
			subtest.Errorf("WrongCompressionSuffix: '%s'", compressedFilePath)
		}
		if !clerk.IsFile(compressedFilePath) {
			subtest.Errorf("CompressedFileShouldExist: %s", compressedFilePath)
		}
		if clerk.FileExists(sourceDir + ".tar") {
			subtest.Errorf("IntermediateTarShouldBeRemoved: %s", sourceDir+".tar")
		}
		if !clerk.IsDir(sourceDir) {
			subtest.Errorf("SourceDirMustBeKept: %s", sourceDir)
		}
	})

	test.Run("TarFormatSkipsSecondStage", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "compressdir_tar")
		tarFormat := "tar"

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		compressedFilePath, err := clerk.CompressDir(sourceDir, &tarFormat)
		if err != nil {
			subtest.Fatalf("CompressDirFailed: %v", err)
		}
		if !strings.HasSuffix(compressedFilePath, ".tar") {
			subtest.Errorf("WrongCompressionSuffix: '%s'", compressedFilePath)
		}
	})
}

func TestDecompressFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("DecompressFileWithFormats", func(subtest *testing.T) {
		keepSourceTrue := true
		keepSourceFalse := false

		testCaseStructs := []struct {
			format           string
			shouldKeepSource *bool
			shouldSucceed    bool
			expectedError    string
		}{
			{"tar", &keepSourceTrue, true, ""},
			{"tar", &keepSourceFalse, true, ""},
			{"tar", nil, true, ""},
			{"gz", &keepSourceTrue, true, ""},
			{"gz", nil, true, ""},
			{"xz", &keepSourceTrue, true, ""},
			{"xz", nil, true, ""},
			{"br", &keepSourceTrue, true, ""},
			{"br", nil, true, ""},
			{"unsupported", nil, false, "UnsupportedCompressionFormat"},
		}

		for _, testCase := range testCaseStructs {
			testFile := filepath.Join(tempDir, "decompress_test.txt")
			preCompressedContent := "content to decompress"

			err := clerk.TouchFile(testFile)
			if err != nil {
				subtest.Fatalf("TouchFileFailed: %v", err)
			}

			err = os.WriteFile(testFile, []byte(preCompressedContent), 0644)
			if err != nil {
				subtest.Fatalf("WriteFileFailed: %v", err)
			}

			compressedFile, err := clerk.CompressFile(testFile, &testCase.format, nil)
			if err != nil && testCase.shouldSucceed {
				subtest.Fatalf("[%s] CompressFileFailed: %v", testCase.format, err)

				if clerk.FileExists(testFile) {
					err = clerk.DeleteFile(testFile)
					if err != nil {
						subtest.Errorf("DeleteOriginalFileFailed: %v", err)
					}
				}

				if clerk.FileExists(compressedFile) {
					err = clerk.DeleteFile(compressedFile)
					if err != nil {
						subtest.Errorf("DeleteCompressedFileFailed: %v", err)
					}
				}

				continue
			}

			decompressedFile, err := clerk.DecompressFile(compressedFile, nil, testCase.shouldKeepSource)
			if err != nil && testCase.shouldSucceed {
				subtest.Errorf("[%s] DecompressFileFailed: %v", testCase.format, err)
			}

			if err == nil && !testCase.shouldSucceed {
				subtest.Errorf("[%s] UnexpectedDecompressionSuccess", testCase.format)
			}

			if decompressedFile == "" && testCase.shouldSucceed {
				subtest.Errorf("[%s] DecompressedFilePathShouldNotBeEmpty", testCase.format)
			}

			if testCase.shouldSucceed && clerk.FileExists(decompressedFile) {
				decompressedContent, err := clerk.ReadFileContent(decompressedFile, nil)
				if err != nil {
					subtest.Fatalf("[%s]ReadDecompressedFileFailed: %v", testCase.format, err)
				}
				if decompressedContent != preCompressedContent {
					subtest.Errorf(
						"[%s] DecompressedContentMismatch: '%s' vs '%s'",
						testCase.format, preCompressedContent, decompressedContent,
					)
				}
			}

			shouldKeep := false
			if testCase.shouldKeepSource != nil {
				shouldKeep = *testCase.shouldKeepSource
			}

			if shouldKeep && !clerk.FileExists(compressedFile) {
				subtest.Errorf("[%s] CompressedFileShouldBeKept", testCase.format)
			}

			if !shouldKeep && clerk.FileExists(compressedFile) {
				subtest.Errorf("[%s] CompressedFileShouldBeRemoved", testCase.format)
			}

			if clerk.FileExists(decompressedFile) {
				err = clerk.DeleteFile(decompressedFile)
				if err != nil {
					subtest.Errorf("DeleteDecompressedFileFailed: %v", err)
				}
			}

			if clerk.FileExists(compressedFile) {
				err = clerk.DeleteFile(compressedFile)
				if err != nil {
					subtest.Errorf("DeleteCompressedFileFailed: %v", err)
				}
			}
		}
	})

	test.Run("DecompressNonExistentFile", func(subtest *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.tar")

		_, err := clerk.DecompressFile(nonExistentFile, nil, nil)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			subtest.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})
}

func TestFileContentRegexSearch(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	testCaseStructs := []struct {
		description      string
		fileContent      string
		patternSource    string
		shouldProvideNil bool
		expectedErrMsg   string
		expectedFindings []FileContentRegexFindings
		expectedErrIsNil bool
	}{
		{
			description:      "MatchesTwoAnchoredLines",
			fileContent:      "alpha=1\nbeta=2\nalpha=3",
			patternSource:    `(?m)^alpha=(\d+)$`,
			expectedErrIsNil: true,
			expectedFindings: []FileContentRegexFindings{
				{Match: "alpha=1", Groups: []string{"1"}, LineNumRange: []int{1, 1}},
				{Match: "alpha=3", Groups: []string{"3"}, LineNumRange: []int{3, 3}},
			},
		},
		{
			description:      "CaptureGroupPopulatesGroups",
			fileContent:      "user=42\nuser=99",
			patternSource:    `(?m)^user=(\d+)$`,
			expectedErrIsNil: true,
			expectedFindings: []FileContentRegexFindings{
				{Match: "user=42", Groups: []string{"42"}, LineNumRange: []int{1, 1}},
				{Match: "user=99", Groups: []string{"99"}, LineNumRange: []int{2, 2}},
			},
		},
		{
			description:      "NonExistentFileReturnsFileNotFound",
			fileContent:      "",
			patternSource:    `^foo$`,
			expectedErrMsg:   "FileNotFound",
			expectedErrIsNil: false,
		},
		{
			description:      "DirectoryPathReturnsTargetIsDirectory",
			fileContent:      "",
			patternSource:    `^foo$`,
			expectedErrMsg:   "TargetIsDirectory",
			expectedErrIsNil: false,
		},
		{
			description:      "EmptyFileReturnsNoMatches",
			fileContent:      "",
			patternSource:    `^alpha=(\d+)$`,
			expectedErrIsNil: true,
			expectedFindings: []FileContentRegexFindings{},
		},
		{
			description:      "LineExceedingScannerBufferSucceedsViaWholeFile",
			fileContent:      strings.Repeat("a", 100*1024),
			patternSource:    `^a+$`,
			expectedErrIsNil: true,
			expectedFindings: []FileContentRegexFindings{
				{Match: strings.Repeat("a", 100*1024), Groups: []string{}, LineNumRange: []int{1, 1}},
			},
		},
		{
			description:      "MultiLineMatchSpansLineRange",
			fileContent:      "alpha=1\nbeta=2\nalpha=3",
			patternSource:    `(?s)alpha=.+alpha=\d+`,
			expectedErrIsNil: true,
			expectedFindings: []FileContentRegexFindings{
				{Match: "alpha=1\nbeta=2\nalpha=3", Groups: []string{}, LineNumRange: []int{1, 3}},
			},
		},
		{
			description:      "MatchTextRepeatedBeforeRealMatchKeepsLineNumbers",
			fileContent:      "afoo\nfoo",
			patternSource:    `(?m)^foo$`,
			expectedErrIsNil: true,
			expectedFindings: []FileContentRegexFindings{
				{Match: "foo", Groups: []string{}, LineNumRange: []int{2, 2}},
			},
		},
		{
			description:      "NilPatternReturnsRegexPatternCannotBeNil",
			fileContent:      "alpha=1",
			shouldProvideNil: true,
			expectedErrMsg:   "RegexPatternCannotBeNil",
			expectedErrIsNil: false,
		},
	}

	for _, testCase := range testCaseStructs {
		test.Run(testCase.description, func(subtest *testing.T) {
			targetFile := filepath.Join(tempDir, testCase.description+".txt")
			createErr := clerk.TouchFile(targetFile)
			if createErr != nil {
				subtest.Fatalf("TouchFileFailed: %v", createErr)
			}

			if testCase.fileContent != "" {
				writeErr := os.WriteFile(
					targetFile, []byte(testCase.fileContent), 0644,
				)
				if writeErr != nil {
					subtest.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}

			if testCase.description == "NonExistentFileReturnsFileNotFound" {
				deleteErr := clerk.DeleteFile(targetFile)
				if deleteErr != nil {
					subtest.Fatalf("DeleteFileFailed: %v", deleteErr)
				}
			}

			if testCase.description == "DirectoryPathReturnsTargetIsDirectory" {
				deleteErr := clerk.DeleteFile(targetFile)
				if deleteErr != nil {
					subtest.Fatalf("DeleteFileFailed: %v", deleteErr)
				}
				createDirErr := clerk.CreateDir(targetFile)
				if createDirErr != nil {
					subtest.Fatalf("CreateDirFailed: %v", createDirErr)
				}
			}

			filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
				targetFile, true,
			)
			if pathErr != nil {
				subtest.Fatalf("NewUnixAbsoluteFilePathFailed: %v", pathErr)
			}

			var regexPattern *regexp.Regexp
			if !testCase.shouldProvideNil {
				regexPattern = regexp.MustCompile(testCase.patternSource)
			}

			regexSearchFindings, searchErr := clerk.FileContentRegexSearch(
				filePath, regexPattern,
			)

			if testCase.expectedErrIsNil {
				if searchErr != nil {
					subtest.Errorf("UnexpectedError: '%s'", searchErr.Error())
					return
				}
			}

			if !testCase.expectedErrIsNil {
				if searchErr == nil {
					subtest.Errorf("MissingExpectedError: %s", testCase.expectedErrMsg)
					return
				}
				if testCase.expectedErrMsg != "" &&
					searchErr.Error() != testCase.expectedErrMsg {
					subtest.Errorf(
						"WrongErrorMessage: '%s' vs '%s'",
						testCase.expectedErrMsg, searchErr.Error(),
					)
				}
				return
			}

			if len(regexSearchFindings) != len(testCase.expectedFindings) {
				subtest.Errorf(
					"WrongFindingsCount: expected=%d actual=%d",
					len(testCase.expectedFindings), len(regexSearchFindings),
				)
				return
			}

			for findingIndex, expectedFinding := range testCase.expectedFindings {
				actualFinding := regexSearchFindings[findingIndex]
				if actualFinding.Match != expectedFinding.Match {
					subtest.Errorf(
						"WrongFindingMatch%d: '%s' vs '%s'",
						findingIndex, expectedFinding.Match, actualFinding.Match,
					)
				}
				if len(actualFinding.LineNumRange) != 2 ||
					len(expectedFinding.LineNumRange) != 2 {
					subtest.Errorf(
						"WrongFindingLineRangeShape%d: expected=%v actual=%v",
						findingIndex,
						expectedFinding.LineNumRange,
						actualFinding.LineNumRange,
					)
					continue
				}
				if actualFinding.LineNumRange[0] != expectedFinding.LineNumRange[0] {
					subtest.Errorf(
						"WrongFindingLineRangeStart%d: '%d' vs '%d'",
						findingIndex,
						expectedFinding.LineNumRange[0],
						actualFinding.LineNumRange[0],
					)
				}
				if actualFinding.LineNumRange[1] != expectedFinding.LineNumRange[1] {
					subtest.Errorf(
						"WrongFindingLineRangeEnd%d: '%d' vs '%d'",
						findingIndex,
						expectedFinding.LineNumRange[1],
						actualFinding.LineNumRange[1],
					)
				}
				if len(actualFinding.Groups) != len(expectedFinding.Groups) {
					subtest.Errorf(
						"WrongFindingGroupCount%d: expected=%d actual=%d",
						findingIndex,
						len(expectedFinding.Groups), len(actualFinding.Groups),
					)
					continue
				}
				for groupIndex, expectedGroup := range expectedFinding.Groups {
					if actualFinding.Groups[groupIndex] != expectedGroup {
						subtest.Errorf(
							"WrongFindingGroup%d-%d: '%s' vs '%s'",
							findingIndex, groupIndex,
							expectedGroup, actualFinding.Groups[groupIndex],
						)
					}
				}
			}

			cleanupErr := clerk.DeleteFile(targetFile)
			if cleanupErr != nil {
				subtest.Errorf("DeleteFileFailed: %v", cleanupErr)
			}

			if testCase.description == "DirectoryPathReturnsTargetIsDirectory" {
				dirCleanupErr := clerk.DeleteDir(targetFile)
				if dirCleanupErr != nil {
					subtest.Errorf("DeleteDirFailed: %v", dirCleanupErr)
				}
			}
		})
	}
}

func TestOverwriteFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	test.Run("OverwriteExistingTarget", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		sourceContent := "new content via overwrite"
		targetContent := "old content to be overwritten"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		if err != nil {
			subtest.Fatalf("UpdateSourceFileContentFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			subtest.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = os.WriteFile(targetFile, []byte(targetContent), 0644)
		if err != nil {
			subtest.Fatalf("UpdateTargetFileContentFailed: %v", err)
		}

		err = clerk.OverwriteFile(sourceFile, targetFile)
		if err != nil {
			subtest.Errorf("OverwriteFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			subtest.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			subtest.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		actualContent, err := clerk.ReadFileContent(targetFile, nil)
		if err != nil {
			subtest.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != sourceContent {
			subtest.Errorf(
				"ContentMismatch: '%s' vs '%s'",
				actualContent, sourceContent,
			)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("OverwriteWithNonExistentTarget", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "source2.txt")
		targetFile := filepath.Join(tempDir, "new_target.txt")
		sourceContent := "content"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			subtest.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		if err != nil {
			subtest.Fatalf("UpdateSourceFileContentFailed: %v", err)
		}

		err = clerk.OverwriteFile(sourceFile, targetFile)
		if err != nil {
			subtest.Errorf("OverwriteFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			subtest.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			subtest.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			subtest.Errorf("DeleteFileFailed: %v", err)
		}
	})

	test.Run("OverwriteNonExistentSource", func(subtest *testing.T) {
		sourceFile := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.OverwriteFile(sourceFile, targetFile)
		if err == nil {
			subtest.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			subtest.Errorf(
				"WrongErrorMessage: '%s' vs '%s'",
				"SourceFileNotFound", err.Error(),
			)
		}
	})

	test.Run("OverwriteSymlinkTargetReplacesUnderlyingFile", func(subtest *testing.T) {
		realFile := filepath.Join(tempDir, "real.txt")
		linkFile := filepath.Join(tempDir, "link.txt")
		sourceFile := filepath.Join(tempDir, "source3.txt")

		if err := os.WriteFile(realFile, []byte("original"), 0644); err != nil {
			subtest.Fatalf("UpdateRealFileContentFailed: %v", err)
		}
		if err := os.Symlink(realFile, linkFile); err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}
		if err := os.WriteFile(sourceFile, []byte("replaced"), 0644); err != nil {
			subtest.Fatalf("UpdateSourceFileContentFailed: %v", err)
		}

		if err := clerk.OverwriteFile(sourceFile, linkFile); err != nil {
			subtest.Fatalf("OverwriteFileFailed: %v", err)
		}

		if !clerk.IsSymlink(linkFile) {
			subtest.Errorf("SymlinkShouldBePreserved: %s", linkFile)
		}
		realContent, err := clerk.ReadFileContent(realFile, nil)
		if err != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", err)
		}
		if realContent != "replaced" {
			subtest.Errorf("ContentMismatch: '%s' vs 'replaced'", realContent)
		}
	})

	test.Run("OverwriteWithDirectorySourceReturnsSourceIsDirectory", func(subtest *testing.T) {
		sourceDir := filepath.Join(tempDir, "sourceDir")
		if err := clerk.CreateDir(sourceDir); err != nil {
			subtest.Fatalf("CreateDirFailed: %v", err)
		}

		err := clerk.OverwriteFile(sourceDir, filepath.Join(tempDir, "any.txt"))
		if err == nil || err.Error() != "SourceIsDirectory" {
			subtest.Errorf("WrongErrorMessage: 'SourceIsDirectory' vs '%v'", err)
		}
	})
}

func TestFileContentRegexReplace(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	testCaseStructs := []struct {
		description      string
		fileContent      string
		patternSource    string
		replacement      string
		shouldProvideNil bool
		expectedErrMsg   string
		expectedErrIsNil bool
		expectedContent  string
		expectedCount    int
	}{
		{
			description:      "SimpleReplacementUpdatesContent",
			fileContent:      "alpha=1\nbeta=2\nalpha=3",
			patternSource:    `(?m)^alpha=(\d+)$`,
			replacement:      `alpha=$1!`,
			expectedErrIsNil: true,
			expectedContent:  "alpha=1!\nbeta=2\nalpha=3!",
			expectedCount:    2,
		},
		{
			description:      "ReplacementWithoutAnchorsReplacesAll",
			fileContent:      "foo foo foo",
			patternSource:    `foo`,
			replacement:      `bar`,
			expectedErrIsNil: true,
			expectedContent:  "bar bar bar",
			expectedCount:    3,
		},
		{
			description:      "NoMatchReturnsZeroCountAndUnchangedContent",
			fileContent:      "hello world",
			patternSource:    `^xyz$`,
			replacement:      `replaced`,
			expectedErrIsNil: true,
			expectedContent:  "hello world",
			expectedCount:    0,
		},
		{
			description:      "EmptyReplacementRemovesMatches",
			fileContent:      "abc abc def",
			patternSource:    `abc `,
			replacement:      ``,
			expectedErrIsNil: true,
			expectedContent:  "def",
			expectedCount:    2,
		},
		{
			description:      "NonExistentFileReturnsFileNotFound",
			fileContent:      "",
			patternSource:    `^foo$`,
			replacement:      `bar`,
			expectedErrMsg:   "FileNotFound",
			expectedErrIsNil: false,
		},
		{
			description:      "EmptyFileReturnsFileEmpty",
			fileContent:      "",
			patternSource:    `^foo$`,
			replacement:      `bar`,
			expectedErrMsg:   "FileEmpty",
			expectedErrIsNil: false,
		},
		{
			description:      "EmptyResultOnNonEmptySourceReturnsReplacementWouldTruncateFile",
			fileContent:      "foo bar",
			patternSource:    `(?s).*`,
			replacement:      ``,
			expectedErrMsg:   "ReplacementWouldTruncateFile",
			expectedErrIsNil: false,
			expectedContent:  "foo bar",
		},
		{
			description:      "StreamingRegexMatchingAllWithEmptyReplacementProducesNewlineOnlyFile",
			fileContent:      strings.Repeat("foo\n", 3000000),
			patternSource:    `(?s).*`,
			replacement:      ``,
			expectedErrIsNil: true,
			expectedContent:  strings.Repeat("\n", 3000000),
			expectedCount:    3000000,
		},
		{
			description:      "NilPatternReturnsRegexPatternCannotBeNil",
			fileContent:      "alpha=1",
			shouldProvideNil: true,
			replacement:      `alpha=2`,
			expectedErrMsg:   "RegexPatternCannotBeNil",
			expectedErrIsNil: false,
		},
	}

	for _, testCase := range testCaseStructs {
		test.Run(testCase.description, func(subtest *testing.T) {
			targetFile := filepath.Join(tempDir, testCase.description+".txt")
			createErr := clerk.TouchFile(targetFile)
			if createErr != nil {
				subtest.Fatalf("TouchFileFailed: %v", createErr)
			}

			if testCase.fileContent != "" {
				writeErr := os.WriteFile(
					targetFile, []byte(testCase.fileContent), 0644,
				)
				if writeErr != nil {
					subtest.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}

			if testCase.description == "NonExistentFileReturnsFileNotFound" {
				deleteErr := clerk.DeleteFile(targetFile)
				if deleteErr != nil {
					subtest.Fatalf("DeleteFileFailed: %v", deleteErr)
				}
			}

			filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
				targetFile, true,
			)
			if pathErr != nil {
				subtest.Fatalf("NewUnixAbsoluteFilePathFailed: %v", pathErr)
			}

			var regexPattern *regexp.Regexp
			if !testCase.shouldProvideNil {
				regexPattern = regexp.MustCompile(testCase.patternSource)
			}

			replacementCount, replaceErr := clerk.FileContentRegexReplace(
				FileRegexReplaceSettings{FilePath: filePath},
				regexPattern, testCase.replacement,
			)

			if testCase.expectedErrIsNil {
				if replaceErr != nil {
					subtest.Errorf("UnexpectedError: '%s'", replaceErr.Error())
					return
				}

				if replacementCount != testCase.expectedCount {
					subtest.Errorf(
						"WrongReplacementCount: expected=%d actual=%d",
						testCase.expectedCount, replacementCount,
					)
				}

				actualContent, readErr := clerk.ReadFileContent(targetFile, nil)
				if readErr != nil {
					subtest.Errorf("ReadFileContentFailed: %v", readErr)
					return
				}

				if actualContent != testCase.expectedContent {
					subtest.Errorf(
						"WrongContent: '%s' vs '%s'",
						actualContent, testCase.expectedContent,
					)
				}

				tempLeftoverMatches, _ := filepath.Glob(
					filepath.Join(filepath.Dir(targetFile), ".*.tk-tmp*"),
				)
				for _, tempLeftoverPath := range tempLeftoverMatches {
					subtest.Errorf(
						"TempFileShouldNotExistAfterReplace: %s",
						tempLeftoverPath,
					)
					_ = clerk.DeleteFile(tempLeftoverPath)
				}

				return
			}

			if replaceErr == nil {
				subtest.Errorf("MissingExpectedError: %s", testCase.expectedErrMsg)
				return
			}
			if testCase.expectedErrMsg != "" &&
				replaceErr.Error() != testCase.expectedErrMsg {
				subtest.Errorf(
					"WrongErrorMessage: '%s' vs '%s'",
					testCase.expectedErrMsg, replaceErr.Error(),
				)
			}

			tempLeftoverMatches, _ := filepath.Glob(
				filepath.Join(filepath.Dir(targetFile), ".*.tk-tmp*"),
			)
			for _, tempLeftoverPath := range tempLeftoverMatches {
				subtest.Errorf(
					"TempFileShouldNotExistAfterFailedReplace: %s",
					tempLeftoverPath,
				)
				_ = clerk.DeleteFile(tempLeftoverPath)
			}

			if testCase.expectedContent != "" {
				actualContentAfterFail, readErr := clerk.ReadFileContent(
					targetFile, nil,
				)
				if readErr != nil {
					subtest.Errorf("ReadFileContentAfterFailFailed: %v", readErr)
					return
				}
				if actualContentAfterFail != testCase.expectedContent {
					subtest.Errorf(
						"OriginalFileShouldBeUnchangedAfterFailedReplace: '%s' vs '%s'",
						testCase.expectedContent, actualContentAfterFail,
					)
				}
			}
		})
	}

	test.Run("RefusesSymlinkTargetByDefault", func(subtest *testing.T) {
		realFile := filepath.Join(tempDir, "replaceSymlinkReal.txt")
		err := os.WriteFile(realFile, []byte("alpha=1"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		linkFile := filepath.Join(tempDir, "replaceSymlinkLink.txt")
		err = os.Symlink(realFile, linkFile)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		regexPattern := regexp.MustCompile(`alpha=(\d+)`)
		_, replaceErr := clerk.FileContentRegexReplace(
			FileRegexReplaceSettings{
				FilePath: absoluteFilePathForTest(subtest, linkFile),
			},
			regexPattern, `alpha=$1!`,
		)
		if !errors.Is(replaceErr, ErrTargetIsSymlink) {
			subtest.Errorf("ExpectedErrTargetIsSymlink, got: %v", replaceErr)
		}
	})

	test.Run("ResolvesSymlinkTargetWhenRequested", func(subtest *testing.T) {
		realFile := filepath.Join(tempDir, "resolveReplaceReal.txt")
		err := os.WriteFile(realFile, []byte("alpha=1"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		linkFile := filepath.Join(tempDir, "resolveReplaceLink.txt")
		err = os.Symlink(realFile, linkFile)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		symlinkPolicy := FileClerkSymlinkPolicyResolve
		regexPattern := regexp.MustCompile(`alpha=(\d+)`)
		replacementCount, replaceErr := clerk.FileContentRegexReplace(
			FileRegexReplaceSettings{
				FilePath:      absoluteFilePathForTest(subtest, linkFile),
				SymlinkPolicy: &symlinkPolicy,
			},
			regexPattern, `alpha=$1!`,
		)
		if replaceErr != nil {
			subtest.Fatalf("FileContentRegexReplaceFailed: %v", replaceErr)
		}
		if replacementCount != 1 {
			subtest.Errorf("WrongReplacementCount: expected=1 actual=%d", replacementCount)
		}

		actualContent, readErr := clerk.ReadFileContent(realFile, nil)
		if readErr != nil {
			subtest.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if actualContent != "alpha=1!" {
			subtest.Errorf("WrongContent: '%s'", actualContent)
		}
		if !clerk.IsSymlink(linkFile) {
			subtest.Errorf("SymlinkWasReplaced")
		}
	})

	test.Run("RefusesWritableDirChainWhenRequired", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "writableReplaceDir")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chmod(dir, 0o777)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}
		targetFile := filepath.Join(dir, "replace.txt")
		err = os.WriteFile(targetFile, []byte("alpha=1"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused
		regexPattern := regexp.MustCompile(`alpha=(\d+)`)
		_, replaceErr := clerk.FileContentRegexReplace(
			FileRegexReplaceSettings{
				FilePath:       absoluteFilePathForTest(subtest, targetFile),
				DirChainPolicy: &sharedWriteRefusedPolicy,
			},
			regexPattern, `alpha=$1!`,
		)
		if !errors.Is(replaceErr, ErrDirectoryWritableByOthers) {
			subtest.Errorf("ExpectedErrDirectoryWritableByOthers, got: %v", replaceErr)
		}
	})

	test.Run("PreservesTargetOwnerAndSpecialModeBits", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			subtest.Fatalf("GidParseFailed: %v", gidErr)
		}

		targetFile := filepath.Join(tempDir, "ownerPreserved.txt")
		err := os.WriteFile(targetFile, []byte("alpha=1"), 0640)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		err = os.Chown(targetFile, nobodyUid, nobodyGid)
		if err != nil {
			subtest.Fatalf("ChownFailed: %v", err)
		}
		err = os.Chmod(targetFile, os.FileMode(0640)|os.ModeSetgid)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}

		regexPattern := regexp.MustCompile(`alpha=(\d+)`)
		_, replaceErr := clerk.FileContentRegexReplace(
			FileRegexReplaceSettings{
				FilePath: absoluteFilePathForTest(subtest, targetFile),
			},
			regexPattern, `alpha=$1!`,
		)
		if replaceErr != nil {
			subtest.Fatalf("FileContentRegexReplaceFailed: %v", replaceErr)
		}

		fileInfo, statErr := os.Stat(targetFile)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			subtest.Errorf(
				"OwnerChanged: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
		if fileInfo.Mode().Perm() != 0640 {
			subtest.Errorf("ModeChanged: expected 0640, got %o", fileInfo.Mode().Perm())
		}
		if fileInfo.Mode()&os.ModeSetgid == 0 {
			subtest.Errorf("SetgidBitLost: %v", fileInfo.Mode())
		}
	})
}

func absoluteFilePathForTest(
	test *testing.T, path string,
) tkValueObject.UnixAbsoluteFilePath {
	test.Helper()
	filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(path, true)
	if pathErr != nil {
		test.Fatalf("FilePathInvalid: %v", pathErr)
	}
	return filePath
}

func absoluteGlobPathForTest(
	test *testing.T, path string,
) tkValueObject.UnixAbsoluteGlobPath {
	test.Helper()
	globPath, globPathErr := tkValueObject.NewUnixAbsoluteGlobPath(path)
	if globPathErr != nil {
		test.Fatalf("GlobPathInvalid: %v", globPathErr)
	}
	return globPath
}

func TestInspectedTargetFileOpener(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	targetFileName, fileNameErr := tkValueObject.NewUnixFileName(
		"verified.txt", true,
	)
	if fileNameErr != nil {
		test.Fatalf("FileNameInvalid: %v", fileNameErr)
	}
	targetFilePath := filepath.Join(tempDir, targetFileName.String())
	writeErr := os.WriteFile(targetFilePath, []byte("alpha=1"), 0644)
	if writeErr != nil {
		test.Fatalf("WriteFileFailed: %v", writeErr)
	}

	dirHandle, openErr := unix.Open(
		tempDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0,
	)
	if openErr != nil {
		test.Fatalf("OpenDirFailed: %v", openErr)
	}
	defer func() { _ = unix.Close(dirHandle) }()

	targetSnapshot, snapshotErr := clerk.fileSnapshotReader(
		dirHandle, targetFileName.String(),
	)
	if snapshotErr != nil {
		test.Fatalf("TargetFileSnapshotReadFailed: %v", snapshotErr)
	}

	test.Run("AcceptsTheInspectedInode", func(subtest *testing.T) {
		fileHandle, openErr := clerk.inspectedTargetFileOpener(
			dirHandle, targetFileName, targetSnapshot, targetFileReadOpenFlags,
		)
		if openErr != nil {
			subtest.Fatalf("UnexpectedOpenError: %v", openErr)
		}
		_ = unix.Close(fileHandle)
	})

	test.Run("RefusesSwappedRegularFile", func(subtest *testing.T) {
		swappedFilePath := filepath.Join(tempDir, "swapped.txt")
		writeErr := os.WriteFile(swappedFilePath, []byte("bravo=2"), 0644)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}
		renameErr := os.Rename(swappedFilePath, targetFilePath)
		if renameErr != nil {
			subtest.Fatalf("RenameFailed: %v", renameErr)
		}

		openFlagCases := []int{targetFileReadOpenFlags, targetFileAppendOpenFlags}
		for _, openFlags := range openFlagCases {
			_, openErr := clerk.inspectedTargetFileOpener(
				dirHandle, targetFileName, targetSnapshot, openFlags,
			)
			if !errors.Is(openErr, ErrTargetFileChanged) {
				subtest.Fatalf("ExpectedErrTargetFileChanged, got: %v", openErr)
			}
		}
	})

	test.Run("RefusesFifoSwap", func(subtest *testing.T) {
		fifoPath := filepath.Join(tempDir, "swapped.fifo")
		fifoErr := unix.Mkfifo(fifoPath, 0o600)
		if fifoErr != nil {
			subtest.Fatalf("MkfifoFailed: %v", fifoErr)
		}
		renameErr := os.Rename(fifoPath, targetFilePath)
		if renameErr != nil {
			subtest.Fatalf("RenameFailed: %v", renameErr)
		}

		openResultChan := make(chan error, 1)
		go func() {
			fileHandle, openErr := clerk.inspectedTargetFileOpener(
				dirHandle, targetFileName, targetSnapshot, targetFileReadOpenFlags,
			)
			if fileHandle != 0 {
				_ = unix.Close(fileHandle)
			}
			openResultChan <- openErr
		}()

		select {
		case openErr := <-openResultChan:
			if !errors.Is(openErr, ErrTargetFileChanged) {
				subtest.Fatalf("ExpectedErrTargetFileChanged, got: %v", openErr)
			}
		case <-time.After(2 * time.Second):
			subtest.Fatalf("OpenBlockedOnFifo")
		}
	})
}

func TestInspectedSubDirOpener(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	inspectedDirName := "inspected"
	inspectedDirPath := filepath.Join(tempDir, inspectedDirName)
	mkdirErr := os.MkdirAll(inspectedDirPath, 0755)
	if mkdirErr != nil {
		test.Fatalf("MkdirFailed: %v", mkdirErr)
	}

	dirHandle, openErr := unix.Open(
		tempDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0,
	)
	if openErr != nil {
		test.Fatalf("OpenDirFailed: %v", openErr)
	}
	defer func() { _ = unix.Close(dirHandle) }()

	inspectedSnapshot, snapshotErr := clerk.fileSnapshotReader(
		dirHandle, inspectedDirName,
	)
	if snapshotErr != nil {
		test.Fatalf("DirSnapshotReadFailed: %v", snapshotErr)
	}

	test.Run("AcceptsTheInspectedDirectory", func(subtest *testing.T) {
		subDirHandle, openErr := clerk.inspectedSubDirOpener(
			dirHandle, inspectedDirName, inspectedSnapshot,
		)
		if openErr != nil {
			subtest.Fatalf("UnexpectedOpenError: %v", openErr)
		}
		_ = unix.Close(subDirHandle)
	})

	test.Run("RefusesSwappedDirectory", func(subtest *testing.T) {
		swappedDirPath := filepath.Join(tempDir, "swapped")
		mkdirErr := os.MkdirAll(swappedDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		renameErr := unix.Rename(swappedDirPath, inspectedDirPath)
		if renameErr != nil {
			subtest.Fatalf("RenameFailed: %v", renameErr)
		}

		_, openErr := clerk.inspectedSubDirOpener(
			dirHandle, inspectedDirName, inspectedSnapshot,
		)
		if !errors.Is(openErr, ErrTargetFileChanged) {
			subtest.Fatalf("ExpectedErrTargetFileChanged, got: %v", openErr)
		}
	})

	test.Run("RefusesSymlinkSwap", func(subtest *testing.T) {
		removeErr := os.Remove(inspectedDirPath)
		if removeErr != nil {
			subtest.Fatalf("RemoveFailed: %v", removeErr)
		}
		symlinkErr := os.Symlink(tempDir, inspectedDirPath)
		if symlinkErr != nil {
			subtest.Fatalf("SymlinkFailed: %v", symlinkErr)
		}

		_, openErr := clerk.inspectedSubDirOpener(
			dirHandle, inspectedDirName, inspectedSnapshot,
		)
		if !errors.Is(openErr, ErrTargetFileChanged) {
			subtest.Fatalf("ExpectedErrTargetFileChanged, got: %v", openErr)
		}
	})

	test.Run("RefusesNonDirectorySwap", func(subtest *testing.T) {
		removeErr := os.Remove(inspectedDirPath)
		if removeErr != nil {
			subtest.Fatalf("RemoveFailed: %v", removeErr)
		}
		writeErr := os.WriteFile(inspectedDirPath, []byte("x"), 0644)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}

		_, openErr := clerk.inspectedSubDirOpener(
			dirHandle, inspectedDirName, inspectedSnapshot,
		)
		if !errors.Is(openErr, ErrTargetFileChanged) {
			subtest.Fatalf("ExpectedErrTargetFileChanged, got: %v", openErr)
		}
	})
}

func TestTempFileNameFactory(test *testing.T) {
	clerk := FileClerk{}

	targetFileName, fileNameErr := tkValueObject.NewUnixFileName(
		"unit.service", true,
	)
	if fileNameErr != nil {
		test.Fatalf("FileNameInvalid: %v", fileNameErr)
	}

	test.Run("KeepsBaseNameWithTempMarker", func(subtest *testing.T) {
		tempFileName := clerk.TempFileNameFactory(targetFileName)

		if !strings.HasPrefix(tempFileName, ".unit.service.") {
			subtest.Errorf("MissingHiddenBaseNamePrefix: %s", tempFileName)
		}
		if !strings.HasSuffix(tempFileName, tempFileNameSuffix) {
			subtest.Errorf("MissingTempSuffix: %s", tempFileName)
		}
	})

	test.Run("DistinctAcrossCalls", func(subtest *testing.T) {
		first := clerk.TempFileNameFactory(targetFileName)
		second := clerk.TempFileNameFactory(targetFileName)
		if first == second {
			subtest.Errorf("IdenticalTempNamesAcrossCalls: %s", first)
		}
	})
}

func TestTempFilePathFactory(test *testing.T) {
	clerk := FileClerk{}

	target := absoluteFilePathForTest(test, "/tmp/example/unit.service")

	test.Run("KeepsDirAndBaseNameWithTempMarker", func(subtest *testing.T) {
		tempPath, tempPathErr := clerk.TempFilePathFactory(target)
		if tempPathErr != nil {
			subtest.Fatalf("UnexpectedError: %v", tempPathErr)
		}

		if filepath.Dir(tempPath) != "/tmp/example" {
			subtest.Errorf("UnexpectedDir: %s", tempPath)
		}
		baseName := filepath.Base(tempPath)
		if !strings.HasPrefix(baseName, ".unit.service.") {
			subtest.Errorf("MissingHiddenBaseNamePrefix: %s", baseName)
		}
		if !strings.HasSuffix(baseName, tempFileNameSuffix) {
			subtest.Errorf("MissingTempSuffix: %s", baseName)
		}
	})

	test.Run("DistinctAcrossCalls", func(subtest *testing.T) {
		first, firstErr := clerk.TempFilePathFactory(target)
		if firstErr != nil {
			subtest.Fatalf("UnexpectedError: %v", firstErr)
		}
		second, secondErr := clerk.TempFilePathFactory(target)
		if secondErr != nil {
			subtest.Fatalf("UnexpectedError: %v", secondErr)
		}
		if first == second {
			subtest.Errorf("IdenticalTempNamesAcrossCalls: %s", first)
		}
	})
}

func unresolvableUnixUserIdFinder(test *testing.T) uint32 {
	test.Helper()

	for candidate := uint32(999999); candidate > 0; candidate-- {
		_, lookupErr := user.LookupId(strconv.FormatUint(uint64(candidate), 10))
		if lookupErr != nil {
			return candidate
		}
	}

	test.Fatal("NoUnresolvableUserIdFound")
	return 0
}

func TestUpsertFile(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	newFileContent := []byte("unit content")

	overwritePolicy := FileClerkOverwritePolicyReplace
	symlinkPolicy := FileClerkSymlinkPolicyResolve

	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		test.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUsername, usernameErr := tkValueObject.NewUnixUsername(
		currentAccount.Username,
	)
	if usernameErr != nil {
		test.Fatalf("CurrentUsernameInvalid: %v", usernameErr)
	}

	test.Run("WritesNewFileWithDefaultMode", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "defaultMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			subtest.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			subtest.Errorf("ContentMismatch: %s", written)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != FileClerkDefaultNewFileMode {
			subtest.Errorf(
				"ModeMismatch: expected %v, got %v",
				FileClerkDefaultNewFileMode, fileInfo.Mode().Perm(),
			)
		}
	})

	test.Run("WritesNewFileWithStatedMode", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "statedMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		statedPermissions := os.FileMode(0640)
		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:    absoluteFilePathForTest(subtest, target),
			Permissions: &statedPermissions,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0640 {
			subtest.Errorf("ModeMismatch: expected 0640, got %v", fileInfo.Mode().Perm())
		}
	})

	test.Run("WritesNewFileWithZeroStatedMode", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "zeroMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		zeroPermissions := os.FileMode(0)
		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:    absoluteFilePathForTest(subtest, target),
			Permissions: &zeroPermissions,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0 {
			subtest.Errorf("ModeMismatch: expected 0000, got %v", fileInfo.Mode().Perm())
		}
	})

	test.Run("WritesNewFileWithProcessOwnership", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "processOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			subtest.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			subtest.Errorf("ContentMismatch: %s", written)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != os.Geteuid() || int(fileStat.Gid) != os.Getegid() {
			subtest.Errorf(
				"OwnershipMismatch: expected %d:%d, got %d:%d",
				os.Geteuid(), os.Getegid(), fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("UsesRunningProcessOwnerWhenRequested", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "runningProcessOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceRunningProcess
		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:    absoluteFilePathForTest(subtest, target),
			OwnerSource: &ownerSource,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != os.Geteuid() || int(fileStat.Gid) != os.Getegid() {
			subtest.Errorf(
				"OwnershipMismatch: expected %d:%d, got %d:%d",
				os.Geteuid(), os.Getegid(), fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("RefusesTempNameAboveNameMax", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "longName")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		longBaseName := strings.Repeat("a", 246) + ".txt"
		statedPermissions := os.FileMode(0644)
		target := filepath.Join(dir, longBaseName)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(subtest, target),
			Permissions:   &statedPermissions,
			OwnerUsername: &currentUsername,
		}, newFileContent)
		if !errors.Is(err, ErrTempFileNameTooLong) {
			subtest.Fatalf("ExpectedErrTempFileNameTooLong, got: %v", err)
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			subtest.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 0 {
			subtest.Errorf("TempFileLeaked: %v", entries)
		}
	})

	test.Run("ReplacesExistingFileWithStatedMode", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "replace")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		statedPermissions := os.FileMode(0640)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(subtest, target),
			OverwritePolicy: &overwritePolicy,
			Permissions:     &statedPermissions,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			subtest.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			subtest.Errorf("ContentMismatch: %s", written)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0640 {
			subtest.Errorf("ModeMismatch: expected 0640, got %v", fileInfo.Mode().Perm())
		}
	})

	test.Run("InheritsTargetModeAndOwnershipOnReplace", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "inherit")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0640)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		originalInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		originalStat, assertOk := originalInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(subtest, target),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			subtest.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			subtest.Errorf("ContentMismatch: %s", written)
		}

		replacedInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		if replacedInfo.Mode().Perm() != 0640 {
			subtest.Errorf(
				"ModeNotInherited: expected 0640, got %v",
				replacedInfo.Mode().Perm(),
			)
		}

		replacedStat, assertOk := replacedInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if replacedStat.Uid != originalStat.Uid || replacedStat.Gid != originalStat.Gid {
			subtest.Errorf(
				"OwnershipNotInherited: expected %d:%d, got %d:%d",
				originalStat.Uid, originalStat.Gid,
				replacedStat.Uid, replacedStat.Gid,
			)
		}
	})

	test.Run("InheritsSpecialModeBitsOnReplace", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "inheritSpecialMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0755)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		specialMode := os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0755
		err = os.Chmod(target, specialMode)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(subtest, target),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		specialModeMask := os.ModeSetuid | os.ModeSetgid | os.ModeSticky
		inheritedSpecialMode := fileInfo.Mode() & specialModeMask
		if inheritedSpecialMode != specialMode&specialModeMask {
			subtest.Errorf(
				"SpecialModeNotInherited: expected %v, got %v",
				specialMode, fileInfo.Mode(),
			)
		}
	})

	test.Run("RefusesExistingFileByDefault", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "refuseTaken")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
		}, newFileContent)
		if !errors.Is(err, ErrTargetFileExists) {
			subtest.Fatalf("ExpectedErrTargetFileExists, got: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			subtest.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != "old" {
			subtest.Errorf("ExistingContentWasOverwritten: %s", written)
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			subtest.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 1 || entries[0].Name() != "unit.service" {
			subtest.Errorf("TempFileLeaked: %v", entries)
		}
	})

	test.Run("RefusesFinalComponentSymlinkByDefault", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "refuseLink")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		canaryPath := filepath.Join(dir, "canary.txt")
		canaryContent := "do-not-touch"
		err = os.WriteFile(canaryPath, []byte(canaryContent), 0600)
		if err != nil {
			subtest.Fatalf("CanaryWriteFailed: %v", err)
		}
		target := filepath.Join(dir, "unit.service")
		err = os.Symlink(canaryPath, target)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
		}, newFileContent)
		if !errors.Is(err, ErrTargetIsSymlink) {
			subtest.Fatalf("ExpectedErrTargetIsSymlink, got: %v", err)
		}
		if !clerk.IsSymlink(target) {
			subtest.Errorf("SymlinkWasReplacedDespiteRefusal")
		}

		canary, canaryErr := os.ReadFile(canaryPath)
		if canaryErr != nil {
			subtest.Fatalf("CanaryReadFailed: %v", canaryErr)
		}
		if string(canary) != canaryContent {
			subtest.Errorf("CanaryWasModifiedThroughSymlink: %s", canary)
		}
	})

	test.Run("ResolvesFinalComponentSymlinkWhenRequested", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "resolveLink")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		canaryPath := filepath.Join(dir, "canary.txt")
		err = os.WriteFile(canaryPath, []byte("old"), 0600)
		if err != nil {
			subtest.Fatalf("CanaryWriteFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.Symlink(canaryPath, target)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(subtest, target),
			SymlinkPolicy:   &symlinkPolicy,
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		if !clerk.IsSymlink(target) {
			subtest.Errorf("TargetSymlinkWasReplacedInsteadOfFollowed")
		}
		canary, canaryErr := os.ReadFile(canaryPath)
		if canaryErr != nil {
			subtest.Fatalf("CanaryReadFailed: %v", canaryErr)
		}
		if string(canary) != string(newFileContent) {
			subtest.Errorf("CanaryContentMismatch: %s", canary)
		}
	})

	test.Run("RefusesSymlinkedParentDirectory", func(subtest *testing.T) {
		outsideDir := filepath.Join(tempDir, "outside", "nested")
		err := os.MkdirAll(outsideDir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		escapeTarget := filepath.Join(tempDir, "outside")
		trapLink := filepath.Join(tempDir, "trap")
		err = os.Symlink(escapeTarget, trapLink)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		escapedFilePath := filepath.Join(trapLink, "nested", "escape.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, escapedFilePath),
		}, newFileContent)
		if !errors.Is(err, ErrSymlinkedPathInvalid) {
			subtest.Fatalf("ExpectedErrSymlinkedPathInvalid, got: %v", err)
		}

		escapedFile := filepath.Join(outsideDir, "escape.service")
		if clerk.FileExists(escapedFile) {
			subtest.Errorf("WriteLandedOutsideIntendedTree: %s", escapedFile)
		}
	})

	test.Run("ResolvesSymlinkedParentChainWhenRequested", func(subtest *testing.T) {
		realParent := filepath.Join(tempDir, "realParent")
		err := os.MkdirAll(realParent, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		linkedParent := filepath.Join(tempDir, "linkedParent")
		err = os.Symlink(realParent, linkedParent)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		target := filepath.Join(linkedParent, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(subtest, target),
			SymlinkPolicy: &symlinkPolicy,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(filepath.Join(realParent, "unit.service"))
		if readErr != nil {
			subtest.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			subtest.Errorf("ContentMismatch: %s", written)
		}
	})

	test.Run("RefusesDegenerateFilePaths", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "degenerate")
		err := os.MkdirAll(filepath.Join(dir, "sub"), 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		testCases := []struct {
			name     string
			filePath string
		}{
			{"TrailingSlash", filepath.Join(dir, "sub") + "/"},
			{"DotAsFinalComponent", filepath.Join(dir, "sub") + "/."},
			{"DotDotAsFinalComponent", filepath.Join(dir, "sub") + "/.."},
		}

		for _, testCase := range testCases {
			subtest.Run(testCase.name, func(subtest *testing.T) {
				err := clerk.UpsertFile(FileUpsertSettings{
					FilePath: absoluteFilePathForTest(subtest, testCase.filePath),
				}, newFileContent)
				if !errors.Is(err, ErrFileNameInvalid) {
					subtest.Errorf(
						"ExpectedErrFileNameInvalid, got: %v [%s]",
						err, testCase.filePath,
					)
				}
			})
		}
	})

	test.Run("RefusesNonDirectoryComponent", func(subtest *testing.T) {
		blockerFile := filepath.Join(tempDir, "blocker")
		err := os.WriteFile(blockerFile, []byte("x"), 0644)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}

		blockedFilePath := filepath.Join(blockerFile, "nested", "blocked.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, blockedFilePath),
		}, newFileContent)
		if !errors.Is(err, ErrTargetNotDirectory) {
			subtest.Errorf("ExpectedErrTargetNotDirectory, got: %v", err)
		}
	})

	test.Run("ToleratesDotAndEmptyComponents", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "dotted", "inner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		dottedFilePath := tempDir + "/./dotted//inner/dotted.txt"
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, dottedFilePath),
		}, newFileContent)
		if err != nil {
			subtest.Errorf("UnexpectedSurprise: %v", err)
		}
	})

	test.Run("ReportsUninspectableDirStep", func(subtest *testing.T) {
		missingFilePath := filepath.Join(tempDir, "missing", "deeper", "file.txt")
		err := clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, missingFilePath),
		}, newFileContent)
		if err == nil {
			subtest.Fatalf("MissingErrorForUninspectableStep")
		}
		if !strings.Contains(err.Error(), "PathCheckFailed") {
			subtest.Errorf("ErrorMissingPathCheckFailed: %v", err)
		}
	})

	test.Run("ReportsUnknownTrustedDirOwner", func(subtest *testing.T) {
		unknownUsername, usernameErr := tkValueObject.NewUnixUsername(
			"no-such-user-xyz-42",
		)
		if usernameErr != nil {
			subtest.Fatalf("UnknownUsernameInvalid: %v", usernameErr)
		}

		filePath := filepath.Join(tempDir, "unknownOwner.txt")
		err := clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, filePath),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				unknownUsername,
			},
		}, newFileContent)
		if err == nil {
			subtest.Fatalf("MissingErrorForUnknownTrustedDirOwner")
		}
		if !strings.Contains(err.Error(), "OwnerLookupFailed") {
			subtest.Errorf("ErrorMissingOwnerLookupFailed: %v", err)
		}
	})

	test.Run("AcceptsUnresolvableTrustedDirOwnerUserId", func(subtest *testing.T) {
		if os.Geteuid() == 0 {
			subtest.Skip("NonRootRequired")
		}

		unresolvableUserId, userIdErr := tkValueObject.NewUnixUserId(
			unresolvableUnixUserIdFinder(subtest),
		)
		if userIdErr != nil {
			subtest.Fatalf("UnresolvableUserIdInvalid: %v", userIdErr)
		}

		filePath := filepath.Join(tempDir, "numericOwner.txt")
		err := clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, filePath),
			TrustedDirOwnerUserIds: []tkValueObject.UnixUserId{
				unresolvableUserId,
			},
		}, newFileContent)
		if !errors.Is(err, ErrDirectoryOwnerInvalid) {
			subtest.Errorf("ExpectedErrDirectoryOwnerInvalid, got: %v", err)
		}
	})

	test.Run("RefusesWritableDirChainWhenRequired", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "writableUpsertDir")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chmod(dir, 0o777)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}

		sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused
		filePath := filepath.Join(dir, "strict.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:       absoluteFilePathForTest(subtest, filePath),
			DirChainPolicy: &sharedWriteRefusedPolicy,
		}, newFileContent)
		if !errors.Is(err, ErrDirectoryWritableByOthers) {
			subtest.Errorf("ExpectedErrDirectoryWritableByOthers, got: %v", err)
		}
	})

	test.Run("ReadersNeverSeePartialContent", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "atomic")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "target.txt")
		targetPath := absoluteFilePathForTest(subtest, target)
		contentSizeBytes := 1024 * 1024
		contentA := strings.Repeat("a", contentSizeBytes)
		contentB := strings.Repeat("b", contentSizeBytes)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        targetPath,
			OverwritePolicy: &overwritePolicy,
		}, []byte(contentA))
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		writesDone := make(chan struct{})
		go func() {
			defer close(writesDone)
			for writeIdx := range 200 {
				content := contentA
				if writeIdx%2 == 1 {
					content = contentB
				}
				writeErr := clerk.UpsertFile(FileUpsertSettings{
					FilePath:        targetPath,
					OverwritePolicy: &overwritePolicy,
				}, []byte(content))
				if writeErr != nil {
					subtest.Errorf("ConcurrentWriteFailed: %v", writeErr)
					return
				}
			}
		}()

	observersLoop:
		for {
			readContent, readErr := os.ReadFile(target)
			if readErr != nil {
				subtest.Errorf("ConcurrentReadFailed: %v", readErr)
				break
			}
			if string(readContent) != contentA && string(readContent) != contentB {
				subtest.Errorf(
					"ReaderObservedPartialContent: length %d", len(readContent),
				)
				break
			}
			select {
			case <-writesDone:
				break observersLoop
			default:
			}
		}
	})

	test.Run("RefusesDirectoryTargetLeavingNoTempFile", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "leakCheck")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		blockingDir := filepath.Join(dir, "taken.txt")
		err = os.Mkdir(blockingDir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(subtest, blockingDir),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if !errors.Is(err, ErrTargetIsDirectory) {
			subtest.Fatalf("ExpectedErrTargetIsDirectory, got: %v", err)
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			subtest.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 1 || entries[0].Name() != "taken.txt" {
			subtest.Errorf("TempFileLeaked: %v", entries)
		}
	})

	test.Run("SetsRequestedOwner", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			subtest.Fatalf("GidParseFailed: %v", gidErr)
		}

		dir := filepath.Join(tempDir, "ownership")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			subtest.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		statedPermissions := os.FileMode(0644)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(subtest, filepath.Join(dir, "owned.txt")),
			Permissions:   &statedPermissions,
			OwnerUsername: &ownerName,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(filepath.Join(dir, "owned.txt"))
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			subtest.Errorf(
				"OwnerMismatch: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("RefusesDirectoryChainOwnedByUnexpectedUser", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		daemon, lookupErr := user.Lookup("daemon")
		if lookupErr != nil {
			subtest.Skipf("DaemonUserMissing: %v", lookupErr)
		}
		_, nobodyErr := user.Lookup("nobody")
		if nobodyErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", nobodyErr)
		}
		daemonUid, uidErr := strconv.Atoi(daemon.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}

		dir := filepath.Join(tempDir, "wrongOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chown(dir, daemonUid, -1)
		if err != nil {
			subtest.Fatalf("ChownFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			subtest.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		target := filepath.Join(dir, "refused.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				ownerName,
			},
		}, newFileContent)
		if !errors.Is(err, ErrDirectoryOwnerInvalid) {
			subtest.Errorf(
				"ExpectedErrDirectoryOwnerInvalid, got: %v", err,
			)
		}
	})

	test.Run("InheritsTargetOwnerOnReplace", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			subtest.Fatalf("GidParseFailed: %v", gidErr)
		}

		dir := filepath.Join(tempDir, "inheritOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "owned.txt")
		err = os.WriteFile(target, []byte("old"), 0640)
		if err != nil {
			subtest.Fatalf("WriteFileFailed: %v", err)
		}
		err = os.Chown(target, nobodyUid, nobodyGid)
		if err != nil {
			subtest.Fatalf("ChownFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(subtest, target),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			subtest.Errorf(
				"OwnerNotInherited: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("UsesContainingDirectoryOwnerWhenRequested", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			subtest.Fatalf("GidParseFailed: %v", gidErr)
		}
		nobodyUserId, userIdErr := tkValueObject.NewUnixUserId(nobodyUid)
		if userIdErr != nil {
			subtest.Fatalf("NobodyUserIdInvalid: %v", userIdErr)
		}

		dir := filepath.Join(tempDir, "dirOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chown(dir, nobodyUid, nobodyGid)
		if err != nil {
			subtest.Fatalf("ChownFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceContainingDirectory
		target := filepath.Join(dir, "owned.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
			TrustedDirOwnerUserIds: []tkValueObject.UnixUserId{
				nobodyUserId,
			},
			OwnerSource: &ownerSource,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			subtest.Errorf(
				"OwnerMismatch: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("StatesOwnerGroup", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		daemon, daemonErr := user.Lookup("daemon")
		if daemonErr != nil {
			subtest.Skipf("DaemonUserMissing: %v", daemonErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		daemonGid, gidErr := strconv.Atoi(daemon.Gid)
		if gidErr != nil {
			subtest.Fatalf("GidParseFailed: %v", gidErr)
		}

		dir := filepath.Join(tempDir, "statedGroup")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			subtest.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		ownerGroupId, groupIdErr := tkValueObject.NewUnixGroupId(daemonGid)
		if groupIdErr != nil {
			subtest.Fatalf("GroupIdInvalid: %v", groupIdErr)
		}

		target := filepath.Join(dir, "owned.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(subtest, target),
			OwnerUsername: &ownerName,
			OwnerGroupId:  &ownerGroupId,
		}, newFileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			subtest.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			subtest.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != daemonGid {
			subtest.Errorf(
				"OwnershipMismatch: expected %d:%d, got %d:%d",
				nobodyUid, daemonGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	test.Run("RefusesStatedOwnerWithContainingDirectorySource", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "ownerConflict")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceContainingDirectory
		target := filepath.Join(dir, "conflict.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(subtest, target),
			OwnerSource:   &ownerSource,
			OwnerUsername: &currentUsername,
		}, newFileContent)
		if !errors.Is(err, ErrOwnerSourceConflict) {
			subtest.Fatalf("ExpectedErrOwnerSourceConflict, got: %v", err)
		}
		if clerk.FileExists(target) {
			subtest.Errorf("TargetWrittenDespiteConflict")
		}
	})

	test.Run("RefusesStatedOwnerWithRunningProcessSource", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "ownerProcessConflict")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceRunningProcess
		target := filepath.Join(dir, "conflict.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(subtest, target),
			OwnerSource:   &ownerSource,
			OwnerUsername: &currentUsername,
		}, newFileContent)
		if !errors.Is(err, ErrOwnerSourceConflict) {
			subtest.Fatalf("ExpectedErrOwnerSourceConflict, got: %v", err)
		}
		if clerk.FileExists(target) {
			subtest.Errorf("TargetWrittenDespiteConflict")
		}
	})

	test.Run("RefusesUnknownPolicyValues", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "unknownPolicy")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unknown.txt")
		targetPath := absoluteFilePathForTest(subtest, target)
		unsupportedOwnerSource := FileClerkOwnerSource("unsupported")
		unsupportedSymlinkPolicy := FileClerkSymlinkPolicy("unsupported")
		unsupportedOverwritePolicy := FileClerkOverwritePolicy("unsupported")
		unsupportedDirChainPolicy := FileClerkDirChainPolicy("unsupported")

		testCases := []struct {
			name          string
			settings      FileUpsertSettings
			expectedError error
		}{
			{
				"DirChainPolicy",
				FileUpsertSettings{
					FilePath:       targetPath,
					DirChainPolicy: &unsupportedDirChainPolicy,
				},
				ErrDirChainPolicyInvalid,
			},
			{
				"SymlinkPolicy",
				FileUpsertSettings{
					FilePath:      targetPath,
					SymlinkPolicy: &unsupportedSymlinkPolicy,
				},
				ErrSymlinkPolicyInvalid,
			},
			{
				"OverwritePolicy",
				FileUpsertSettings{
					FilePath:        targetPath,
					OverwritePolicy: &unsupportedOverwritePolicy,
				},
				ErrOverwritePolicyInvalid,
			},
			{
				"OwnerSource",
				FileUpsertSettings{
					FilePath:    targetPath,
					OwnerSource: &unsupportedOwnerSource,
				},
				ErrOwnerSourceInvalid,
			},
		}

		for _, testCase := range testCases {
			subtest.Run(testCase.name, func(subtest *testing.T) {
				err := clerk.UpsertFile(testCase.settings, newFileContent)
				if !errors.Is(err, testCase.expectedError) {
					subtest.Errorf("Expected%v, got: %v", testCase.expectedError, err)
				}
			})
		}

		if clerk.FileExists(target) {
			subtest.Errorf("TargetWrittenDespiteInvalidSettings")
		}
	})

	test.Run("RefusesUnprivilegedOwnerChange", func(subtest *testing.T) {
		if os.Geteuid() == 0 {
			subtest.Skip("NonRootRequired")
		}

		dir := filepath.Join(tempDir, "unprivilegedOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		foreignOwnerUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid() + 1)
		if userIdErr != nil {
			subtest.Fatalf("ForeignOwnerUserIdInvalid: %v", userIdErr)
		}
		foreignOwnerGroupId, groupIdErr := tkValueObject.NewUnixGroupId(os.Getegid())
		if groupIdErr != nil {
			subtest.Fatalf("ForeignOwnerGroupIdInvalid: %v", groupIdErr)
		}

		target := filepath.Join(dir, "owned.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:     absoluteFilePathForTest(subtest, target),
			OwnerUserId:  &foreignOwnerUserId,
			OwnerGroupId: &foreignOwnerGroupId,
		}, newFileContent)
		if !errors.Is(err, ErrFileOwnerChangeFailed) {
			subtest.Fatalf("ExpectedErrFileOwnerChangeFailed, got: %v", err)
		}
		if clerk.FileExists(target) {
			subtest.Errorf("TargetWrittenDespiteOwnerChangeFailure")
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			subtest.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 0 {
			subtest.Errorf("TempFileLeaked: %v", entries)
		}
	})
}

func TestFileUpsertOwnerResolver(test *testing.T) {
	clerk := FileClerk{}

	runningProcessUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid())
	if userIdErr != nil {
		test.Fatalf("RunningProcessUserIdInvalid: %v", userIdErr)
	}
	runningProcessGroupId, groupIdErr := tkValueObject.NewUnixGroupId(os.Getegid())
	if groupIdErr != nil {
		test.Fatalf("RunningProcessGroupIdInvalid: %v", groupIdErr)
	}

	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		test.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUsername, usernameErr := tkValueObject.NewUnixUsername(
		currentAccount.Username,
	)
	if usernameErr != nil {
		test.Fatalf("CurrentUsernameInvalid: %v", usernameErr)
	}
	currentUserId, userIdErr := tkValueObject.NewUnixUserId(currentAccount.Uid)
	if userIdErr != nil {
		test.Fatalf("CurrentUserIdInvalid: %v", userIdErr)
	}
	currentPrimaryGroupId, groupIdErr := tkValueObject.NewUnixGroupId(
		currentAccount.Gid,
	)
	if groupIdErr != nil {
		test.Fatalf("CurrentPrimaryGroupIdInvalid: %v", groupIdErr)
	}

	existingFileOwnerUserId := tkValueObject.UnixUserId(1234)
	existingFileOwnerGroupId := tkValueObject.UnixGroupId(1235)
	containingDirOwnerUserId := tkValueObject.UnixUserId(2345)
	containingDirOwnerGroupId := tkValueObject.UnixGroupId(2346)
	statedOwnerUserId := tkValueObject.UnixUserId(3456)
	statedGroupId := tkValueObject.UnixGroupId(3457)

	targetSnapshotExisting := fileStatSnapshot{
		Exists:       true,
		OwnerUserId:  existingFileOwnerUserId,
		OwnerGroupId: existingFileOwnerGroupId,
	}
	targetSnapshotMissing := fileStatSnapshot{}
	containingDirStat := unix.Stat_t{
		Uid: uint32(containingDirOwnerUserId),
		Gid: uint32(containingDirOwnerGroupId),
	}

	testCases := []struct {
		name            string
		ownerSource     FileClerkOwnerSource
		ownerUsername   *tkValueObject.UnixUsername
		ownerUserId     *tkValueObject.UnixUserId
		ownerGroupId    *tkValueObject.UnixGroupId
		targetSnapshot  fileStatSnapshot
		expectedUserId  tkValueObject.UnixUserId
		expectedGroupId tkValueObject.UnixGroupId
		expectedError   error
	}{
		{
			name:            "StatedAccountUsesPrimaryGroup",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			ownerUsername:   &currentUsername,
			targetSnapshot:  targetSnapshotMissing,
			expectedUserId:  currentUserId,
			expectedGroupId: currentPrimaryGroupId,
		},
		{
			name:            "StatedUserIdAndGroupWin",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			ownerUserId:     &statedOwnerUserId,
			ownerGroupId:    &statedGroupId,
			targetSnapshot:  targetSnapshotExisting,
			expectedUserId:  statedOwnerUserId,
			expectedGroupId: statedGroupId,
		},
		{
			name:            "ExistingFileSourceInheritsTargetOwnership",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			targetSnapshot:  targetSnapshotExisting,
			expectedUserId:  existingFileOwnerUserId,
			expectedGroupId: existingFileOwnerGroupId,
		},
		{
			name:            "ExistingFileSourceFallsBackToRunningProcess",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			targetSnapshot:  targetSnapshotMissing,
			expectedUserId:  runningProcessUserId,
			expectedGroupId: runningProcessGroupId,
		},
		{
			name:            "ContainingDirectorySourceUsesDirectoryOwnership",
			ownerSource:     FileClerkOwnerSourceContainingDirectory,
			targetSnapshot:  targetSnapshotMissing,
			expectedUserId:  containingDirOwnerUserId,
			expectedGroupId: containingDirOwnerGroupId,
		},
		{
			name:            "RunningProcessSourceUsesProcessOwnership",
			ownerSource:     FileClerkOwnerSourceRunningProcess,
			targetSnapshot:  targetSnapshotMissing,
			expectedUserId:  runningProcessUserId,
			expectedGroupId: runningProcessGroupId,
		},
		{
			name:            "StatedGroupOverridesInheritedGroup",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			ownerGroupId:    &statedGroupId,
			targetSnapshot:  targetSnapshotExisting,
			expectedUserId:  existingFileOwnerUserId,
			expectedGroupId: statedGroupId,
		},
		{
			name:           "StatedAccountConflictsWithContainingDirectory",
			ownerSource:    FileClerkOwnerSourceContainingDirectory,
			ownerUsername:  &currentUsername,
			targetSnapshot: targetSnapshotMissing,
			expectedError:  ErrOwnerSourceConflict,
		},
		{
			name:           "StatedAccountConflictsWithRunningProcess",
			ownerSource:    FileClerkOwnerSourceRunningProcess,
			ownerUsername:  &currentUsername,
			targetSnapshot: targetSnapshotMissing,
			expectedError:  ErrOwnerSourceConflict,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			ownership, err := clerk.fileUpsertOwnerResolver(
				testCase.ownerSource, testCase.ownerUsername, testCase.ownerUserId,
				testCase.ownerGroupId, testCase.targetSnapshot, containingDirStat,
			)

			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					subtest.Fatalf("Expected%v, got: %v", testCase.expectedError, err)
				}
				return
			}
			if err != nil {
				subtest.Fatalf("UnexpectedError: %v", err)
			}
			if ownership.UserId != testCase.expectedUserId {
				subtest.Errorf(
					"UserIdMismatch: expected %d, got %d",
					testCase.expectedUserId, ownership.UserId,
				)
			}
			if ownership.GroupId != testCase.expectedGroupId {
				subtest.Errorf(
					"GroupIdMismatch: expected %d, got %d",
					testCase.expectedGroupId, ownership.GroupId,
				)
			}
		})
	}
}

func TestFileUpsertPermissionsResolver(test *testing.T) {
	clerk := FileClerk{}

	statedPermissions := os.FileMode(0640)
	zeroPermissions := os.FileMode(0)
	existingPermissions := os.FileMode(0604)

	targetSnapshotExisting := fileStatSnapshot{
		Exists:      true,
		Permissions: existingPermissions,
	}

	testCases := []struct {
		name                 string
		statedPermissionsPtr *os.FileMode
		targetSnapshot       fileStatSnapshot
		expectedPermissions  os.FileMode
	}{
		{
			"StatedPermissionsWin",
			&statedPermissions, targetSnapshotExisting, statedPermissions,
		},
		{
			"StatedZeroPermissionsWin",
			&zeroPermissions, targetSnapshotExisting, zeroPermissions,
		},
		{
			"InheritsExistingPermissions",
			nil, targetSnapshotExisting, existingPermissions,
		},
		{
			"DefaultsForMissingTarget",
			nil, fileStatSnapshot{}, FileClerkDefaultNewFileMode,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			permissions := clerk.fileUpsertPermissionsResolver(
				testCase.statedPermissionsPtr, testCase.targetSnapshot,
			)
			if permissions != testCase.expectedPermissions {
				subtest.Errorf(
					"PermissionsMismatch: expected %v, got %v",
					testCase.expectedPermissions, permissions,
				)
			}
		})
	}
}

func TestUnixFileModeConverter(test *testing.T) {
	clerk := FileClerk{}

	specialModeMask := os.ModeSetuid | os.ModeSetgid | os.ModeSticky

	testCases := []struct {
		name             string
		rawMode          uint32
		expectedFileMode os.FileMode
	}{
		{"PermissionBits", unix.S_IFREG | 0o640, os.FileMode(0o640)},
		{"SetuidBit", unix.S_IFREG | 0o4755, os.FileMode(0o755) | os.ModeSetuid},
		{"SetgidBit", unix.S_IFREG | 0o2755, os.FileMode(0o755) | os.ModeSetgid},
		{"StickyBit", unix.S_IFREG | 0o1755, os.FileMode(0o755) | os.ModeSticky},
		{
			"AllSpecialBits",
			unix.S_IFREG | 0o7755,
			os.FileMode(0o755) | specialModeMask,
		},
		{"DropsFileTypeBits", unix.S_IFDIR | 0o750, os.FileMode(0o750)},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			fileMode := clerk.unixFileModeConverter(testCase.rawMode)
			if fileMode != testCase.expectedFileMode {
				subtest.Errorf(
					"FileModeMismatch: expected %v, got %v",
					testCase.expectedFileMode, fileMode,
				)
			}
		})
	}
}

func TestOpenRedirectProofDirChain(test *testing.T) {
	clerk := FileClerk{}
	tempDir := test.TempDir()

	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		test.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUserId, userIdErr := tkValueObject.NewUnixUserId(currentAccount.Uid)
	if userIdErr != nil {
		test.Fatalf("CurrentUserIdInvalid: %v", userIdErr)
	}
	currentTrustedDirOwnerIds := trustedDirOwnerIdSet{
		currentUserId.Uint64(): struct{}{},
	}

	sharedWriteAllowedPolicy := FileClerkDirChainPolicySharedWriteAllowed
	sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused

	test.Run("RefusesParentTraversalComponents", func(subtest *testing.T) {
		traversalPath := tkValueObject.UnixAbsoluteFilePath(
			tempDir + "/../traversal.txt",
		)
		_, dirChainErr := clerk.openRedirectProofDirChain(
			traversalPath, currentTrustedDirOwnerIds, sharedWriteAllowedPolicy,
		)
		if !errors.Is(dirChainErr, ErrDirPathTraversalInvalid) {
			subtest.Errorf("ExpectedErrDirPathTraversalInvalid, got: %v", dirChainErr)
		}
	})

	test.Run("RefusesWritableComponentWhenRequired", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "writable")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chmod(dir, 0o777)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}

		_, strictChainErr := clerk.openRedirectProofDirChain(
			absoluteFilePathForTest(subtest, dir), currentTrustedDirOwnerIds,
			sharedWriteRefusedPolicy,
		)
		if !errors.Is(strictChainErr, ErrDirectoryWritableByOthers) {
			subtest.Errorf(
				"ExpectedErrDirectoryWritableByOthers, got: %v", strictChainErr,
			)
		}

		lenientHandle, lenientChainErr := clerk.openRedirectProofDirChain(
			absoluteFilePathForTest(subtest, dir), currentTrustedDirOwnerIds,
			sharedWriteAllowedPolicy,
		)
		if lenientChainErr != nil {
			subtest.Fatalf("UnexpectedErrorForLenientChain: %v", lenientChainErr)
		}
		_ = unix.Close(lenientHandle)
	})

	test.Run("AcceptsStickyWritableComponentWhenRequired", func(subtest *testing.T) {
		dir := filepath.Join(tempDir, "stickyWritable")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chmod(dir, os.ModeSticky|0o777)
		if err != nil {
			subtest.Fatalf("ChmodFailed: %v", err)
		}

		dirHandle, strictChainErr := clerk.openRedirectProofDirChain(
			absoluteFilePathForTest(subtest, dir), currentTrustedDirOwnerIds,
			sharedWriteRefusedPolicy,
		)
		if strictChainErr != nil {
			subtest.Fatalf("UnexpectedSurprise: %v", strictChainErr)
		}
		_ = unix.Close(dirHandle)
	})

	test.Run("HandleStaysOnOriginalInodeAfterPathSwap", func(subtest *testing.T) {
		parentDir := filepath.Join(tempDir, "pinning")
		originalDir := filepath.Join(parentDir, "original")
		err := os.MkdirAll(originalDir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}

		dirHandle, dirChainErr := clerk.openRedirectProofDirChain(
			absoluteFilePathForTest(subtest, originalDir), currentTrustedDirOwnerIds,
			sharedWriteAllowedPolicy,
		)
		if dirChainErr != nil {
			subtest.Fatalf("RedirectProofFailed: %v", dirChainErr)
		}
		defer func() { _ = unix.Close(dirHandle) }()

		movedDir := filepath.Join(parentDir, "moved")
		err = os.Rename(originalDir, movedDir)
		if err != nil {
			subtest.Fatalf("RenameFailed: %v", err)
		}
		escapeDir := filepath.Join(parentDir, "escape")
		err = os.MkdirAll(escapeDir, 0755)
		if err != nil {
			subtest.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Symlink(escapeDir, originalDir)
		if err != nil {
			subtest.Fatalf("SymlinkFailed: %v", err)
		}

		stageHandle, openErr := unix.Openat(
			dirHandle, "staged.txt",
			unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL, 0o600,
		)
		if openErr != nil {
			subtest.Fatalf("OpenatFailed: %v", openErr)
		}
		_ = unix.Close(stageHandle)

		if !clerk.FileExists(filepath.Join(movedDir, "staged.txt")) {
			subtest.Errorf("WriteDidNotLandOnScoutedInode")
		}
		if clerk.FileExists(filepath.Join(escapeDir, "staged.txt")) {
			subtest.Errorf("WriteFollowedReplacementSymlink")
		}
	})
}

func TestTrustedDirOwnerIdSetResolver(test *testing.T) {
	clerk := FileClerk{}

	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		test.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUsername, usernameErr := tkValueObject.NewUnixUsername(
		currentAccount.Username,
	)
	if usernameErr != nil {
		test.Fatalf("CurrentUsernameInvalid: %v", usernameErr)
	}
	currentUserId, userIdErr := tkValueObject.NewUnixUserId(currentAccount.Uid)
	if userIdErr != nil {
		test.Fatalf("CurrentUserIdInvalid: %v", userIdErr)
	}
	runningProcessUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid())
	if userIdErr != nil {
		test.Fatalf("RunningProcessUserIdInvalid: %v", userIdErr)
	}
	unknownUsername, usernameErr := tkValueObject.NewUnixUsername(
		"no-such-user-xyz-42",
	)
	if usernameErr != nil {
		test.Fatalf("UnknownUsernameInvalid: %v", usernameErr)
	}
	arbitraryUserId := tkValueObject.UnixUserId(4242)

	testCases := []struct {
		name            string
		usernames       []tkValueObject.UnixUsername
		userIds         []tkValueObject.UnixUserId
		expectedUserIds []tkValueObject.UnixUserId
		expectedError   string
	}{
		{
			name:            "EmptySetFallsBackToRunningProcessAccount",
			expectedUserIds: []tkValueObject.UnixUserId{runningProcessUserId},
		},
		{
			name:            "ResolvesUsernameToUserId",
			usernames:       []tkValueObject.UnixUsername{currentUsername},
			expectedUserIds: []tkValueObject.UnixUserId{currentUserId},
		},
		{
			name:            "KeepsUserIdWithoutLookup",
			userIds:         []tkValueObject.UnixUserId{arbitraryUserId},
			expectedUserIds: []tkValueObject.UnixUserId{arbitraryUserId},
		},
		{
			name:      "MixedUsernameAndUserIdBothCount",
			usernames: []tkValueObject.UnixUsername{currentUsername},
			userIds:   []tkValueObject.UnixUserId{arbitraryUserId},
			expectedUserIds: []tkValueObject.UnixUserId{
				currentUserId, arbitraryUserId,
			},
		},
		{
			name:          "UnknownUsernameFails",
			usernames:     []tkValueObject.UnixUsername{unknownUsername},
			expectedError: "OwnerLookupFailed",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			trustedDirOwnerIds, err := clerk.trustedDirOwnerIdSetResolver(
				testCase.usernames, testCase.userIds,
			)
			if testCase.expectedError != "" {
				errorMissing := err == nil ||
					!strings.Contains(err.Error(), testCase.expectedError)
				if errorMissing {
					subtest.Fatalf(
						"Expected%sError, got: %v", testCase.expectedError, err,
					)
				}
				return
			}
			if err != nil {
				subtest.Fatalf("UnexpectedError: %v", err)
			}
			if len(trustedDirOwnerIds) != len(testCase.expectedUserIds) {
				subtest.Errorf(
					"SetSizeMismatch: expected %d, got %d",
					len(testCase.expectedUserIds), len(trustedDirOwnerIds),
				)
			}
			for _, expectedUserId := range testCase.expectedUserIds {
				_, isTrusted := trustedDirOwnerIds[expectedUserId.Uint64()]
				if !isTrusted {
					subtest.Errorf("MissingTrustedOwnerId: %d", expectedUserId)
				}
			}
		})
	}
}

func TestOpenRedirectProofDirChainTrustedOwnerSets(test *testing.T) {
	if os.Geteuid() != 0 {
		test.Skip("RootPrivilegesRequired")
	}

	clerk := FileClerk{}
	tempDir := test.TempDir()

	candidateAccountNames := []string{"nobody", "daemon", "bin", "mail", "games"}
	ownerAccountNames := []string{}
	for _, candidateAccountName := range candidateAccountNames {
		_, lookupErr := user.Lookup(candidateAccountName)
		if lookupErr != nil {
			continue
		}
		ownerAccountNames = append(ownerAccountNames, candidateAccountName)
		if len(ownerAccountNames) == 3 {
			break
		}
	}
	if len(ownerAccountNames) < 3 {
		test.Skip("ThreeOwnerAccountsRequired")
	}

	ownerUserIds := make([]tkValueObject.UnixUserId, 0, len(ownerAccountNames))
	ownerUids := make([]int, 0, len(ownerAccountNames))
	for _, ownerAccountName := range ownerAccountNames {
		ownerAccount, lookupErr := user.Lookup(ownerAccountName)
		if lookupErr != nil {
			test.Skipf("AccountMissing: %s: %v", ownerAccountName, lookupErr)
		}
		ownerUid, uidParseErr := strconv.Atoi(ownerAccount.Uid)
		if uidParseErr != nil {
			test.Fatalf("UidParseFailed: %s: %v", ownerAccountName, uidParseErr)
		}
		ownerUserId, userIdErr := tkValueObject.NewUnixUserId(ownerUid)
		if userIdErr != nil {
			test.Fatalf("OwnerUserIdInvalid: %s: %v", ownerAccountName, userIdErr)
		}
		ownerUids = append(ownerUids, ownerUid)
		ownerUserIds = append(ownerUserIds, ownerUserId)
	}

	firstOwnerDir := filepath.Join(tempDir, "firstOwner")
	secondOwnerDir := filepath.Join(firstOwnerDir, "secondOwner")
	thirdOwnerDir := filepath.Join(secondOwnerDir, "thirdOwner")
	mkdirErr := os.MkdirAll(thirdOwnerDir, 0755)
	if mkdirErr != nil {
		test.Fatalf("MkdirFailed: %v", mkdirErr)
	}
	ownedDirs := []string{firstOwnerDir, secondOwnerDir, thirdOwnerDir}
	for dirIndex, ownedDir := range ownedDirs {
		chownErr := os.Chown(ownedDir, ownerUids[dirIndex], -1)
		if chownErr != nil {
			test.Fatalf("ChownFailed: %s: %v", ownedDir, chownErr)
		}
	}

	sharedWriteAllowedPolicy := FileClerkDirChainPolicySharedWriteAllowed

	testCases := []struct {
		name              string
		walkPath          tkValueObject.UnixAbsoluteFilePath
		trustedUserIds    []tkValueObject.UnixUserId
		expectedError     error
		expectedComponent string
	}{
		{
			name:           "AcceptsChainWithEveryOwnerNamed",
			walkPath:       absoluteFilePathForTest(test, thirdOwnerDir),
			trustedUserIds: ownerUserIds,
		},
		{
			name:              "RefusesFirstOwnerWhenUnnamed",
			walkPath:          absoluteFilePathForTest(test, thirdOwnerDir),
			trustedUserIds:    ownerUserIds[1:],
			expectedError:     ErrDirectoryOwnerInvalid,
			expectedComponent: "firstOwner",
		},
		{
			name:              "RefusesThirdOwnerWhenUnnamed",
			walkPath:          absoluteFilePathForTest(test, thirdOwnerDir),
			trustedUserIds:    ownerUserIds[:2],
			expectedError:     ErrDirectoryOwnerInvalid,
			expectedComponent: "thirdOwner",
		},
		{
			name:              "RefusesChainWhenTrustedSetEmpty",
			walkPath:          absoluteFilePathForTest(test, thirdOwnerDir),
			expectedError:     ErrDirectoryOwnerInvalid,
			expectedComponent: "firstOwner",
		},
		{
			name:           "AcceptsSingleOwnerChainWhenNamed",
			walkPath:       absoluteFilePathForTest(test, firstOwnerDir),
			trustedUserIds: ownerUserIds[:1],
		},
		{
			name:              "RefusesSingleOwnerChainWhenUnnamed",
			walkPath:          absoluteFilePathForTest(test, firstOwnerDir),
			trustedUserIds:    ownerUserIds[1:2],
			expectedError:     ErrDirectoryOwnerInvalid,
			expectedComponent: "firstOwner",
		},
		{
			name:           "AcceptsRootOwnedChainWithUnrelatedTrustedOwner",
			walkPath:       absoluteFilePathForTest(test, tempDir),
			trustedUserIds: ownerUserIds[:1],
		},
		{
			name:     "AcceptsRootOwnedChainWithEmptySet",
			walkPath: absoluteFilePathForTest(test, tempDir),
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			trustedDirOwnerIds := trustedDirOwnerIdSet{}
			for _, trustedUserId := range testCase.trustedUserIds {
				trustedDirOwnerIds[trustedUserId.Uint64()] = struct{}{}
			}

			dirHandle, err := clerk.openRedirectProofDirChain(
				testCase.walkPath, trustedDirOwnerIds, sharedWriteAllowedPolicy,
			)
			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					subtest.Fatalf("Expected%v, got: %v", testCase.expectedError, err)
				}
				if !strings.Contains(err.Error(), testCase.expectedComponent) {
					subtest.Errorf(
						"ErrorDoesNotNameComponent: %s: %v",
						testCase.expectedComponent, err,
					)
				}
				return
			}
			if err != nil {
				subtest.Fatalf("UnexpectedError: %v", err)
			}
			_ = unix.Close(dirHandle)
		})
	}
}

func TestUpsertFileTrustedDirOwners(test *testing.T) {
	if os.Geteuid() != 0 {
		test.Skip("RootPrivilegesRequired")
	}

	clerk := FileClerk{}
	tempDir := test.TempDir()

	nobodyAccount, nobodyErr := user.Lookup("nobody")
	if nobodyErr != nil {
		test.Skipf("NobodyUserMissing: %v", nobodyErr)
	}
	daemonAccount, daemonErr := user.Lookup("daemon")
	if daemonErr != nil {
		test.Skipf("DaemonUserMissing: %v", daemonErr)
	}
	nobodyUid, nobodyUidErr := strconv.Atoi(nobodyAccount.Uid)
	if nobodyUidErr != nil {
		test.Fatalf("NobodyUidParseFailed: %v", nobodyUidErr)
	}
	daemonUid, daemonUidErr := strconv.Atoi(daemonAccount.Uid)
	if daemonUidErr != nil {
		test.Fatalf("DaemonUidParseFailed: %v", daemonUidErr)
	}

	nobodyUsername, usernameErr := tkValueObject.NewUnixUsername(
		nobodyAccount.Username,
	)
	if usernameErr != nil {
		test.Fatalf("NobodyUsernameInvalid: %v", usernameErr)
	}
	daemonUserId, userIdErr := tkValueObject.NewUnixUserId(daemonUid)
	if userIdErr != nil {
		test.Fatalf("DaemonUserIdInvalid: %v", userIdErr)
	}

	nobodyDir := filepath.Join(tempDir, "nobodyOwner")
	daemonDir := filepath.Join(nobodyDir, "daemonOwner")
	mkdirErr := os.MkdirAll(daemonDir, 0755)
	if mkdirErr != nil {
		test.Fatalf("MkdirFailed: %v", mkdirErr)
	}
	nobodyChownErr := os.Chown(nobodyDir, nobodyUid, -1)
	if nobodyChownErr != nil {
		test.Fatalf("NobodyChownFailed: %v", nobodyChownErr)
	}
	daemonChownErr := os.Chown(daemonDir, daemonUid, -1)
	if daemonChownErr != nil {
		test.Fatalf("DaemonChownFailed: %v", daemonChownErr)
	}

	test.Run("WritesThroughChainWithUsernameAndUserIdTrusted", func(subtest *testing.T) {
		target := filepath.Join(daemonDir, "php-webserver.conf")
		fileContent := []byte("display_errors = On")
		err := clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				nobodyUsername,
			},
			TrustedDirOwnerUserIds: []tkValueObject.UnixUserId{daemonUserId},
		}, fileContent)
		if err != nil {
			subtest.Fatalf("UpsertFileFailed: %v", err)
		}

		writtenContent, readErr := os.ReadFile(target)
		if readErr != nil {
			subtest.Fatalf("ReadFileFailed: %v", readErr)
		}
		if string(writtenContent) != string(fileContent) {
			subtest.Errorf("ContentMismatch: %s", writtenContent)
		}
	})

	test.Run("RefusesChainWhenOneOwnerUnnamed", func(subtest *testing.T) {
		target := filepath.Join(daemonDir, "refused.conf")
		err := clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(subtest, target),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				nobodyUsername,
			},
		}, []byte("display_errors = Off"))
		if !errors.Is(err, ErrDirectoryOwnerInvalid) {
			subtest.Fatalf("ExpectedErrDirectoryOwnerInvalid, got: %v", err)
		}
		if !strings.Contains(err.Error(), "daemonOwner") {
			subtest.Errorf("ErrorDoesNotNameOffendingComponent: %v", err)
		}
		if clerk.FileExists(target) {
			subtest.Errorf("RefusedTargetWasCreated")
		}
	})
}

func TestListDir(test *testing.T) {
	clerk := FileClerk{}

	// Note: Setup/teardown are intentionally file-local — test independence
	// requires each file to own its preconditions, even if it duplicates code.
	listDirTestTreeFactory := func() string {
		test.Helper()

		tempDir := test.TempDir()
		subDir := filepath.Join(tempDir, "sub")
		elsewhereDir := filepath.Join(tempDir, "elsewhere")
		mkdirErr := os.MkdirAll(subDir, 0755)
		if mkdirErr != nil {
			test.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		mkdirErr = os.MkdirAll(elsewhereDir, 0755)
		if mkdirErr != nil {
			test.Fatalf("MkdirFailed: %v", mkdirErr)
		}

		writeErr := os.WriteFile(
			filepath.Join(tempDir, "a.txt"), []byte("hello"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(tempDir, "b.log"), []byte("x"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(subDir, "inner.txt"), []byte("world"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(elsewhereDir, "hidden.txt"), []byte("!"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}

		symlinkTargetDir := test.TempDir()
		writeErr = os.WriteFile(
			filepath.Join(symlinkTargetDir, "symlinkTargetCanary.txt"),
			[]byte("canary"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		symlinkErr := os.Symlink(
			symlinkTargetDir, filepath.Join(subDir, "linkToOutside"),
		)
		if symlinkErr != nil {
			test.Fatalf("SymlinkFailed: %v", symlinkErr)
		}

		return tempDir
	}
	tempDir := listDirTestTreeFactory()

	test.Run("ListsRootEntriesSortedByName", func(subtest *testing.T) {
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, tempDir),
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}

		expectedNames := []string{"a.txt", "b.log", "elsewhere", "sub"}
		if len(unixFiles) != len(expectedNames) {
			subtest.Fatalf("UnexpectedEntryCount: %d", len(unixFiles))
		}
		for entryIndex, expectedName := range expectedNames {
			actualName := unixFiles[entryIndex].Name.String()
			if actualName != expectedName {
				subtest.Errorf("UnexpectedEntryName: %s", actualName)
			}
		}
	})

	test.Run("ReadsEntryAttributes", func(subtest *testing.T) {
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, tempDir),
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}

		for _, unixFile := range unixFiles {
			switch unixFile.Name.String() {
			case "a.txt":
				expectedPath := filepath.Join(tempDir, "a.txt")
				if unixFile.Path.String() != expectedPath {
					subtest.Errorf("UnexpectedFilePath: %s", unixFile.Path.String())
				}
				if unixFile.MimeType.String() != "text/plain" {
					subtest.Errorf("UnexpectedMimeType: %s", unixFile.MimeType.String())
				}
				if unixFile.Extension == nil ||
					unixFile.Extension.String() != "txt" {
					subtest.Errorf("UnexpectedExtension: %v", unixFile.Extension)
				}
				if unixFile.Size.Uint64() != 5 {
					subtest.Errorf("UnexpectedSize: %d", unixFile.Size.Uint64())
				}
				if unixFile.IsSymlink {
					subtest.Errorf("RegularFileMarkedAsSymlink")
				}
			case "sub":
				if unixFile.MimeType.String() != "directory" {
					subtest.Errorf("UnexpectedMimeType: %s", unixFile.MimeType.String())
				}
				if unixFile.Extension != nil {
					subtest.Errorf("DirectoryHasExtension: %s", unixFile.Extension.String())
				}
				if unixFile.IsSymlink {
					subtest.Errorf("DirectoryMarkedAsSymlink")
				}
			}
		}
	})

	test.Run("RespectsMaxDepth", func(subtest *testing.T) {
		depthZeroFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, tempDir),
			MaxDepth: 0,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(depthZeroFiles) != 4 {
			subtest.Errorf("UnexpectedDepthZeroEntryCount: %d", len(depthZeroFiles))
		}

		depthOneFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, tempDir),
			MaxDepth: 1,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(depthOneFiles) != 7 {
			subtest.Errorf("UnexpectedDepthOneEntryCount: %d", len(depthOneFiles))
		}
	})

	test.Run("CapsEntries", func(subtest *testing.T) {
		maxFiles := uint64(2)
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, tempDir),
			MaxFiles: &maxFiles,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != 2 {
			subtest.Errorf("UnexpectedCappedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("CapsEntriesAcrossSubdirectories", func(subtest *testing.T) {
		nestedCapDirPath := subtest.TempDir()
		for _, subDirName := range []string{"first", "second"} {
			subDirPath := filepath.Join(nestedCapDirPath, subDirName)
			mkdirErr := os.MkdirAll(subDirPath, 0755)
			if mkdirErr != nil {
				subtest.Fatalf("MkdirFailed: %v", mkdirErr)
			}
			for _, fileName := range []string{"one.txt", "two.txt"} {
				writeErr := os.WriteFile(
					filepath.Join(subDirPath, fileName), []byte("x"), 0644,
				)
				if writeErr != nil {
					subtest.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}
		}

		maxFiles := uint64(2)
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, nestedCapDirPath),
			MaxFiles: &maxFiles,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != int(maxFiles) {
			subtest.Errorf("UnexpectedCappedEntryCount: %d", len(unixFiles))
		}
	})

	rootLinkPath := filepath.Join(tempDir, "rootLink")
	symlinkErr := os.Symlink(tempDir, rootLinkPath)
	if symlinkErr != nil {
		test.Fatalf("SymlinkFailed: %v", symlinkErr)
	}

	test.Run("RefusesSymlinkedRootByDefault", func(subtest *testing.T) {
		_, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, rootLinkPath),
		})
		if !errors.Is(err, ErrSymlinkedPathInvalid) {
			subtest.Fatalf("ExpectedErrSymlinkedPathInvalid, got: %v", err)
		}
	})

	test.Run("ResolvesSymlinkedRootWithPolicy", func(subtest *testing.T) {
		resolvePolicy := FileClerkSymlinkPolicyResolve
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:       absoluteFilePathForTest(subtest, rootLinkPath),
			SymlinkPolicy: &resolvePolicy,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != 5 {
			subtest.Errorf("UnexpectedResolvedEntryCount: %d", len(unixFiles))
		}
		for _, unixFile := range unixFiles {
			if unixFile.Name.String() != "a.txt" {
				continue
			}
			expectedPath := filepath.Join(tempDir, "a.txt")
			if unixFile.Path.String() != expectedPath {
				subtest.Errorf("PathStillGoesThroughSymlink: %s", unixFile.Path.String())
			}
		}
	})

	test.Run("FailsOnMissingDir", func(subtest *testing.T) {
		missingDirPath := filepath.Join(tempDir, "missing")
		_, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, missingDirPath),
		})
		if !errors.Is(err, ErrDirMissing) {
			subtest.Fatalf("ExpectedErrDirMissing, got: %v", err)
		}
	})

	test.Run("FailsOnEmptyDirPath", func(subtest *testing.T) {
		_, err := clerk.ListDir(FileListDirSettings{})
		if !errors.Is(err, ErrDirMissing) {
			subtest.Fatalf("ExpectedErrDirMissing, got: %v", err)
		}
	})

	test.Run("SkipsEntriesWithRejectedNames", func(subtest *testing.T) {
		rejectedNamePath := filepath.Join(tempDir, "~")
		writeErr := os.WriteFile(rejectedNamePath, []byte("x"), 0644)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}

		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, tempDir),
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		for _, unixFile := range unixFiles {
			if unixFile.Name.String() == "~" {
				subtest.Errorf("RejectedNameWasListed")
			}
		}
	})

	test.Run("ListsEmptyDir", func(subtest *testing.T) {
		emptyDirPath := filepath.Join(tempDir, "emptyDir")
		mkdirErr := os.MkdirAll(emptyDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}

		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, emptyDirPath),
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != 0 {
			subtest.Errorf("UnexpectedEntryCount: %d", len(unixFiles))
		}
	})

	capTreeDirPath := filepath.Join(tempDir, "capTree")
	capTreeMkdirErr := os.MkdirAll(capTreeDirPath, 0755)
	if capTreeMkdirErr != nil {
		test.Fatalf("MkdirFailed: %v", capTreeMkdirErr)
	}
	expectedUncappedEntryCount := int(FileClerkDefaultMaxFiles) + 2
	for fileIndex := range expectedUncappedEntryCount {
		capTreeFileName := "entry-" + strconv.Itoa(fileIndex) + ".txt"
		capTreeWriteErr := os.WriteFile(
			filepath.Join(capTreeDirPath, capTreeFileName), []byte("x"), 0644,
		)
		if capTreeWriteErr != nil {
			test.Fatalf("WriteFileFailed: %v", capTreeWriteErr)
		}
	}

	test.Run("DefaultsToMaxFilesWhenUnset", func(subtest *testing.T) {
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, capTreeDirPath),
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if uint64(len(unixFiles)) != FileClerkDefaultMaxFiles {
			subtest.Errorf("UnexpectedDefaultCappedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("ReadsWithoutCapWhenMaxFilesIsZero", func(subtest *testing.T) {
		zeroMaxFiles := uint64(0)
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, capTreeDirPath),
			MaxFiles: &zeroMaxFiles,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != expectedUncappedEntryCount {
			subtest.Errorf("UnexpectedUncappedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("RefusesWritableDirChainWhenRequired", func(subtest *testing.T) {
		writableDirPath := filepath.Join(tempDir, "writableListDir")
		mkdirErr := os.MkdirAll(writableDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		chmodErr := os.Chmod(writableDirPath, 0o777)
		if chmodErr != nil {
			subtest.Fatalf("ChmodFailed: %v", chmodErr)
		}

		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, writableDirPath),
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != 0 {
			subtest.Errorf("UnexpectedEntryCount: %d", len(unixFiles))
		}

		sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused
		_, err = clerk.ListDir(FileListDirSettings{
			DirPath:        absoluteFilePathForTest(subtest, writableDirPath),
			DirChainPolicy: &sharedWriteRefusedPolicy,
		})
		if !errors.Is(err, ErrDirectoryWritableByOthers) {
			subtest.Errorf("ExpectedErrDirectoryWritableByOthers, got: %v", err)
		}
	})

	test.Run("ReportsUnknownTrustedDirOwner", func(subtest *testing.T) {
		unknownUsername, usernameErr := tkValueObject.NewUnixUsername(
			"no-such-user-xyz-42",
		)
		if usernameErr != nil {
			subtest.Fatalf("UnknownUsernameInvalid: %v", usernameErr)
		}

		_, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, tempDir),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				unknownUsername,
			},
		})
		if err == nil {
			subtest.Fatalf("MissingErrorForUnknownTrustedDirOwner")
		}
		if !strings.Contains(err.Error(), "OwnerLookupFailed") {
			subtest.Errorf("ErrorMissingOwnerLookupFailed: %v", err)
		}
	})

	test.Run("AcceptsUnresolvableTrustedDirOwnerUserId", func(subtest *testing.T) {
		if os.Geteuid() == 0 {
			subtest.Skip("NonRootRequired")
		}

		unresolvableUserId, userIdErr := tkValueObject.NewUnixUserId(
			unresolvableUnixUserIdFinder(subtest),
		)
		if userIdErr != nil {
			subtest.Fatalf("UnresolvableUserIdInvalid: %v", userIdErr)
		}

		_, err := clerk.ListDir(FileListDirSettings{
			DirPath: absoluteFilePathForTest(subtest, tempDir),
			TrustedDirOwnerUserIds: []tkValueObject.UnixUserId{
				unresolvableUserId,
			},
		})
		if !errors.Is(err, ErrDirectoryOwnerInvalid) {
			subtest.Errorf("ExpectedErrDirectoryOwnerInvalid, got: %v", err)
		}
	})

	test.Run("SkipsDescentWhenSubDirOpenFails", func(subtest *testing.T) {
		if os.Geteuid() == 0 {
			subtest.Skip("NonRootRequired")
		}

		lockedParentDirPath := filepath.Join(tempDir, "lockedParent")
		lockedDirPath := filepath.Join(lockedParentDirPath, "locked")
		mkdirErr := os.MkdirAll(lockedDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		chmodErr := os.Chmod(lockedDirPath, 0o000)
		if chmodErr != nil {
			subtest.Fatalf("ChmodFailed: %v", chmodErr)
		}

		depthOne := uint16(1)
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, lockedParentDirPath),
			MaxDepth: depthOne,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != 1 {
			subtest.Fatalf("UnexpectedEntryCount: %d", len(unixFiles))
		}
		if unixFiles[0].Path.String() != lockedDirPath {
			subtest.Errorf("UnexpectedEntryPath: %s", unixFiles[0].Path.String())
		}
	})

	test.Run("SkipsEntryWhenDirEntryStatFails", func(subtest *testing.T) {
		if os.Geteuid() == 0 {
			subtest.Skip("NonRootRequired")
		}

		unsearchableParentDirPath := filepath.Join(tempDir, "unsearchableParent")
		unsearchableDirPath := filepath.Join(unsearchableParentDirPath, "unsearchable")
		mkdirErr := os.MkdirAll(unsearchableDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		entryPath := filepath.Join(unsearchableDirPath, "entry.txt")
		writeErr := os.WriteFile(entryPath, []byte("x"), 0644)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}
		chmodErr := os.Chmod(unsearchableDirPath, 0o444)
		if chmodErr != nil {
			subtest.Fatalf("ChmodFailed: %v", chmodErr)
		}
		defer func() {
			_ = os.Chmod(unsearchableDirPath, 0o755)
		}()

		depthOne := uint16(1)
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, unsearchableParentDirPath),
			MaxDepth: depthOne,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}
		if len(unixFiles) != 1 {
			subtest.Fatalf("UnexpectedEntryCount: %d", len(unixFiles))
		}
		if unixFiles[0].Path.String() != unsearchableDirPath {
			subtest.Errorf("UnexpectedEntryPath: %s", unixFiles[0].Path.String())
		}
	})

	test.Run("SortsDuplicateNamesByPath", func(subtest *testing.T) {
		dupRootDirPath := filepath.Join(tempDir, "dupRoot")
		dupDirBPath := filepath.Join(dupRootDirPath, "dupB")
		dupDirAPath := filepath.Join(dupRootDirPath, "dupA")
		mkdirErr := os.MkdirAll(dupDirBPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		mkdirErr = os.MkdirAll(dupDirAPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		writeErr := os.WriteFile(
			filepath.Join(dupDirBPath, "same.txt"), []byte("b"), 0644,
		)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(dupDirAPath, "same.txt"), []byte("a"), 0644,
		)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}

		depthOne := uint16(1)
		unixFiles, err := clerk.ListDir(FileListDirSettings{
			DirPath:  absoluteFilePathForTest(subtest, dupRootDirPath),
			MaxDepth: depthOne,
		})
		if err != nil {
			subtest.Fatalf("ListDirFailed: %v", err)
		}

		duplicatePaths := []string{}
		for _, unixFile := range unixFiles {
			if unixFile.Name.String() == "same.txt" {
				duplicatePaths = append(duplicatePaths, unixFile.Path.String())
			}
		}
		if len(duplicatePaths) != 2 {
			subtest.Fatalf("UnexpectedDuplicateCount: %d", len(duplicatePaths))
		}
		expectedFirstPath := filepath.Join(dupDirAPath, "same.txt")
		if duplicatePaths[0] != expectedFirstPath {
			subtest.Errorf("DuplicateNamesNotSortedByPath: %s", duplicatePaths[0])
		}
	})
}

func TestFind(test *testing.T) {
	clerk := FileClerk{}

	// Note: Setup/teardown are intentionally file-local — test independence
	// requires each file to own its preconditions, even if it duplicates code.
	findTestTreeFactory := func() string {
		test.Helper()

		tempDir := test.TempDir()
		subDir := filepath.Join(tempDir, "sub")
		elsewhereDir := filepath.Join(tempDir, "elsewhere")
		mkdirErr := os.MkdirAll(subDir, 0755)
		if mkdirErr != nil {
			test.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		mkdirErr = os.MkdirAll(elsewhereDir, 0755)
		if mkdirErr != nil {
			test.Fatalf("MkdirFailed: %v", mkdirErr)
		}

		writeErr := os.WriteFile(
			filepath.Join(tempDir, "a.txt"), []byte("hello"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(tempDir, "b.log"), []byte("x"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(subDir, "inner.txt"), []byte("world"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(elsewhereDir, "hidden.txt"), []byte("!"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}

		symlinkTargetDir := test.TempDir()
		writeErr = os.WriteFile(
			filepath.Join(symlinkTargetDir, "symlinkTargetCanary.txt"),
			[]byte("canary"), 0644,
		)
		if writeErr != nil {
			test.Fatalf("WriteFileFailed: %v", writeErr)
		}
		symlinkErr := os.Symlink(
			symlinkTargetDir, filepath.Join(subDir, "linkToOutside"),
		)
		if symlinkErr != nil {
			test.Fatalf("SymlinkFailed: %v", symlinkErr)
		}

		return tempDir
	}
	tempDir := findTestTreeFactory()

	test.Run("FiltersByNamePattern", func(subtest *testing.T) {
		namePattern := regexp.MustCompile(`\.txt$`)
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, tempDir),
			MaxDepth:            1,
			NamePattern:         namePattern,
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}

		expectedPaths := []string{
			filepath.Join(tempDir, "a.txt"),
			filepath.Join(tempDir, "elsewhere", "hidden.txt"),
			filepath.Join(tempDir, "sub", "inner.txt"),
		}
		if len(unixFiles) != len(expectedPaths) {
			subtest.Fatalf("UnexpectedMatchCount: %d", len(unixFiles))
		}
		for matchIndex, expectedPath := range expectedPaths {
			actualPath := unixFiles[matchIndex].Path.String()
			if actualPath != expectedPath {
				subtest.Errorf("UnexpectedMatchPath: %s", actualPath)
			}
		}
	})

	test.Run("MatchesEveryEntryWithoutPattern", func(subtest *testing.T) {
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, tempDir),
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 4 {
			subtest.Errorf("UnexpectedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("CapsMatchedEntries", func(subtest *testing.T) {
		flatDir := subtest.TempDir()
		for fileIndex := range 5 {
			fileName := "file" + strconv.Itoa(fileIndex) + ".txt"
			writeErr := os.WriteFile(
				filepath.Join(flatDir, fileName), []byte("x"), 0644,
			)
			if writeErr != nil {
				subtest.Fatalf("WriteFileFailed: %v", writeErr)
			}
		}

		maxFiles := uint64(2)
		namePattern := regexp.MustCompile(`\.txt$`)
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, flatDir),
			MaxFiles:            &maxFiles,
			NamePattern:         namePattern,
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 2 {
			subtest.Errorf("UnexpectedCappedMatchCount: %d", len(unixFiles))
		}
	})

	test.Run("NeverDescendsIntoSymlinkedDir", func(subtest *testing.T) {
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, tempDir),
			MaxDepth:            2,
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}

		symlinkReported := false
		for _, unixFile := range unixFiles {
			if unixFile.Name.String() == "symlinkTargetCanary.txt" {
				subtest.Errorf("SymlinkedDirWasDescended: %s", unixFile.Path.String())
			}
			if unixFile.Name.String() == "linkToOutside" {
				symlinkReported = true
				if !unixFile.IsSymlink {
					subtest.Errorf("SymlinkNotMarkedAsSymlink")
				}
			}
		}
		if !symlinkReported {
			subtest.Errorf("SymlinkNotReported")
		}
	})

	test.Run("FailsOnMissingRoot", func(subtest *testing.T) {
		missingRootPath := filepath.Join(tempDir, "missing")
		_, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, missingRootPath),
		})
		if !errors.Is(err, ErrDirMissing) {
			subtest.Fatalf("ExpectedErrDirMissing, got: %v", err)
		}
	})

	test.Run("WalksPlainStartingPath", func(subtest *testing.T) {
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPath: absoluteFilePathForTest(subtest, tempDir),
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 4 {
			subtest.Errorf("UnexpectedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("FailsOnMissingStartingPath", func(subtest *testing.T) {
		missingPath := filepath.Join(tempDir, "missingStartingPath")
		_, err := clerk.Find(FileFindSettings{
			StartingPath: absoluteFilePathForTest(subtest, missingPath),
		})
		if !errors.Is(err, ErrDirMissing) {
			subtest.Fatalf("ExpectedErrDirMissing, got: %v", err)
		}
	})

	test.Run("FailsWhenBothStartingPathsSet", func(subtest *testing.T) {
		_, err := clerk.Find(FileFindSettings{
			StartingPath:        absoluteFilePathForTest(subtest, tempDir),
			StartingPathPattern: absoluteGlobPathForTest(subtest, tempDir),
		})
		if !errors.Is(err, ErrStartingPathConflict) {
			subtest.Fatalf("ExpectedErrStartingPathConflict, got: %v", err)
		}
	})

	test.Run("FailsWhenNoStartingPathSet", func(subtest *testing.T) {
		_, err := clerk.Find(FileFindSettings{})
		if !errors.Is(err, ErrStartingPathMissing) {
			subtest.Fatalf("ExpectedErrStartingPathMissing, got: %v", err)
		}
	})

	test.Run("ReadsWithoutCapWhenMaxFilesIsZero", func(subtest *testing.T) {
		zeroMaxFiles := uint64(0)
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, tempDir),
			MaxFiles:            &zeroMaxFiles,
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 4 {
			subtest.Errorf("UnexpectedUncappedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("RefusesWritableDirChainWhenRequired", func(subtest *testing.T) {
		writableDirPath := filepath.Join(tempDir, "writableFindDir")
		mkdirErr := os.MkdirAll(writableDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		chmodErr := os.Chmod(writableDirPath, 0o777)
		if chmodErr != nil {
			subtest.Fatalf("ChmodFailed: %v", chmodErr)
		}

		sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused
		_, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, writableDirPath),
			DirChainPolicy:      &sharedWriteRefusedPolicy,
		})
		if !errors.Is(err, ErrDirectoryWritableByOthers) {
			subtest.Errorf("ExpectedErrDirectoryWritableByOthers, got: %v", err)
		}
	})

	test.Run("ReportsUnknownTrustedDirOwner", func(subtest *testing.T) {
		unknownUsername, usernameErr := tkValueObject.NewUnixUsername(
			"no-such-user-xyz-42",
		)
		if usernameErr != nil {
			subtest.Fatalf("UnknownUsernameInvalid: %v", usernameErr)
		}

		_, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, tempDir),
			TrustedDirOwnerUsernames: []tkValueObject.UnixUsername{
				unknownUsername,
			},
		})
		if err == nil {
			subtest.Fatalf("MissingErrorForUnknownTrustedDirOwner")
		}
		if !strings.Contains(err.Error(), "OwnerLookupFailed") {
			subtest.Errorf("ErrorMissingOwnerLookupFailed: %v", err)
		}
	})

	test.Run("ExpandsWildcardRootPattern", func(subtest *testing.T) {
		patternRootDirPath := subtest.TempDir()
		for _, accountName := range []string{"alpha", "beta"} {
			docsDirPath := filepath.Join(patternRootDirPath, accountName, "docs")
			mkdirErr := os.MkdirAll(docsDirPath, 0755)
			if mkdirErr != nil {
				subtest.Fatalf("MkdirFailed: %v", mkdirErr)
			}
			for _, fileName := range []string{"one.txt", "two.txt"} {
				writeErr := os.WriteFile(
					filepath.Join(docsDirPath, fileName), []byte("x"), 0644,
				)
				if writeErr != nil {
					subtest.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}
		}

		startingPathPattern := filepath.Join(patternRootDirPath, "*", "docs")
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, startingPathPattern),
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}

		expectedPaths := []string{
			filepath.Join(patternRootDirPath, "alpha", "docs", "one.txt"),
			filepath.Join(patternRootDirPath, "alpha", "docs", "two.txt"),
			filepath.Join(patternRootDirPath, "beta", "docs", "one.txt"),
			filepath.Join(patternRootDirPath, "beta", "docs", "two.txt"),
		}
		if len(unixFiles) != len(expectedPaths) {
			subtest.Fatalf("UnexpectedMatchCount: %d", len(unixFiles))
		}
		for matchIndex, expectedPath := range expectedPaths {
			actualPath := unixFiles[matchIndex].Path.String()
			if actualPath != expectedPath {
				subtest.Errorf("UnexpectedMatchPath: %s", actualPath)
			}
		}
	})

	test.Run("ReturnsEmptyWhenWildcardMatchesNothing", func(subtest *testing.T) {
		startingPathPattern := filepath.Join(tempDir, "*", "noSuchDir")
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, startingPathPattern),
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 0 {
			subtest.Errorf("UnexpectedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("CapsEntriesAcrossWildcardRoots", func(subtest *testing.T) {
		patternRootDirPath := subtest.TempDir()
		for _, accountName := range []string{"alpha", "beta"} {
			docsDirPath := filepath.Join(patternRootDirPath, accountName, "docs")
			mkdirErr := os.MkdirAll(docsDirPath, 0755)
			if mkdirErr != nil {
				subtest.Fatalf("MkdirFailed: %v", mkdirErr)
			}
			for _, fileName := range []string{"one.txt", "two.txt"} {
				writeErr := os.WriteFile(
					filepath.Join(docsDirPath, fileName), []byte("x"), 0644,
				)
				if writeErr != nil {
					subtest.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}
		}

		maxFiles := uint64(3)
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(
				subtest, filepath.Join(patternRootDirPath, "*", "docs"),
			),
			MaxFiles: &maxFiles,
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != int(maxFiles) {
			subtest.Errorf("UnexpectedCappedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("SkipsNonDirectoryMatches", func(subtest *testing.T) {
		patternRootDirPath := subtest.TempDir()
		accountDirPath := filepath.Join(patternRootDirPath, "alpha")
		mkdirErr := os.MkdirAll(accountDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		writeErr := os.WriteFile(
			filepath.Join(accountDirPath, "one.txt"), []byte("x"), 0644,
		)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}
		writeErr = os.WriteFile(
			filepath.Join(patternRootDirPath, "stray.txt"), []byte("x"), 0644,
		)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}

		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(
				subtest, filepath.Join(patternRootDirPath, "*"),
			),
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 1 {
			subtest.Errorf("UnexpectedEntryCount: %d", len(unixFiles))
		}
	})

	test.Run("SkipsUnreadableWildcardRoots", func(subtest *testing.T) {
		patternRootDirPath := subtest.TempDir()
		alphaDirPath := filepath.Join(patternRootDirPath, "alpha")
		mkdirErr := os.MkdirAll(alphaDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		writeErr := os.WriteFile(
			filepath.Join(alphaDirPath, "one.txt"), []byte("x"), 0644,
		)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}
		betaDirPath := filepath.Join(patternRootDirPath, "beta")
		mkdirErr = os.MkdirAll(betaDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		chmodErr := os.Chmod(betaDirPath, 0o777)
		if chmodErr != nil {
			subtest.Fatalf("ChmodFailed: %v", chmodErr)
		}

		sharedWriteRefusedPolicy := FileClerkDirChainPolicySharedWriteRefused
		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(
				subtest, filepath.Join(patternRootDirPath, "*"),
			),
			DirChainPolicy: &sharedWriteRefusedPolicy,
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 1 {
			subtest.Fatalf("UnexpectedEntryCount: %d", len(unixFiles))
		}
		expectedPath := filepath.Join(alphaDirPath, "one.txt")
		if unixFiles[0].Path.String() != expectedPath {
			subtest.Errorf("UnexpectedMatchPath: %s", unixFiles[0].Path.String())
		}
	})

	test.Run("SkipsUnstattableWildcardMatches", func(subtest *testing.T) {
		if os.Geteuid() == 0 {
			subtest.Skip("NonRootRequired")
		}
		patternRootDirPath := subtest.TempDir()
		goodDocsDirPath := filepath.Join(patternRootDirPath, "good", "docs")
		mkdirErr := os.MkdirAll(goodDocsDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		writeErr := os.WriteFile(
			filepath.Join(goodDocsDirPath, "one.txt"), []byte("x"), 0644,
		)
		if writeErr != nil {
			subtest.Fatalf("WriteFileFailed: %v", writeErr)
		}

		unsearchableDirPath := filepath.Join(patternRootDirPath, "unsearchable")
		innerDirPath := filepath.Join(unsearchableDirPath, "inner")
		mkdirErr = os.MkdirAll(innerDirPath, 0755)
		if mkdirErr != nil {
			subtest.Fatalf("MkdirFailed: %v", mkdirErr)
		}
		chmodErr := os.Chmod(unsearchableDirPath, 0o444)
		if chmodErr != nil {
			subtest.Fatalf("ChmodFailed: %v", chmodErr)
		}
		subtest.Cleanup(func() {
			_ = os.Chmod(unsearchableDirPath, 0o755)
		})

		unixFiles, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(
				subtest, filepath.Join(patternRootDirPath, "*", "*"),
			),
		})
		if err != nil {
			subtest.Fatalf("FindFailed: %v", err)
		}
		if len(unixFiles) != 1 {
			subtest.Fatalf("UnexpectedEntryCount: %d", len(unixFiles))
		}
		expectedPath := filepath.Join(goodDocsDirPath, "one.txt")
		if unixFiles[0].Path.String() != expectedPath {
			subtest.Errorf("UnexpectedMatchPath: %s", unixFiles[0].Path.String())
		}
	})

	test.Run("FailsOnMalformedWildcardPattern", func(subtest *testing.T) {
		_, err := clerk.Find(FileFindSettings{
			StartingPathPattern: absoluteGlobPathForTest(subtest, filepath.Join(tempDir, "[")),
		})
		if !errors.Is(err, ErrGlobPatternInvalid) {
			subtest.Fatalf("ExpectedErrGlobPatternInvalid, got: %v", err)
		}
	})
}

func TestCachedOwnerNameResolvers(test *testing.T) {
	clerk := FileClerk{}

	test.Run("ReturnsCachedUsernameWithoutLookup", func(subtest *testing.T) {
		missingUserId, userIdErr := tkValueObject.NewUnixUserId(
			unresolvableUnixUserIdFinder(subtest),
		)
		if userIdErr != nil {
			subtest.Fatalf("UserIdInvalid: %v", userIdErr)
		}
		cachedUsername, usernameErr := tkValueObject.NewUnixUsername("ghost")
		if usernameErr != nil {
			subtest.Fatalf("UsernameInvalid: %v", usernameErr)
		}
		ownerNames := &ownerNameCache{
			usernames: map[tkValueObject.UnixUserId]tkValueObject.UnixUsername{
				missingUserId: cachedUsername,
			},
			groupNames: map[tkValueObject.UnixGroupId]tkValueObject.UnixGroupName{},
		}

		username, err := clerk.cachedUsernameResolver(ownerNames, missingUserId)
		if err != nil {
			subtest.Fatalf("CachedUsernameLookupFailed: %v", err)
		}
		if username != cachedUsername {
			subtest.Errorf("UnexpectedCachedUsername: %s", username)
		}
	})

	test.Run("StoresLookupResultForLaterHits", func(subtest *testing.T) {
		processUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid())
		if userIdErr != nil {
			subtest.Fatalf("UserIdInvalid: %v", userIdErr)
		}
		ownerNames := &ownerNameCache{
			usernames:  map[tkValueObject.UnixUserId]tkValueObject.UnixUsername{},
			groupNames: map[tkValueObject.UnixGroupId]tkValueObject.UnixGroupName{},
		}

		_, err := clerk.cachedUsernameResolver(ownerNames, processUserId)
		if err != nil {
			subtest.Fatalf("UsernameLookupFailed: %v", err)
		}
		if _, isCached := ownerNames.usernames[processUserId]; !isCached {
			subtest.Errorf("UsernameNotCached")
		}
	})

	test.Run("CachesLookupErrorForLaterHits", func(subtest *testing.T) {
		missingUserId, userIdErr := tkValueObject.NewUnixUserId(
			unresolvableUnixUserIdFinder(subtest),
		)
		if userIdErr != nil {
			subtest.Fatalf("UserIdInvalid: %v", userIdErr)
		}
		ownerNames := &ownerNameCache{
			usernames:       map[tkValueObject.UnixUserId]tkValueObject.UnixUsername{},
			groupNames:      map[tkValueObject.UnixGroupId]tkValueObject.UnixGroupName{},
			usernameErrors:  map[tkValueObject.UnixUserId]error{},
			groupNameErrors: map[tkValueObject.UnixGroupId]error{},
		}

		_, firstErr := clerk.cachedUsernameResolver(ownerNames, missingUserId)
		if firstErr == nil {
			subtest.Fatalf("ExpectedUsernameLookupError")
		}
		if _, isCached := ownerNames.usernameErrors[missingUserId]; !isCached {
			subtest.Errorf("UsernameLookupErrorNotCached")
		}

		_, secondErr := clerk.cachedUsernameResolver(ownerNames, missingUserId)
		if secondErr != firstErr {
			subtest.Errorf("CachedUsernameLookupErrorNotReturned: %v", secondErr)
		}
	})
}

func TestUnixFilePermissionsFormatter(test *testing.T) {
	clerk := FileClerk{}

	testCases := []struct {
		name                string
		fileMode            os.FileMode
		expectedPermissions string
	}{
		{"ZeroMode", 0, "000"},
		{"PlainMode", 0o644, "644"},
		{"SetuidMode", os.ModeSetuid | 0o755, "4755"},
		{"SetgidMode", os.ModeSetgid | 0o775, "2775"},
		{"StickyMode", os.ModeSticky | 0o777, "1777"},
		{"AllSpecialBits", os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o770, "7770"},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(subtest *testing.T) {
			permissions, permissionsErr := clerk.unixFilePermissionsFormatter(
				testCase.fileMode,
			)
			if permissionsErr != nil {
				subtest.Fatalf("PermissionsFormatFailed: %v", permissionsErr)
			}
			if permissions.String() != testCase.expectedPermissions {
				subtest.Errorf(
					"UnexpectedPermissions: got %s, expected %s",
					permissions.String(), testCase.expectedPermissions,
				)
			}
		})
	}

	test.Run("RoundTripsThroughToFileMode", func(subtest *testing.T) {
		roundTripModes := []os.FileMode{
			0,
			0o644,
			0o755,
			os.ModeSetuid | 0o755,
			os.ModeSetgid | 0o775,
			os.ModeSticky | 0o777,
			os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o770,
		}

		for _, fileMode := range roundTripModes {
			permissions, permissionsErr := clerk.unixFilePermissionsFormatter(
				fileMode,
			)
			if permissionsErr != nil {
				subtest.Fatalf("PermissionsFormatFailed: %v", permissionsErr)
			}
			roundTrippedMode, modeErr := permissions.ToFileMode()
			if modeErr != nil {
				subtest.Fatalf("ToFileModeFailed: %v", modeErr)
			}
			if roundTrippedMode != fileMode {
				subtest.Errorf(
					"RoundTripMismatch: got %v, expected %v",
					roundTrippedMode, fileMode,
				)
			}
		}
	})
}
