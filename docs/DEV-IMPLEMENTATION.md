---
readWhen: writing or editing code
description: Code style, Go-specific rules, and testing guidelines.
---

# Implementation Guidelines

## Before Starting

1. **Check branch**: Ensure you're on an appropriate branch (`feat-*`, `fix-*`, `refactor-*`, `docs-*`, etc.)
2. **Sync dependencies**: Run `go get -u` and `go mod tidy` to ensure go.mod matches source code

## Critical Rules

- **NEVER** use 'else' statements (except in UI layer)
- **NEVER** use single-letter variable names (except "t" for test cases)
- **ALWAYS** read the `.context.md` file in any directory you are editing
- **ALWAYS** attempt to keep lines under 85 characters when possible, exceeding only when not possible

## Git

- Use conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`, etc.
- Branch names follow the same prefixes: `feat-*`, `fix-*`, `refactor-*`, `docs-*`, `test-*`, `chore-*`
- Commit messages should be short - a single phrase with a few words
- If you need to describe too much, you should have committed earlier

## Code Style Rules

- Avoid comments unless strictly necessary
- Method names SHOULD focus on the conceptual purpose rather than implementation details
- Variable names MUST convey the intention or purpose rather than describing their content
- Variable names SHOULD reflect the primary flow, not conditional outcomes
- Use purposeful named return values whenever the method returns multiple values
- Boolean variables SHOULD start with "Is", "Should", "Has" etc. prefixes
- Prefer "Fail", "Error", "Invalid" suffixes instead of "FailedTo", "Cannot", "UnableTo" prefixes
- Prefer "Read" prefix or "Factory" suffix instead of generic "Get" suffix
- Prefer suffixes instead of prefixes for struct and method names to preserve alphabetical order context
- Use mechanism-describing nouns like Builder, Parser, or Factory as method name suffixes
- Avoid redundant prefixes in struct field names when the context is already clear
- Struct fields SHOULD be ordered by importance, followed by alphabetical order
- Struct required fields SHOULD be placed before optional (pointer) fields

## Go(lang) Specific Rules

- Prefer slog.Error or slog.Debug instead of log.Printf depending on the gravity of the log
- `slog.Any` MUST NOT be used with `slog.Error` (user-facing logs); allowed only with `slog.Debug`
- When using struct constructors (New), use multiple arguments per line
- Sequential method parameters of the same type SHOULD be grouped on the same line
- Auxiliary methods SHOULD be ordered according to their appearance/call order in the main method
- Avoid unnecessary line breaks, especially in simple struct initializations that fit on a single line

## Testing Rules

- Value objects, infrastructure, and use cases with complex logic **MUST** have unit tests
- Unit tests SHOULD use testCases as much as possible
- Unit tests error messages SHOULD be descriptive and provide context about what operation failed
- Use `t.Fatalf` to interrupt tests immediately on critical errors (setup failures, unexpected errors, assertion failures)
- Use `t.Errorf` only in for loops or situations where the test should continue running after an error
- When an error prevents the test from proceeding meaningfully, always use `t.Fatalf` instead of `t.Errorf`

## Layer-Specific Rules

Layer-specific rules are in the `.context.md` file of each directory. Always read them before editing.
