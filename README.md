# [Infinite Toolkit _(TK)_](https://github.com/goinfinite/tk) &middot; [![/r/goinfinite](https://img.shields.io/badge/%2Fr%2Fgoinfinite-FF4500?logo=reddit&logoColor=ffffff)](https://www.reddit.com/r/goinfinite/) [![Discussions](https://img.shields.io/badge/discussions-751A3D?logo=github)](https://github.com/orgs/goinfinite/discussions) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=goinfinite_tk&metric=alert_status)](https://sonarcloud.io/project/overview?id=goinfinite_tk) [![License](https://img.shields.io/badge/license-MIT-teal.svg)](https://github.com/goinfinite/tk/blob/main/LICENSE.md)

Infinite Toolkit _(TK)_ is a Clean Architecture toolkit for Go. It provides validated value objects, infrastructure helpers for files, shell, network, crypto, and databases, and presentation utilities for API and CLI input and output.

While developed primarily for Infinite ecosystem projects, this open-source library is available for general use under the MIT license.

If you're looking for UI components, please refer to the [Infinite UI](https://github.com/goinfinite/ui) repository.

> [!TIP]
> **Working with an AI agent?** Point it to [`SKILL.md`](SKILL.md) before it writes code that imports TK. The skill maps the three layers and routes the agent to the right layer README. Installed projects find the same file at `$(go env GOMODCACHE)/github.com/goinfinite/tk@<version>/SKILL.md`.

> [!IMPORTANT]
> **Human Reviewed**: AI models assist development, but senior developers review every line for coherence, readability, and maintainability.

## Installation

TK requires Go 1.27.1 or later. Install it with:

```bash
go get github.com/goinfinite/tk
```

See [CHANGELOG.md](CHANGELOG.md) for release history.

## Usage

TK is a toolkit, not a framework. Use only the components you need. They are organized in three Clean Architecture layers, each documented in its own README:

- **[Domain](src/domain/README.md)** — validated value objects, entities, DTOs, repository interfaces, use cases, and the Activity Record Management subsystem.
- **[Infrastructure](src/infra/README.md)** — file, shell, network, crypto, and database helpers, plus repository implementations.
- **[Presentation](src/presentation/README.md)** — request parsers, response wrappers, and Echo middleware.

Each layer README lists its components and shows usage snippets.
