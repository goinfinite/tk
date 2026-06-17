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
5. `src/infra/db/paginationQueryBuilder.go` — builds paginated GORM queries (page-number or last-seen-id mode, sorting, total count)
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

1. `src/infra/synthesizer.go` — `SynthesizePrivateKey` generates an RSA/ECDSA/Ed25519 key; `SynthesizeSelfSignedCert` creates a self-signed certificate from a private key
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

1. `src/infra/shell.go` — `NewShell` configures a command; `Run` executes it with timeout enforcement, optional user switching, and stdout/stderr capture
2. `src/infra/shellEscape.go` — `Quote` escapes shell arguments for safe interpolation

---

## DNS Lookup

Resolves DNS records with configurable resolvers and timeout.

**Flow:**

1. `src/infra/dnsLookup.go` — `NewDnsLookup` configures resolver IPs, hostname, and record type; `Lookup` performs the resolution

---

## File Operations

Provides filesystem utilities: existence checks, read/write, copy, move, compress/decompress, permission management.

**Flow:**

1. `src/infra/fileClerk.go` — `FileClerk` struct with methods for all filesystem operations

---

## Data Deserialization

Deserializes JSON and YAML from files or readers into maps.

**Flow:**

1. `src/infra/deserializer.go` — `DataDeserializeFile` reads a file and deserializes based on extension; `dataDeserializer` handles reader-based deserialization

---

## Random String / Password Generation

Generates random strings with configurable charsets and cryptographic passwords.

**Flow:**

1. `src/infra/synthesizer.go` — `SynthesizeRandomString` generates from custom charset; `SynthesizePassword` generates passwords meeting complexity requirements

---

## Server IP Address Detection

Reads the server's private and public IP addresses.

**Flow:**

1. `src/infra/serverIpAddress.go` — `ReadServerPrivateIpAddress` via `hostname -I`; `ReadServerPublicIpAddress` via DNS lookup to external resolver

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

1. `src/presentation/responseWrappers.go` — `ApiResponseWrapper` for HTTP JSON responses; `LiaisonCliResponseRenderer` for terminal output with chroma syntax highlighting; `SimpleCliResponseRenderer(isSuccess, message)` for simplified CLI usage — maps isSuccess to a LiaisonResponse status and delegates to LiaisonCliResponseRenderer for JSON envelope output

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

1. `src/presentation/middleware/logHandler.go` — `LogHandler.Init` reads LOG_LEVEL env var and configures slog with zerolog backend; supports Debug, Info, Warn, Error levels with TTY-aware formatting

---

## Panic Recovery (API and CLI)

Catches panics, logs stack traces, and returns safe error responses.

**Flow:**

1. `src/presentation/middleware/panicHandler.go` — `ApiPanicHandler` is Echo middleware that catches panics, writes stack traces to `logs/panic.log`, filters domain-layer frames, and returns HTTP 500 with masked error for untrusted clients; `CliPanicHandler` does the same for CLI via `defer`

---

## Honeypot IP Blocking

Intercepts security scanner probes with fake-vulnerability payloads, then blocks source IP with 24h TTL. All subsequent requests from banned IPs redirect to xkcd.com.

**Flow:**

1. `src/presentation/honeypotMiddleware.go` — `HoneypotMiddlewareSettings` struct with VO-typed fields (ActivePathCount HoneypotActivePathCount, AggressivenessMode HoneypotAggressivenessMode, MaxEntries HoneypotMaxEntries, MaxStreamSizeBytes HoneypotMaxStreamSizeBytes, RedirectUrl Url, StatsInterval HoneypotStatsInterval); `HoneypotPathMapping` uses UrlPath and MimeType VOs; `NewHoneypotMiddleware(settings, activityRecordCmdRepo, activityRecordQueryRepo)` returns echo.MiddlewareFunc; env var parsing done directly via VO constructors in constructor; `resolveAggressivenessMode` package-level function handles deprecated mode fallback; `HoneypotActivePathCount`, `HoneypotMaxEntries`, `HoneypotMaxStreamSizeBytes`, `HoneypotStatsInterval` value objects in `src/domain/valueObject/`
2. Middleware intercepts honeypot paths via map lookup (not route table): checks all requests against path map, returns fake payload + creates HoneypotHit ActivityRecord
3. Middleware checks all requests against ban list: queries ActivityRecordQueryRepo for HoneypotHit records with CreatedAfterAt filter (24h window), returns 302 if banned
4. `src/infra/activityRecord/activityRecordCmdRepo.go` — creates ActivityRecord with RecordCode="HoneypotHit", RecordLevel="SECURITY", OperatorIpAddress extracted via RequesterIpExtractor
5. `src/infra/activityRecord/activityRecordQueryRepo.go` — queries ActivityRecord with CreatedAfterAt filter for TTL enforcement
6. `src/presentation/requesterIpExtractor.go` — extracts untrusted IP from X-Forwarded-For/X-Real-IP headers (if trusted proxy) or RemoteAddr
7. Fake payloads: 25 paths including /.env (dotenv), /wp-config.php (PHP constants), /actuator/env (JSON), /backup.sql (SQL), /.git/config (git config), /server-status (Apache HTML), /phpmyadmin (login form), etc.
8. `src/infra/db/transientDatabaseService.go` — in-memory SQLite key-value store for hit-count tracking; KeyValueModel with Key, Value, CreatedAt fields; methods: Has, Read, ReadAll, Set, Count

**Environment Variables:**

- `HONEYPOT_AGGRESSIVENESS` — aggressiveness mode (immediate/balanced/tolerant/observe), default balanced
- `HONEYPOT_ACTIVE_PATHS` — number of active honeypot paths, default 30, floor 30
- `HONEYPOT_MAX_ENTRIES` — max transient DB entries, default 5000, floor 100, ceiling 50000
- `HONEYPOT_MAX_STREAM_SIZE` — max stream size in bytes, default 20MB, floor 5MB
- `HONEYPOT_STATS_INTERVAL` — stats aggregation interval, default 30m, floor 5m

---
