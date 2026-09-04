# Changelog

```log
0.3.3 - 2026/09/03
fix: CLI logs no longer corrupt the machine-readable stdout channel; all logs go to stderr — interactive debug keeps its console formatting — so stdout carries only the JSON response in every session mode (LOG_LEVEL=debug included); LOG_LEVEL matching is now case-insensitive
fix: CLI response renderer writes its error diagnostics (ResponseEncodingError, SyntaxHighlighting*Error) to stderr instead of stdout
fix: LogHandler.SetLevel logs a failed LOG_LEVEL env write instead of dropping it silently
refactor: extract the stdout-terminal check into tkInfra.IsStdoutTerminal and use it from both the log handler and the CLI response renderer so they cannot drift
fix: Shell enforces ExecutionTimeoutSecs via a Go context deadline (SIGTERM on expiry) instead of the external timeout binary; a command that exits 124 on its own is no longer misreported as CommandDeadlineExceeded (a real timeout still surfaces as exit 124 with CommandDeadlineExceeded)
refactor: extract the timeout policy to exported constants ShellExecutionTimeoutDefaultSecs, ShellExecutionTimeoutHardLimitSecs, ShellExecutionTimeoutGraceSecs, and ShellCommandTimeoutExitCode, applied by Shell.executionTimeoutResolver; SIGTERM children that ignore the signal are escalated to SIGKILL after the grace via Cmd.WaitDelay
fix: Shell.Run closes its stdout/stderr capture files — executionPlanner never populated the executionPlan file-handler fields, so the handles leaked on every run with StdoutFilePath/StderrFilePath; a failed stderr capture creation now also releases the already-opened stdout capture file
feat: add ShouldUseCleanEnv to ShellSettings; when set, the child runs from a minimal environment — PATH of ~/.local/bin (the resolved user's own, target's on Username switch) plus the standard system directories, HOME (parent's for same-user runs, target user's home on Username switch), PWD when a working directory is set, DEBIAN_FRONTEND, and explicit Envs — so the parent's other variables cannot leak; independent of Username, so it also covers running as the same user with a clean environment
feat: add WriteNewFile to FileClerk — creates a file exclusively with the given content and exact permissions (umask cannot shrink them) and removes the half-written file on failure
fix: FileClerk.FileExists reports false when the check itself fails (EACCES/EIO no longer mean "exists")
fix: FileClerk.ReadFileContent returns ErrFileTooLarge instead of silently handing back truncated content when the file exceeds the size cap
fix: FileClerk.CreateFile becomes TouchFile with the touch(1) contract — an existing path gets its timestamps refreshed, a missing one gets an exclusively-created empty file — so callers can drop their FileExists pre-check and a taken path is never truncated (public API rename: ez call sites must migrate)
fix: FileClerk.CopyFile preserves the source file's mode (exec bit survives) and creates the target exclusively, so a symlink at the target is no longer written through
fix: FileClerk.MoveFile relocates via renameat2(2) with RENAME_NOREPLACE — the target guard and the move are one atomic kernel operation, so a target can no longer be silently replaced; cross-device moves now work via mv(1)-style copy+delete (requires golang.org/x/sys, now a direct dependency)
fix: FileClerk.DeleteFile/DeleteDir reject wrong-type targets (ErrTargetIsDirectory / ErrTargetNotDirectory) instead of no-op'ing; DeleteFile now removes symlinks
fix: FileClerk.UpdateFilePermissions rejects symlinks (ErrTargetIsSymlink) so chmod never lands on the link's target — the guard moved into the open itself (O_NOFOLLOW handle, fchmod), which means non-root callers now need read permission on the target (root unaffected); the permissions parameter is now *os.FileMode instead of *int
refactor: FileClerk.CompressFile gains shouldKeepSourceFilePtr mirroring DecompressFile — every format shares one source-deletion contract (the external tool always keeps its input, FileClerk removes it unless keeping is requested; previously tar kept the source while gzip/xz/zip/brotli silently deleted it); this also fixes CompressDir returning an ENOENT error on success
test: add TestWriteNewFile, TestCompressDir, and regression coverage for exclusive create/copy/move, source-mode preservation, symlink guards, ErrFileTooLarge, and wrong-type delete rejections
feat: add LiaisonResponseStatus accepted (202/exit 0), conflict (409/exit 65), and serviceUnavailable (503/exit 69); both renderers map them per the curated table
test: add TestLogLevelParser, TestIsStdoutTerminal, TestChildEnvironment (inherited+PWD, Username-alone-inherits, clean-env PATH/HOME/PWD policy incl. target-user home on switch, Envs-win end-to-end over parent and clean base), StdoutFileCapture/CaptureFileHandlesDoNotLeak (fd-count regression), and timeout (CommandTimeoutEnforced, NaturalExitCode124KeepsCommandStdErr) coverage; extend both renderer tests with the three new statuses
chore: gofmt paginationQueryBuilder_test.go (pre-existing formatting drift)
chore: update go to 1.27.1 and refresh dependencies; net/mail now accepts bracketed IPv6 domains, so MailAddress follows the stdlib and accepts them too

0.3.2 - 2026/07/30
fix: ReadFileContent accepts symlinks (replace IsFile guard with os.IsNotExist mapping; os.Open follows the chain)
fix: lower ReadFileContent default cap from 1GiB to 500MiB to bound in-memory string allocation
docs: document 500MiB cap on ReadFileContent and recommend streaming (io.Reader/bufio.Scanner) for larger files
refactor: extract ReadFileContent default cap to ReadFileContentDefaultMaxSizeBytes constant
fix: correct 10MB to 10MiB unit in regexSearchWholeFile inline comment
test: add ReadSymlinkToFile and ReadDanglingSymlink coverage to TestReadFileContent

0.3.1 - 2026/07/30
feat: add ShouldBypassLocalResolver to DnsLookupSettings; bypass /etc/hosts by issuing raw dnsmessage UDP query for A/AAAA and parsing the response in-process
feat: add directIpAddressResolver that builds DNS messages via golang.org/x/net/dns/dnsmessage and parses A/AAAA resource records
refactor: rename resolverFactory to netResolverBuilder and queryDnsRecords to defaultDnsRecordsResolver; extract dnsRecordsResolver dispatcher that routes IP record types to the direct path and others to the system resolver
chore: promote golang.org/x/net from indirect to direct dependency
test: add directIpAddressResolver A/AAAA tests, ShouldBypassLocalResolver construction test, and bypass-aware Execute tests (bypass verified against /etc/hosts pointing goinfinite.dev to 127.0.0.1)
test: refactor dnsLookup_test into 3 top-level table-driven funcs; verify ShouldBypassLocalResolver via localhost (local resolver returns 127.0.0.1, raw 8.8.8.8 query does not)
fix: randomize DNS transaction id via crypto/rand; surface response RCODE, truncation, and id-mismatch failures as ErrDnsLookupResponse* sentinels in direct resolver path (was silent empty+nils with static id=1)
refactor: split dnsResponseIpAddressesExtractor into dnsMessageValidator (parse + header validation) and dnsMessageIpAddrExtractor (record extraction); rename dnsMessagePackBuilder to dnsMessagePacker; drop dead `queryError = err` after guards in defaultDnsRecordsResolver
chore: extract dnsStandardPort = "53" constant (was hardcoded in two dial sites)
test: LocalhostBypassedSkipsLocalLookup accepts DnsLookupResponseNameError as evidence bypass succeeded (8.8.8.8 returns NXDOMAIN for localhost under bypass, not empty+nils)

0.3.0 - 2026/07/27
feat: add FileContentRegexSearch and FileContentRegexReplace to FileClerk (size-based routing at 10MiB, atomic .tmp+rename, follow-symlinks, would-empty-result guard)
feat: add FileClerk.OverwriteFile (atomic source-over-target rename resolving symlink chains via filepath.EvalSymlinks)
refactor: replace RegexPattern VO with native *regexp.Regexp; return []FileContentRegexFindings with 1-based inclusive LineNumRange
refactor: tighten FileContentRegex error contract — ErrSourceIsDirectory and ErrTargetIsDirectory distinguished; ErrReplacementWouldTruncateFile via tempfile stat; streaming-fallback slog.Warn lives at the dispatch site; DeleteFileContent delegates to TruncateFileContent
docs: align .context.md, README.md, and FEATURE-MAP with 0.3.0 surface; add Human Reviewed callout to README
fix: regex search derives match offsets from engine; replace preserves terminators and mode
fix: capture deferred Close errors in CopyFile and UpdateFileContent
fix: close file handles before reporting success in write-bearing operations
fix: resolve symlink target before creating regexReplace temp file (cross-device rename)
docs: explain why regexSearchWholeFile keeps two-pass regex scan

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
