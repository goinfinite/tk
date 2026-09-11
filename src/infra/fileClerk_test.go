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

func TestFileExists(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("ExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "existing.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.FileExists(testFile) {
			t.Errorf("FileExistsShouldReturnTrue: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		if clerk.FileExists(nonExistentFile) {
			t.Errorf("FileExistsShouldReturnFalse: %s", nonExistentFile)
		}
	})

	t.Run("ExistingDirectory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		if !clerk.FileExists(testDir) {
			t.Errorf("FileExistsShouldReturnTrueForDir: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})
}

func TestIsFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("RegularFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			t.Errorf("IsFileShouldReturnTrue: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("Directory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		if clerk.IsFile(testDir) {
			t.Errorf("IsFileShouldReturnFalseForDir: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")
		if clerk.IsFile(nonExistentPath) {
			t.Errorf("IsFileShouldReturnFalseForNonExistent: %s", nonExistentPath)
		}
	})
}

func TestIsDir(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("RegularDirectory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		if !clerk.IsDir(testDir) {
			t.Errorf("IsDirShouldReturnTrue: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})

	t.Run("RegularFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		if clerk.IsDir(testFile) {
			t.Errorf("IsDirShouldReturnFalseForFile: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")
		if clerk.IsDir(nonExistentPath) {
			t.Errorf("IsDirShouldReturnFalseForNonExistent: %s", nonExistentPath)
		}
	})
}

func TestTouchFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CreateNewFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "newfile.txt")
		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Errorf("TouchFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			t.Errorf("FileShouldExistAfterCreation: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("TouchFileInNonExistentDir", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent", "subdir")
		testFile := filepath.Join(nonExistentDir, "file.txt")
		err := clerk.TouchFile(testFile)
		if err == nil {
			t.Errorf("MissingExpectedError: TouchFileInNonExistentDir")
		}
	})

	t.Run("TouchExistingFileRefreshesTimestamps", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "existing_touch.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		pastTimestamp := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		err = os.Chtimes(testFile, pastTimestamp, pastTimestamp)
		if err != nil {
			t.Fatalf("ChtimesFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			t.Errorf("TouchFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if !fileInfo.ModTime().After(pastTimestamp) {
			t.Errorf("TimestampsNotRefreshed: mtime %v", fileInfo.ModTime())
		}
	})

	t.Run("TouchExistingFileKeepsContent", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "kept_content.txt")

		err := clerk.WriteNewFile(testFile, "do not truncate me", 0644)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			t.Errorf("TouchFileFailed: %v", err)
		}

		content, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			t.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "do not truncate me" {
			t.Errorf("ContentWasTruncated: '%s'", content)
		}
	})

	t.Run("TouchDanglingSymlinkFails", func(t *testing.T) {
		targetPath := filepath.Join(tempDir, "dangling_touch_target")
		symlinkPath := filepath.Join(tempDir, "dangling_touch.txt")

		err := os.Symlink(targetPath, symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.TouchFile(symlinkPath)
		if err == nil || err.Error() != "TargetIsSymlink" {
			t.Errorf("MissingExpectedError: TargetIsSymlink, got %v", err)
		}
		if clerk.FileExists(targetPath) {
			t.Errorf("TouchFollowedDanglingSymlink")
		}
	})
}

func TestWriteNewFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("WritesContentWithExactPermissions", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "written.txt")

		err := clerk.WriteNewFile(testFile, "hello clerk", 0666)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0666 {
			t.Errorf(
				"PermissionsMismatch: umask must not shrink granted mode, got %04o",
				fileInfo.Mode().Perm(),
			)
		}

		content, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			t.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "hello clerk" {
			t.Errorf("ContentMismatch: '%s' vs '%s'", content, "hello clerk")
		}
	})

	t.Run("EmptyContentWritesEmptyFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "empty.txt")

		err := clerk.WriteNewFile(testFile, "", 0644)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Size() != 0 {
			t.Errorf("FileShouldBeEmpty: size %d", fileInfo.Size())
		}
	})

	t.Run("ExistingPathFails", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "occupied.txt")

		err := clerk.WriteNewFile(testFile, "first", 0644)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = clerk.WriteNewFile(testFile, "second", 0644)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}

		content, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			t.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "first" {
			t.Errorf("ExistingContentWasOverwritten: '%s'", content)
		}
	})

	t.Run("DanglingSymlinkPathFails", func(t *testing.T) {
		symlinkPath := filepath.Join(tempDir, "dangling_link.txt")

		err := os.Symlink(filepath.Join(tempDir, "nowhere"), symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.WriteNewFile(symlinkPath, "payload", 0644)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if clerk.FileExists(filepath.Join(tempDir, "nowhere")) {
			t.Errorf("WriteFollowedDanglingSymlink")
		}
	})
}

func TestCopyFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CopyExistingFile", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		testContent := "test content for copy"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, targetFile)
		if err != nil {
			t.Errorf("CopyFileFailed: %v", err)
		}

		if !clerk.IsFile(targetFile) {
			t.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		targetContent, err := clerk.ReadFileContent(targetFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if targetContent != testContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", targetContent, testContent)
		}

		err = clerk.DeleteFile(sourceFile)
		if err != nil {
			t.Errorf("DeleteSourceFileFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	t.Run("CopyNonExistentFile", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.CopyFile(sourceFile, targetFile)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})

	t.Run("CopyToExistingTarget", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "existing_target.txt")

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, targetFile)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}

		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}

		err = clerk.DeleteFile(sourceFile)
		if err != nil {
			t.Errorf("DeleteSourceFileFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	t.Run("CopyPreservesSourceModeIncludingExecBit", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "executable.sh")
		targetFile := filepath.Join(tempDir, "executable_copy.sh")

		err := clerk.WriteNewFile(sourceFile, "#!/bin/sh\necho hi\n", 0755)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, targetFile)
		if err != nil {
			t.Fatalf("CopyFileFailed: %v", err)
		}

		targetInfo, statErr := os.Stat(targetFile)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if targetInfo.Mode().Perm() != 0755 {
			t.Errorf(
				"ModeMismatch: exec bit must survive copy, got %04o",
				targetInfo.Mode().Perm(),
			)
		}
	})

	t.Run("CopyToDanglingSymlinkFails", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "copy_source.txt")
		symlinkPath := filepath.Join(tempDir, "dangling_target.txt")

		err := clerk.WriteNewFile(sourceFile, "payload", 0644)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		err = os.Symlink(filepath.Join(tempDir, "nowhere_copy"), symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.CopyFile(sourceFile, symlinkPath)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}
		if clerk.FileExists(filepath.Join(tempDir, "nowhere_copy")) {
			t.Errorf("CopyFollowedDanglingSymlink")
		}
	})
}

func TestMoveFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("MoveExistingFile", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		testContent := "test content for move"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.MoveFile(sourceFile, targetFile)
		if err != nil {
			t.Errorf("MoveFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			t.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			t.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		targetContent, err := clerk.ReadFileContent(targetFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if targetContent != testContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", targetContent, testContent)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	t.Run("MoveNonExistentFile", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.MoveFile(sourceFile, targetFile)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})

	t.Run("MoveToExistingTargetFails", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "move_source.txt")
		targetFile := filepath.Join(tempDir, "move_existing_target.txt")

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.MoveFile(sourceFile, targetFile)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if err != nil && err.Error() != "TargetFileAlreadyExists" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetFileAlreadyExists", err.Error())
		}

		if !clerk.IsFile(sourceFile) || !clerk.IsFile(targetFile) {
			t.Errorf("FailedMoveMustLeaveBothFilesIntact")
		}
	})

	t.Run("MoveToDanglingSymlinkFails", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "move_dangling_source.txt")
		symlinkPath := filepath.Join(tempDir, "move_dangling_target.txt")

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = os.Symlink(filepath.Join(tempDir, "move_dangling_nowhere"), symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.MoveFile(sourceFile, symlinkPath)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetFileAlreadyExists")
		}
		if !clerk.IsFile(sourceFile) {
			t.Errorf("SourceFileMustSurviveFailedMove: %s", sourceFile)
		}
	})

	t.Run("MoveFileOntoItselfSucceeds", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "move_onto_itself.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.MoveFile(testFile, testFile)
		if err != nil {
			t.Errorf("MoveFileOntoItselfFailed: %v", err)
		}
		if !clerk.IsFile(testFile) {
			t.Errorf("FileShouldStillExist: %s", testFile)
		}
	})

	t.Run("MoveFileCrossDeviceFallsBackToCopyAndDelete", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "xdev_source.txt")

		err := clerk.WriteNewFile(sourceFile, "cross device payload", 0644)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		otherDeviceBase := directoryOnDifferentDevice(t, tempDir)
		otherDeviceDir, mkErr := os.MkdirTemp(otherDeviceBase, "tk-xdev-")
		if mkErr != nil {
			t.Fatalf("MkdirTempFailed: %v", mkErr)
		}
		defer func() { _ = os.RemoveAll(otherDeviceDir) }()

		targetFile := filepath.Join(otherDeviceDir, "xdev_target.txt")

		err = clerk.MoveFile(sourceFile, targetFile)
		if err != nil {
			t.Fatalf("MoveFileCrossDeviceFailed: %v", err)
		}
		if clerk.FileExists(sourceFile) {
			t.Errorf("SourceShouldBeRemovedAfterCrossDeviceMove: %s", sourceFile)
		}

		content, readErr := clerk.ReadFileContent(targetFile, nil)
		if readErr != nil {
			t.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if content != "cross device payload" {
			t.Errorf("ContentMismatch: '%s'", content)
		}
	})
}

// directoryOnDifferentDevice finds a world-writable directory living on
// another device than referenceDir, so MoveFile must take its EXDEV path.
func directoryOnDifferentDevice(t *testing.T, referenceDir string) string {
	t.Helper()

	referenceInfo, statErr := os.Stat(referenceDir)
	if statErr != nil {
		t.Fatalf("StatFailed: %v", statErr)
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

	t.Skip("NoDirectoryOnDifferentDevice")
	return ""
}

func TestDeleteFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("DeleteExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "todelete.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			t.Fatalf("FileShouldExistBeforeDeletion: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}

		if clerk.IsFile(testFile) {
			t.Errorf("FileShouldNotExistAfterDeletion: %s", testFile)
		}
	})

	t.Run("DeleteNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		err := clerk.DeleteFile(nonExistentFile)
		if err != nil {
			t.Errorf("DeleteNonExistentFileFailed: %v", err)
		}
	})

	t.Run("DeleteDirectoryFails", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "undeletable_dir")

		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.DeleteFile(testDir)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetIsDirectory")
		}
		if err != nil && err.Error() != "TargetIsDirectory" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetIsDirectory", err.Error())
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})

	t.Run("DeleteSymlinkRemovesLinkOnly", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "symlink_target_delete.txt")
		symlinkPath := filepath.Join(tempDir, "symlink_delete.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.Symlink(targetFile, symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(symlinkPath)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}

		if clerk.IsSymlink(symlinkPath) {
			t.Errorf("SymlinkShouldBeGone: %s", symlinkPath)
		}
		if !clerk.IsFile(targetFile) {
			t.Errorf("SymlinkTargetMustSurvive: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestReadFileContent(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("ReadExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "content.txt")
		expectedContent := "Hello, World!\nThis is test content."

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(expectedContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != expectedContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, expectedContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("ReadNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := clerk.ReadFileContent(nonExistentFile, nil)
		if err == nil {
			t.Errorf("MissingExpectedError: FileNotFound")
		}

		if err != nil && err.Error() != "FileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "FileNotFound", err.Error())
		}
	})

	t.Run("ReadExceedingMaxContentSizeFails", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "limited.txt")
		fullContent := "This is a longer content that should be truncated when reading with size limit."
		maxSize := int64(20)

		err := os.WriteFile(testFile, []byte(fullContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		_, err = clerk.ReadFileContent(testFile, &maxSize)
		if err == nil {
			t.Fatalf("MissingExpectedError: FileTooLarge")
		}
		if err.Error() != "FileTooLarge" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "FileTooLarge", err.Error())
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("ReadExactlyAtMaxContentSize", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "exact_limit.txt")
		fullContent := "exactly-twenty-chars"
		maxSize := int64(len(fullContent))

		err := os.WriteFile(testFile, []byte(fullContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, &maxSize)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}
		if actualContent != fullContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, fullContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("ReadWithNegativeMaxSizeFails", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "negative_limit.txt")
		negativeSize := int64(-1)

		err := os.WriteFile(testFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		_, err = clerk.ReadFileContent(testFile, &negativeSize)
		if err == nil || err.Error() != "FileTooLarge" {
			t.Errorf("WrongErrorForNegativeCap: %v", err)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("ReadWithDefaultMaxSize", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "default_size.txt")
		testContent := "Content with default size limit"

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != testContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, testContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("ReadSymlinkToFile", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "symlink_target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")
		expectedContent := "content behind the symlink"

		err := clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = os.WriteFile(targetFile, []byte(expectedContent), 0644)
		if err != nil {
			t.Fatalf("UpdateTargetFileContentFailed: %v", err)
		}

		err = os.Symlink(targetFile, symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(symlinkPath, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != expectedContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, expectedContent)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			t.Errorf("RemoveSymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	t.Run("ReadDanglingSymlink", func(t *testing.T) {
		deletedTarget := filepath.Join(tempDir, "deleted_target.txt")
		symlinkPath := filepath.Join(tempDir, "dangling.txt")

		err := clerk.TouchFile(deletedTarget)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.Symlink(deletedTarget, symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(deletedTarget)
		if err != nil {
			t.Fatalf("DeleteTargetFileFailed: %v", err)
		}

		_, err = clerk.ReadFileContent(symlinkPath, nil)
		if err == nil {
			t.Errorf("MissingExpectedError: FileNotFound")
		}

		if err != nil && err.Error() != "FileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "FileNotFound", err.Error())
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			t.Errorf("RemoveSymlinkFailed: %v", err)
		}
	})
}

func TestAppendFileContent(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CreatesMissingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "created.txt")
		content := "Created by append"

		err := clerk.AppendFileContent(
			absoluteFilePathForTest(t, testFile), content,
		)
		if err != nil {
			t.Fatalf("AppendFileContentFailed: %v", err)
		}

		actualContent, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			t.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		if actualContent != content {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, content)
		}
	})

	t.Run("AppendsToExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "append.txt")
		initialContent := "Initial"
		appendContent := " Appended"

		err := os.WriteFile(testFile, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.AppendFileContent(
			absoluteFilePathForTest(t, testFile), appendContent,
		)
		if err != nil {
			t.Fatalf("AppendFileContentFailed: %v", err)
		}

		actualContent, readErr := clerk.ReadFileContent(testFile, nil)
		if readErr != nil {
			t.Fatalf("ReadFileContentFailed: %v", readErr)
		}
		expectedContent := initialContent + appendContent
		if actualContent != expectedContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, expectedContent)
		}
	})
}

func TestTruncateFileContent(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("TruncateContentFromFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "clear.txt")
		initialContent := "Content to be cleared"

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.TruncateFileContent(absoluteFilePathForTest(t, testFile))
		if err != nil {
			t.Errorf("TruncateFileContentFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != "" {
			t.Errorf("ContentShouldBeEmpty: '%s'", actualContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})
}

func TestCreateDir(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CreateNewDirectory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "newdir")

		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Errorf("UnexpectedError: '%s'", err.Error())
		}

		if !clerk.IsDir(testDir) {
			t.Errorf("DirectoryShouldExist: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})

	t.Run("CreateNestedDirectory", func(t *testing.T) {
		nestedDir := filepath.Join(tempDir, "parent", "child", "grandchild")

		err := clerk.CreateDir(nestedDir)
		if err != nil {
			t.Errorf("CreateDirFailed: %v", err)
		}

		if !clerk.IsDir(nestedDir) {
			t.Errorf("NestedDirectoryShouldExist: %s", nestedDir)
		}

		parentDir := filepath.Join(tempDir, "parent")
		err = clerk.DeleteDir(parentDir)
		if err != nil {
			t.Errorf("DeleteParentDirFailed: %v", err)
		}
	})

	t.Run("CreateExistingDirectory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "existing")

		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.CreateDir(testDir)
		if err != nil {
			t.Errorf("CreateDirFailed: %v", err)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})
}

func TestCopyDir(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CopyDirectoryWithFiles", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "sourcedir")
		targetDir := filepath.Join(tempDir, "targetdir")
		testFile := filepath.Join(sourceDir, "testfile.txt")
		testContent := "test content"

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			t.Fatalf("CreateSourceDirFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("CreateTestFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.CopyDir(sourceDir, targetDir)
		if err != nil {
			t.Errorf("CopyDirFailed: %v", err)
		}

		if !clerk.IsDir(targetDir) {
			t.Errorf("TargetDirectoryShouldExist: %s", targetDir)
		}

		copiedFile := filepath.Join(targetDir, "testfile.txt")
		if !clerk.IsFile(copiedFile) {
			t.Errorf("CopiedFileShouldExist: %s", copiedFile)
		}

		copiedContent, err := clerk.ReadFileContent(copiedFile, nil)
		if err != nil {
			t.Errorf("ReadCopiedFileContentFailed: %v", err)
		}

		if copiedContent != testContent {
			t.Errorf("CopiedContentMismatch: '%s' vs '%s'", copiedContent, testContent)
		}

		err = clerk.DeleteDir(sourceDir)
		if err != nil {
			t.Errorf("DeleteSourceDirFailed: %v", err)
		}

		err = clerk.DeleteDir(targetDir)
		if err != nil {
			t.Errorf("DeleteTargetDirFailed: %v", err)
		}
	})

	t.Run("CopyNonExistentDirectory", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "nonexistent")
		targetDir := filepath.Join(tempDir, "target")

		err := clerk.CopyDir(sourceDir, targetDir)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceDirNotFound")
		}

		if err != nil && err.Error() != "SourceDirNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceDirNotFound", err.Error())
		}
	})
}

func TestMoveDir(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("MoveDirectoryWithFiles", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "movesource")
		targetDir := filepath.Join(tempDir, "movetarget")
		testFile := filepath.Join(sourceDir, "movefile.txt")
		testContent := "content to move"

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			t.Fatalf("CreateSourceDirFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("CreateTestFileFailed: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.MoveDir(sourceDir, targetDir)
		if err != nil {
			t.Errorf("MoveDirFailed: %v", err)
		}

		if clerk.IsDir(sourceDir) {
			t.Errorf("SourceDirectoryShouldNotExist: %s", sourceDir)
		}

		if !clerk.IsDir(targetDir) {
			t.Errorf("TargetDirectoryShouldExist: %s", targetDir)
		}

		movedFile := filepath.Join(targetDir, "movefile.txt")
		if !clerk.IsFile(movedFile) {
			t.Errorf("MovedFileShouldExist: %s", movedFile)
		}

		movedContent, err := clerk.ReadFileContent(movedFile, nil)
		if err != nil {
			t.Errorf("ReadMovedFileContentFailed: %v", err)
		}

		if movedContent != testContent {
			t.Errorf("MovedContentMismatch: '%s' vs '%s'", movedContent, testContent)
		}

		err = clerk.DeleteDir(targetDir)
		if err != nil {
			t.Errorf("DeleteTargetDirFailed: %v", err)
		}
	})

	t.Run("MoveNonExistentDirectory", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "nonexistent")
		targetDir := filepath.Join(tempDir, "target")

		err := clerk.MoveDir(sourceDir, targetDir)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceDirNotFound")
		}

		if err != nil && err.Error() != "SourceDirNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceDirNotFound", err.Error())
		}
	})
}

func TestDeleteDir(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("DeleteExistingDirectory", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "deleteme")
		testFile := filepath.Join(testDir, "file.txt")

		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		if !clerk.IsDir(testDir) {
			t.Fatalf("DirectoryShouldExistBeforeDeletion: %s", testDir)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}

		if clerk.IsDir(testDir) {
			t.Errorf("DirectoryShouldNotExistAfterDeletion: %s", testDir)
		}
	})

	t.Run("DeleteNonExistentDirectory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")

		err := clerk.DeleteDir(nonExistentDir)
		if err != nil {
			t.Errorf("DeleteNonExistentDirFailed: %v", err)
		}
	})

	t.Run("DeleteFileThroughDeleteDirFails", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "not_a_dir.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.DeleteDir(testFile)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetNotDirectory")
		}
		if err != nil && err.Error() != "TargetNotDirectory" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetNotDirectory", err.Error())
		}

		if !clerk.IsFile(testFile) {
			t.Errorf("FileMustSurviveRejectedDelete: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})
}

func TestIsSymlink(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("RegularFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		if clerk.IsSymlink(testFile) {
			t.Errorf("IsSymlinkShouldReturnFalseForRegularFile: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")

		if clerk.IsSymlink(nonExistentPath) {
			t.Errorf("IsSymlinkShouldReturnFalseForNonExistent: %s", nonExistentPath)
		}
	})
}

func TestCreateSymlink(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CreateSymlinkToFile", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CreateSymlink(targetFile, symlinkPath, false)
		if err != nil {
			t.Errorf("CreateSymlinkFailed: %v", err)
		}

		if !clerk.IsSymlink(symlinkPath) {
			t.Errorf("SymlinkShouldExist: %s", symlinkPath)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			t.Errorf("RemoveSymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	t.Run("CreateSymlinkToNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.CreateSymlink(nonExistentFile, symlinkPath, false)
		if err == nil {
			t.Errorf("MissingExpectedError: SourcePathNotFound")
		}

		if err != nil && err.Error() != "SourcePathNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourcePathNotFound", err.Error())
		}
	})
}

func TestIsSymlinkTo(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("ValidSymlinkTarget", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CreateSymlink(targetFile, symlinkPath, false)
		if err != nil {
			t.Fatalf("CreateSymlinkFailed: %v", err)
		}

		if !clerk.IsSymlinkTo(symlinkPath, targetFile) {
			t.Errorf("IsSymlinkToShouldReturnTrue: %s -> %s", symlinkPath, targetFile)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			t.Errorf("RemoveSymlinkFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})

	t.Run("RegularFileNotSymlink", func(t *testing.T) {
		regularFile := filepath.Join(tempDir, "regular.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.TouchFile(regularFile)
		if err != nil {
			t.Fatalf("CreateRegularFileFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		if clerk.IsSymlinkTo(regularFile, targetFile) {
			t.Errorf("IsSymlinkToShouldReturnFalseForRegularFile: %s", regularFile)
		}

		err = clerk.DeleteFile(regularFile)
		if err != nil {
			t.Errorf("DeleteRegularFileFailed: %v", err)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestRemoveSymlink(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("RemoveExistingSymlink", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		symlinkPath := filepath.Join(tempDir, "symlink.txt")

		err := clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = clerk.CreateSymlink(targetFile, symlinkPath, false)
		if err != nil {
			t.Fatalf("CreateSymlinkFailed: %v", err)
		}

		err = clerk.RemoveSymlink(symlinkPath)
		if err != nil {
			t.Errorf("RemoveSymlinkFailed: %v", err)
		}

		if clerk.IsSymlink(symlinkPath) {
			t.Errorf("SymlinkShouldNotExistAfterRemoval: %s", symlinkPath)
		}

		if !clerk.IsFile(targetFile) {
			t.Errorf("TargetFileShouldStillExist: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestUpdateFilePermissions(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("UpdateFilePermissionsWithCustomValue", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "permissions.txt")
		customPermissions := os.FileMode(0600)

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(testFile, &customPermissions)
		if err != nil {
			t.Errorf("UpdateFilePermissionsFailed: %v", err)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("UpdateFilePermissionsWithDefaultValue", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "default_permissions.txt")

		err := clerk.TouchFile(testFile)
		if err != nil {
			t.Fatalf("TouchFileFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(testFile, nil)
		if err != nil {
			t.Errorf("UpdateFilePermissionsFailed: %v", err)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("UpdateDirectoryPermissions", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "permissions_dir")

		err := clerk.CreateDir(testDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(testDir, nil)
		if err != nil {
			t.Errorf("UpdateFilePermissionsFailed: %v", err)
		}

		err = clerk.DeleteDir(testDir)
		if err != nil {
			t.Errorf("DeleteDirFailed: %v", err)
		}
	})

	t.Run("UpdateSymlinkPermissionsFails", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "chmod_target.txt")
		symlinkPath := filepath.Join(tempDir, "chmod_link.txt")
		restrictivePermissions := os.FileMode(0600)

		err := os.WriteFile(targetFile, []byte("target"), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = os.Symlink(targetFile, symlinkPath)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.UpdateFilePermissions(symlinkPath, &restrictivePermissions)
		if err == nil {
			t.Errorf("MissingExpectedError: TargetIsSymlink")
		}
		if err != nil && err.Error() != "TargetIsSymlink" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "TargetIsSymlink", err.Error())
		}

		targetInfo, statErr := os.Stat(targetFile)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if targetInfo.Mode().Perm() == 0600 {
			t.Errorf("ChmodLandedOnSymlinkTarget")
		}
	})
}

func TestCompressFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CompressFileWithFormats", func(t *testing.T) {
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
				t.Fatalf("TouchFileFailed: %v", err)
			}

			err = os.WriteFile(testFile, []byte(testContent), 0644)
			if err != nil {
				t.Fatalf("WriteFileFailed: %v", err)
			}

			compressedFilePath, err := clerk.CompressFile(testFile, testCase.format, nil)
			if err != nil && testCase.shouldSucceed {
				t.Errorf("[%s] CompressFileFailed: %v", formatName, err)
			}

			if err != nil && !testCase.shouldSucceed {
				if testCase.expectedError != "" && err.Error() != testCase.expectedError {
					t.Errorf(
						"[%s] WrongErrorMessage: '%s' vs '%s'",
						formatName, testCase.expectedError, err.Error(),
					)
				}
			}

			if err == nil && !testCase.shouldSucceed {
				t.Errorf("[%s] UnexpectedCompressionSuccess: %s", formatName, testCase.expectedError)
			}

			if err == nil && !strings.HasSuffix(compressedFilePath, testCase.expectedSuffix) {
				t.Errorf(
					"[%s] WrongCompressionSuffix: '%s' vs '%s'",
					formatName, testCase.expectedSuffix, compressedFilePath,
				)
			}

			err = clerk.DeleteFile(testFile)
			if err != nil {
				t.Errorf("DeleteOriginalFileFailed: %v", err)
			}
		}
	})

	t.Run("CompressNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := clerk.CompressFile(nonExistentFile, nil, nil)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})

	t.Run("SourceDeletionContractIsFormatIndependent", func(t *testing.T) {
		keepSourceTrue := true
		formats := []string{"tar", "gz", "zip", "xz", "br"}

		for _, format := range formats {
			formatName := format

			t.Run("Keep_"+formatName, func(t *testing.T) {
				testFile := filepath.Join(tempDir, "keep_"+formatName+".txt")
				err := clerk.WriteNewFile(testFile, "content for keep "+formatName, 0644)
				if err != nil {
					t.Fatalf("WriteNewFileFailed: %v", err)
				}

				compressedFile, err := clerk.CompressFile(testFile, &formatName, &keepSourceTrue)
				if err != nil {
					t.Fatalf("[%s] CompressFileFailed: %v", formatName, err)
				}
				if !clerk.IsFile(testFile) {
					t.Errorf("[%s] SourceMustBeKept: %s", formatName, testFile)
				}

				_ = clerk.DeleteFile(testFile)
				_ = clerk.DeleteFile(compressedFile)
			})

			t.Run("Remove_"+formatName, func(t *testing.T) {
				testFile := filepath.Join(tempDir, "remove_"+formatName+".txt")
				err := clerk.WriteNewFile(testFile, "content for remove "+formatName, 0644)
				if err != nil {
					t.Fatalf("WriteNewFileFailed: %v", err)
				}

				compressedFile, err := clerk.CompressFile(testFile, &formatName, nil)
				if err != nil {
					t.Fatalf("[%s] CompressFileFailed: %v", formatName, err)
				}
				if clerk.FileExists(testFile) {
					t.Errorf("[%s] SourceMustBeRemoved: %s", formatName, testFile)
				}

				_ = clerk.DeleteFile(compressedFile)
			})
		}
	})
}

func TestCompressDir(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("DefaultBrotliStageSucceedsAndCleansIntermediateTar", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "compressdir_br")
		nestedFile := filepath.Join(sourceDir, "nested.txt")

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		err = clerk.WriteNewFile(nestedFile, "nested content", 0644)
		if err != nil {
			t.Fatalf("WriteNewFileFailed: %v", err)
		}

		compressedFilePath, err := clerk.CompressDir(sourceDir, nil)
		if err != nil {
			t.Fatalf("CompressDirFailed: %v", err)
		}

		if !strings.HasSuffix(compressedFilePath, ".tar.br") {
			t.Errorf("WrongCompressionSuffix: '%s'", compressedFilePath)
		}
		if !clerk.IsFile(compressedFilePath) {
			t.Errorf("CompressedFileShouldExist: %s", compressedFilePath)
		}
		if clerk.FileExists(sourceDir + ".tar") {
			t.Errorf("IntermediateTarShouldBeRemoved: %s", sourceDir+".tar")
		}
		if !clerk.IsDir(sourceDir) {
			t.Errorf("SourceDirMustBeKept: %s", sourceDir)
		}
	})

	t.Run("TarFormatSkipsSecondStage", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "compressdir_tar")
		tarFormat := "tar"

		err := clerk.CreateDir(sourceDir)
		if err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		compressedFilePath, err := clerk.CompressDir(sourceDir, &tarFormat)
		if err != nil {
			t.Fatalf("CompressDirFailed: %v", err)
		}
		if !strings.HasSuffix(compressedFilePath, ".tar") {
			t.Errorf("WrongCompressionSuffix: '%s'", compressedFilePath)
		}
	})
}

func TestDecompressFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("DecompressFileWithFormats", func(t *testing.T) {
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
				t.Fatalf("TouchFileFailed: %v", err)
			}

			err = os.WriteFile(testFile, []byte(preCompressedContent), 0644)
			if err != nil {
				t.Fatalf("WriteFileFailed: %v", err)
			}

			compressedFile, err := clerk.CompressFile(testFile, &testCase.format, nil)
			if err != nil && testCase.shouldSucceed {
				t.Fatalf("[%s] CompressFileFailed: %v", testCase.format, err)

				if clerk.FileExists(testFile) {
					err = clerk.DeleteFile(testFile)
					if err != nil {
						t.Errorf("DeleteOriginalFileFailed: %v", err)
					}
				}

				if clerk.FileExists(compressedFile) {
					err = clerk.DeleteFile(compressedFile)
					if err != nil {
						t.Errorf("DeleteCompressedFileFailed: %v", err)
					}
				}

				continue
			}

			decompressedFile, err := clerk.DecompressFile(compressedFile, nil, testCase.shouldKeepSource)
			if err != nil && testCase.shouldSucceed {
				t.Errorf("[%s] DecompressFileFailed: %v", testCase.format, err)
			}

			if err == nil && !testCase.shouldSucceed {
				t.Errorf("[%s] UnexpectedDecompressionSuccess", testCase.format)
			}

			if decompressedFile == "" && testCase.shouldSucceed {
				t.Errorf("[%s] DecompressedFilePathShouldNotBeEmpty", testCase.format)
			}

			if testCase.shouldSucceed && clerk.FileExists(decompressedFile) {
				decompressedContent, err := clerk.ReadFileContent(decompressedFile, nil)
				if err != nil {
					t.Fatalf("[%s]ReadDecompressedFileFailed: %v", testCase.format, err)
				}
				if decompressedContent != preCompressedContent {
					t.Errorf(
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
				t.Errorf("[%s] CompressedFileShouldBeKept", testCase.format)
			}

			if !shouldKeep && clerk.FileExists(compressedFile) {
				t.Errorf("[%s] CompressedFileShouldBeRemoved", testCase.format)
			}

			if clerk.FileExists(decompressedFile) {
				err = clerk.DeleteFile(decompressedFile)
				if err != nil {
					t.Errorf("DeleteDecompressedFileFailed: %v", err)
				}
			}

			if clerk.FileExists(compressedFile) {
				err = clerk.DeleteFile(compressedFile)
				if err != nil {
					t.Errorf("DeleteCompressedFileFailed: %v", err)
				}
			}
		}
	})

	t.Run("DecompressNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.tar")

		_, err := clerk.DecompressFile(nonExistentFile, nil, nil)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
		}
	})
}

func TestFileContentRegexSearch(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

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
		t.Run(testCase.description, func(t *testing.T) {
			targetFile := filepath.Join(tempDir, testCase.description+".txt")
			createErr := clerk.TouchFile(targetFile)
			if createErr != nil {
				t.Fatalf("TouchFileFailed: %v", createErr)
			}

			if testCase.fileContent != "" {
				writeErr := os.WriteFile(
					targetFile, []byte(testCase.fileContent), 0644,
				)
				if writeErr != nil {
					t.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}

			if testCase.description == "NonExistentFileReturnsFileNotFound" {
				deleteErr := clerk.DeleteFile(targetFile)
				if deleteErr != nil {
					t.Fatalf("DeleteFileFailed: %v", deleteErr)
				}
			}

			if testCase.description == "DirectoryPathReturnsTargetIsDirectory" {
				deleteErr := clerk.DeleteFile(targetFile)
				if deleteErr != nil {
					t.Fatalf("DeleteFileFailed: %v", deleteErr)
				}
				createDirErr := clerk.CreateDir(targetFile)
				if createDirErr != nil {
					t.Fatalf("CreateDirFailed: %v", createDirErr)
				}
			}

			filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
				targetFile, true,
			)
			if pathErr != nil {
				t.Fatalf("NewUnixAbsoluteFilePathFailed: %v", pathErr)
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
					t.Errorf("UnexpectedError: '%s'", searchErr.Error())
					return
				}
			}

			if !testCase.expectedErrIsNil {
				if searchErr == nil {
					t.Errorf("MissingExpectedError: %s", testCase.expectedErrMsg)
					return
				}
				if testCase.expectedErrMsg != "" &&
					searchErr.Error() != testCase.expectedErrMsg {
					t.Errorf(
						"WrongErrorMessage: '%s' vs '%s'",
						testCase.expectedErrMsg, searchErr.Error(),
					)
				}
				return
			}

			if len(regexSearchFindings) != len(testCase.expectedFindings) {
				t.Errorf(
					"WrongFindingsCount: expected=%d actual=%d",
					len(testCase.expectedFindings), len(regexSearchFindings),
				)
				return
			}

			for findingIndex, expectedFinding := range testCase.expectedFindings {
				actualFinding := regexSearchFindings[findingIndex]
				if actualFinding.Match != expectedFinding.Match {
					t.Errorf(
						"WrongFindingMatch%d: '%s' vs '%s'",
						findingIndex, expectedFinding.Match, actualFinding.Match,
					)
				}
				if len(actualFinding.LineNumRange) != 2 ||
					len(expectedFinding.LineNumRange) != 2 {
					t.Errorf(
						"WrongFindingLineRangeShape%d: expected=%v actual=%v",
						findingIndex,
						expectedFinding.LineNumRange,
						actualFinding.LineNumRange,
					)
					continue
				}
				if actualFinding.LineNumRange[0] != expectedFinding.LineNumRange[0] {
					t.Errorf(
						"WrongFindingLineRangeStart%d: '%d' vs '%d'",
						findingIndex,
						expectedFinding.LineNumRange[0],
						actualFinding.LineNumRange[0],
					)
				}
				if actualFinding.LineNumRange[1] != expectedFinding.LineNumRange[1] {
					t.Errorf(
						"WrongFindingLineRangeEnd%d: '%d' vs '%d'",
						findingIndex,
						expectedFinding.LineNumRange[1],
						actualFinding.LineNumRange[1],
					)
				}
				if len(actualFinding.Groups) != len(expectedFinding.Groups) {
					t.Errorf(
						"WrongFindingGroupCount%d: expected=%d actual=%d",
						findingIndex,
						len(expectedFinding.Groups), len(actualFinding.Groups),
					)
					continue
				}
				for groupIndex, expectedGroup := range expectedFinding.Groups {
					if actualFinding.Groups[groupIndex] != expectedGroup {
						t.Errorf(
							"WrongFindingGroup%d-%d: '%s' vs '%s'",
							findingIndex, groupIndex,
							expectedGroup, actualFinding.Groups[groupIndex],
						)
					}
				}
			}

			cleanupErr := clerk.DeleteFile(targetFile)
			if cleanupErr != nil {
				t.Errorf("DeleteFileFailed: %v", cleanupErr)
			}

			if testCase.description == "DirectoryPathReturnsTargetIsDirectory" {
				dirCleanupErr := clerk.DeleteDir(targetFile)
				if dirCleanupErr != nil {
					t.Errorf("DeleteDirFailed: %v", dirCleanupErr)
				}
			}
		})
	}
}

func TestOverwriteFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("OverwriteExistingTarget", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		sourceContent := "new content via overwrite"
		targetContent := "old content to be overwritten"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		if err != nil {
			t.Fatalf("UpdateSourceFileContentFailed: %v", err)
		}

		err = clerk.TouchFile(targetFile)
		if err != nil {
			t.Fatalf("CreateTargetFileFailed: %v", err)
		}

		err = os.WriteFile(targetFile, []byte(targetContent), 0644)
		if err != nil {
			t.Fatalf("UpdateTargetFileContentFailed: %v", err)
		}

		err = clerk.OverwriteFile(sourceFile, targetFile)
		if err != nil {
			t.Errorf("OverwriteFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			t.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			t.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		actualContent, err := clerk.ReadFileContent(targetFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != sourceContent {
			t.Errorf(
				"ContentMismatch: '%s' vs '%s'",
				actualContent, sourceContent,
			)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("OverwriteWithNonExistentTarget", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source2.txt")
		targetFile := filepath.Join(tempDir, "new_target.txt")
		sourceContent := "content"

		err := clerk.TouchFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		if err != nil {
			t.Fatalf("UpdateSourceFileContentFailed: %v", err)
		}

		err = clerk.OverwriteFile(sourceFile, targetFile)
		if err != nil {
			t.Errorf("OverwriteFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			t.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			t.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("OverwriteNonExistentSource", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "nonexistent.txt")
		targetFile := filepath.Join(tempDir, "target.txt")

		err := clerk.OverwriteFile(sourceFile, targetFile)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			t.Errorf(
				"WrongErrorMessage: '%s' vs '%s'",
				"SourceFileNotFound", err.Error(),
			)
		}
	})

	t.Run("OverwriteSymlinkTargetReplacesUnderlyingFile", func(t *testing.T) {
		realFile := filepath.Join(tempDir, "real.txt")
		linkFile := filepath.Join(tempDir, "link.txt")
		sourceFile := filepath.Join(tempDir, "source3.txt")

		if err := os.WriteFile(realFile, []byte("original"), 0644); err != nil {
			t.Fatalf("UpdateRealFileContentFailed: %v", err)
		}
		if err := os.Symlink(realFile, linkFile); err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}
		if err := os.WriteFile(sourceFile, []byte("replaced"), 0644); err != nil {
			t.Fatalf("UpdateSourceFileContentFailed: %v", err)
		}

		if err := clerk.OverwriteFile(sourceFile, linkFile); err != nil {
			t.Fatalf("OverwriteFileFailed: %v", err)
		}

		if !clerk.IsSymlink(linkFile) {
			t.Errorf("SymlinkShouldBePreserved: %s", linkFile)
		}
		realContent, err := clerk.ReadFileContent(realFile, nil)
		if err != nil {
			t.Fatalf("ReadFileContentFailed: %v", err)
		}
		if realContent != "replaced" {
			t.Errorf("ContentMismatch: '%s' vs 'replaced'", realContent)
		}
	})

	t.Run("OverwriteWithDirectorySourceReturnsSourceIsDirectory", func(t *testing.T) {
		sourceDir := filepath.Join(tempDir, "sourceDir")
		if err := clerk.CreateDir(sourceDir); err != nil {
			t.Fatalf("CreateDirFailed: %v", err)
		}

		err := clerk.OverwriteFile(sourceDir, filepath.Join(tempDir, "any.txt"))
		if err == nil || err.Error() != "SourceIsDirectory" {
			t.Errorf("WrongErrorMessage: 'SourceIsDirectory' vs '%v'", err)
		}
	})
}

func TestFileContentRegexReplace(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

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
		t.Run(testCase.description, func(t *testing.T) {
			targetFile := filepath.Join(tempDir, testCase.description+".txt")
			createErr := clerk.TouchFile(targetFile)
			if createErr != nil {
				t.Fatalf("TouchFileFailed: %v", createErr)
			}

			if testCase.fileContent != "" {
				writeErr := os.WriteFile(
					targetFile, []byte(testCase.fileContent), 0644,
				)
				if writeErr != nil {
					t.Fatalf("WriteFileFailed: %v", writeErr)
				}
			}

			if testCase.description == "NonExistentFileReturnsFileNotFound" {
				deleteErr := clerk.DeleteFile(targetFile)
				if deleteErr != nil {
					t.Fatalf("DeleteFileFailed: %v", deleteErr)
				}
			}

			filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(
				targetFile, true,
			)
			if pathErr != nil {
				t.Fatalf("NewUnixAbsoluteFilePathFailed: %v", pathErr)
			}

			var regexPattern *regexp.Regexp
			if !testCase.shouldProvideNil {
				regexPattern = regexp.MustCompile(testCase.patternSource)
			}

			replacementCount, replaceErr := clerk.FileContentRegexReplace(
				filePath, regexPattern, testCase.replacement,
			)

			if testCase.expectedErrIsNil {
				if replaceErr != nil {
					t.Errorf("UnexpectedError: '%s'", replaceErr.Error())
					return
				}

				if replacementCount != testCase.expectedCount {
					t.Errorf(
						"WrongReplacementCount: expected=%d actual=%d",
						testCase.expectedCount, replacementCount,
					)
				}

				actualContent, readErr := clerk.ReadFileContent(targetFile, nil)
				if readErr != nil {
					t.Errorf("ReadFileContentFailed: %v", readErr)
					return
				}

				if actualContent != testCase.expectedContent {
					t.Errorf(
						"WrongContent: '%s' vs '%s'",
						actualContent, testCase.expectedContent,
					)
				}

				tempLeftoverMatches, _ := filepath.Glob(
					filepath.Join(filepath.Dir(targetFile), ".*.tk-tmp*"),
				)
				for _, tempLeftoverPath := range tempLeftoverMatches {
					t.Errorf(
						"TempFileShouldNotExistAfterReplace: %s",
						tempLeftoverPath,
					)
					_ = clerk.DeleteFile(tempLeftoverPath)
				}

				return
			}

			if replaceErr == nil {
				t.Errorf("MissingExpectedError: %s", testCase.expectedErrMsg)
				return
			}
			if testCase.expectedErrMsg != "" &&
				replaceErr.Error() != testCase.expectedErrMsg {
				t.Errorf(
					"WrongErrorMessage: '%s' vs '%s'",
					testCase.expectedErrMsg, replaceErr.Error(),
				)
			}

			tempLeftoverMatches, _ := filepath.Glob(
				filepath.Join(filepath.Dir(targetFile), ".*.tk-tmp*"),
			)
			for _, tempLeftoverPath := range tempLeftoverMatches {
				t.Errorf(
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
					t.Errorf("ReadFileContentAfterFailFailed: %v", readErr)
					return
				}
				if actualContentAfterFail != testCase.expectedContent {
					t.Errorf(
						"OriginalFileShouldBeUnchangedAfterFailedReplace: '%s' vs '%s'",
						testCase.expectedContent, actualContentAfterFail,
					)
				}
			}
		})
	}
}

func absoluteFilePathForTest(
	t *testing.T, path string,
) tkValueObject.UnixAbsoluteFilePath {
	t.Helper()
	filePath, pathErr := tkValueObject.NewUnixAbsoluteFilePath(path, true)
	if pathErr != nil {
		t.Fatalf("FilePathInvalid: %v", pathErr)
	}
	return filePath
}

func TestTempFileNameFactory(t *testing.T) {
	clerk := FileClerk{}

	targetFileName, fileNameErr := tkValueObject.NewUnixFileName(
		"unit.service", true,
	)
	if fileNameErr != nil {
		t.Fatalf("FileNameInvalid: %v", fileNameErr)
	}

	t.Run("KeepsBaseNameWithTempMarker", func(t *testing.T) {
		tempFileName := clerk.TempFileNameFactory(targetFileName)

		if !strings.HasPrefix(tempFileName, ".unit.service.") {
			t.Errorf("MissingHiddenBaseNamePrefix: %s", tempFileName)
		}
		if !strings.HasSuffix(tempFileName, tempFileNameSuffix) {
			t.Errorf("MissingTempSuffix: %s", tempFileName)
		}
	})

	t.Run("DistinctAcrossCalls", func(t *testing.T) {
		first := clerk.TempFileNameFactory(targetFileName)
		second := clerk.TempFileNameFactory(targetFileName)
		if first == second {
			t.Errorf("IdenticalTempNamesAcrossCalls: %s", first)
		}
	})
}

func TestTempFilePathFactory(t *testing.T) {
	clerk := FileClerk{}

	target := absoluteFilePathForTest(t, "/tmp/example/unit.service")

	t.Run("KeepsDirAndBaseNameWithTempMarker", func(t *testing.T) {
		tempPath, tempPathErr := clerk.TempFilePathFactory(target)
		if tempPathErr != nil {
			t.Fatalf("UnexpectedError: %v", tempPathErr)
		}

		if filepath.Dir(tempPath) != "/tmp/example" {
			t.Errorf("UnexpectedDir: %s", tempPath)
		}
		baseName := filepath.Base(tempPath)
		if !strings.HasPrefix(baseName, ".unit.service.") {
			t.Errorf("MissingHiddenBaseNamePrefix: %s", baseName)
		}
		if !strings.HasSuffix(baseName, tempFileNameSuffix) {
			t.Errorf("MissingTempSuffix: %s", baseName)
		}
	})

	t.Run("DistinctAcrossCalls", func(t *testing.T) {
		first, firstErr := clerk.TempFilePathFactory(target)
		if firstErr != nil {
			t.Fatalf("UnexpectedError: %v", firstErr)
		}
		second, secondErr := clerk.TempFilePathFactory(target)
		if secondErr != nil {
			t.Fatalf("UnexpectedError: %v", secondErr)
		}
		if first == second {
			t.Errorf("IdenticalTempNamesAcrossCalls: %s", first)
		}
	})
}

func TestUpsertFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	newFileContent := []byte("unit content")

	overwritePolicy := FileClerkOverwritePolicyReplace
	symlinkPolicy := FileClerkSymlinkPolicyResolve

	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		t.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUsername, usernameErr := tkValueObject.NewUnixUsername(
		currentAccount.Username,
	)
	if usernameErr != nil {
		t.Fatalf("CurrentUsernameInvalid: %v", usernameErr)
	}

	t.Run("WritesNewFileWithDefaultMode", func(t *testing.T) {
		dir := filepath.Join(tempDir, "defaultMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(t, target),
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			t.Errorf("ContentMismatch: %s", written)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != FileClerkDefaultNewFileMode {
			t.Errorf(
				"ModeMismatch: expected %v, got %v",
				FileClerkDefaultNewFileMode, fileInfo.Mode().Perm(),
			)
		}
	})

	t.Run("WritesNewFileWithStatedMode", func(t *testing.T) {
		dir := filepath.Join(tempDir, "statedMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		statedPermissions := os.FileMode(0640)
		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:    absoluteFilePathForTest(t, target),
			Permissions: &statedPermissions,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0640 {
			t.Errorf("ModeMismatch: expected 0640, got %v", fileInfo.Mode().Perm())
		}
	})

	t.Run("WritesNewFileWithZeroStatedMode", func(t *testing.T) {
		dir := filepath.Join(tempDir, "zeroMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		zeroPermissions := os.FileMode(0)
		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:    absoluteFilePathForTest(t, target),
			Permissions: &zeroPermissions,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0 {
			t.Errorf("ModeMismatch: expected 0000, got %v", fileInfo.Mode().Perm())
		}
	})

	t.Run("WritesNewFileWithProcessOwnership", func(t *testing.T) {
		dir := filepath.Join(tempDir, "processOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(t, target),
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			t.Errorf("ContentMismatch: %s", written)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != os.Geteuid() || int(fileStat.Gid) != os.Getegid() {
			t.Errorf(
				"OwnershipMismatch: expected %d:%d, got %d:%d",
				os.Geteuid(), os.Getegid(), fileStat.Uid, fileStat.Gid,
			)
		}
	})

	t.Run("UsesRunningProcessOwnerWhenRequested", func(t *testing.T) {
		dir := filepath.Join(tempDir, "runningProcessOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceRunningProcess
		target := filepath.Join(dir, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:    absoluteFilePathForTest(t, target),
			OwnerSource: &ownerSource,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != os.Geteuid() || int(fileStat.Gid) != os.Getegid() {
			t.Errorf(
				"OwnershipMismatch: expected %d:%d, got %d:%d",
				os.Geteuid(), os.Getegid(), fileStat.Uid, fileStat.Gid,
			)
		}
	})

	t.Run("RefusesTempNameAboveNameMax", func(t *testing.T) {
		dir := filepath.Join(tempDir, "longName")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		longBaseName := strings.Repeat("a", 246) + ".txt"
		statedPermissions := os.FileMode(0644)
		target := filepath.Join(dir, longBaseName)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(t, target),
			Permissions:   &statedPermissions,
			OwnerUsername: &currentUsername,
		}, newFileContent)
		if !errors.Is(err, ErrTempFileNameTooLong) {
			t.Fatalf("ExpectedErrTempFileNameTooLong, got: %v", err)
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 0 {
			t.Errorf("TempFileLeaked: %v", entries)
		}
	})

	t.Run("ReplacesExistingFileWithStatedMode", func(t *testing.T) {
		dir := filepath.Join(tempDir, "replace")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		statedPermissions := os.FileMode(0640)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(t, target),
			OverwritePolicy: &overwritePolicy,
			Permissions:     &statedPermissions,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			t.Errorf("ContentMismatch: %s", written)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if fileInfo.Mode().Perm() != 0640 {
			t.Errorf("ModeMismatch: expected 0640, got %v", fileInfo.Mode().Perm())
		}
	})

	t.Run("InheritsTargetModeAndOwnershipOnReplace", func(t *testing.T) {
		dir := filepath.Join(tempDir, "inherit")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0640)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		originalInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		originalStat, assertOk := originalInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(t, target),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			t.Errorf("ContentMismatch: %s", written)
		}

		replacedInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		if replacedInfo.Mode().Perm() != 0640 {
			t.Errorf(
				"ModeNotInherited: expected 0640, got %v",
				replacedInfo.Mode().Perm(),
			)
		}

		replacedStat, assertOk := replacedInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if replacedStat.Uid != originalStat.Uid || replacedStat.Gid != originalStat.Gid {
			t.Errorf(
				"OwnershipNotInherited: expected %d:%d, got %d:%d",
				originalStat.Uid, originalStat.Gid,
				replacedStat.Uid, replacedStat.Gid,
			)
		}
	})

	t.Run("InheritsSpecialModeBitsOnReplace", func(t *testing.T) {
		dir := filepath.Join(tempDir, "inheritSpecialMode")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0755)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}
		specialMode := os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0755
		err = os.Chmod(target, specialMode)
		if err != nil {
			t.Fatalf("ChmodFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(t, target),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		specialModeMask := os.ModeSetuid | os.ModeSetgid | os.ModeSticky
		inheritedSpecialMode := fileInfo.Mode() & specialModeMask
		if inheritedSpecialMode != specialMode&specialModeMask {
			t.Errorf(
				"SpecialModeNotInherited: expected %v, got %v",
				specialMode, fileInfo.Mode(),
			)
		}
	})

	t.Run("RefusesExistingFileByDefault", func(t *testing.T) {
		dir := filepath.Join(tempDir, "refuseTaken")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.WriteFile(target, []byte("old"), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(t, target),
		}, newFileContent)
		if !errors.Is(err, ErrTargetFileExists) {
			t.Fatalf("ExpectedErrTargetFileExists, got: %v", err)
		}

		written, readErr := os.ReadFile(target)
		if readErr != nil {
			t.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != "old" {
			t.Errorf("ExistingContentWasOverwritten: %s", written)
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 1 || entries[0].Name() != "unit.service" {
			t.Errorf("TempFileLeaked: %v", entries)
		}
	})

	t.Run("RefusesFinalComponentSymlinkByDefault", func(t *testing.T) {
		dir := filepath.Join(tempDir, "refuseLink")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		canaryPath := filepath.Join(dir, "canary.txt")
		canaryContent := "do-not-touch"
		err = os.WriteFile(canaryPath, []byte(canaryContent), 0600)
		if err != nil {
			t.Fatalf("CanaryWriteFailed: %v", err)
		}
		target := filepath.Join(dir, "unit.service")
		err = os.Symlink(canaryPath, target)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(t, target),
		}, newFileContent)
		if !errors.Is(err, ErrTargetIsSymlink) {
			t.Fatalf("ExpectedErrTargetIsSymlink, got: %v", err)
		}
		if !clerk.IsSymlink(target) {
			t.Errorf("SymlinkWasReplacedDespiteRefusal")
		}

		canary, canaryErr := os.ReadFile(canaryPath)
		if canaryErr != nil {
			t.Fatalf("CanaryReadFailed: %v", canaryErr)
		}
		if string(canary) != canaryContent {
			t.Errorf("CanaryWasModifiedThroughSymlink: %s", canary)
		}
	})

	t.Run("ResolvesFinalComponentSymlinkWhenRequested", func(t *testing.T) {
		dir := filepath.Join(tempDir, "resolveLink")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		canaryPath := filepath.Join(dir, "canary.txt")
		err = os.WriteFile(canaryPath, []byte("old"), 0600)
		if err != nil {
			t.Fatalf("CanaryWriteFailed: %v", err)
		}

		target := filepath.Join(dir, "unit.service")
		err = os.Symlink(canaryPath, target)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(t, target),
			SymlinkPolicy:   &symlinkPolicy,
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		if !clerk.IsSymlink(target) {
			t.Errorf("TargetSymlinkWasReplacedInsteadOfFollowed")
		}
		canary, canaryErr := os.ReadFile(canaryPath)
		if canaryErr != nil {
			t.Fatalf("CanaryReadFailed: %v", canaryErr)
		}
		if string(canary) != string(newFileContent) {
			t.Errorf("CanaryContentMismatch: %s", canary)
		}
	})

	t.Run("RefusesSymlinkedParentDirectory", func(t *testing.T) {
		outsideDir := filepath.Join(tempDir, "outside", "nested")
		err := os.MkdirAll(outsideDir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		escapeTarget := filepath.Join(tempDir, "outside")
		trapLink := filepath.Join(tempDir, "trap")
		err = os.Symlink(escapeTarget, trapLink)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		escapedFilePath := filepath.Join(trapLink, "nested", "escape.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(t, escapedFilePath),
		}, newFileContent)
		if !errors.Is(err, ErrSymlinkedPathInvalid) {
			t.Fatalf("ExpectedErrSymlinkedPathInvalid, got: %v", err)
		}

		escapedFile := filepath.Join(outsideDir, "escape.service")
		if clerk.FileExists(escapedFile) {
			t.Errorf("WriteLandedOutsideIntendedTree: %s", escapedFile)
		}
	})

	t.Run("ResolvesSymlinkedParentChainWhenRequested", func(t *testing.T) {
		realParent := filepath.Join(tempDir, "realParent")
		err := os.MkdirAll(realParent, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		linkedParent := filepath.Join(tempDir, "linkedParent")
		err = os.Symlink(realParent, linkedParent)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		target := filepath.Join(linkedParent, "unit.service")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(t, target),
			SymlinkPolicy: &symlinkPolicy,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		written, readErr := os.ReadFile(filepath.Join(realParent, "unit.service"))
		if readErr != nil {
			t.Fatalf("ReadFailed: %v", readErr)
		}
		if string(written) != string(newFileContent) {
			t.Errorf("ContentMismatch: %s", written)
		}
	})

	t.Run("RefusesDegenerateFilePaths", func(t *testing.T) {
		dir := filepath.Join(tempDir, "degenerate")
		err := os.MkdirAll(filepath.Join(dir, "sub"), 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
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
			t.Run(testCase.name, func(t *testing.T) {
				err := clerk.UpsertFile(FileUpsertSettings{
					FilePath: absoluteFilePathForTest(t, testCase.filePath),
				}, newFileContent)
				if !errors.Is(err, ErrFileNameInvalid) {
					t.Errorf(
						"ExpectedErrFileNameInvalid, got: %v [%s]",
						err, testCase.filePath,
					)
				}
			})
		}
	})

	t.Run("RefusesNonDirectoryComponent", func(t *testing.T) {
		blockerFile := filepath.Join(tempDir, "blocker")
		err := os.WriteFile(blockerFile, []byte("x"), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		blockedFilePath := filepath.Join(blockerFile, "nested", "blocked.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath: absoluteFilePathForTest(t, blockedFilePath),
		}, newFileContent)
		if !errors.Is(err, ErrTargetNotDirectory) {
			t.Errorf("ExpectedErrTargetNotDirectory, got: %v", err)
		}
	})

	t.Run("ReadersNeverSeePartialContent", func(t *testing.T) {
		dir := filepath.Join(tempDir, "atomic")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "target.txt")
		targetPath := absoluteFilePathForTest(t, target)
		contentSizeBytes := 1024 * 1024
		contentA := strings.Repeat("a", contentSizeBytes)
		contentB := strings.Repeat("b", contentSizeBytes)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        targetPath,
			OverwritePolicy: &overwritePolicy,
		}, []byte(contentA))
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
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
					t.Errorf("ConcurrentWriteFailed: %v", writeErr)
					return
				}
			}
		}()

	observersLoop:
		for {
			readContent, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Errorf("ConcurrentReadFailed: %v", readErr)
				break
			}
			if string(readContent) != contentA && string(readContent) != contentB {
				t.Errorf(
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

	t.Run("RefusesDirectoryTargetLeavingNoTempFile", func(t *testing.T) {
		dir := filepath.Join(tempDir, "leakCheck")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		blockingDir := filepath.Join(dir, "taken.txt")
		err = os.Mkdir(blockingDir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(t, blockingDir),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if !errors.Is(err, ErrTargetIsDirectory) {
			t.Fatalf("ExpectedErrTargetIsDirectory, got: %v", err)
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 1 || entries[0].Name() != "taken.txt" {
			t.Errorf("TempFileLeaked: %v", entries)
		}
	})

	t.Run("SetsRequestedOwner", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			t.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			t.Fatalf("GidParseFailed: %v", gidErr)
		}

		dir := filepath.Join(tempDir, "ownership")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			t.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		statedPermissions := os.FileMode(0644)
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(t, filepath.Join(dir, "owned.txt")),
			Permissions:   &statedPermissions,
			OwnerUsername: &ownerName,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(filepath.Join(dir, "owned.txt"))
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			t.Errorf(
				"OwnerMismatch: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	t.Run("RefusesDirectoryChainOwnedByUnexpectedUser", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		daemon, lookupErr := user.Lookup("daemon")
		if lookupErr != nil {
			t.Skipf("DaemonUserMissing: %v", lookupErr)
		}
		_, nobodyErr := user.Lookup("nobody")
		if nobodyErr != nil {
			t.Skipf("NobodyUserMissing: %v", nobodyErr)
		}
		daemonUid, uidErr := strconv.Atoi(daemon.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}

		dir := filepath.Join(tempDir, "wrongOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chown(dir, daemonUid, -1)
		if err != nil {
			t.Fatalf("ChownFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			t.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		target := filepath.Join(dir, "refused.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:                absoluteFilePathForTest(t, target),
			TrustedDirOwnerUsername: &ownerName,
		}, newFileContent)
		if !errors.Is(err, ErrDirectoryOwnerInvalid) {
			t.Errorf(
				"ExpectedErrDirectoryOwnerInvalid, got: %v", err,
			)
		}
	})

	t.Run("InheritsTargetOwnerOnReplace", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			t.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			t.Fatalf("GidParseFailed: %v", gidErr)
		}

		dir := filepath.Join(tempDir, "inheritOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "owned.txt")
		err = os.WriteFile(target, []byte("old"), 0640)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}
		err = os.Chown(target, nobodyUid, nobodyGid)
		if err != nil {
			t.Fatalf("ChownFailed: %v", err)
		}

		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:        absoluteFilePathForTest(t, target),
			OverwritePolicy: &overwritePolicy,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			t.Errorf(
				"OwnerNotInherited: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	t.Run("UsesContainingDirectoryOwnerWhenRequested", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			t.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyGid, gidErr := strconv.Atoi(nobody.Gid)
		if gidErr != nil {
			t.Fatalf("GidParseFailed: %v", gidErr)
		}
		nobodyUserId, userIdErr := tkValueObject.NewUnixUserId(nobodyUid)
		if userIdErr != nil {
			t.Fatalf("NobodyUserIdInvalid: %v", userIdErr)
		}

		dir := filepath.Join(tempDir, "dirOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chown(dir, nobodyUid, nobodyGid)
		if err != nil {
			t.Fatalf("ChownFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceContainingDirectory
		target := filepath.Join(dir, "owned.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:              absoluteFilePathForTest(t, target),
			TrustedDirOwnerUserId: &nobodyUserId,
			OwnerSource:           &ownerSource,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != nobodyGid {
			t.Errorf(
				"OwnerMismatch: expected %d:%d, got %d:%d",
				nobodyUid, nobodyGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	t.Run("StatesOwnerGroup", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			t.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		daemon, daemonErr := user.Lookup("daemon")
		if daemonErr != nil {
			t.Skipf("DaemonUserMissing: %v", daemonErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}
		daemonGid, gidErr := strconv.Atoi(daemon.Gid)
		if gidErr != nil {
			t.Fatalf("GidParseFailed: %v", gidErr)
		}

		dir := filepath.Join(tempDir, "statedGroup")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			t.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		ownerGroupId, groupIdErr := tkValueObject.NewUnixGroupId(daemonGid)
		if groupIdErr != nil {
			t.Fatalf("GroupIdInvalid: %v", groupIdErr)
		}

		target := filepath.Join(dir, "owned.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(t, target),
			OwnerUsername: &ownerName,
			OwnerGroupId:  &ownerGroupId,
		}, newFileContent)
		if err != nil {
			t.Fatalf("UpsertFileFailed: %v", err)
		}

		fileInfo, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("StatFailed: %v", statErr)
		}
		fileStat, assertOk := fileInfo.Sys().(*syscall.Stat_t)
		if !assertOk {
			t.Fatalf("StatAssertionFailed")
		}
		if int(fileStat.Uid) != nobodyUid || int(fileStat.Gid) != daemonGid {
			t.Errorf(
				"OwnershipMismatch: expected %d:%d, got %d:%d",
				nobodyUid, daemonGid, fileStat.Uid, fileStat.Gid,
			)
		}
	})

	t.Run("RefusesStatedOwnerWithContainingDirectorySource", func(t *testing.T) {
		dir := filepath.Join(tempDir, "ownerConflict")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceContainingDirectory
		target := filepath.Join(dir, "conflict.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(t, target),
			OwnerSource:   &ownerSource,
			OwnerUsername: &currentUsername,
		}, newFileContent)
		if !errors.Is(err, ErrOwnerSourceConflict) {
			t.Fatalf("ExpectedErrOwnerSourceConflict, got: %v", err)
		}
		if clerk.FileExists(target) {
			t.Errorf("TargetWrittenDespiteConflict")
		}
	})

	t.Run("RefusesStatedOwnerWithRunningProcessSource", func(t *testing.T) {
		dir := filepath.Join(tempDir, "ownerProcessConflict")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		ownerSource := FileClerkOwnerSourceRunningProcess
		target := filepath.Join(dir, "conflict.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:      absoluteFilePathForTest(t, target),
			OwnerSource:   &ownerSource,
			OwnerUsername: &currentUsername,
		}, newFileContent)
		if !errors.Is(err, ErrOwnerSourceConflict) {
			t.Fatalf("ExpectedErrOwnerSourceConflict, got: %v", err)
		}
		if clerk.FileExists(target) {
			t.Errorf("TargetWrittenDespiteConflict")
		}
	})

	t.Run("RefusesUnknownPolicyValues", func(t *testing.T) {
		dir := filepath.Join(tempDir, "unknownPolicy")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		target := filepath.Join(dir, "unknown.txt")
		targetPath := absoluteFilePathForTest(t, target)
		unsupportedOwnerSource := FileClerkOwnerSource("unsupported")
		unsupportedSymlinkPolicy := FileClerkSymlinkPolicy("unsupported")
		unsupportedOverwritePolicy := FileClerkOverwritePolicy("unsupported")

		testCases := []struct {
			name          string
			settings      FileUpsertSettings
			expectedError error
		}{
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
			t.Run(testCase.name, func(t *testing.T) {
				err := clerk.UpsertFile(testCase.settings, newFileContent)
				if !errors.Is(err, testCase.expectedError) {
					t.Errorf("Expected%v, got: %v", testCase.expectedError, err)
				}
			})
		}

		if clerk.FileExists(target) {
			t.Errorf("TargetWrittenDespiteInvalidSettings")
		}
	})

	t.Run("RefusesUnprivilegedOwnerChange", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("NonRootRequired")
		}

		dir := filepath.Join(tempDir, "unprivilegedOwner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		foreignOwnerUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid() + 1)
		if userIdErr != nil {
			t.Fatalf("ForeignOwnerUserIdInvalid: %v", userIdErr)
		}
		foreignOwnerGroupId, groupIdErr := tkValueObject.NewUnixGroupId(os.Getegid())
		if groupIdErr != nil {
			t.Fatalf("ForeignOwnerGroupIdInvalid: %v", groupIdErr)
		}

		target := filepath.Join(dir, "owned.txt")
		err = clerk.UpsertFile(FileUpsertSettings{
			FilePath:     absoluteFilePathForTest(t, target),
			OwnerUserId:  &foreignOwnerUserId,
			OwnerGroupId: &foreignOwnerGroupId,
		}, newFileContent)
		if !errors.Is(err, ErrFileOwnerChangeFailed) {
			t.Fatalf("ExpectedErrFileOwnerChangeFailed, got: %v", err)
		}
		if clerk.FileExists(target) {
			t.Errorf("TargetWrittenDespiteOwnerChangeFailure")
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			t.Fatalf("ReadDirFailed: %v", readErr)
		}
		if len(entries) != 0 {
			t.Errorf("TempFileLeaked: %v", entries)
		}
	})
}

func TestFileUpsertOwnerResolver(t *testing.T) {
	clerk := FileClerk{}

	runningProcessUserId, userIdErr := tkValueObject.NewUnixUserId(os.Geteuid())
	if userIdErr != nil {
		t.Fatalf("RunningProcessUserIdInvalid: %v", userIdErr)
	}
	runningProcessGroupId, groupIdErr := tkValueObject.NewUnixGroupId(os.Getegid())
	if groupIdErr != nil {
		t.Fatalf("RunningProcessGroupIdInvalid: %v", groupIdErr)
	}

	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		t.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUsername, usernameErr := tkValueObject.NewUnixUsername(
		currentAccount.Username,
	)
	if usernameErr != nil {
		t.Fatalf("CurrentUsernameInvalid: %v", usernameErr)
	}
	currentUserId, userIdErr := tkValueObject.NewUnixUserId(currentAccount.Uid)
	if userIdErr != nil {
		t.Fatalf("CurrentUserIdInvalid: %v", userIdErr)
	}
	currentPrimaryGroupId, groupIdErr := tkValueObject.NewUnixGroupId(
		currentAccount.Gid,
	)
	if groupIdErr != nil {
		t.Fatalf("CurrentPrimaryGroupIdInvalid: %v", groupIdErr)
	}

	existingFileOwnerUserId := tkValueObject.UnixUserId(1234)
	existingFileOwnerGroupId := tkValueObject.UnixGroupId(1235)
	containingDirOwnerUserId := tkValueObject.UnixUserId(2345)
	containingDirOwnerGroupId := tkValueObject.UnixGroupId(2346)
	statedOwnerUserId := tkValueObject.UnixUserId(3456)
	statedGroupId := tkValueObject.UnixGroupId(3457)

	targetStateExisting := targetFileState{
		Exists:       true,
		OwnerUserId:  existingFileOwnerUserId,
		OwnerGroupId: existingFileOwnerGroupId,
	}
	targetStateMissing := targetFileState{}
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
		targetState     targetFileState
		expectedUserId  tkValueObject.UnixUserId
		expectedGroupId tkValueObject.UnixGroupId
		expectedError   error
	}{
		{
			name:            "StatedAccountUsesPrimaryGroup",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			ownerUsername:   &currentUsername,
			targetState:     targetStateMissing,
			expectedUserId:  currentUserId,
			expectedGroupId: currentPrimaryGroupId,
		},
		{
			name:            "StatedUserIdAndGroupWin",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			ownerUserId:     &statedOwnerUserId,
			ownerGroupId:    &statedGroupId,
			targetState:     targetStateExisting,
			expectedUserId:  statedOwnerUserId,
			expectedGroupId: statedGroupId,
		},
		{
			name:            "ExistingFileSourceInheritsTargetOwnership",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			targetState:     targetStateExisting,
			expectedUserId:  existingFileOwnerUserId,
			expectedGroupId: existingFileOwnerGroupId,
		},
		{
			name:            "ExistingFileSourceFallsBackToRunningProcess",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			targetState:     targetStateMissing,
			expectedUserId:  runningProcessUserId,
			expectedGroupId: runningProcessGroupId,
		},
		{
			name:            "ContainingDirectorySourceUsesDirectoryOwnership",
			ownerSource:     FileClerkOwnerSourceContainingDirectory,
			targetState:     targetStateMissing,
			expectedUserId:  containingDirOwnerUserId,
			expectedGroupId: containingDirOwnerGroupId,
		},
		{
			name:            "RunningProcessSourceUsesProcessOwnership",
			ownerSource:     FileClerkOwnerSourceRunningProcess,
			targetState:     targetStateMissing,
			expectedUserId:  runningProcessUserId,
			expectedGroupId: runningProcessGroupId,
		},
		{
			name:            "StatedGroupOverridesInheritedGroup",
			ownerSource:     FileClerkOwnerSourceExistingFile,
			ownerGroupId:    &statedGroupId,
			targetState:     targetStateExisting,
			expectedUserId:  existingFileOwnerUserId,
			expectedGroupId: statedGroupId,
		},
		{
			name:          "StatedAccountConflictsWithContainingDirectory",
			ownerSource:   FileClerkOwnerSourceContainingDirectory,
			ownerUsername: &currentUsername,
			targetState:   targetStateMissing,
			expectedError: ErrOwnerSourceConflict,
		},
		{
			name:          "StatedAccountConflictsWithRunningProcess",
			ownerSource:   FileClerkOwnerSourceRunningProcess,
			ownerUsername: &currentUsername,
			targetState:   targetStateMissing,
			expectedError: ErrOwnerSourceConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ownership, err := clerk.fileUpsertOwnerResolver(
				testCase.ownerSource, testCase.ownerUsername, testCase.ownerUserId,
				testCase.ownerGroupId, testCase.targetState, containingDirStat,
			)

			if testCase.expectedError != nil {
				if !errors.Is(err, testCase.expectedError) {
					t.Fatalf("Expected%v, got: %v", testCase.expectedError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnexpectedError: %v", err)
			}
			if ownership.UserId != testCase.expectedUserId {
				t.Errorf(
					"UserIdMismatch: expected %d, got %d",
					testCase.expectedUserId, ownership.UserId,
				)
			}
			if ownership.GroupId != testCase.expectedGroupId {
				t.Errorf(
					"GroupIdMismatch: expected %d, got %d",
					testCase.expectedGroupId, ownership.GroupId,
				)
			}
		})
	}
}

func TestFileUpsertPermissionsResolver(t *testing.T) {
	clerk := FileClerk{}

	statedPermissions := os.FileMode(0640)
	zeroPermissions := os.FileMode(0)
	existingPermissions := os.FileMode(0604)

	targetStateExisting := targetFileState{
		Exists:      true,
		Permissions: existingPermissions,
	}

	testCases := []struct {
		name                 string
		statedPermissionsPtr *os.FileMode
		targetState          targetFileState
		expectedPermissions  os.FileMode
	}{
		{
			"StatedPermissionsWin",
			&statedPermissions, targetStateExisting, statedPermissions,
		},
		{
			"StatedZeroPermissionsWin",
			&zeroPermissions, targetStateExisting, zeroPermissions,
		},
		{
			"InheritsExistingPermissions",
			nil, targetStateExisting, existingPermissions,
		},
		{
			"DefaultsForMissingTarget",
			nil, targetFileState{}, FileClerkDefaultNewFileMode,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			permissions := clerk.fileUpsertPermissionsResolver(
				testCase.statedPermissionsPtr, testCase.targetState,
			)
			if permissions != testCase.expectedPermissions {
				t.Errorf(
					"PermissionsMismatch: expected %v, got %v",
					testCase.expectedPermissions, permissions,
				)
			}
		})
	}
}

func TestUnixFileModeConverter(t *testing.T) {
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
		t.Run(testCase.name, func(t *testing.T) {
			fileMode := clerk.unixFileModeConverter(testCase.rawMode)
			if fileMode != testCase.expectedFileMode {
				t.Errorf(
					"FileModeMismatch: expected %v, got %v",
					testCase.expectedFileMode, fileMode,
				)
			}
		})
	}
}

func TestVerifyDirPathRedirectSafety(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()
	currentAccount, accountErr := user.Current()
	if accountErr != nil {
		t.Fatalf("CurrentUserLookupFailed: %v", accountErr)
	}
	currentUsername, usernameErr := tkValueObject.NewUnixUsername(
		currentAccount.Username,
	)
	if usernameErr != nil {
		t.Fatalf("CurrentUsernameInvalid: %v", usernameErr)
	}
	currentUserId, userIdErr := tkValueObject.NewUnixUserId(currentAccount.Uid)
	if userIdErr != nil {
		t.Fatalf("CurrentUserIdInvalid: %v", userIdErr)
	}

	t.Run("AllClearChain", func(t *testing.T) {
		dir := filepath.Join(tempDir, "scout", "inner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		err = clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, dir), &currentUsername, nil,
		)
		if err != nil {
			t.Errorf("UnexpectedSurprise: %v", err)
		}
	})

	t.Run("AcceptsOmittedOwnerUsingProcessAccount", func(t *testing.T) {
		err := clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, tempDir), nil, nil,
		)
		if err != nil {
			t.Errorf("UnexpectedSurprise: %v", err)
		}
	})

	t.Run("AcceptsUserIdInsteadOfUsername", func(t *testing.T) {
		err := clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, tempDir), nil, &currentUserId,
		)
		if err != nil {
			t.Errorf("UnexpectedSurprise: %v", err)
		}
	})

	t.Run("AcceptsUserIdWithoutAccountEntry", func(t *testing.T) {
		unresolvableUserIdFinder := func() uint32 {
			for candidate := uint32(999999); candidate > 0; candidate-- {
				_, lookupErr := user.LookupId(strconv.FormatUint(uint64(candidate), 10))
				if lookupErr != nil {
					return candidate
				}
			}

			t.Fatal("NoUnresolvableUserIdFound")
			return 0
		}

		unresolvableUserIdVo, userIdErr := tkValueObject.NewUnixUserId(
			unresolvableUserIdFinder(),
		)
		if userIdErr != nil {
			t.Fatalf("UnresolvableUserIdInvalid: %v", userIdErr)
		}

		err := clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, tempDir), nil, &unresolvableUserIdVo,
		)
		if err != nil && !errors.Is(err, ErrDirectoryOwnerInvalid) {
			t.Errorf(
				"ExpectedOwnershipComparisonNotAccountLookup, got: %v", err,
			)
		}
	})

	t.Run("HandleStaysOnOriginalInodeAfterPathSwap", func(t *testing.T) {
		parentDir := filepath.Join(tempDir, "pinning")
		originalDir := filepath.Join(parentDir, "original")
		err := os.MkdirAll(originalDir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		dirHandle, dirChainErr := clerk.openRedirectProofDirChain(
			absoluteFilePathForTest(t, originalDir), currentUserId,
		)
		if dirChainErr != nil {
			t.Fatalf("RedirectProofFailed: %v", dirChainErr)
		}
		defer func() { _ = unix.Close(dirHandle) }()

		movedDir := filepath.Join(parentDir, "moved")
		err = os.Rename(originalDir, movedDir)
		if err != nil {
			t.Fatalf("RenameFailed: %v", err)
		}
		escapeDir := filepath.Join(parentDir, "escape")
		err = os.MkdirAll(escapeDir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Symlink(escapeDir, originalDir)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		stageHandle, openErr := unix.Openat(
			dirHandle, "staged.txt",
			unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL, 0o600,
		)
		if openErr != nil {
			t.Fatalf("OpenatFailed: %v", openErr)
		}
		_ = unix.Close(stageHandle)

		if !clerk.FileExists(filepath.Join(movedDir, "staged.txt")) {
			t.Errorf("WriteDidNotLandOnScoutedInode")
		}
		if clerk.FileExists(filepath.Join(escapeDir, "staged.txt")) {
			t.Errorf("WriteFollowedReplacementSymlink")
		}
	})

	t.Run("RefusesParentTraversalComponents", func(t *testing.T) {
		traversalPath := tempDir + "/.."
		err := clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, traversalPath), &currentUsername, nil,
		)
		if !errors.Is(err, ErrDirPathTraversalInvalid) {
			t.Errorf("ExpectedErrDirPathTraversalInvalid, got: %v", err)
		}
	})

	t.Run("ToleratesDotAndEmptyComponents", func(t *testing.T) {
		dir := filepath.Join(tempDir, "dotted", "inner")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}

		dottedPath := tempDir + "/./dotted//inner"
		err = clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, dottedPath), &currentUsername, nil,
		)
		if err != nil {
			t.Errorf("UnexpectedSurprise: %v", err)
		}
	})

	t.Run("ReportsSymlinkedComponent", func(t *testing.T) {
		linkTarget := filepath.Join(tempDir, "scoutTarget")
		err := os.MkdirAll(filepath.Join(linkTarget, "nested"), 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}
		trapLink := filepath.Join(tempDir, "scoutTrap")
		err = os.Symlink(linkTarget, trapLink)
		if err != nil {
			t.Fatalf("SymlinkFailed: %v", err)
		}

		scoutPath := filepath.Join(trapLink, "nested")
		err = clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, scoutPath), &currentUsername, nil,
		)
		if !errors.Is(err, ErrSymlinkedPathInvalid) {
			t.Errorf("ExpectedErrSymlinkedPathInvalid, got: %v", err)
		}
	})

	t.Run("ReportsFileWhereDirExpected", func(t *testing.T) {
		blockerFile := filepath.Join(tempDir, "scoutBlocker")
		err := os.WriteFile(blockerFile, []byte("x"), 0644)
		if err != nil {
			t.Fatalf("WriteFileFailed: %v", err)
		}

		blockedDirPath := filepath.Join(blockerFile, "nested")
		err = clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, blockedDirPath), &currentUsername, nil,
		)
		if !errors.Is(err, ErrTargetNotDirectory) {
			t.Errorf("ExpectedErrTargetNotDirectory, got: %v", err)
		}
	})

	t.Run("ReportsUninspectableStep", func(t *testing.T) {
		missing := filepath.Join(tempDir, "scoutMissing", "deeper")
		err := clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, missing), &currentUsername, nil,
		)
		if err == nil {
			t.Fatalf("MissingErrorForUninspectableStep")
		}
		if !strings.Contains(err.Error(), "PathCheckFailed") {
			t.Errorf("ErrorMissingPathCheckFailed: %v", err)
		}
	})

	t.Run("RefusesStrangerOwnedComponent", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		daemon, lookupErr := user.Lookup("daemon")
		if lookupErr != nil {
			t.Skipf("DaemonUserMissing: %v", lookupErr)
		}
		_, nobodyErr := user.Lookup("nobody")
		if nobodyErr != nil {
			t.Skipf("NobodyUserMissing: %v", nobodyErr)
		}
		daemonUid, uidErr := strconv.Atoi(daemon.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}

		dir := filepath.Join(tempDir, "scoutStranger")
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("MkdirFailed: %v", err)
		}
		err = os.Chown(dir, daemonUid, -1)
		if err != nil {
			t.Fatalf("ChownFailed: %v", err)
		}

		ownerName, ownerNameErr := tkValueObject.NewUnixUsername("nobody")
		if ownerNameErr != nil {
			t.Fatalf("OwnerNameInvalid: %v", ownerNameErr)
		}
		err = clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, dir), &ownerName, nil,
		)
		if !errors.Is(err, ErrDirectoryOwnerInvalid) {
			t.Errorf(
				"ExpectedErrDirectoryOwnerInvalid, got: %v", err,
			)
		}
	})

	t.Run("ReportsUnknownOwner", func(t *testing.T) {
		unknownName, unknownNameErr := tkValueObject.NewUnixUsername(
			"no-such-user-xyz-42",
		)
		if unknownNameErr != nil {
			t.Fatalf("UnknownNameInvalid: %v", unknownNameErr)
		}
		err := clerk.VerifyDirPathRedirectSafety(
			absoluteFilePathForTest(t, tempDir), &unknownName, nil,
		)
		if err == nil {
			t.Fatalf("MissingErrorForUnknownOwner")
		}
		if !strings.Contains(err.Error(), "OwnerLookupFailed") {
			t.Errorf("ErrorMissingOwnerLookupFailed: %v", err)
		}
	})
}
