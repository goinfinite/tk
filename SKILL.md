---
name: tk-usage
description: Use when working in a Go project that imports github.com/goinfinite/tk — presents the three layers, routes you to the right layer README, and states the usage conventions.
version: 1.0.0
lastUpdated: 2026-09-10
---

## Purpose

Infinite Toolkit _(TK)_ is the shared foundation for Go projects in the Infinite ecosystem. It provides a Clean Architecture toolkit: validated value objects, infrastructure helpers for files, shell, network, crypto, and databases, and presentation utilities for API and CLI input and output. The components are tested and shared across projects, so reusing them beats reimplementing them. Before you write a validator, a helper, or a primitive, check whether TK already provides it. This skill maps the three layers and points you to the layer README that documents each component with a usage snippet.

## Procedure

### 1. Locate the installed module

Run this command from the project root:

```sh
go list -m -f '{{.Dir}}' github.com/goinfinite/tk
```

The output is the module directory inside the Go module cache. Read the source there. Do not copy files into the project. Do not edit the cache.

### 2. Learn what each layer offers

TK follows Clean Architecture. Each layer has a README with the full component list and a usage snippet per component:

- **Domain** (`src/domain/README.md`) — validated value objects (for example IpAddress, UnixFilePath, ActivityRecordCode), entities, DTOs, repository interfaces, use cases, and the Activity Record Management subsystem. Start here for types, validation, and business rules.
- **Infrastructure** (`src/infra/README.md`) — file operations, shell execution, DNS lookup, encryption, random and certificate generation, server IP reading, GORM pagination, and the trail and transient database services. Start here for I/O and external systems.
- **Presentation** (`src/presentation/README.md`) — HTTP request input reading, pagination, time, and string-slice parsers, API and CLI response wrappers, and Echo middleware. Start here for input parsing and output formatting.

### 3. Open the README of the layer you need

Pick the layer that matches the task, then open its README and follow its component list:

- Modeling or validating a domain value → `src/domain/README.md`
- Touching the OS, network, files, crypto, or a database → `src/infra/README.md`
- Parsing request input or shaping a response → `src/presentation/README.md`

Most tasks touch one layer, so read only the READMEs that apply.

### 4. Read deeper only when needed

- `docs/FEATURE-MAP.md` — end-to-end flows when the task spans a whole feature.
- `.context.md` in the directory you import from — file-level constraints.
- `README.md` — the project overview and layer index.

The module directory matches the version in `go.mod`, so its documentation matches the API you compile against.

### 5. Follow the usage conventions

- Convert untrusted data into a value object before business logic. Request input, environment variables, file content, and database rows all count as untrusted.
- Compare sentinel errors with `errors.Is`. A helper returns the error; the caller decides to stop or log it.
- Keep dependencies pointing inward: `presentation -> infra -> domain`. The domain layer never imports infra or presentation.
- Use the import aliases already present in the project. The ecosystem convention prefixes the package: `tkValueObject`, `tkInfra`, `tkInfraDb`, `tkInfraDbModel`, `tkPresentation`, `tkDto`, `tkEntity`, `tkRepository`, `tkUseCase`, `tkVoUtil`.

## Guardrails

- Never copy TK code into the project. Import it.
- Never edit files inside the Go module cache. The cache is read-only input.
- Do not add a third-party dependency for a primitive TK already provides.
- Do not reimplement validation that a value object constructor already enforces.
- Check `go.mod` for the imported version before using a recently added component. The module cache holds the exact version you compile against.
- Do not import a TK package from the wrong layer. For example, domain code must not import `src/infra`.
- Do not read all three layer READMEs for a single-component task. Read only the layer you need.
