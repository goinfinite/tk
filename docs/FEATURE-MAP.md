# Feature Map

> Auto-maintained index of every user-facing feature and the code path that implements it. Updated alongside the code — not after the fact.

## Create Activity Record

Persists an audit trail entry recording a system event. Designed as a fire-and-forget side effect — errors are logged but never returned to the caller.

**Flow:**

1. `src/domain/dto/createActivityRecord.go` — input DTO carrying record code, level, message, operator info, and affected resources
2. `src/domain/useCase/createActivityRecord.go` — orchestrates the create operation; delegates to the cmd repo and logs errors without propagating them
3. `src/domain/repository/activityRecordCmdRepo.go` — interface declaring the `Create` method
4. `src/infra/activityRecord/activityRecordCmdRepo.go` — GORM implementation: transforms DTO to database model and persists via trail database
5. `src/infra/db/model/activityRecord.go` — GORM model struct for the activity_records table
6. `src/infra/db/model/activityRecordAffectedResource.go` — GORM model for associated affected resources (one-to-many)
7. `src/infra/db/trailDatabaseService.go` — SQLite database connection and auto-migration

---

## Read Activity Records

Queries activity records with filtering and pagination support. Returns a paginated response with matching records.

**Flow:**

1. `src/domain/dto/readActivityRecords.go` — request DTO with optional filters (record code, level, time range, operator, affected resource) and response DTO wrapping pagination + entity slice
2. `src/domain/useCase/readActivityRecords.go` — orchestrates the read; defines default pagination; delegates to the query repo
3. `src/domain/repository/activityRecordQueryRepo.go` — interface declaring `Read` (paginated list) and `ReadFirst`
4. `src/infra/activityRecord/activityRecordQueryRepo.go` — GORM implementation: builds filtered query, applies pagination, loads associated resources, transforms models to entities
5. `src/infra/db/paginationQueryBuilder.go` — builds paginated GORM queries (page-number or last-seen-id mode, sorting, total count); callers MUST set `.Model(...)` on the query — sort fields and the cursor default resolve through the parsed schema, and `.Table(...)`-only queries fail with ParseStatementSchemaError
6. `src/infra/db/model/activityRecord.go` — GORM model with `ToEntity()` conversion to domain entity
7. `src/domain/entity/activityRecord.go` — domain entity returned in the response

---

## Delete Activity Record

Removes an activity record by ID or filter criteria.

**Flow:**

1. `src/domain/dto/deleteActivityRecord.go` — input DTO with deletion filters
2. `src/domain/useCase/deleteActivityRecord.go` — orchestrates deletion; wraps infra errors as domain errors
3. `src/domain/repository/activityRecordCmdRepo.go` — interface declaring the `Delete` method
4. `src/infra/activityRecord/activityRecordCmdRepo.go` — GORM implementation: transforms DTO to model conditions and deletes

---

## X.509 Certificate Parsing

Parses PEM-encoded X.509 certificates into a richly typed domain entity with all standard fields.

**Flow:**

1. `src/domain/entity/x509Certificate.go` — entity with `NewX509Certificate` constructor that accepts a PEM string, parses it via Go's `crypto/x509`, and populates all fields (subject, issuer, SANs, key usage, policies, etc.) as value objects

---

## Self-Signed Certificate Generation

Generates a private key and self-signed X.509 certificate for TLS bootstrap.

**Flow:**

1. `src/infra/synthesizer.go` — `PrivateKeyPemFactory` generates an RSA/ECDSA/DSA/Ed25519 key PEM; `CertificatePemFactory` and `CACertificatePemFactory` create self-signed certificate PEM pairs
2. `src/infra/readThrough.go` — `CertPairFilePathsReader` attempts to read cert/key paths from environment variables, falls back to generating a self-signed pair and writing it to disk

---

## AES-GCM Encryption / Decryption

Encrypts and decrypts data using AES-GCM with base64-encoded secret keys.

**Flow:**

1. `src/infra/cypher.go` — `NewCypherSecretKey` generates a 32-byte key; `NewCypher` initializes from an existing key; `Encrypt`/`Decrypt` perform AES-GCM operations

---

## Shell Command Execution

Runs subprocess commands with configurable timeout, user, working directory, and environment.

**Flow:**

1. `src/infra/shell.go` — `NewShell` configures a command; `Run` executes it under a context deadline (SIGTERM, SIGKILL after a grace; exit 124 `CommandDeadlineExceeded`) unless `ShouldDisableTimeout` removes the deadline entirely, with optional user switching by `Username` or `UserId` (Username wins when both are set; an unresolvable target fails unless `ShouldIgnoreUsernameLookupError` is set), and stdout/stderr capture; when `ShouldUseCleanEnv` is set the child runs from a minimal environment (PATH of the resolved user's `~/.local/bin` plus the standard system directories, HOME — parent's for same-user runs or the target user's home on user switch, PWD when a working directory is set, DEBIAN_FRONTEND, and explicit Envs) so the parent's other variables cannot leak
2. `src/infra/shellEscape.go` — `Quote` escapes shell arguments for safe interpolation

---

## DNS Lookup

Resolves DNS records with configurable resolvers, timeouts, and an optional local-resolver bypass. The bypass path skips Go's net.Resolver (so /etc/hosts and resolv.conf are not consulted) and issues raw dnsmessage UDP queries to the configured resolver; it applies only to A and AAAA record types.

**Flow:**

1. `src/infra/dnsLookup.go` — `NewDnsLookup` accepts a `DnsLookupSettings` (resolvers, timeouts, bypass flag); `Execute(hostname, *recordType)` performs the resolution with primary-then-secondary fallback and logs each failed attempt

---

## File Operations

Provides filesystem utilities: existence checks, read/write, copy, move, compress/decompress, permission management, regex search.

**Flow:**

1. `src/infra/fileClerk.go` — `FileClerk` struct with methods for all filesystem operations; creation is exclusive (`WriteNewFile`/`CopyFile` reject taken paths with `ErrTargetFileExists`, `CopyFile` preserves the source mode; `TouchFile` carries the touch(1) idempotent contract), `MoveFile` relocates via renameat2(2) RENAME_NOREPLACE without replacing a foreign target (same-file moves succeed as no-ops; cross-device moves fall back to mv(1)-style copy+delete); `CompressFile` and `DecompressFile` share one `shouldKeepSourceFilePtr` contract where the external tool always keeps its input and FileClerk performs the deletion; `FileClerk.OverwriteFile` atomically replaces a target file (resolving symlink chains via filepath.EvalSymlinks); `VerifyDirPathRedirectSafety(dirPath, ownerUsernamePtr, ownerUserIdPtr)` walks a directory chain with O_PATH|O_NOFOLLOW and reports the first surprise (`ErrDirPathTraversalInvalid`, `ErrSymlinkedPathInvalid`, `ErrTargetNotDirectory`, `ErrDirectoryOwnerInvalid`, `PathCheckFailed`; an omitted owner defaults to the process account, and dot and empty components are skipped); `UpsertFile` takes a `FileUpsertSettings` struct and the content bytes, resolves the owner account (defaulting to the process account) to uid/gid value objects, scouts the target's parent chain, stages a private temp file (hidden `.name.<entropy>.tk-tmp`), and swaps it in via openat/renameat2 relative to the scouted handle, so no component is re-resolved after the check; it replaces the target (symlink included, never followed) only when `ShouldOverwrite` is set, writes through symlinked final components and parent chains when `ShouldFollowSymlinks` is set, refuses a symlinked parent chain with `ErrSymlinkedPathInvalid` otherwise, refuses a taken target with `ErrTargetFileExists`, an unset `Permissions` mode with `ErrFilePermissionsInvalid`, a non-file-name final component with `ErrFileNameInvalid`, and a temp name beyond NAME_MAX with `ErrTempFileNameTooLong`, and chowns before it chmods (chown clears setuid/setgid bits); regex replace and `UpsertFile` share one `writeFileAtomically` staging path; `FileContentRegexSearch` and `FileContentRegexReplace` use size-based routing at 10MiB (whole-file pass / bufio.Scanner streaming), follow symlinks, and reject directories and would-empty results (`ErrTargetIsDirectory`, `ErrSourceIsDirectory`, `ErrReplacementWouldTruncateFile`)

---

## Data Deserialization

Deserializes JSON and YAML from files or readers into maps.

**Flow:**

1. `src/infra/deserializer.go` — `DataDeserializeFile` reads a file and deserializes based on extension; `dataDeserializer` handles reader-based deserialization

---

## Password Generation

Generates cryptographic random integers, passwords with charset guarantees, and dummy identities.

**Flow:**

1. `src/infra/synthesizer.go` — `RandomIntegerGenerator` draws a uniform crypto/rand integer between two inclusive bounds; `PasswordFactory` draws crypto/rand passwords, placing each requested character class at its own guaranteed position; `CharsetPresenceGuarantor` ensures a charset appears in a byte slice; `UsernameFactory`/`MailAddressFactory` generate dummy identities

---

## Server IP Address Detection

Reads the server's private and public IP addresses.

**Flow:**

1. `src/infra/serverIpAddress.go` — `ReadServerPrivateIpAddress` via `hostname -I`; `ReadServerPublicIpAddress` honors `SERVER_PUBLIC_IP_ADDR` env var, then fans out across multiple public IP resolvers

---

## Trusted IPs Reader

Reads a list of trusted IP addresses from the TRUSTED_IPS environment variable.

**Flow:**

1. `src/infra/trustedIpsReader.go` — `TrustedIpsReader` parses comma-separated IPs from the environment variable into validated value objects

---

## API Request Input Reading

Reads and merges HTTP request input from path parameters, query strings, and request body (JSON or form) into a single map.

**Flow:**

1. `src/presentation/requestInputReader.go` — `ApiRequestInputReader.Reader` merges request body, query params, route params, operator context, and multipart file uploads; supports dot-notation keys for hierarchical maps

---

## API / CLI Response Formatting

Wraps responses in a standard envelope for API consumers and provides syntax-highlighted JSON output for CLI.

**Flow:**

1. `src/presentation/responseWrappers.go` — `ApiResponseWrapper` for HTTP JSON responses; `LiaisonCliResponseRenderer` for terminal output with chroma syntax highlighting; `SimpleCliResponseRenderer(isSuccess, message)` for simplified CLI usage — maps isSuccess to a LiaisonResponse status and delegates to LiaisonCliResponseRenderer for JSON envelope output; `LiaisonApiResponseEmitter` and `LiaisonCliResponseRenderer` translate each curated `LiaisonResponseStatus` (success, created, accepted/202, multiStatus, userError, unauthorized, forbidden, notFound, timeout, conflict/409, rateLimited, infraError, unknownError, serviceUnavailable/503) to its HTTP code and sysexits CLI code

---

## Pagination Parsing

Parses pagination parameters from untrusted input into a typed Pagination DTO.

**Flow:**

1. `src/presentation/paginationParser.go` — `PaginationParser` extracts page number, items per page, sort by, sort direction, and last-seen-id from an input map

---

## Environment Variable Inspection

Loads .env files, validates that required environment variables are set, and auto-fills derivable values (e.g., server IP).

**Flow:**

1. `src/presentation/envsInspector.go` — `NewEnvsInspector` configures required and auto-fillable vars; `InspectEnvs` loads the .env file and validates; `AutoFillRequiredEnvVars` populates derivable values

---

## Log Level Configuration

Configures structured logging level at application startup.

**Flow:**

1. `src/presentation/middleware/logHandler.go` — `LogHandler.Init` reads LOG_LEVEL env var and configures slog with zerolog backend; supports Debug, Info, Warn, Error levels case-insensitively; logs always go to stderr so the CLI's stdout carries only the JSON response, and interactive debug sessions get the console writer formatting (tkInfra.IsStdoutTerminal)

---

## Panic Recovery (API and CLI)

Catches panics, logs stack traces, and returns safe error responses.

**Flow:**

1. `src/presentation/middleware/panicHandler.go` — `ApiPanicHandler` is Echo middleware that catches panics, writes stack traces to `logs/panic.log`, filters domain-layer frames, and returns HTTP 500 with masked error for untrusted clients; `CliPanicHandler` does the same for CLI via `defer`

---
