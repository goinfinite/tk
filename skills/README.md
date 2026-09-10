# Skills

Agentic workflows distributed via Infinite Toolkit (TK) for use in dependent projects. These skills guide agents working in dependent projects. Some generate deterministic artifacts (shell scripts, configs) that can run independently once generated; others orient agents before they write code.

Agents working on projects that import TK can reference these skills directly from the Go module cache — no copying required.

## Available Skills

- `openapi-test/SKILL.md` — OpenAPI/Swagger spec testing with agent-assisted script generation
- `tk-usage/SKILL.md` — orientation and component-selection guide for agents using TK in dependent projects

## Using These Skills

Skills live in the Go module cache. Reference them by version from your project's `go.mod`:

```sh
$(go env GOMODCACHE)/github.com/goinfinite/tk@vX.Y.Z/skills/openapi-test/SKILL.md
```

Replace `vX.Y.Z` with the version your project imports. Generated artifacts (config files, test scripts) are created in your project directory — not in the TK module cache.

## When to Extract a Skill

Extract a skill when:

- A testing or automation procedure is complex enough that an agent needs explicit guidance to produce correct artifacts.
- A workflow must be repeatable across specs or environments without re-exploration.

Do not extract when the procedure is short and intuitive.

## Directory Layout

Each skill lives in its own directory. The main file MUST be named `SKILL.md`, following the OpenCode standard:

```text
skills/
├── openapi-test/
│   └── SKILL.md
└── tk-usage/
    └── SKILL.md
```

Directory names are lowercase and hyphenated: `openapi-test`, `db-migration`. The directory name MUST match the skill's `name` field.

## Schema (v0.1.0 // 2026-09-10)

### Frontmatter

- `name` (Required) — lowercase, hyphen-separated skill identifier; MUST match the directory name
- `description` (Required) — one sentence covering what the skill does AND when to use it; write in third person and front-load trigger keywords
- `version` (Required) — semantic version
- `lastUpdated` (Required) — last modification date

### Body

- `Purpose` (Required) — one paragraph on what problem this solves
- `Procedure` (Required) — numbered execution steps with artifact descriptions
- `Guardrails` (Optional) — skill-specific pitfalls and common mistakes
