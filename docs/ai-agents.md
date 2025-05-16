# General Rules

- NEVER create new features that weren't explicitly requested, even if they seem like useful additions.
- Value objects, infrastructure and use cases (with complex logic) MUST have unit tests.
- Unit tests SHOULD use testCases as much as possible.

# Code Style Rules

- Use clear names for functions and use named return values.
- NEVER use 'else' statements unless it's the UI layer.
- NEVER use single letter variable names. Use descriptive names, but avoid long names.
- Use Ptr suffix on variables when parsing optional fields (usually pointers on DTOs).
- Avoid comments unless strictly necessary.
- Prefer value objects as custom primitive types rather than structs when possible.
- Use PascalCase format for the entire error message whenever possible.
- Prefer "Fail", "Error", "Invalid" suffixes instead of "FailedTo", "Cannot", "UnableTo" prefixes.
- Prefer "Read" prefix or "Factory" suffix instead of "Get" suffix (depending on context).

# Golang Specific Rules

- Prefer using slog.Error or slog.Debug instead of log.Printf depending on the gravity of the log.
- Value objects accept interface{}/any directly without pre-conversion.
- When using constructors, don't use one line per argument.
