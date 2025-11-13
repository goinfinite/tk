# [Infinite Toolkit _(TK)_](https://github.com/goinfinite/tk) &middot; [![/r/goinfinite](https://img.shields.io/badge/%2Fr%2Fgoinfinite-FF4500?logo=reddit&logoColor=ffffff)](https://www.reddit.com/r/goinfinite/) [![Discussions](https://img.shields.io/badge/discussions-751A3D?logo=github)](https://github.com/orgs/goinfinite/discussions) [![Report Card](https://img.shields.io/badge/report-A%2B-brightgreen)](https://goreportcard.com/report/github.com/goinfinite/tk) [![License](https://img.shields.io/badge/license-MIT-teal.svg)](https://github.com/goinfinite/tk/blob/main/LICENSE.md)

Infinite Toolkit _(TK)_ offers a comprehensive suite of core components for Infinite projects. The library includes value objects, utilities for Clean Architecture layers, service abstractions, and other foundational elements.

While developed primarily for Infinite ecosystem projects, this open-source library is available for general use under the MIT license.

If you're looking for UI components, please refer to the [Infinite UI](https://github.com/goinfinite/ui) repository.

## Installation

To use Infinite Toolkit _(TK)_ in your project, you can install it using Go modules. Run the following command in your terminal:

```bash
go get github.com/goinfinite/tk
```

## Components

### Infrastructure Helpers

Infinite Toolkit _(TK)_ provides various infrastructure helpers for common tasks:

- **Deserializer**: Deserialize JSON and YAML files into maps for configuration handling.

  ```go
  deserializedMap, deserializationErr := StringDeserializer(`{"name": "test", "value": 123}`, SerializationFormatJson)

  deserializedMap, deserializationErr := StringDeserializer("name: test\nvalue: 123", SerializationFormatYaml)

  deserializedMap, deserializationErr := FileDeserializer("config.json")
  ```

- **FileClerk**: Perform file operations including existence checks, creation, copying, reading content, and symlink handling.

  ```go
  clerk := FileClerk{}

  // ExistenceChecks
  isFileExists := clerk.FileExists("example.txt")
  isRegularFile := clerk.IsFile("example.txt")
  isDirectoryExists := clerk.IsDir("example_dir")
  isSymlink := clerk.IsSymlink("symlink.txt")
  isSymlinkToTarget := clerk.IsSymlinkTo("symlink.txt", "target.txt")

  // FileCreation
  fileCreationErr := clerk.CreateFile("example.txt")

  // FileContentOperations
  maxContentSize := int64(1024)
  fileContent, fileReadingErr := clerk.ReadFileContent("example.txt", &maxContentSize)
  shouldOverwrite := true
  fileUpdateErr := clerk.UpdateFileContent("example.txt", "new content", shouldOverwrite)
  fileContentDeletionErr := clerk.DeleteFileContent("example.txt")
  fileTruncationErr := clerk.TruncateFileContent("example.txt")

  // FileManipulation
  fileCopyErr := clerk.CopyFile("source.txt", "destination.txt")
  fileMoveErr := clerk.MoveFile("old.txt", "new.txt")
  fileRenameErr := clerk.RenameFile("old.txt", "new.txt")
  fileDeletionErr := clerk.DeleteFile("example.txt")

  // FileAdvancedOperations
  fileOwnershipUpdateErr := clerk.UpdateFileOwnership("example.txt", 1000, 1000)
  filePermissions := 0755
  filePermissionsUpdateErr := clerk.UpdateFilePermissions("example.txt", &filePermissions)

  // CompressionOperations
  compressionFormat := "gzip"
  compressedFilePath, compressionErr := clerk.CompressFile("example.txt", &compressionFormat)
  decompressionTargetPath := "decompressed.txt"
  shouldKeepSourceFile := false
  decompressedFilePath, decompressionErr := clerk.DecompressFile(
    "example.txt.tar", &decompressionTargetPath, &shouldKeepSourceFile,
  )

  // DirectoryOperations
  directoryCreationErr := clerk.CreateDir("example_dir")
  directoryCopyErr := clerk.CopyDir("source_dir", "dest_dir")
  directoryMoveErr := clerk.MoveDir("old_dir", "new_dir")
  directoryDeletionErr := clerk.DeleteDir("example_dir")
  directoryCompressionFormat := "brotli"
  directoryCompressionErr := clerk.CompressDir("example_dir", &directoryCompressionFormat)
  directoryDecompressionTargetPath := "decompressed_dir"
  shouldKeepSourceDir := true
  directoryDecompressionErr := clerk.DecompressDir(
    "example_dir.tar", &directoryDecompressionTargetPath, &shouldKeepSourceDir,
  )

  // SymlinkOperations
  shouldOverwriteSymlink := false
  symlinkCreationErr := clerk.CreateSymlink(
    "target.txt", "symlink.txt", shouldOverwriteSymlink,
  )
  symlinkRemovalErr := clerk.RemoveSymlink("symlink.txt")
  ```

- **Shell**: Execute system commands with configurable user, timeout, environment variables, and output redirection to files.

  ```go
  shell := NewShell(ShellSettings{
      Command: "echo",
      Args:    []string{"hello world"},
  })
  commandOutput, executionErr := shell.Run()
  fmt.Println(commandOutput)
  ```

- **Synthesizer**: Generate secure passwords with charset guarantees, random usernames/emails, and self-signed TLS certificates.

  ```go
  synthesizer := &Synthesizer{}

  securePassword := synthesizer.PasswordFactory(12, true)

  randomUsername := synthesizer.UsernameFactory()

  randomEmail := synthesizer.MailAddressFactory(nil)

  commonName, _ := tkValueObject.NewFqdn("goinfinite.net")
  altNames := []tkValueObject.Fqdn{commonName}
  certificatePair, certGenerationErr := synthesizer.SelfSignedCertificatePairFactory(&commonName, altNames)

  certPem, keyPem, certPemGenerationErr := synthesizer.SelfSignedCertificatePairPemFactory(&commonName, altNames)
  ```

- **ServerIpAddress**: Retrieve the server's private and public IP addresses.

  ```go
  privateIpAddress, privateIpReadingErr := ReadServerPrivateIpAddress()

  publicIpAddress, publicIpReadingErr := ReadServerPublicIpAddress()
  ```

- **TrustedIpsReader**: Parse a comma-separated list of trusted IP addresses from the `TRUSTED_IPS` environment variable.

  ```go
  trustedIpAddresses, trustedIpsReadingErr := TrustedIpsReader()
  ```

- **ReadThrough**: Read-through utilities for TLS certificate pairs from `CERTIFICATE_PAIR_CERT_PATH` and `CERTIFICATE_PAIR_KEY_PATH` env vars, generating self-signed certificates in `PKI_DIR` if not provided.

  ```go
  readThrough := &ReadThrough{}

  certificateFilePath, keyFilePath, certPairReadingErr := readThrough.CertPairFilePathsReader()
  ```

- **LogHandler**: Configure logging levels via `LOG_LEVEL` environment variable and initialize structured logging with slog and Zerolog.
- **PaginationQueryBuilder**: Build paginated database queries with support for page number, items per page, last seen ID, sorting, and total count.

  ```go
  databaseQuery := db.Model(&YourModel{})

  requestPagination := tkDto.Pagination{
      PageNumber:   0,
      ItemsPerPage: 10,
  }

  paginatedQuery, responsePagination, paginationBuildingErr := PaginationQueryBuilder(databaseQuery, requestPagination)

  modelRecords := []YourModel{}
  queryExecutionErr := paginatedQuery.Find(&modelRecords).Error
  ```

- **TrailDatabaseService**: Initialize and migrate a SQLite trail database for activity records using GORM, configurable via `TRAIL_DATABASE_FILE_PATH` environment variable.

  ```go
  os.Setenv("TRAIL_DATABASE_FILE_PATH", "/path/to/trail.db")

  trailDatabaseService, serviceInitializationErr := NewTrailDatabaseService([]any{&YourAdditionalModel{}})

  activityRecords := []ActivityRecord{}
  trailDatabaseService.Handler.Model(&ActivityRecord{}).Find(&activityRecords)
  ```

### Presentation Middlewares

For web applications built with Echo:

- **PanicHandler**: Handle panics in HTTP requests, log filtered stack traces (excluding domain layers), and respond with error messages; uses `TRUSTED_IPS` env var to mask sensitive information. It can be used with CLI applications as well.
- **RequiredParamsInspector**: Normalizes parameters from HTTP requests into a `map[string]any` regardless of the request type (JSON, form data, multipart files, path or query params).
- **ApiRequestInputReader**: Read and parse JSON, form data, or multipart files from Echo HTTP requests into structured data.

### Presentation Helpers

- **EnvsInspector**: Inspect and validate environment variables from .env files loaded via `ENV_FILE_PATH` env var, supporting required and auto-fillable variables.
- **PaginationParser**: Parse pagination parameters like pageNumber, itemsPerPage, lastSeenId, sortBy, and sortDirection from HTTP requests.
- **StringSliceVoParser**: Convert comma-separated, semicolon-separated, or array strings into value object slices.
- **TimeParamsParser**: Parse date ranges, timestamps, and relative times from request parameters.
- **ResponseWrapper**: A wrapper struct for liaison responses and when needed, used to emit API and CLI responses.

### Value Objects

The library offers a diverse range of value objects (VO) to represent domain entities. Each VO is designed to guarantee type safety and provide validation. Examples include Email, Password, URL, IPAddress, UnixFilePath, HttpMethod, CountryCode, CurrencyCode, SystemResourceIdentifier, and so on. These components are thoroughly tested, ensuring 100% coverage.

### Value Object Utilities

- **InterfaceTo**: Safely convert interface{} to primitive types (bool, string, int, float, etc.) using reflection, handling various input formats with error checking.

### DTOs

- **Pagination**: General pagination DTO with page number, items per page, last seen ID, sort by, and sort direction.

### Activity Record Management

Infinite Toolkit _(TK)_ provides a comprehensive activity record management system for auditing and logging user actions, following Clean Architecture principles:

#### Infrastructure

- **ActivityRecordCmdRepo**: Command repository for creating and deleting activity records in the trail database.
- **ActivityRecordQueryRepo**: Query repository for reading activity records from the trail database with pagination support.

#### Domain

##### Entities

- **ActivityRecord**: Represents an activity record with record ID, level, code, affected resources, details, operator account ID, IP address, and creation time.

##### Use Cases

- **CreateActivityRecord**: Persists an activity record as a non-blocking side effect, logging errors without failing the primary operation.
- **DeleteActivityRecord**: Deletes an activity record by ID.
- **ReadActivityRecords**: Retrieves activity records with pagination and filtering options.

##### DTOs

- **CreateActivityRecord**: Data transfer object for creating activity records.
- **DeleteActivityRecord**: Data transfer object for deleting activity records.
- **ReadActivityRecords**: Data transfer object for reading activity records with pagination.

##### Repositories

- **ActivityRecordCmdRepo**: Interface for command operations (create, delete) on activity records.
- **ActivityRecordQueryRepo**: Interface for query operations (read) on activity records.
