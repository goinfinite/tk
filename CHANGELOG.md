# Changelog

```log
0.3.4 - unreleased
feat: add TransientDatabaseService, a shared in-memory SQLite key-value store. NewTransientDatabaseService opens file::memory:?cache=shared and migrates the KeyValue model. Has counts a key and logs a failed count, Read returns its value or ErrKeyNotFound, and Set upserts with ON CONFLICT DO UPDATE. Every instance in the process shares the same data.
feat: add FileClerk.VerifyDirPathRedirectSafety(dirPath, ownerUsernamePtr, ownerUserIdPtr). It walks each directory step without following symlinks. The first problem fails the walk: '..' (ErrDirPathTraversalInvalid), a symlink (ErrSymlinkedPathInvalid), a non-directory (ErrTargetNotDirectory), a foreign owner (ErrDirectoryOwnerInvalid), an unknown account (OwnerLookupFailed), or an uninspectable step (PathCheckFailed). It skips dot and empty components. A numeric UID is compared directly without an account lookup; with no owner given, it expects the process account.
feat: add FileClerk.UpsertFile(settings, content). It holds the verified parent-directory handle open, then creates and swaps the temp file via openat/renameat2. No component is re-resolved after the check, so a swapped path cannot redirect the write. A taken target fails with ErrTargetFileExists unless ShouldOverwrite is set. A symlinked final component is replaced, never followed, unless ShouldFollowSymlinks is set. A symlinked parent chain fails with ErrSymlinkedPathInvalid unless that flag is set. Unset Permissions fails with ErrFilePermissionsInvalid; an invalid final name with ErrFileNameInvalid; an overlong temp name with ErrTempFileNameTooLong. With an owner named, chown runs before chmod. The regex replacers reuse writeFileAtomically but follow symlinks and skip the chain check.
feat: add FileClerk.TempFileNameFactory(targetFileName) and FileClerk.TempFilePathFactory(targetFilePath). They build the hidden `.name.<entropy>.tk-tmp` staging name and path. Both take value objects, and TempFilePathFactory returns (string, error).
refactor: remove FileClerk.RenameFile and FileClerk.DeleteFileContent, one-line aliases of MoveFile and TruncateFileContent.
refactor: replace FileClerk.UpdateFileContent with AppendFileContent(UnixAbsoluteFilePath, content), which only appends. TruncateFileContent now takes a UnixAbsoluteFilePath and calls os.Truncate instead of writing empty content.
refactor: FileClerk.UpsertFile owner is optional. Without one, the chain check expects the process account and chown is skipped, so the file keeps the creating process's ownership. ErrDirectoryOwnerMissing removed.
refactor: PasswordFactory now draws with crypto/rand. Each requested character class occupies its own position, so one guarantee never erases another. CharsetPresenceGuarantor, UsernameFactory, and MailAddressFactory draw with crypto/rand too, so synthesizer.go no longer uses math/rand. A failed crypto/rand draw panics instead of emitting a predictable secret.
refactor: PublicIpAddressResolver orders resolver endpoints with crypto/rand instead of math/rand.Perm.
feat: add Synthesizer.RandomIntegerGenerator(lowestValue, highestValue), which draws a uniform crypto/rand integer between two inclusive bounds. An inverted range panics with RandomIntegerRangeInvalid.
feat: add UserId (*uint32) to ShellSettings. The command runs as that account, UserId 0 included, and Username wins when both are set. An unresolvable target fails unless ShouldIgnoreUsernameLookupError is set; an ignored failure logs ShellTargetAccountUnresolvedRunningAsCurrentAccount. Add ShouldDisableTimeout to remove the execution deadline entirely for runs like day-long backups.
fix: PaginationQueryBuilder resolves ORDER BY fields through the parsed GORM statement schema, so CamelCase input maps to the model's real column. Callers MUST set .Model(...); queries built on .Table(...) alone fail with ParseStatementSchemaError. Unknown fields fail with UnknownPaginationSortFieldError instead of emitting invalid SQL.
fix: PaginationQueryBuilder defaults the last-seen-id cursor column to the schema's primary key instead of the literal id. A model without a primary key fails with ModelWithoutPrimaryKeyError. Both the cursor and ORDER BY columns are quoted through the statement, so neither can inject SQL.
chore: drop github.com/iancoleman/strcase; sort-field mapping now uses GORM's naming strategy.
fix: UnixAbsoluteFilePath and UnixRelativeFilePath no longer treat a dotfile's leading dot as an extension separator; '.hidden' is preserved and '.config.json' yields '.config'.
refactor: extract the page-count rule to PaginationPagesTotalResolver(itemsTotal, itemsPerPage). A partial page counts as a page. Zero page size fails with ErrItemsPerPageCannotBeZero; a page count beyond uint32 fails with ErrPagesTotalOverflow, both checked only in the resolver.
fix: EnvsInspector separates an auto-filled variable from a .env file that lacks a trailing newline.
test: cover UpsertFile, AppendFileContent, VerifyDirPathRedirectSafety, the temp-name factories, PaginationPagesTotalResolver, PasswordFactory class guarantees, schema-resolved sorting, custom primary key cursors, and the ShouldDisableTimeout and UserId shell settings.
test: modernize trailDatabaseService_test.go. t.Setenv replaces the manual env save/restore and clears the errcheck findings; critical failures use t.Fatalf; extra-model migration is verified through NewTrailDatabaseService and the created table, not the private dbMigrate.

0.3.3 - 2026/09/03
fix: CLI logs no longer corrupt the machine-readable stdout channel. All logs go to stderr, so stdout carries only the JSON response in every session mode, LOG_LEVEL=debug included. LOG_LEVEL matching is case-insensitive.
fix: CLI response renderer writes its error diagnostics (ResponseEncodingError, SyntaxHighlighting*Error) to stderr.
fix: LogHandler.SetLevel logs a failed LOG_LEVEL env write instead of dropping it silently.
refactor: extract the stdout-terminal check into tkInfra.IsStdoutTerminal; the log handler and CLI response renderer share it.
fix: Shell enforces ExecutionTimeoutSecs with a Go context deadline (SIGTERM on expiry) instead of the timeout binary. A command that exits 124 on its own is no longer misreported as CommandDeadlineExceeded; a real timeout still exits 124 with that error.
refactor: extract the timeout policy to exported constants ShellExecutionTimeoutDefaultSecs, ShellExecutionTimeoutHardLimitSecs, ShellExecutionTimeoutGraceSecs, and ShellCommandTimeoutExitCode, applied by Shell.executionTimeoutResolver. Children that ignore SIGTERM get SIGKILL after the grace via Cmd.WaitDelay.
fix: Shell.Run closes its stdout/stderr capture files. executionPlanner never populated the executionPlan file-handler fields, so handles leaked on every run with StdoutFilePath/StderrFilePath. A failed stderr capture now releases the already-opened stdout capture.
feat: add ShouldUseCleanEnv to ShellSettings. The child then runs from a minimal environment: PATH of the resolved user's ~/.local/bin plus the standard system directories, HOME (parent's for same-user runs, target's on a Username switch), PWD when a working directory is set, DEBIAN_FRONTEND, and explicit Envs. Parent variables cannot leak. It works with or without a Username switch.
feat: add FileClerk.WriteNewFile. It creates a file exclusively with the given content and exact permissions (umask cannot shrink them) and removes the half-written file on failure.
fix: FileClerk.FileExists reports false when the check itself fails (EACCES/EIO no longer mean "exists").
fix: FileClerk.ReadFileContent returns ErrFileTooLarge instead of truncated content when the file exceeds the size cap.
fix: rename FileClerk.CreateFile to TouchFile with the touch(1) contract: refresh timestamps on an existing path, exclusively create a missing one. A taken path is never truncated. Public API rename; ez call sites must migrate.
fix: FileClerk.CopyFile preserves the source mode (exec bit survives) and creates the target exclusively, so a target symlink is no longer written through.
fix: FileClerk.MoveFile relocates via renameat2(2) with RENAME_NOREPLACE, so the target guard and the move are one atomic operation and a target cannot be silently replaced. Cross-device moves fall back to copy+delete (golang.org/x/sys is now a direct dependency).
fix: FileClerk.DeleteFile/DeleteDir reject wrong-type targets (ErrTargetIsDirectory / ErrTargetNotDirectory) instead of no-op'ing. DeleteFile now removes symlinks.
fix: FileClerk.UpdateFilePermissions rejects symlinks (ErrTargetIsSymlink), so chmod never lands on the link's target. The guard is now the open itself (O_NOFOLLOW handle, fchmod), so non-root callers need read permission on the target (root is exempt). The permissions parameter is now *os.FileMode instead of *int.
refactor: FileClerk.CompressFile gains shouldKeepSourceFilePtr, mirroring DecompressFile. Every format now shares one source-deletion contract: the external tool keeps its input, and FileClerk removes it unless keeping is requested. Previously tar kept the source while gzip/xz/zip/brotli deleted it. This also fixes CompressDir returning ENOENT on success.
test: add TestWriteNewFile, TestCompressDir, and regression coverage for exclusive create/copy/move, source-mode preservation, symlink guards, ErrFileTooLarge, and wrong-type delete rejections.
feat: add LiaisonResponseStatus accepted (202/exit 0), conflict (409/exit 65), and serviceUnavailable (503/exit 69); both renderers map them.
test: add TestLogLevelParser, TestIsStdoutTerminal, TestChildEnvironment (inherited+PWD, Username-alone-inherits, clean-env PATH/HOME/PWD policy, Envs-win end-to-end), StdoutFileCapture, CaptureFileHandlesDoNotLeak (fd-count regression), and timeout (CommandTimeoutEnforced, NaturalExitCode124KeepsCommandStdErr) coverage. Both renderer tests cover the three new statuses.
chore: gofmt paginationQueryBuilder_test.go.
chore: update Go to 1.27.1 and refresh dependencies. net/mail now accepts bracketed IPv6 domains, so MailAddress does too.

0.3.2 - 2026/07/30
fix: ReadFileContent accepts symlinks; os.Open follows the chain. The IsFile guard is replaced with os.IsNotExist mapping.
fix: lower the ReadFileContent default cap from 1GiB to 500MiB to bound in-memory string allocation.
docs: document the 500MiB ReadFileContent cap and recommend streaming for larger files.
refactor: extract the ReadFileContent default cap to the ReadFileContentDefaultMaxSizeBytes constant.
fix: correct 10MB to 10MiB in the regexSearchWholeFile comment.
test: add ReadSymlinkToFile and ReadDanglingSymlink coverage to TestReadFileContent.

0.3.1 - 2026/07/30
feat: add ShouldBypassLocalResolver to DnsLookupSettings. It bypasses /etc/hosts by sending a raw dnsmessage UDP query for A/AAAA and parsing the response in-process.
feat: add directIpAddressResolver. It builds DNS messages with golang.org/x/net/dns/dnsmessage and parses A/AAAA records.
refactor: rename resolverFactory to netResolverBuilder and queryDnsRecords to defaultDnsRecordsResolver. The new dnsRecordsResolver routes IP record types to the direct path and the rest to the system resolver.
chore: promote golang.org/x/net from indirect to direct dependency.
test: add directIpAddressResolver A/AAAA tests, a ShouldBypassLocalResolver construction test, and bypass-aware Execute tests.
test: refactor dnsLookup_test into three table-driven funcs. Bypass is verified via localhost: the local resolver returns 127.0.0.1, the raw 8.8.8.8 query does not.
fix: randomize the DNS transaction id with crypto/rand. The direct resolver path now surfaces RCODE, truncation, and id-mismatch failures as ErrDnsLookupResponse* sentinels instead of silent empty results with a static id.
refactor: split dnsResponseIpAddressesExtractor into dnsMessageValidator (parse and header validation) and dnsMessageIpAddrExtractor (record extraction). Rename dnsMessagePackBuilder to dnsMessagePacker. Drop a dead assignment in defaultDnsRecordsResolver.
chore: extract the dnsStandardPort constant, replacing two hardcoded dial ports.
test: LocalhostBypassedSkipsLocalLookup accepts DnsLookupResponseNameError as evidence that bypass succeeded.

0.3.0 - 2026/07/27
feat: add FileClerk.FileContentRegexSearch and FileContentRegexReplace. They route by size at 10MiB, replace atomically via temp file plus rename, follow symlinks, and guard against an empty result.
feat: add FileClerk.OverwriteFile, an atomic source-over-target rename that resolves symlink chains.
refactor: replace the RegexPattern VO with native *regexp.Regexp and return []FileContentRegexFindings with a 1-based inclusive LineNumRange.
refactor: tighten the FileContentRegex error contract. ErrSourceIsDirectory and ErrTargetIsDirectory are now distinct. ErrReplacementWouldTruncateFile comes from a tempfile stat. The streaming-fallback log lives at the dispatch site. DeleteFileContent delegates to TruncateFileContent.
docs: align .context.md, README.md, and FEATURE-MAP with the 0.3.0 surface; add a Human Reviewed callout to the README.
fix: regex search derives match offsets from the engine; replace preserves terminators and mode.
fix: capture deferred Close errors in CopyFile and UpdateFileContent.
fix: close file handles before reporting success in write-bearing operations.
fix: resolve the symlink target before creating the regexReplace temp file (cross-device rename).
docs: explain why regexSearchWholeFile keeps two regex passes.

0.2.9 - 2026/07/20
feat: add FileContentRegexSearch to FileClerk for streaming line-by-line regex search
feat: add RegexPattern value object for validated regex compilation
feat: honor SERVER_PUBLIC_IP_ADDR env var as primary source of truth in ReadServerPublicIpAddress
fix: bound public IP resolver requests with cancellable contexts
refactor: extract FileClerk error sentinels to package-level vars
refactor: replace ReadServerPublicIpAddress DNS+curl flow with native HTTP fan-out across multiple public IP resolvers
refactor: split DnsLookup config into DnsLookupSettings; move hostname and recordType to Execute()
docs: mention FileContentRegexSearch and SERVER_PUBLIC_IP_ADDR env var in README

0.2.8 - 2026/06/30
feat: add agent skills system with OpenAPI testing skill
docs: add skills for agents section to README
chore: add agent tool configs to gitignore

0.2.7 - 2026/04/08
feat: add TrustedCidrsReader for TRUSTED_CIDRS env var
feat: add RequesterIpExtractor with XFF-aware IP extraction
feat: add HttpHeader value object with validation
feat: add IsLinkLocal to IpAddress and Contains to CidrBlock
refactor: replace ExtractIPDirect with RequesterIpExtractor at all call sites
refactor: simplify RequesterIpExtractor consumers for (IpAddress, error) return
refactor: drop echo from RequesterIpExtractor, unified header extraction with Direct/RemoteAddr keywords
fix: replace negation logic in CidrBlock and IpAddress with explicit checks (CWE-20, CWE-284)
fix: fail-closed on invalid IP_EXTRACT_DISABLE_TRUST and trim whitespace in CIDR parsing
fix: add loopback check to CidrBlock.IsPublic, fail-closed IpExtractHeaderReader, Header.Values for multi-header XFF, echo.NewHTTPError wrapping
refactor: move RequesterIpExtractor to presentation, unify TrustedIpsReader into TrustedCidrsReader

0.2.6 - 2026/03/24
fix: switch UnixAbsoluteFilePath regex from allowlist to blacklist with security test suite
fix: switch UnixFileName regex from allowlist to blacklist with tilde and bracket support
fix: add tilde expansion pattern detection to UnixRelativeFilePath

0.2.5 - 2026/03/13
fix: replace PasswordFactory with NewCypherSecretKey in EnvsInspector

0.2.4 - 2026/03/12
fix: add missing os.Exit on interactive terminal path in LiaisonCliResponseRenderer
refactor: replace binary exit codes with BSD sysexits.h conventions in CLI renderer

0.2.3 - 2026/03/09
feat: add SimpleCliResponseRenderer for simplified CLI error output
fix: enforce required first character in UnixFileName regex patterns
fix: allow glob wildcard in UnixFileName VO and forbid consecutive asterisks

0.2.2 - 2026/03/09
refactor: add type assertion short-circuit to VO constructors

0.2.1 - 2026/03/09
refactor: remove hardcoded "id" column assumption from PaginationQueryBuilder
fix: complete IsLocal() loopback detection for IPv4 range and IPv6
fix: strip IPv6 zone ID suffix in IpAddress value object
fix: replace RealIP() with echo.ExtractIPDirect() in presentation layer
fix: extract operatorAccountId from echo context in request input reader
chore: remove all agents related files
chore: update dependencies
fix: export Shell error const

0.2.0 - 2026/01/09
feat: x509 certificate value objects and entity
feat: add PrivateKeyPemFactory to Synthesizer (supports RSA, ECDSA, DSA, Ed25519)
feat: add CertificatePemFactory to Synthesizer
feat: add CACertificatePemFactory to Synthesizer
fix: remove all hardcoded private keys/certificates from test files

0.1.9 - 2026/01/07
feat: add NamedGroupsExtractor to tkVoUtil pkg
feat: allow different schemes on url value object
docs: improve agent workflow documentation

0.1.8 - 2025/12/30
feat: add pure go dns lookup
feat: add user agent and dns record type value objects
refactor: replace dig with DnsLookup and curl fallback (serverIpAddress)

0.1.7 - 2025/11/21
feat: add cypher infra helper
fix: add missing suffix on record code regex

0.1.6 - 2025/11/17
feat: add system resource type and id vo
feat: add more liaison response statuses
fix!: replace accountId with sri on request input reader

0.1.5 - 2025/11/14
feat: add component reader to sri vo
fix!: replace accountId with sri on activity record feat
docs: improve line breaks on code snippets

0.1.4 - 2025/11/13
test: add CreatedAfterAt tests for activity record query repo
test: add panic handler middleware
fix: prevent regen of existent cert pair
fix: force gorm to use UTC
fix: use UTC on UnixTime vo alt constructors
fix: move recover() to defer func in panic handlers
docs: add examples for all components

0.1.3 - 2025/11/10
fix: fix pagination query builder order statement declaration bug
fix: missing gofmt run on a few VO unit tests
fix: init query repo on activity record cmd construct
docs: add activity record mgmt and enhance existing docs
chore: upgrade deps

0.1.2 - 2025/11/07
feat: add trail database service
feat: add activity record entity, vos and models
feat: add activity record use cases, repositories and implementations
feat: add system resource identifier vo
feat: add account id vo
feat: add weak password
fix!: set message as last field in responseWrappers
fix!: add api prefix to request input reader
fix: run mod tidy after first test sample on responseWrappers_test.go

0.1.1 - 2025/11/06
feat: add request input reader
feat: add trusted ip reader
feat: add pagination parser
feat: add time params parser
feat: add response wrappers
fix: turn last seen id vo regex stricter

0.1.0 - 2025/11/05
feat: add SelfSignedCertificatePairFactory to Synthesizer
feat: add CertPairFilePathsReader to ReadThrough
feat: add envsInspector presentation helper
fix: decompress using source dir as working dir
fix: keep only utf8 chars on StripUnsafe

0.0.9 - 2025/11/03
feat: split unix file path into relative and absolute vos
feat: add panic handler middleware
feat: add log handler middleware
chore: add echo as dependency
chore: add zerolog as dependency

0.0.8 - 2025/10/31
feat: import, refactor and create unit tests for common vos from OS/Ez/Bz projects
fix: move regex must compile to pkg level

0.0.7 - 2025/06/17
feat: add FileClerk
feat: add CompressionFormat vo
fix: add stdout and stderr file handlers to shell

0.0.6 - 2025/06/11
feat: add Shell
feat: add ReadServerPublic/PrivateIpAddresses
feat: add IsBetween() for UnixTime

0.0.5 - 2025/06/02
chore: remove RequestInputParser
fix: StringSliceValueObjectParser nil and empty string check

0.0.4 - 2025/06/01
feat: add UnixTime vo
feat: add RequiredParamsInspector

0.0.3 - 2025/05/31
feat: add RequestInputParser

0.0.2 - 2025/05/16
feat: add deserializer

0.0.1 - 2025/05/11
feat: initial release
```
