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

## Honeypot Ban Decision

Determines whether a requester IP should be banned based on cumulative hit count and aggressiveness mode. Escalates through graduated actions (payload → mixed → stream → ban) depending on mode.

**Flow:**

1. `src/presentation/middleware/honeypotMiddleware.go` — middleware extracts operator IP and invokes ban decision use case
2. `src/domain/useCase/readHoneypotBanDecision.go` — resolves ban threshold and suggested action from aggressiveness mode
3. `src/domain/dto/readHoneypotBanDecision.go` — request/response DTOs for ban decision queries
4. `src/domain/repository/honeypotQueryRepo.go` — interface declaring `ReadBanDecision`
5. `src/infra/honeypot/honeypotQueryRepo.go` — queries hit count from transient database
6. `src/infra/db/transientDatabaseService.go` — in-memory SQLite database for honeypot data

---

## Honeypot Hit Creation

Records a honeypot hit when a requester accesses a monitored path. Fire-and-forget — errors are logged but never returned to the caller.

**Flow:**

1. `src/presentation/middleware/honeypotMiddleware.go` — middleware `recordHit` method creates hit entry
2. `src/domain/useCase/createHoneypotHit.go` — fire-and-forget hit recording via cmd repo (errors logged with slog.Error)
3. `src/domain/dto/createHoneypotHit.go` — input DTO carrying requester IP, path, class, and cumulative hit count
4. `src/domain/repository/honeypotCmdRepo.go` — interface declaring `Create` method
5. `src/infra/honeypot/honeypotCmdRepo.go` — GORM implementation: persists hit to transient database
6. `src/infra/db/model/honeypotHit.go` — GORM model for honeypot_hits table with unique index on requester_ip_address

---

## Honeypot Stats Report

Queries aggregated honeypot hit statistics: total hits, unique IPs, hit distribution by path class, and ban rate.

**Flow:**

1. `src/domain/useCase/readHoneypotStatsReport.go` — delegates to query repo for aggregated stats; wraps errors
2. `src/domain/dto/readHoneypotStatsReport.go` — request/response DTOs for stats queries
3. `src/domain/repository/honeypotQueryRepo.go` — interface declaring `ReadStatsReport`
4. `src/infra/honeypot/honeypotQueryRepo.go` — queries aggregated stats from transient database
5. `src/infra/db/transientDatabaseService.go` — in-memory SQLite database for honeypot data

---

## Honeypot Maintenance

Periodic cleanup of expired honeypot entries (TTL-based) and enforcement of maximum entry cap. Runs as a watchdog goroutine inside the middleware.

**Flow:**

1. `src/presentation/middleware/honeypotMiddleware.go` — watchdog goroutine fires maintenance at configured StatsInterval
2. `src/domain/useCase/runHoneypotMaintenance.go` — calls DeleteExpired then EnforceMaxEntries in sequence
3. `src/domain/dto/runHoneypotMaintenance.go` — input DTO with MaxEntries and BanDuration parameters
4. `src/domain/repository/honeypotCmdRepo.go` — interface declaring `DeleteExpired` and `EnforceMaxEntries`
5. `src/infra/honeypot/honeypotCmdRepo.go` — GORM implementation: TTL-based deletion and entry cap enforcement

---

## Honeypot Middleware

Intercepts HTTP requests to monitored honeypot paths and serves deceptive responses. Supports multiple response strategies: static payload, bandwidth exhaust stream, AI-generated trap content, mixed/redirect, and 403 ban.

**Flow:**

1. `src/presentation/middleware/honeypotMiddleware.go` — Echo middleware entry point: request interception, lifecycle management (Start/Stop), watchdog goroutine
2. `src/presentation/middleware/honeypot/honeypotPathMapping.go` — resolves request URL paths to honeypot path class
3. `src/presentation/middleware/honeypot/honeypotPathPool.go` — pool of all possible honeypot paths by class
4. `src/presentation/middleware/honeypot/honeypotPathSelector.go` — selects active paths based on settings and random seed
5. `src/presentation/middleware/honeypot/honeypotSettingsParser.go` — parses and validates raw settings input
6. `src/presentation/middleware/honeypot/streamHandler.go` — serves bandwidth exhaust stream and AI trap continuous stream
7. `src/presentation/middleware/honeypot/aiTrapGenerator.go` — generates AI-simulated administrative interface content
8. `src/presentation/middleware/honeypot/mixedResponseHandler.go` — serves mixed/redirect responses (e.g., HTTP 302 to external targets)
9. `src/presentation/middleware/honeypot/payloadLoader.go` — loads static payload files from embedded filesystem
10. `src/domain/useCase/readHoneypotBanDecision.go` — ban decision use case called per-request to determine response strategy

---
