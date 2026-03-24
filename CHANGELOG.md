# Changelog

```log
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
