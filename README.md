# [Infinite Toolkit _(TK)_](https://github.com/goinfinite/tk) &middot; [![/r/goinfinite](https://img.shields.io/badge/%2Fr%2Fgoinfinite-FF4500?logo=reddit&logoColor=ffffff)](https://www.reddit.com/r/goinfinite/) [![Discussions](https://img.shields.io/badge/discussions-751A3D?logo=github)](https://github.com/orgs/goinfinite/discussions) [![Report Card](https://img.shields.io/badge/report-A%2B-brightgreen)](https://goreportcard.com/report/github.com/goinfinite/tk) [![License](https://img.shields.io/badge/license-MIT-teal.svg)](https://github.com/goinfinite/tk/blob/main/LICENSE.md)

Infinite Toolkit _(TK)_ offers a comprehensive suite of core components for Infinite projects. The library includes value objects, utilities for Clean Architecture layers, service abstractions, and other foundational elements.

While developed primarily for Infinite ecosystem projects, this open-source library is available for general use under the MIT license.

If you're looking for UI components, please refer to the [Infinite UI](https://github.com/goinfinite/ui) repository.

> [!TIP]
> **Working with an AI agent?** Point it to [`skills/tk-usage/SKILL.md`](skills/tk-usage/SKILL.md) before it writes code that imports TK. Installed projects find the same file at `$(go env GOMODCACHE)/github.com/goinfinite/tk@<version>/skills/tk-usage/SKILL.md`.

> [!IMPORTANT]
> **Human Reviewed**: AI models assist development, but senior developers review every line for coherence, readability, and maintainability.

## Installation

To use Infinite Toolkit _(TK)_ in your project, you can install it using Go modules. Run the following command in your terminal:

```bash
go get github.com/goinfinite/tk
```

## Components

Infinite Toolkit _(TK)_ is organized in three Clean Architecture layers:

- **[Domain](src/domain/README.md)** — validated value objects, entities, DTOs, repository interfaces, use cases, and the Activity Record Management subsystem.
- **[Infrastructure](src/infra/README.md)** — file, shell, network, crypto, and database helpers, plus repository implementations.
- **[Presentation](src/presentation/README.md)** — request parsers, response wrappers, and Echo middleware.

Each layer README documents its components with usage snippets.

## Skills for Agents

TK provides reusable agent skills for common workflows — testing, automation, and operational tasks. These skills are designed for agents working on projects that import TK as a dependency.

**Available skills:**
- **OpenAPI Testing** — generates deterministic shell scripts from Swagger/OpenAPI specs. Agents explore endpoints interactively, validate payloads, and write working curl commands. Subsequent runs diff the spec against the script and test only what changed.
- **TK Usage** — orients agents in projects that import TK: presents the three layers, routes to the right layer README, and states the usage conventions.

**For agents:** Reference skills from the Go module cache at `$(go env GOMODCACHE)/github.com/goinfinite/tk@*/skills/<skill-name>/SKILL.md`. See `skills/README.md` in the TK source for full documentation.
