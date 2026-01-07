---
readWhen: planning or discussing implementation strategies
description: Guidelines for planning implementations, creating PRs, and documenting changes.
---

# Planning Guidelines

## Critical Rules

- **NEVER** create features that weren't explicitly requested
- **ALWAYS** question ambiguity - it generates doubts about developer intention
- **ALWAYS** document the "before" and "after" states explicitly in the plan

## Planning Thought Process

Before writing any code, go through this mental checklist:

### 1. Understand Current State ("Before")

- What exists today? How does it work?
- Which limitations or problems exist with the current approach?
- What code/files will be affected?
- Read relevant `.context.md` files for business rules

### 2. Define Target State ("After")

- What behavior and capability changes will users/developers see?
- What will become possible?
- Write this explicitly: "After completion, users will be able to..."

### 3. Identify the Delta

- What exactly needs to change to get from "before" to "after"?
- Which layers are affected? (domain → infra → presentation)
- What are the dependencies between changes?
- Are there breaking changes? Migration needs?

### 4. Assess Complexity

- Is this a single coherent change or multiple independent changes?
- Can this be done in one PR or does it need phases?
- **Rule of thumb**: If a plan exceeds ~15 files or ~1000 lines, consider splitting it
- If the plan feels too complex to explain simply, it should be divided

## Creating Implementation Plans

### Plan Structure

Each plan should contain:

```markdown
## Goal

[One sentence describing what this achieves]

## Current State (Before)

[How things work today, what limitations exist]

## Target State (After)

[What will be possible after completion, what changes users/developers will see]

## Affected Areas

- [ ] Domain: [entities, DTOs, value objects, use cases affected]
- [ ] Infrastructure: [repositories, external APIs affected]
- [ ] Presentation: [controllers, liaisons, middleware affected]

## Implementation Phases

[If multiple phases needed, list them with dependencies]

## Phase N: [Name]

- Files to create/modify: [list]
- Estimated changes: [X files, ~Y lines]
- Dependencies: [what must be done first]
- Tests: [what tests will be added]
```

### Splitting Complex Plans

When a plan is too large:

1. Identify natural boundaries (by layer, by feature, by entity)
2. Create separate plans that can be reviewed independently
3. Document dependencies between plans
4. Each plan should leave the codebase in a working state

## Plan Storage

Completed plans are archived in `docs/history/` (create if it doesn't exist) with naming convention:

```
docs/history/X-Y-Z-phase-N-short-description.md
```

Where:

- `X-Y-Z` is the target semver version
- `phase-N` is the phase number (for multi-phase implementations)

Examples:

- `0-0-2-phase-1-add-user-entity.md`
- `0-0-2-phase-2-add-user-repository.md`
- `0-0-3-phase-1-refactor-authentication.md`

The plan file should include the date inside the document.

This provides:

- Audit trail of architectural decisions
- Learning resource for understanding why things were built a certain way
- Reference for similar future implementations
- Version history tracking

## Updating Documentation

After completing an implementation, update documentation to reflect changes:

**`README.md`** (user-facing):

- New features or capabilities
- Changed behavior
- New integration options

**`docs/DEVELOPMENT.md`** (developer-facing):

- New entities added
- New dependencies added to `go.mod`
- New environment variables
- Architecture changes

## Versioning

Each PR increments the project version (semver):

- `0.0.1` → `0.0.2` → `0.0.3` etc.

Update these files:

- `src/infra/envs/envs.go`: Update project version constant
- `src/presentation/api/api.go`: Update project version on swagger info
- `CHANGELOG.md`: Add entry describing what changed in this version

## General Planning Rules

- Prefer native methods over third-party libraries when possible
- During delete operations, validate all constraints upfront rather than discovering them mid-operation
- Consider backward compatibility and migration paths
