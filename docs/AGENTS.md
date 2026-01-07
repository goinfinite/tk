---
readWhen: always
description: Entrypoint for AI agents. Read this first, then load context-specific docs as needed.
---

# Infinite Standard Agents Guidelines

## @source [tk/docs/AGENTS.md](https://github.com/goinfinite/tk/blob/main/docs/AGENTS.md)

## Quick Start

1. **Always read this file first** - it's the entrypoint for all agent workflows
2. **Read `.context.md`** in any directory you're about to edit
3. **Load phase-specific docs** based on your current task (see below)

## Workflow Phases

### Planning Phase

**When:** Discussing features, creating implementation plans, reviewing PRs

**Read:** `docs/DEV-PLANNING.md`

Contains:

- Planning thought process (before/after states, identifying delta)
- Plan structure template
- Guidelines for splitting complex plans
- Plan archival in `docs/history/`

### Implementation Phase

**When:** Writing or editing code

**Read:** `docs/DEV-IMPLEMENTATION.md`

Contains:

- Code style rules
- Go-specific rules
- Testing requirements

## Critical Rules (Always Apply)

- **NEVER** create features that weren't explicitly requested
- **NEVER** add dependencies unless explicitly told by the developer
- **ALWAYS** question ambiguity - it generates doubts about developer intention
- **ALWAYS** read the `.context.md` file in any directory you are editing

## Documentation Structure

```
docs/
├── AGENTS.md              # This file - entrypoint
├── DEV-PLANNING.md        # Planning guidelines
├── DEV-IMPLEMENTATION.md  # Code style and testing rules
├── DEVELOPMENT.md         # Architecture and environment setup
└── history/               # Archived plans (do not read unless actively working on that plan)
```

**Note:** `docs/history/` contains archived implementation plans. Only read a plan file when you are actively working on implementing that specific plan.

## Directory Context Files

Every directory may contain a `.context.md` file with:

- Layer-specific rules
- Business logic constraints
- Project-specific conventions

**Always read before editing files in that directory.**

## Continuous Improvement

When you learn something new that should be permanently remembered for this project:

- **Planning learnings**: Update `docs/DEV-PLANNING.md`
- **Implementation learnings**: Update `docs/DEV-IMPLEMENTATION.md`
- **Architecture learnings**: Update `docs/DEVELOPMENT.md`
- **Layer-specific learnings**: Update the relevant `.context.md` file

This ensures our documentation improves as we work together, making future sessions more effective.
