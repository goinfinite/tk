# Skills

Agentic workflows distributed via Infinite Toolkit (TK) for use in dependent projects. These skills produce deterministic artifacts (shell scripts, configs) that can run independently once generated.

Agents working on projects that import TK can reference these skills directly from the Go module cache — no copying required.

## Available Skills

- `openapi-test.md` — OpenAPI/Swagger spec testing with agent-assisted script generation

## Using These Skills

Skills live in the Go module cache. Reference them by version from your project's `go.mod`:

```sh
$(go env GOMODCACHE)/github.com/goinfinite/tk@vX.Y.Z/skills/openapi-test.md
```

Replace `vX.Y.Z` with the version your project imports. Generated artifacts (config files, test scripts) are created in your project directory — not in the TK module cache.

## When to Extract a Skill

Extract a skill when:

- A testing or automation procedure is complex enough that an agent needs explicit guidance to produce correct artifacts.
- A workflow must be repeatable across specs or environments without re-exploration.

Do not extract when the procedure is short and intuitive.

## File Naming

Lowercase, hyphenated: `openapi-test.md`, `db-migration.md`

## Schema (v0.0.2 // 2026-06-30)

### Frontmatter

- `shortDescription` (Required) — what the skill does in one sentence
- `version` (Required) — semantic version
- `lastUpdated` (Required) — last modification date

### Body

- `Purpose` (Required) — one paragraph on what problem this solves
- `Procedure` (Required) — numbered execution steps with artifact descriptions
- `Guardrails` (Optional) — skill-specific pitfalls and common mistakes
