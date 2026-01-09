---
readWhen: writing or editing code
description: Code style, Go-specific rules, and testing guidelines.
---

# Implementation Guidelines

## Before Starting

1. **Check branch**: Ensure you're on an appropriate branch (`feat-*`, `fix-*`, `refactor-*`, `docs-*`, etc.)
2. **Sync dependencies**: Run `go get -u` and `go mod tidy` to ensure go.mod matches source code
3. **Save implementation plan**: Create a plan document in `docs/history/` directory before starting work

## Critical Rules

- **NEVER** use 'else' statements (except in UI layer)
- **NEVER** use single-letter variable names (except "t" for test cases)
- **ALWAYS** read the `.context.md` file in any directory you are editing
- **ALWAYS** save implementation plans to `docs/history/` directory before starting work

## Naming Conventions

### Variable Names

- Variable names MUST convey the intention or purpose rather than describing their content
- Variable names SHOULD reflect the primary flow, not conditional outcomes
- Variable names MUST be descriptive and avoid generic names like "cert", "data", "result"
- Single-word variable names are usually too generic and SHOULD be avoided (e.g., use `dsaParameters` instead of `params`, `ellipticCurve` instead of `curve`, `decodedPemBlock` instead of `block`)
- Exceptions to single-word rule: standard Go conventions like `err` for errors, test variable `t`, and when the single word clearly conveys purpose in context
- Optional pointer variables MUST have "Ptr" suffix (e.g., `organizationPtr *X509Organization`)
- Avoid ambiguous variable names - qualify with context when needed (e.g., `stdlibCert` not `parsedCert` to clarify it's from Go stdlib)
- Create descriptive intermediate variables for mysterious commands (e.g., `sha256FingerprintHex := hex.EncodeToString(sha256HashBytes[:])` instead of inline)
- Create descriptive boolean variables for non-self-explanatory conditions (e.g., `caHasMaxPathLengthConstraint := stdlibCert.MaxPathLen >= 0`)

### Boolean Variables

- Boolean variables SHOULD start with "Is", "Should", "Has" etc. prefixes
- Boolean names MUST express intention/purpose, not describe content (e.g., `caHasMaxPathLengthConstraint` not `maxPathLengthIsSet`)

### Method Names

- Method names SHOULD focus on the conceptual purpose rather than implementation details
- Prefer "Read" prefix or "Factory" suffix instead of generic "Get" suffix
- Prefer suffixes instead of prefixes for struct and method names to preserve alphabetical order context
- Use mechanism-describing nouns like Builder, Parser, or Factory as method name suffixes

### Error Names

- Prefer "Fail", "Error", "Invalid" suffixes instead of "FailedTo", "Cannot", "UnableTo" prefixes
- Examples: "ParseCertificateFailed" not "FailedToParseCertificate"

### Struct Conventions

- Avoid redundant prefixes in struct field names when the context is already clear
- Struct fields SHOULD be ordered by importance, followed by alphabetical order
- Struct required fields SHOULD be placed before optional (pointer) fields
- Structs and types SHOULD NOT be exposed at package level unless they are used by multiple methods across different files or layers
- Prefer defining auxiliary types inside methods when they are only used within that method's scope

### Named Return Values

- Use purposeful named return values whenever the method returns multiple values

## Code Formatting

- Avoid comments unless strictly necessary
- ALWAYS attempt to keep lines under 85 characters when possible, exceeding only when not possible

## Git Workflow

- Use conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`, etc.
- Branch names follow the same prefixes: `feat-*`, `fix-*`, `refactor-*`, `docs-*`, `test-*`, `chore-*`
- Commit messages should be short - a single phrase with a few words
- If you need to describe too much, you should have committed earlier

## Go(lang) Specific Rules

### Logging

- Prefer slog.Error or slog.Debug instead of log.Printf depending on the gravity of the log
- `slog.Any` MUST NOT be used with `slog.Error` (user-facing logs); allowed only with `slog.Debug`

### Code Organization

- Auxiliary methods SHOULD be ordered according to their appearance/call order in the main method
- When using struct constructors (New), use multiple arguments per line
- Sequential method parameters of the same type SHOULD be grouped on the same line
- Avoid unnecessary line breaks, especially in simple struct initializations that fit on a single line

## Testing Rules

- Value objects, infrastructure, and use cases with complex logic **MUST** have unit tests
- Unit tests SHOULD use testCases as much as possible
- Unit tests error messages SHOULD be descriptive and provide context about what operation failed
- Use `t.Fatalf` to interrupt tests immediately on critical errors (setup failures, unexpected errors, assertion failures)
- Use `t.Errorf` only in for loops or situations where the test should continue running after an error
- When an error prevents the test from proceeding meaningfully, always use `t.Fatalf` instead of `t.Errorf`
- **NEVER** hardcode certificates, private keys, or secrets in test files - use commands (e.g., `openssl`) to generate them at test time or analyze existing test data

## Layer-Specific Rules

Layer-specific rules are in the `.context.md` file of each directory. Always read them before editing.
