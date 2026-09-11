# Infrastructure

Infrastructure layer of Infinite Toolkit _(TK)_. It implements I/O: file, shell, network, crypto, and database helpers, plus the repository implementations. It depends on the domain layer and never on presentation. Part of [Infinite Toolkit](../../README.md).

## Helpers

- **Deserializer**: Deserialize JSON and YAML files into maps for configuration handling.

  ```go
  deserializedMap, deserializationErr := StringDeserializer(
    `{"name": "test", "value": 123}`, SerializationFormatJson,
  )

  deserializedMap, deserializationErr := StringDeserializer(
    "name: test\nvalue: 123", SerializationFormatYaml,
  )

  deserializedMap, deserializationErr := FileDeserializer("config.json")
  ```

- **FileClerk**: Perform file operations including existence checks, creation, copying, reading content, regex search and replace, atomic overwrite-rename, collision-safe temp file naming, and symlink handling.

  ```go
  clerk := FileClerk{}

  // ExistenceChecks
  isFileExists := clerk.FileExists("example.txt")
  isRegularFile := clerk.IsFile("example.txt")
  isDirectoryExists := clerk.IsDir("example_dir")
  isSymlink := clerk.IsSymlink("symlink.txt")
  isSymlinkToTarget := clerk.IsSymlinkTo("symlink.txt", "target.txt")

  // FileCreation
  fileCreationErr := clerk.TouchFile("example.txt")

  // FileContentOperations
  maxContentSize := int64(1024)
  fileContent, fileReadingErr := clerk.ReadFileContent("example.txt", &maxContentSize)
  regexSearchFilePath, regexSearchFilePathErr := tkValueObject.NewUnixAbsoluteFilePath("example.txt", false)
  regexPattern := regexp.MustCompile(`(?m)^error: (.+)$`)
  regexSearchFindings, regexSearchErr := clerk.FileContentRegexSearch(regexSearchFilePath, regexPattern)
  regexReplacement := `warn: $1`
  replacementCount, regexReplaceErr := clerk.FileContentRegexReplace(
    regexSearchFilePath, regexPattern, regexReplacement,
  )
  fileAppendErr := clerk.AppendFileContent(regexSearchFilePath, "new content")
  fileTruncationErr := clerk.TruncateFileContent(regexSearchFilePath)

  // FileManipulation
  fileCopyErr := clerk.CopyFile("source.txt", "destination.txt")
  fileMoveErr := clerk.MoveFile("old.txt", "new.txt")
  fileOverwriteErr := clerk.OverwriteFile("source.tmp", "destination.txt")
  fileDeletionErr := clerk.DeleteFile("example.txt")

  // FileAdvancedOperations
  fileOwnershipUpdateErr := clerk.UpdateFileOwnership("example.txt", 1000, 1000)
  filePermissions := 0755
  filePermissionsUpdateErr := clerk.UpdateFilePermissions("example.txt", &filePermissions)
  ownerUsername, ownerUsernameErr := tkValueObject.NewUnixUsername("example")
  systemdUserDir, systemdUserDirErr := tkValueObject.NewUnixAbsoluteFilePath(
    "/home/example/.config/systemd/user", false,
  )
  dirPathRedirectSafetyErr := clerk.VerifyDirPathRedirectSafety(
    systemdUserDir, &ownerUsername, nil,
  )
  serviceFilePath, serviceFilePathErr := tkValueObject.NewUnixAbsoluteFilePath(
    "/home/example/.config/systemd/user/service.service", false,
  )
  serviceFilePermissions := os.FileMode(0644)
  overwritePolicy := FileClerkOverwritePolicyReplace
  fileUpsertErr := clerk.UpsertFile(FileUpsertSettings{
    FilePath:                serviceFilePath,
    OverwritePolicy:         &overwritePolicy,
    TrustedDirOwnerUsername: &ownerUsername,
    Permissions:             &serviceFilePermissions,
    OwnerUsername:           &ownerUsername,
  }, []byte("unit content"))

  // CompressionOperations
  compressionFormat := "gzip"
  compressedFilePath, compressionErr := clerk.CompressFile(
    "example.txt", &compressionFormat,
  )
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
  directoryCompressionErr := clerk.CompressDir(
    "example_dir", &directoryCompressionFormat,
  )
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

  **UpsertFile Notes**

  `UpsertFile` writes through a held parent-directory handle and defaults to the safe policies: it refuses symlinked paths, refuses to replace an existing target, creates new files as `FileClerkDefaultNewFileMode` (0600), and inherits the target's mode and owner on replace. Callers opt in with `FileClerkSymlinkPolicyResolve`, `FileClerkOverwritePolicyReplace`, and a `FileClerkOwnerSource`; `TrustedDirOwner*` gates the directory chain. A stated owner wins over the existing-file source and conflicts with the other sources (`ErrOwnerSourceConflict`); a stated group or mode wins, and an omitted mode inherits the target or defaults to 0600 on create. A directory target fails with `ErrTargetIsDirectory`, and a process that cannot set the resolved owner fails with `ErrFileOwnerChangeFailed`.

- **Shell**: Execute system commands with configurable user, timeout, environment variables, and output redirection to files.

  ```go
  shell := NewShell(ShellSettings{
      Command: "echo",
      Args:    []string{"hello world"},
  })
  commandOutput, executionErr := shell.Run()
  fmt.Println(commandOutput)
  ```

- **Synthesizer**: Generate cryptographically secure random integers, passwords with charset guarantees, filler usernames/emails, private keys, and TLS certificates (including CA certificates).

  ```go
  synthesizer := &Synthesizer{}

  randomInteger := synthesizer.RandomIntegerGenerator(1, 100)

  password := synthesizer.PasswordFactory(16, true)

  randomUsername := synthesizer.UsernameFactory()
  randomEmail := synthesizer.MailAddressFactory(nil)

  rsaKeyPem, rsaErr := synthesizer.PrivateKeyPemFactory(PrivateKeySettings{
    Algorithm: tkValueObject.PrivateKeyAlgorithmRSA,
    BitSize:   2048,
  })

  ecdsaKeyPem, ecdsaErr := synthesizer.PrivateKeyPemFactory(PrivateKeySettings{
    Algorithm: tkValueObject.PrivateKeyAlgorithmECDSA,
    BitSize:   256, // Supports 256, 384, 521
  })

  commonName, _ := tkValueObject.NewFqdn("goinfinite.net")
  aliasName, _ := tkValueObject.NewFqdn("goinfinite.com.br")
  altNames := []tkValueObject.Fqdn{aliasName}

  certPair, certGenErr := synthesizer.SelfSignedCertificatePairFactory(
    &commonName, altNames,
  )

  certPem, keyPem, certPemGenErr := synthesizer.SelfSignedCertificatePairPemFactory(
    &commonName, altNames,
  )

  certPem, keyPem, certErr := synthesizer.CertificatePemFactory(CertificateSettings{
    CommonName: &commonName,
    AltNames:   altNames,
  })

  maxPathLen := 2
  caCertPem, caKeyPem, caErr := synthesizer.CACertificatePemFactory(CertificateSettings{
    CommonName:       &commonName,
    MaxPathLengthPtr: &maxPathLen,
  })
  ```

- **ServerIpAddress**: Retrieve the server's private and public IP addresses. The public IP read honors `SERVER_PUBLIC_IP_ADDR` as the primary source of truth before any HTTP lookup.

  ```go
  privateIpAddress, privateIpReadingErr := ReadServerPrivateIpAddress()

  publicIpAddress, publicIpReadingErr := ReadServerPublicIpAddress()
  ```

- **DnsLookup**: Perform DNS queries for various record types using custom resolvers with fallback support. Setting `ShouldBypassLocalResolver: true` skips Go's net.Resolver and issues raw dnsmessage UDP queries to bypass `/etc/hosts` and `resolv.conf` (applies to A/AAAA only).

  ```go
  hostname, _ := tkValueObject.NewUnixHostname("example.com")
  dnsLookup := NewDnsLookup(DnsLookupSettings{
      ShouldBypassLocalResolver: true,
  })

  dnsRecords, lookupErr := dnsLookup.Execute(
      hostname, &tkValueObject.DnsRecordTypeA,
  )
  ```

- **TrustedCidrsReader**: Parse comma-separated trusted entries from both `TRUSTED_IPS` and `TRUSTED_CIDRS` environment variables. Accepts plain IP addresses (converted to `/32` for IPv4 and `/128` for IPv6) and CIDR notation in either variable. Invalid entries are logged and skipped. Returns `[]CidrBlock`.

  ```go
  trustedCidrBlocks, trustedCidrsReadingErr := TrustedCidrsReader()
  ```

- **ReadThrough**: Read-through utilities for TLS certificate pairs from `CERTIFICATE_PAIR_CERT_PATH` and `CERTIFICATE_PAIR_KEY_PATH` env vars, generating self-signed certificates in `PKI_DIR` if not provided.

  ```go
  readThrough := &ReadThrough{}

  certFilePath, keyFilePath, certPairReadingErr := readThrough.CertPairFilePathsReader()
  ```

- **Cypher**: Encrypt and decrypt strings using AES-GCM for authenticated encryption with base64 encoding.

  ```go
  encodedSecretKey, keyGenerationErr := NewCypherSecretKey()

  cypher, cypherCreationErr := NewCypher(encodedSecretKey)

  encryptedText, encryptionErr := cypher.Encrypt("plain text")

  decryptedText, decryptionErr := cypher.Decrypt(encryptedText)
  ```

- **PaginationQueryBuilder**: Build paginated database queries with support for page number, items per page, last seen ID, sorting, and total count.

  ```go
  databaseQuery := db.Model(&YourModel{})

  requestPagination := tkDto.Pagination{
      PageNumber:   0,
      ItemsPerPage: 10,
  }

  paginatedQuery, responsePagination, paginationBuildingErr := PaginationQueryBuilder(
    databaseQuery, requestPagination, "",
  )

  modelRecords := []YourModel{}
  queryExecutionErr := paginatedQuery.Find(&modelRecords).Error
  ```

- **TrailDatabaseService**: Initialize and migrate a SQLite trail database for activity records using GORM, configurable via `TRAIL_DATABASE_FILE_PATH` environment variable.

  ```go
  os.Setenv("TRAIL_DATABASE_FILE_PATH", "/path/to/trail.db")

  trailDatabaseService, serviceInitializationErr := NewTrailDatabaseService(
    []any{&YourAdditionalModel{}},
  )

  activityRecords := []ActivityRecord{}
  trailDatabaseService.Handler.Model(&ActivityRecord{}).Find(&activityRecords)
  ```

- **TransientDatabaseService**: Initialize a shared in-memory SQLite key-value store with `Set`, `Read`, and `Has`. Every instance in the process shares the same data, which vanishes when the process ends; `Read` returns `ErrKeyNotFound` for a missing key.

  ```go
  transientDatabaseService, serviceInitializationErr := NewTransientDatabaseService()

  setErr := transientDatabaseService.Set("key", "value")

  value, readErr := transientDatabaseService.Read("key")

  keyExists := transientDatabaseService.Has("key")
  ```

## Repositories

- **activityRecord/**: GORM implementations of the domain `ActivityRecordCmdRepo` and `ActivityRecordQueryRepo` interfaces.
