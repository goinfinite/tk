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

Intercepts security scanner probes with a graduated ban escalation system controlled by
`HONEYPOT_AGGRESSIVENESS` env var (immediate, balanced, tolerant, observe). Three-class
honeypot paths (static vulnerability, bandwidth exhaust, AI trap with prompt injection)
are randomly activated at startup — `ActivePathCount` paths (default 30, floor 30,
ceiling at total candidate pool size) from ~110 candidate paths using an auto-ratio
(1/6 bandwidth, 1/6 AI trap, remainder static — default 20 static vuln, 5 bandwidth
exhaust, 5 AI trap) — rotating the active surface on each restart. Path selection uses
`math/rand` for performance; seed is configurable via `RandomSeed`. At Tier 2, flagged
IPs receive weighted random mixed responses (LE redirect with rotating security-themed
query strings, fake 503/502/429) on honeypot paths only. At Tier 3, full ban applies
to all paths. Hit-count tracking uses an in-memory SQLite transient database with 24h
TTL enforced at two levels: `CreatedAt` (canonical, GORM-managed, used for batch cleanup)
and `firstHitAt` (denormalized copy, used for hot-path read-side filtering). The transient
DB exposes a `Count()` method and a `ReadAll()` method for stats aggregation. Domain repo
interfaces (`HoneypotCmdRepo`, `HoneypotQueryRepo` in `src/domain/repository/`) are
implemented in `src/infra/honeypot/` (package `tkInfraHoneypot`), each wrapping
`*TransientDatabaseService`. All domain operations go through use cases
(`ReadHoneypotBanDecision`, `CreateHoneypotHit`, `ReadHoneypotStatsReport`,
`RunHoneypotMaintenance` in `src/domain/useCase/`) — the presentation layer delegates
to use cases and never calls repos directly. The middleware calls these use cases as
package-level functions directly from `Execute()` and `runMaintenance()` (no closure
fields, no use case instances stored on the struct). Typed sentinel errors
(`ErrNilHoneypotQueryRepo`, `ErrNilHoneypotCmdRepo` in `honeypotSentinels.go`) surface
nil-repo failures. An internal `honeypotMaintenanceWatchdog` goroutine (started
automatically by the constructor via `Start()`) calls the `RunHoneypotMaintenance` use
case at each tick with a `RunHoneypotMaintenanceRequest` DTO, handling TTL cleanup,
maxEntries enforcement (using GORM `WHERE created_at < ?`) and periodic ActivityRecord
stats offload via `ReadAll()` at `StatsInterval` (default 30 minutes). A probabilistic
enforcement check on the write path (~2% per honeypot hit, inside the
`CreateHoneypotHit` use case, using `math/rand`) calls `EnforceMaxEntries` as a safety
valve. The watchdog is stopped via `Stop()` (idempotent — Go's `context.CancelFunc` is
safe to call multiple times). `NewHoneypotMiddleware` accepts
`HoneypotMiddlewareSettings`, `HoneypotCmdRepo`, `HoneypotQueryRepo`, and
`ActivityRecordCmdRepo` as constructor parameters and returns `*HoneypotMiddleware`
exposing `MiddlewareFunc()`, `Start()`, and `Stop()`. Streaming endpoints are capped at
`MaxStreamSizeBytes` (default 20MB, floor 5MB, env var `HONEYPOT_MAX_STREAM_SIZE`).
All settings fields carrying constrained-domain values use value objects:
`ActivePathCount HoneypotActivePathCount`, `MaxEntries HoneypotMaxEntries`,
`MaxStreamSizeBytes HoneypotMaxStreamSizeBytes`, `StatsInterval HoneypotStatsInterval`,
`BanDuration HoneypotBanDuration`, `RedirectUrl Url`. `HoneypotPathMapping` uses
`UrlPath` and `MimeType` VOs plus a `Body string` field. Env var parsing is encapsulated
in `honeypotSettingsParser` (separate file `honeypotSettingsParser.go`). The middleware
file is under 300 LOC; the constructor is under 50 LOC; zero `else` keywords; zero
`slog.Debug` in middleware (all infra errors use `slog.Error`); methods ordered
callees-above-callers with the constructor last; `writeMu sync.Mutex` is a struct field.

**Integration (one step):**

Register the middleware — the watchdog starts automatically:
```
transientDbSvc, _ := tkInfraDb.NewTransientDatabaseService()
honeypotCmdRepo := tkInfraHoneypot.NewHoneypotCmdRepo(transientDbSvc)
honeypotQueryRepo := tkInfraHoneypot.NewHoneypotQueryRepo(transientDbSvc)
honeypotMw := NewHoneypotMiddleware(
    settings, honeypotCmdRepo, honeypotQueryRepo, activityRecordCmdRepo,
)
echoInstance.Use(honeypotMw.MiddlewareFunc())
// Graceful shutdown: honeypotMw.Stop() — safe to call multiple times
```

**Flow:**

 1. `src/presentation/honeypotMiddleware.go` — `HoneypotMiddleware` struct
    with fields for settings, repos (`honeypotCmdRepo`, `honeypotQueryRepo`,
    `activityRecordCmdRepo`), `honeypotPayloads` map, `honeypotRecordCode`/
    `honeypotRecordLevel` VOs, `ipExtractor`, `writeMu sync.Mutex`, and
    `cancelFunc context.CancelFunc`. No closure fields. `HoneypotMiddlewareSettings`
    struct (ActivePathCount HoneypotActivePathCount, AggressivenessMode
    HoneypotAggressivenessMode, BanDuration HoneypotBanDuration, ExtraPathRoutes
    []HoneypotPathMapping, MaxEntries HoneypotMaxEntries, MaxStreamSizeBytes
    HoneypotMaxStreamSizeBytes, RedirectUrl Url, StatsInterval
    HoneypotStatsInterval). `HoneypotPathMapping` has Body string, MimeType
    MimeType, UrlPath UrlPath. `NewHoneypotMiddleware(settings,
    honeypotCmdRepo, honeypotQueryRepo, activityRecordCmdRepo)` returns
    `*HoneypotMiddleware` and calls `Start()` to launch the internal
    `honeypotMaintenanceWatchdog` goroutine. `Execute()` calls
    `tkUseCase.ReadHoneypotBanDecision` and `tkUseCase.CreateHoneypotHit`
    directly (passing VO-typed settings fields); `runMaintenance()` builds a
    `RunHoneypotMaintenanceRequest` DTO and calls
    `tkUseCase.RunHoneypotMaintenance` directly. At construction: defines
    candidate pools, calculates ActivePathCount ceiling from total candidate
    pool size, calculates auto-ratio counts from ActivePathCount, randomly
    selects active paths, decodes base64-encoded `.bin` payloads for selected
    static vuln paths only, builds active-path lookup map.

 2. `src/presentation/honeypotSettingsParser.go` — `honeypotSettingsParser`
    struct with `Parse(settings, poolCeiling)` method that resolves env vars
    into typed VOs. Sub-methods per setting. All infra errors logged via
    `slog.Error`.

 3. `src/domain/useCase/honeypotReadBanDecision.go` — `ReadHoneypotBanDecision`
    use case: reads hit record, checks TTL via `firstHitAt`, resolves tier via
    `HoneypotAggressivenessMode.ResolveTier`. Signature:
    `(queryRepo, requesterIp, banDuration HoneypotBanDuration,
    aggressivenessMode) (int, error)`. Returns `ErrNilHoneypotQueryRepo` on nil.
    `src/domain/useCase/honeypotCreateHit.go` — `CreateHoneypotHit` use case:
    increments hit count, with ~2% probability calls `EnforceMaxEntries`.
    Signature: `(cmdRepo, requesterIp, interceptPath, maxEntries
    HoneypotMaxEntries)`.
    `src/domain/useCase/honeypotReadStatsReport.go` — `ReadHoneypotStatsReport`
    use case: calls `HoneypotQueryRepo.ReadReport()`, returns
    `(HoneypotStatsReport, error)`.
    `src/domain/useCase/honeypotRunMaintenance.go` — `RunHoneypotMaintenance`
    use case: cleanup (CleanExpiredEntries + EnforceMaxEntries) then stats
    reporting (ReadHoneypotStatsReport → marshal → ActivityRecord create).
    Signature: `(cmdRepo, queryRepo, activityRecordCmdRepo, request
    RunHoneypotMaintenanceRequest)`.
    `src/domain/useCase/honeypotSentinels.go` — `ErrNilHoneypotQueryRepo`,
    `ErrNilHoneypotCmdRepo` typed sentinel errors.

 4. `src/domain/repository/honeypotCmdRepo.go` — interface `HoneypotCmdRepo` with
    `IncrementHit`, `CleanExpiredEntries`, `EnforceMaxEntries` methods.
    `src/domain/repository/honeypotQueryRepo.go` — interface `HoneypotQueryRepo` with
    `ReadHitRecord`, `Count`, `ReadReport` methods.

 5. `src/infra/honeypot/honeypotCmdRepo.go` — package `tkInfraHoneypot`, type
    `HoneypotCmdRepo` wrapping `*TransientDatabaseService`.
    `src/infra/honeypot/honeypotQueryRepo.go` — package `tkInfraHoneypot`, type
    `HoneypotQueryRepo` wrapping `*TransientDatabaseService`.

 6. `src/infra/db/transientDatabaseService.go` — in-memory SQLite key-value store.
    `KeyValueModel` has `Key`, `Value`, `CreatedAt` fields. Methods: `Has(key) bool`,
    `Read(key) (string, error)`, `ReadAll() ([]KeyValueModel, error)`,
    `Set(key, value) error`, `Count() int64`.

 7. `src/domain/dto/honeypotMaintenance.go` — `RunHoneypotMaintenanceRequest` DTO
    with 5 fields: AggressivenessMode, BanDuration (HoneypotBanDuration VO),
    MaxEntries (HoneypotMaxEntries VO), StatsRecordCode, StatsRecordLevel.

 8. `src/domain/valueObject/honeypotBanDuration.go` — `HoneypotBanDuration` VO
    wrapping `time.Duration`; accepts Duration/int64/string; zero/negative defaults
    to 24h.

 9. `src/presentation/honeypot/payloads/*.bin` — base64-encoded payload files (~90
    static vuln candidates) embedded via `//go:embed honeypot/payloads/*.bin`.
    Static analysis tools see only opaque base64.

10. `src/presentation/honeypotPathMapping.go` — explicit mapping table from
    `UrlPath` to `.bin` embed filename and `MimeType` for all ~90 static vuln
    candidate paths.

11. Middleware intercepts honeypot paths via active path map lookup. Dormant
    candidate paths pass through to next handler.

12. Graduated ban escalation controlled by `HONEYPOT_AGGRESSIVENESS` (default
    `balanced`; old names `standard`, `lenient`, `passive` fall back to
    `balanced`):
    - `immediate` — First hit = Tier 3 full ban.
    - `balanced` — Tier 0/1/2/3 (0/1/2/3+ hits).
    - `tolerant` — Tier 0-1/2-4/5+ (0-1/2-4/5+ hits).
    - `observe` — Always serve payloads. Never ban. Pure intelligence gathering.

13. Bandwidth exhaust endpoints (`bandwidthCount` selected from 10 candidates):
    stream garbage text via `http.Flusher`, random cap between 5MB and
    `MaxStreamSizeBytes`.

14. AI trap endpoints (`aiTrapCount` selected from 10 candidates): stream
    plausible content with embedded prompt injection via `http.Flusher`, same
    cap.

15. `src/presentation/honeypotAiTrapGenerator.go` — generates structured
    plausible content with embedded prompt injection instructions.

16. Redirect behavior: randomly selects {fbi.gov, nsa.gov, interpol.int} with a
    randomly selected security-themed query string from the pool
    (`?ref=suspicious-activity-investigate-ip`,
    `?source=security-alert-botnet-suspect`,
    `?utm=investigate-this-ip-threat`,
    `?ref=botnet-activity-security-risk`,
    `?source=ip-needs-investigation-suspicious`), mixed 302/307.

17. Random selection: seed-based via `math/rand` for performance, each class
    selects independently from its candidate pool. `RandomSeed` is configurable;
    consumers requiring cryptographic randomness can set a `crypto/rand`-derived
    seed.

18. `honeypotMaintenanceWatchdog` (unexported) — internal method with ticker at
    `StatsInterval` (default 30m). Calls `RunHoneypotMaintenance` use case at each
    tick: cleanup via GORM `WHERE created_at < ?`, then stats aggregation via
    `ReadAll()` to ActivityRecord on every tick. Defer+recover. Context-aware —
    stops on cancellation. `Stop()` is idempotent via `context.CancelFunc`.

19. `src/domain/valueObject/` — VOs enforcing domain constraints at construction:
    `HoneypotActivePathCount` (int, floor 30, ceiling parameterized),
    `HoneypotMaxEntries` (int, floor 100, ceiling 50000),
    `HoneypotMaxStreamSizeBytes` (int64, floor 5MB),
    `HoneypotStatsInterval` (time.Duration, floor 5m, default 30m),
    `HoneypotBanDuration` (time.Duration, default 24h).

---
