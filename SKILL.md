---
name: tk-usage
description: Use when working in a Go project that imports github.com/goinfinite/tk — presents the three layers, routes you to the right layer README, and states the usage conventions.
version: 1.0.1
lastUpdated: 2026-09-11
---

## Purpose

Infinite Toolkit _(TK)_ is the shared foundation for Go projects in the Infinite ecosystem. It provides a Clean Architecture toolkit of validated value objects, infrastructure helpers, and presentation utilities. The components are tested and shared across projects. Reusing them beats reimplementing them. Before you write a validator, a helper, or a primitive, check whether TK already provides it. This skill maps the three layers and points you to the layer README that documents each component.

## Procedure

### 1. Locate the installed module

Run this command from the project root:

```sh
go list -m -f '{{.Dir}}' github.com/goinfinite/tk
```

The output is the module directory inside the Go module cache. Read the source there. Do not copy files into the project. Do not edit the cache.

### 2. Pick the layer and open its README

TK follows Clean Architecture. Each layer README lists the layer's components and shows usage snippets. Pick the layer that matches the task:

- Modeling or validating a domain value → `src/domain/README.md`
- Touching the OS, network, files, crypto, or a database → `src/infra/README.md`
- Parsing request input or shaping a response → `src/presentation/README.md`

Most tasks touch one layer, so read only the READMEs that apply.

### 3. Read deeper only when needed

These documents describe TK's internals. Read them only when the task reaches the matching subsystem:

- `docs/FEATURE-MAP.md` — end-to-end flows for a TK feature.
- `.context.md` in the package you import — constraints on that TK package's files.
- `README.md` in the module root — the project overview and layer index.

The module directory matches the version in `go.mod`, so its documentation matches the API you compile against.

### 4. Follow the usage conventions

- Convert untrusted data into a value object before business logic. Request input, environment variables, file content, and database rows all count as untrusted.
- Compare sentinel errors with `errors.Is`. A helper returns the error. The caller decides to stop or log it.
- Keep dependencies pointing inward: `presentation -> infra -> domain`. Infra is the shared helper layer, so presentation may import it. The domain layer never imports infra or presentation.
- Use the import aliases already present in the project. The ecosystem convention prefixes the package: `tkValueObject`, `tkInfra`, `tkInfraDb`, `tkInfraDbModel`, `tkPresentation`, `tkDto`, `tkEntity`, `tkRepository`, `tkUseCase`, `tkVoUtil`.

## Guardrails

- Never copy TK code into the project. Import it.
- Never edit files inside the Go module cache. The cache is read-only input.
- Do not add a third-party dependency for a primitive TK already provides.
- Do not reimplement validation that a value object constructor already enforces.
- Do not import a TK package from the wrong layer. For example, domain code must not import `src/infra`.
- Do not read all three layer READMEs for a single-component task. Read only the layer you need.
