# Infinite Standard Agents Guidelines

## @source [tk/docs/AGENTS.md](https://github.com/goinfinite/tk/blob/main/docs/AGENTS.md)

## General Rules

- NEVER create features that weren't explicitly requested, even if they seem like useful additions.
- Ambiguity SHOULD always be questioned as it generates doubts about the intention of the developer.
- Value objects, infrastructure and use cases (with complex logic) MUST have unit tests.
- Unit tests SHOULD use testCases as much as possible.
- During delete operations, you MUST validate all constraints upfront rather than discovering them mid-operation.
- Prefer native methods over third-party libraries when possible.

## Code Style Rules

- Avoid comments unless strictly necessary.
- Method names SHOULD focus on the conceptual purpose rather than implementation details.
- NEVER use 'else' statements unless it's the UI layer.
- NEVER use single letter variable names. Use clear and descriptive names, but avoid long names.
- Variables names MUST convey the intention or purpose rather than describing their content, avoiding unnecessary synonyms.
- Variable names SHOULD reflect the primary flow, not conditional outcomes.
- Use purposeful named return values whenever the method returns multiple values.
- Use Ptr suffix on variables when parsing optional fields (usually pointers on DTOs).
- Prefer value objects as custom primitive types rather than structs when possible.
- Boolean variables SHOULD start with "Is", "Should", "Has" etc prefixes.
- Use PascalCase format for the entire error message whenever possible.
- Prefer "Fail", "Error", "Invalid" suffixes instead of "FailedTo", "Cannot", "UnableTo" prefixes.
- Prefer "Read" prefix or "Factory" suffix instead of "Get" suffix (depending on context).
- Prefer suffixes instead of prefixes for struct and method names to preserve alphabetical order context.
- Avoid redundant prefixes in struct field names when the context is already clear from the struct name or surrounding code.
- Struct fields SHOULD be ordered by importance, followed by alphabetical order.
- Struct required fields SHOULD be placed before optional (pointer) fields.
- Watch for repeated creation of identical structs or configurations inside loops to avoid unnecessary allocations and improve performance.
- When logging errors in loops, always include identifying information (like, but not limited to, IDs) about the specific iteration to aid in debugging and operational visibility.

## Go(lang) Specific Rules

- Prefer using slog.Error or slog.Debug instead of log.Printf depending on the gravity of the log.
- Value objects accept interface{}/any directly without the need for pre-assertion.
- When using struct constructors (New), use multiple arguments per line.
- Sequential method parameters of the same type SHOULD be combined together on the same line.
- When a use case or utility has multiple related auxiliary functions, they SHOULD be grouped under a struct.
- Auxiliary methods SHOULD be ordered according to their appearance/call order in the main method, not alphabetically or by perceived importance.
