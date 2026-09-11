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
5. `src/infra/db/paginationQueryBuilder.go` — builds paginated GORM queries (page-number or last-seen-id mode); callers MUST set `.Model(...)` because sort fields and the cursor default resolve through the parsed schema
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

1. `src/infra/synthesizer.go` — `PrivateKeyPemFactory` generates a key PEM; `CertificatePemFactory` and `CACertificatePemFactory` create self-signed certificate PEM pairs
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

1. `src/infra/shell.go` — `NewShell` configures a command; `Run` executes it under an optional context deadline (SIGTERM, then SIGKILL; exit 124 `CommandDeadlineExceeded`), with optional user switching and clean-environment support, capturing stdout/stderr
2. `src/infra/shellEscape.go` — `Quote` escapes shell arguments for safe interpolation

---

## DNS Lookup

Resolves DNS records with configurable resolvers, timeouts, and an optional local-resolver bypass. The bypass path skips Go's net.Resolver (so /etc/hosts and resolv.conf are not consulted) and issues raw dnsmessage UDP queries to the configured resolver; it applies only to A and AAAA record types.

**Flow:**

1. `src/infra/dnsLookup.go` — `NewDnsLookup` accepts a `DnsLookupSettings` (resolvers, timeouts, bypass flag); `Execute(hostname, *recordType)` performs the resolution with primary-then-secondary fallback and logs each failed attempt

---

## File Operations

Provides filesystem utilities: existence checks, read/write, copy, move, compress/decompress, permission management, regex search/replace.

**Flow:**

1. `src/infra/fileClerk.go` — `FileClerk` provides filesystem operations: exclusive create, atomic move, upsert and append, copy, compress/decompress, permission management, and regex search/replace

---

## Data Deserialization

Deserializes JSON and YAML from files or readers into maps.

**Flow:**

1. `src/infra/deserializer.go` — `DataDeserializeFile` reads a file and deserializes based on extension; `dataDeserializer` handles reader-based deserialization

---

## Password Generation

Generates cryptographic random integers, passwords with charset guarantees, and dummy identities.

**Flow:**

1. `src/infra/synthesizer.go` — `RandomIntegerGenerator`, `PasswordFactory` (with `CharsetPresenceGuarantor`), and `UsernameFactory`/`MailAddressFactory` draw crypto/rand integers, passwords, and dummy identities

---

## Server IP Address Detection

Reads the server's private and public IP addresses.

**Flow:**

1. `src/infra/serverIpAddress.go` — `ReadServerPrivateIpAddress` via `hostname -I`; `ReadServerPublicIpAddress` honors `SERVER_PUBLIC_IP_ADDR` env var, then fans out across multiple public IP resolvers

---

## Trusted IPs Reader

Reads trusted IP addresses and CIDR blocks from the TRUSTED_IPS and TRUSTED_CIDRS environment variables.

**Flow:**

1. `src/infra/trustedCidrsReader.go` — `TrustedCidrsReader` parses comma-separated entries from both env vars into validated `CidrBlock` value objects

---

## Requester IP Extraction

Extracts the real requester IP from an HTTP request by walking an ordered header chain and skipping trusted addresses.

**Flow:**

1. `src/presentation/requesterIpExtractor.go` — `NewRequesterIpExtractor` reads the `IP_EXTRACT_HEADER` chain and trusted CIDRs from `TrustedCidrsReader`; `Execute(*http.Request)` walks each header right-to-left and returns the first untrusted IP, falling back to `RemoteAddr`

---

## API Request Input Reading

Reads and merges HTTP request input from path parameters, query strings, and request body (JSON or form) into a single map.

**Flow:**

1. `src/presentation/requestInputReader.go` — `ApiRequestInputReader.Reader` merges request body, query params, route params, operator context, and multipart file uploads; supports dot-notation keys for hierarchical maps

---

## API / CLI Response Formatting

Wraps responses in a standard envelope for API consumers and provides syntax-highlighted JSON output for CLI.

**Flow:**

1. `src/presentation/responseWrappers.go` — `ApiResponseWrapper` for HTTP JSON; `LiaisonCliResponseRenderer` and `SimpleCliResponseRenderer` for terminal output; each `LiaisonResponseStatus` maps to an HTTP code and a sysexits CLI code

---

## Pagination Parsing

Parses pagination parameters from untrusted input into a typed Pagination DTO.

**Flow:**

1. `src/presentation/paginationParser.go` — `PaginationParser` extracts page number, items per page, sort by, sort direction, and last-seen-id from an input map

---

## Required Params Inspection

Checks that all required parameters are present in a request input map before further parsing.

**Flow:**

1. `src/presentation/requiredParamsInspector.go` — `RequiredParamsInspector(inputMap, requiredNames)` returns an error listing any missing parameter names

---

## String Slice Value Object Parsing

Parses raw request input (string, slice, or single value) into a slice of typed value objects.

**Flow:**

1. `src/presentation/stringSliceVoParser.go` — `StringSliceValueObjectParser` splits `;`/`,` strings or iterates slices, runs each element through a value-object constructor, and skips invalid values

---

## Time Params Parsing

Parses optional time parameters from untrusted input into `UnixTime` pointers.

**Flow:**

1. `src/presentation/timeParamsParser.go` — `TimeParamsParser(names, inputMap)` returns a `*UnixTime` for each name present in the input; an unparsable value yields a nil entry

---

## Environment Variable Inspection

Loads .env files, validates that required environment variables are set, and auto-fills missing auto-fillable variables with generated secret keys.

**Flow:**

1. `src/presentation/envsInspector.go` — `NewEnvsInspector` configures required and auto-fillable vars; `Inspect` loads the .env file, validates required vars, and appends generated secret keys for missing auto-fillable vars

---

## Log Level Configuration

Configures structured logging level at application startup.

**Flow:**

1. `src/presentation/middleware/logHandler.go` — `LogHandler.Init` reads LOG_LEVEL and configures slog over a zerolog backend; logs go to stderr so CLI stdout carries only the JSON response

---

## Panic Recovery (API and CLI)

Catches panics, logs stack traces, and returns safe error responses.

**Flow:**

1. `src/presentation/middleware/panicHandler.go` — `ApiPanicHandler` is Echo middleware that catches panics, writes stack traces to `logs/panic.log`, filters domain-layer frames, and returns HTTP 500 with masked error for untrusted clients; `CliPanicHandler` does the same for CLI via `defer`

---

## Transient Key-Value Store

Keeps ephemeral key-value pairs in a shared in-memory SQLite database. The data vanishes when the process ends.

**Flow:**

1. `src/infra/db/transientDatabaseService.go` — `NewTransientDatabaseService` opens the shared in-memory database and migrates the `KeyValue` model; `Has`, `Read`, and `Set` check, fetch, and upsert entries; `Read` returns `ErrKeyNotFound` for a missing key

---
