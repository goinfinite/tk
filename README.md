# [Infinite Toolkit _(TK)_](https://github.com/goinfinite/tk) &middot; [![/r/goinfinite](https://img.shields.io/badge/%2Fr%2Fgoinfinite-FF4500?logo=reddit&logoColor=ffffff)](https://www.reddit.com/r/goinfinite/) [![Discussions](https://img.shields.io/badge/discussions-751A3D?logo=github)](https://github.com/orgs/goinfinite/discussions) [![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=goinfinite_tk&metric=sqale_rating)](https://sonarcloud.io/project/overview?id=goinfinite_tk) [![License](https://img.shields.io/badge/license-MIT-teal.svg)](https://github.com/goinfinite/tk/blob/main/LICENSE.md)

Infinite Toolkit _(TK)_ offers a comprehensive suite of core components for Infinite projects. The library includes value objects, utilities for Clean Architecture layers, service abstractions, and other foundational elements.

While developed primarily for Infinite ecosystem projects, this open-source library is available for general use under the MIT license.

If you're looking for UI components, please refer to the [Infinite UI](https://github.com/goinfinite/ui) repository.

> [!TIP]
> **Working with an AI agent?** Point it to [`SKILL.md`](SKILL.md) before it writes code that imports TK. The skill maps the three layers and routes the agent to the right layer README. Installed projects find the same file at `$(go env GOMODCACHE)/github.com/goinfinite/tk@<version>/SKILL.md`.

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
