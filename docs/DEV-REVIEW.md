---
readWhen: reviewing code after implementation
description: Code review process including self-review and CodeRabbit integration.
---

# Code Review Guidelines

## Context Files to Consider

When reviewing code, verify compliance with rules defined in:

- `docs/DEV-IMPLEMENTATION.md` - Code style, naming conventions, Go-specific rules, testing requirements
- `docs/DEVELOPMENT.md` - Architecture, layer structure, dependency rules
- `.context.md` files in edited directories - Layer-specific rules and constraints

## Review Process

### 1. Self-Review Checklist

Before running external review tools, perform a manual self-review:

#### Rules Compliance

- [ ] Re-read all sections in `docs/DEV-IMPLEMENTATION.md`
- [ ] Verify all rules from that document are followed throughout the code
- [ ] Check layer-specific `.context.md` rules are followed

#### Additional Checks

- [ ] No inline type conversions that should be value object methods
- [ ] Variable names don't collide with package names in scope (e.g., `x509CertEntity` not `x509Cert` when using `x509` package)
- [ ] No package-level auxiliary functions that should be value object constructors (locality of behavior)
- [ ] Incomplete implementations have TODO comments or are fully implemented

#### Documentation

- [ ] Implementation plan saved to `docs/history/` before work started
- [ ] Create or update error correction document in `docs/history/` directory in the format `YYYY-MM-DD-error-correction.md` with a simple list of common agent errors that must be used to create or improve rules in `docs/DEV-IMPLEMENTATION.md` and `docs/DEV-REVIEW.md`.

#### Build Verification

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] No linter warnings
- [ ] `go mod tidy` completed successfully

### 2. CodeRabbit Review

After self-review passes, run CodeRabbit locally:

```bash
coderabbit review --config docs/DEV-REVIEW.md --prompt-only 2>&1
```

### 3. Issue Resolution Workflow

For each issue flagged by CodeRabbit:

1. **Evaluate validity** - Determine if the issue is a real problem or a false positive
2. **Test-first approach** - If valid, write a failing unit test that exposes the issue
3. **Fix the code** - Implement the fix to make the test pass
4. **Re-run review** - Confirm the issue is resolved

This test-first approach ensures:

- The issue is real and reproducible
- We have regression prevention for the future
- The fix actually addresses the problem
