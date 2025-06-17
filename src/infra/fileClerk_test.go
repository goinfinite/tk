package tkInfra

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFileExists(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("ExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "existing.txt")
		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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
		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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
		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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

func TestCreateFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("CreateNewFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "newfile.txt")
		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Errorf("CreateFileFailed: %v", err)
		}

		if !clerk.IsFile(testFile) {
			t.Errorf("FileShouldExistAfterCreation: %s", testFile)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("CreateFileInNonExistentDir", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent", "subdir")
		testFile := filepath.Join(nonExistentDir, "file.txt")
		err := clerk.CreateFile(testFile)
		if err == nil {
			t.Errorf("MissingExpectedError: CreateFileInNonExistentDir")
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

		err := clerk.CreateFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(sourceFile, testContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
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

		err := clerk.CreateFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateSourceFileFailed: %v", err)
		}

		err = clerk.CreateFile(targetFile)
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
}

func TestMoveFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("MoveExistingFile", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		targetFile := filepath.Join(tempDir, "target.txt")
		testContent := "test content for move"

		err := clerk.CreateFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(sourceFile, testContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
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
}

func TestRenameFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("RenameExistingFile", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "original.txt")
		targetFile := filepath.Join(tempDir, "renamed.txt")

		err := clerk.CreateFile(sourceFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.RenameFile(sourceFile, targetFile)
		if err != nil {
			t.Errorf("RenameFileFailed: %v", err)
		}

		if clerk.IsFile(sourceFile) {
			t.Errorf("SourceFileShouldNotExist: %s", sourceFile)
		}

		if !clerk.IsFile(targetFile) {
			t.Errorf("TargetFileShouldExist: %s", targetFile)
		}

		err = clerk.DeleteFile(targetFile)
		if err != nil {
			t.Errorf("DeleteTargetFileFailed: %v", err)
		}
	})
}

func TestDeleteFile(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("DeleteExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "todelete.txt")

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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
}

func TestReadFileContent(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("ReadExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "content.txt")
		expectedContent := "Hello, World!\nThis is test content."

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, expectedContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
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

	t.Run("ReadWithMaxContentSizeLimit", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "limited.txt")
		fullContent := "This is a longer content that should be truncated when reading with size limit."
		maxSize := int64(20)

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, fullContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, &maxSize)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if int64(len(actualContent)) > maxSize {
			t.Errorf("ContentSizeExceedsLimit:  %d vs %d", len(actualContent), maxSize)
		}

		if len(actualContent) == 0 {
			t.Errorf("ContentShouldNotBeEmpty")
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("ReadWithDefaultMaxSize", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "default_size.txt")
		testContent := "Content with default size limit"

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, testContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
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
}

func TestUpdateFileContent(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("OverwriteContent", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "update.txt")
		initialContent := "Initial content"
		newContent := "New content"

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, initialContent, true)
		if err != nil {
			t.Errorf("UpdateFileContentFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, newContent, true)
		if err != nil {
			t.Errorf("UpdateFileContentFailed: %v", err)
		}

		actualContent, err := clerk.ReadFileContent(testFile, nil)
		if err != nil {
			t.Errorf("ReadFileContentFailed: %v", err)
		}

		if actualContent != newContent {
			t.Errorf("ContentMismatch: '%s' vs '%s'", actualContent, newContent)
		}

		err = clerk.DeleteFile(testFile)
		if err != nil {
			t.Errorf("DeleteFileFailed: %v", err)
		}
	})

	t.Run("AppendContent", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "append.txt")
		initialContent := "Initial"
		appendContent := " Appended"
		expectedContent := initialContent + appendContent

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, initialContent, true)
		if err != nil {
			t.Errorf("UpdateFileContentFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, appendContent, false)
		if err != nil {
			t.Errorf("UpdateFileContentFailed: %v", err)
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
}

func TestDeleteFileContent(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("DeleteContentFromFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "clear.txt")
		initialContent := "Content to be cleared"

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, initialContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
		}

		err = clerk.DeleteFileContent(testFile)
		if err != nil {
			t.Errorf("DeleteFileContentFailed: %v", err)
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

		err = clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateTestFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, testContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
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

		err = clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateTestFileFailed: %v", err)
		}

		err = clerk.UpdateFileContent(testFile, testContent, true)
		if err != nil {
			t.Fatalf("UpdateFileContentFailed: %v", err)
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

		err = clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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
}

func TestIsSymlink(t *testing.T) {
	clerk := FileClerk{}
	tempDir := t.TempDir()

	t.Run("RegularFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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

		err := clerk.CreateFile(targetFile)
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

		err := clerk.CreateFile(targetFile)
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

		err := clerk.CreateFile(regularFile)
		if err != nil {
			t.Fatalf("CreateRegularFileFailed: %v", err)
		}

		err = clerk.CreateFile(targetFile)
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

		err := clerk.CreateFile(targetFile)
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
		customPermissions := int(0600)

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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

		err := clerk.CreateFile(testFile)
		if err != nil {
			t.Fatalf("CreateFileFailed: %v", err)
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

			err := clerk.CreateFile(testFile)
			if err != nil {
				t.Fatalf("CreateFileFailed: %v", err)
			}

			err = clerk.UpdateFileContent(testFile, testContent, true)
			if err != nil {
				t.Fatalf("UpdateFileContentFailed: %v", err)
			}

			compressedFilePath, err := clerk.CompressFile(testFile, testCase.format)
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

		_, err := clerk.CompressFile(nonExistentFile, nil)
		if err == nil {
			t.Errorf("MissingExpectedError: SourceFileNotFound")
		}

		if err != nil && err.Error() != "SourceFileNotFound" {
			t.Errorf("WrongErrorMessage: '%s' vs '%s'", "SourceFileNotFound", err.Error())
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
			testContent := "content to decompress"

			err := clerk.CreateFile(testFile)
			if err != nil {
				t.Fatalf("CreateFileFailed: %v", err)
			}

			err = clerk.UpdateFileContent(testFile, testContent, true)
			if err != nil {
				t.Fatalf("UpdateFileContentFailed: %v", err)
			}

			compressedFile, err := clerk.CompressFile(testFile, &testCase.format)
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
