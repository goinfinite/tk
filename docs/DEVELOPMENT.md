---
readWhen: always
description: Architecture, development commands, environment variables, and testing conventions.
---

# Development Guide

## Workflow

When working on this project, read the appropriate documentation for your current phase:

- **Planning phase**: Read `docs/DEV-PLANNING.md`
- **Implementation phase**: Read `docs/DEV-IMPLEMENTATION.md`

## Architecture

### Clean Architecture Layers

The codebase follows Clean Architecture (Domain-Driven Design):

```
src/
├── domain/          # Business logic (entities, use cases, DTOs, value objects, repository interfaces)
├── infra/           # Infrastructure implementations (database, external APIs, repositories)
└── presentation/    # HTTP/API layer (controllers, middleware, routing, liaisons)
```

**Key Principles**:

- Domain layer defines repository interfaces; infra layer implements them
- Dependencies point inward: presentation → domain ← infra
- Use cases (`domain/useCase/`) orchestrate business logic
- DTOs (`domain/dto/`) are temporary structs for transferring data between layers
- Entities (`domain/entity/`) are persistent structs representing stored data
- Value objects (`domain/valueObject/`) enforce domain constraints through validation
- Common domain errors defined in `src/domain/useCase/index.go`

### Presentation Layer Structure

- **API** (`src/presentation/api/`): REST API controllers and routing
- **CLI** (`src/presentation/cli/`): Command-line interface
- **UI** (`src/presentation/ui/`): Web interface
- **Liaisons** (`src/presentation/liaison/`): Bridge between presentation sub-layers (API, UI) and use cases
  - Handle untrusted input validation
  - Instantiate repositories and use cases
  - Transform use case results into liaison responses
  - Example: `AccountLiaison` validates input, calls `ReadAccount` use case, returns `LiaisonResponse`
- **Middleware** (`src/presentation/api/middleware/`): Request processing (auth, database injection, headers)
- **Init** (`src/presentation/init/`): Shared initialization for presentation sub-layers (database services, etc.)

### Infrastructure Layer

- **Repositories** (`src/infra/`): Implement domain repository interfaces
  - Query repos: Read operations
  - Cmd repos: Write operations (with run mode support)
- **Database** (`src/infra/db/`): Database models and services (GORM + SQLite)
- **Helpers** (`src/infra/helper/`): External integrations and utilities
- **Envs** (`src/infra/envs/`): Environment variable constants

## Development Commands

<specific-to-goinfinite-tk>
In most Infinite projects, the following instructions will be true. In this project, however, `templ`, `air`, and `swag` are not used.
</specific-to-goinfinite-tk>

### Build

```bash
make build
```

Generates templ files and builds the binary to `./bin/name-of-the-project` (CGO_ENABLED=0, linux/amd64).

### Development Server

```bash
make dev
# or
air serve
```

Uses Air for live reloading during development. Configuration in `.air.toml`.

### Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./src/infra/account/...

# Run a specific test
go test -run TestFunctionName ./src/infra/account/...

# Run tests with verbose output
go test -v ./...
```

### Swagger Documentation

```bash
make swag
```

Parses Swagger annotations in controller files (`src/presentation/api/controller/*.go`) and generates documentation in `src/presentation/api/docs`. The `--pdl 3` flag gathers external components from `goinfinite/tk` and `goinfinite/ui` packages.

**Access**: When the project is running, visit [`https://localhost:PROJECT_PORT/api/swagger/`](https://localhost:PROJECT_PORT/api/swagger/)

**Reference**: [Swaggo attribute documentation](https://github.com/swaggo/swag#attribute)

## Testing Conventions

- Test files use `_test.go` suffix
- Tests excluded from Air builds (see `.air.toml`)
- Repository tests verify both query and command operations
- Value object tests validate constraints and parsing

## Environment Variables

### Required

None as of now.

### Optional

None as of now.

### Development-Only

None as of now.
