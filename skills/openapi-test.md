---
shortDescription: Tests OpenAPI/Swagger specs by generating and maintaining deterministic shell scripts.
version: 1.0.0
lastUpdated: 2026-06-30
---

## Purpose

Generates and maintains deterministic curl-based shell scripts that test OpenAPI endpoints. First run: explore and record. Subsequent runs: diff spec, retest only changes.

## Artifacts

Generated artifacts live in `tests/api/`, named after the spec's base filename:

- **Script:** `tests/api/<base>-test.sh` (committed)
- **Config:** `tests/api/.<base>-test.env` (gitignored — dot-prefixed)

Ensure `.gitignore` includes the pattern:

```bash
grep -qF 'tests/api/.*.env' .gitignore 2>/dev/null || echo 'tests/api/.*.env' >> .gitignore
```

### Config Format

Config fields are derived from the spec's auth endpoint schema. The agent generates a template on first run and asks the user to fill in values. Auth-derived fields follow `API_AUTH_<FIELD>` (uppercased auth parameter names). Other fields use `API_<PURPOSE>`. Example:

```env
API_BASE_URL=https://localhost:6674
API_AUTH_EMAIL=admin@example.com
API_AUTH_PASSWORD=changeme
API_SKIP_TLS_VERIFY=true
```

> `API_SKIP_TLS_VERIFY=true` disables certificate validation — use only for local/dev with self-signed certs. Never for production.

## Procedure

### Phase 1: Locate Swagger Files

Search the current project for OpenAPI/Swagger specs:

```bash
find . -type f \( \
  -name 'swagger.yaml' -o -name 'swagger.json' \
  -o -name 'openapi.yaml' -o -name 'openapi.json' \
\) 2>/dev/null | grep -v vendor | grep -v node_modules
```

If multiple specs are found, ask the user which one to test. If none, report and stop.

### Phase 2: Parse the Spec

Handle both Swagger 2.0 and OpenAPI 3.x field structures:

- **Schemas:** Swagger 2.0 → `definitions`; OpenAPI 3.x → `components.schemas`
- **Security schemes:** Swagger 2.0 → `securityDefinitions`; OpenAPI 3.x → `components.securitySchemes`
- **Request body:** Swagger 2.0 → `parameters` with `in: body`; OpenAPI 3.x → `requestBody.content`
- **Content types:** Swagger 2.0 → global `produces`/`consumes`; OpenAPI 3.x → per-operation `content`
- **`$ref` prefix:** Swagger 2.0 → `#/definitions/`; OpenAPI 3.x → `#/components/schemas/`
- **Auth tokens:** Swagger 2.0 → `securityDefinitions` with `type: apiKey` or `type: basic`; OpenAPI 3.x → `components.securitySchemes` with `type: http`, `type: apiKey`, or `type: oauth2`

When building the endpoint inventory, normalize across both versions:
- Resolve `$ref` pointers to their target schema regardless of prefix.
- For Swagger 2.0 `in: body` parameters, treat the `schema` as the request body. For OpenAPI 3.x, use `requestBody.content.<media-type>.schema`.
- For response schemas: Swagger 2.0 uses `responses.<code>.schema`; OpenAPI 3.x uses `responses.<code>.content.<media-type>.schema`.

Build an endpoint inventory from `paths`: for each `METHOD /path`, extract parameters (path, query, body/requestBody), required security, expected response codes, and response schemas.

Identify the auth mechanism from security definitions/schemes. Determine which endpoints require authentication and which are public.

### Phase 3: Read Existing Artifacts

If the test script exists, parse it to extract: which endpoints are already tested, the curl commands for each, and any helper functions. The parsing rules:

- Each endpoint is a function named after `operationId` in camelCase (or camelCase of operation summary if `operationId` is absent).
- Endpoint functions contain `curl` commands with the HTTP method, URL, headers, and body.
- The main flow at the bottom of the script calls these functions in dependency order — it reveals which endpoints are tested and in what sequence.
- Config values appear as `source "${scriptDir}/.<base>-test.env"` near the top.

If the config file exists, read it. If not, generate a template derived from the auth endpoint's required parameters and ask the user to fill in values.

### Phase 4: Diff and Sync

Parse the script to extract tested endpoints (function names map to `operationId`). Compare against the spec's endpoint inventory:

- **New:** `METHOD /path` in spec but not in script → needs exploration
- **Changed:** in both but parameters differ (names, types, required flags) → needs re-validation
- **Unchanged:** in both with matching parameters → reuse existing curl
- **Removed:** in script but not in spec → delete function

Re-test New and Changed. Keep Unchanged. Delete Removed. Reassemble and execute. When splitting across multiple scripts, diff each script against its tag/group subset independently.

### Phase 5: Resolve Config

Read the config file. Verify all required values are present and not set to placeholders. If any required value is missing, ask the user before proceeding.

### Phase 6: Discover and Authenticate

If no endpoints require authentication (all are public per the spec's security definitions), skip this phase and proceed to Phase 7.

Otherwise, find the auth endpoint. Look for public POST endpoints that return a token or session identifier. Check the spec's security schemes to understand the auth mechanism (Bearer token, API key, OAuth2, etc.). If the auth mechanism is not a POST endpoint or uses non-standard patterns (API key in header, OAuth2 client credentials), ask the user which endpoint handles auth and how the token is passed.

Build the login request using config credentials. Execute it and extract the token. Look for the token in common response fields: `token`, `access_token`, `data.token`, `result.token`, or the field indicated by the security scheme. If not found, ask the user.

All subsequent requests use the token per the spec's security scheme (typically `Authorization: Bearer ${authToken}` or `X-API-Key: ${apiKey}`).

If login fails, stop and report. The auth function is always called first in the main flow (not first in file — helpers and cleanup are defined above it).

Auth is infrastructure, not an endpoint test. It does not increment `passedCount`/`failedCount`.

Auth function template (Bearer token):

```sh
authenticate() {
  local httpStatus responseBody
  httpStatus=$(curl ${curlInsecure} --connect-timeout 5 --max-time 30 -S -s \
    -o responseBody.tmp -w "%{http_code}" -X POST \
    "${API_BASE_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${API_AUTH_EMAIL}\",\"password\":\"${API_AUTH_PASSWORD}\"}")
  responseBody=$(cat responseBody.tmp)
  rm -f responseBody.tmp
  if [ "$httpStatus" != "200" ]; then
    echo "FAIL: authentication — HTTP ${httpStatus}" >&2
    echo "$responseBody" >&2
    exit 1
  fi
  authToken=$(echo "$responseBody" | jq -r '.token // .data.token // .access_token // empty')
  if [ -z "$authToken" ]; then
    echo "FAIL: authentication — no token in response" >&2
    echo "$responseBody" >&2
    exit 1
  fi
  echo "PASS: authentication" >&2
}
```

Adapt the field names and token extraction path to match the spec. For API key auth, the function sets `apiKey` instead and subsequent requests use `-H "X-API-Key: ${apiKey}"`.

If the spec defines multiple security schemes (e.g., API key for reads, OAuth2 for writes), check which endpoints use which scheme. If endpoints use different schemes, generate separate auth functions (e.g., `authenticateWithApiKey`, `authenticateWithOAuth`) and call the appropriate one per endpoint. Document which credentials map to which scheme in the config file.

### Phase 7: Test Endpoints

For each new or changed endpoint:

#### Generate Payload

Build a request body from the schema. Respect `format`, `pattern`, `minLength`/`maxLength`, `minimum`/`maximum`, `enum`, `minItems`/`maxItems`, `additionalProperties`. Resolve `$ref` recursively. For `enum`, use the first allowed value. For `string` with no hints, use a realistic sample. Generate values for query and path parameters matching their types and constraints.

#### Handle Special Cases

- **Pagination:** include pagination params with defaults (page 1, 20 items). Test page 1 only — verify structure and pagination fields. Report missing pagination docs in Spec Feedback.
- **File uploads:** for `multipart/form-data`, use `curl -F "<fieldName>=@${filePath}"` where `<fieldName>` comes from the spec's parameter name for the operation. Clean up after the test.
- **Nested resources:** parent ID must come from a previous create or read.

#### Execute and Validate

Send the request via curl. Capture status code and response body. Validate:

- Status code matches the spec's documented success code for this operation
- Response body conforms to the spec's response schema (when documented)
- If the spec defines a response envelope, validate its structure

#### Iterate on Errors

On failure, inspect the response. A **failure** is: wrong HTTP status, schema validation error, or connection timeout. A **success** requires all three: correct status, valid schema, usable body.

Error iteration (apply in order):
- **401** → re-authenticate
- **404** → create dependency first
- **409** → use existing or clean up
- **422** → fix payload fields against spec
- **429** → exponential backoff (1s, 2s, 4s); max 3 retries
- **502/503** → transient; wait 2s, retry up to 3 times
- **Other 4xx** → inspect error response, adjust payload
- **Other 5xx** → log and mark FAILED

After fixing, regenerate the curl command. Do not retry unchanged.

Capture the working curl command as a shell function.

### Phase 8: Build and Execute Script

Assemble the script following the Shell Conventions above. Skeleton:

```sh
#!/usr/bin/env bash
# @description: Tests <spec-name> API endpoints
# @usage: ./<base>-test.sh
# @output: Pass/fail summary on stderr, IDs on stdout
# @requires: curl, jq
# @version: 1.0.0
# @updated: 2026-06-30
set -euo pipefail

command -v curl >/dev/null || { echo "curl is required" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

scriptDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${scriptDir}/.<base>-test.env"

curlInsecure=""
if [ "${API_SKIP_TLS_VERIFY:-false}" = "true" ]; then
  curlInsecure="--insecure"
fi

declare -a createdResourceIds=()

cleanupResources() {
  if [ ${#createdResourceIds[@]} -eq 0 ]; then return; fi
  echo "Cleaning up ${#createdResourceIds[@]} created resources..." >&2
  for resourceId in "${createdResourceIds[@]}"; do
    local curlArgs=(
      ${curlInsecure} --connect-timeout 5 --max-time 10 -S -s
      -o /dev/null -w "" -X DELETE
      "${API_BASE_URL}/<resource-path>/${resourceId}"
    )
    if [ -n "${authToken:-}" ]; then
      curlArgs+=(-H "Authorization: Bearer ${authToken}")
    fi
    if [ -n "${apiKey:-}" ]; then
      curlArgs+=(-H "X-API-Key: ${apiKey}")
    fi
    curl "${curlArgs[@]}" 2>/dev/null || \
      echo "WARN: failed to cleanup ${resourceId}" >&2
  done
}
trap 'cleanupResources' EXIT

# --- Auth ---
authenticate() { ... }

# --- Helpers ---
extractId() { jq -r '.id // .data.id // empty'; }

passedCount=0
failedCount=0
skippedCount=0
validateResponseStatus() {
  local endpointName="$1" expectedStatus="$2" httpStatus="$3" responseBody="$4"
  if [ "$httpStatus" != "$expectedStatus" ]; then
    echo "FAIL: ${endpointName} — expected ${expectedStatus}, got ${httpStatus}" >&2
    echo "$responseBody" | head -c 500 >&2
    ((failedCount++)) || true
    return 0
  fi
  echo "PASS: ${endpointName}" >&2
  ((passedCount++)) || true
}
validateResponseSchema() {
  local endpointName="$1" responseBody="$2" jqFilter="$3"
  if ! echo "$responseBody" | jq -e "$jqFilter" >/dev/null 2>&1; then
    echo "FAIL: ${endpointName} — schema validation failed (${jqFilter})" >&2
    echo "$responseBody" | head -c 500 >&2
    ((failedCount++)) || true
  fi
}
skipEndpoint() {
  local endpointName="$1" reason="$2"
  echo "SKIP: ${endpointName} — ${reason}" >&2
  ((skippedCount++)) || true
}

# --- Endpoint functions (grouped by tag, ordered: reads, creates, updates) ---
listResources() { ... }
getResource() { ... }
createResource() { ... }
updateResource() { ... }

# --- Main flow ---
authenticate
listResources
createResource
# updateResource (if applicable)

# --- Results ---
echo "---" >&2
echo "PASSED: ${passedCount} | FAILED: ${failedCount} | SKIPPED: ${skippedCount}" >&2
```

Fill in each `...` block per the endpoint function rules below. Every endpoint function follows the same pattern — only the URL, method, headers, body, and expected status differ.

Each endpoint function:

- Named after `operationId` in camelCase (fall back to operation summary when `operationId` is absent — use the first meaningful words, drop articles, camelCase: "Get all users" → `getAllUsers`)
- Sources path and query params from variables or previous function outputs
- Sends curl request, captures response
- Validates status code
- Extracts and echoes IDs needed by dependent endpoints to stdout
- Prints pass or fail with endpoint name to stderr (so it does not interfere with `$(fn)` capture)

Every endpoint function follows this canonical pattern:

```sh
endpointName() {
  local httpStatus responseBody
  httpStatus=$(curl ${curlInsecure} --connect-timeout 5 --max-time 30 -S -s \
    -o responseBody.tmp -w "%{http_code}" \
    -H "Authorization: Bearer ${authToken}" \
    "${API_BASE_URL}/<path>")
  responseBody=$(cat responseBody.tmp)
  rm -f responseBody.tmp
  validateResponseStatus "endpointName" "<expectedStatus>" "$httpStatus" "$responseBody"
}
```

Adapt per endpoint: add `-X METHOD` for POST/PUT/PATCH, add `-H "Content-Type: ..."` and `-d '<payload>'` for bodies, add `-F "<field>=@${filePath}"` for uploads. Extract IDs when needed:

```sh
  newResourceId=$(echo "$responseBody" | extractId)
  if [ -n "$newResourceId" ]; then
    createdResourceIds+=("$newResourceId")
    echo "$newResourceId"
  fi
```

Adapt URLs, methods, status codes, headers, and bodies per the spec. The `validateResponseStatus` helper prints PASS/FAIL to stderr and tracks counts.

Main flow respects dependencies:

1. Authenticate
2. Read existing resources (collect IDs)
3. Create resources (capture new IDs)
4. Update resources (using collected IDs)

Spec-defined DELETE endpoints are excluded from the main test flow (see Guardrails). The cleanup trap calls DELETE endpoints as infrastructure — they're test teardown, not test cases.

The trap handler (`trap 'cleanupResources' EXIT`) runs on any exit — success, failure, or signal. It uses the existing `authToken`; for short test runs this is sufficient. If cleanup 401s, the warning is logged and remaining resources continue cleanup.

Execute the script. If any endpoint fails, return to Phase 7 for that endpoint and iterate. After 3 failed iterations for the same endpoint, mark it as FAILED and continue with remaining endpoints.

Report: total endpoints tested, passed/failed/skipped counts, and for failures: endpoint name, expected status, received status, response excerpt.

During testing, collect spec issues and report them as a **Spec Feedback** section at the end — even if all tests pass. Categories:

- **schema mismatch** — response has fields not in spec, or spec fields missing from response
- **wrong type** — field type differs between spec and response
- **undocumented error** — API returns error codes not in spec's `responses`
- **missing description** — endpoint or field with no `description`
- **missing example** — no `example` values (informational only)
- **undocumented required field** — API rejects fields the spec marks optional
- **missing enum value** — API accepts values not in spec's `enum`
- **ghost endpoint** — spec endpoint returns 404/405

Format per endpoint: `### METHOD /path (N issues)` with `[category] description` lines.

If the spec has more than 30 endpoints, split the script by tag or resource group (e.g., `users-test.sh`, `subscriptions-test.sh`). Each script is self-contained with its own auth function and config sourcing. If a spec has no tags, group by path prefix (first two segments). Ask the user before splitting — some projects prefer a single script. The 30-endpoint threshold keeps scripts readable and debuggable; beyond that, failures are harder to isolate and execution time becomes unwieldy.

## Shell Conventions

### Header and Strictness

Scripts MUST include `@description`, `@usage`, `@requires`, `@version`, `@updated` header tags. MUST start with `set -euo pipefail`.

### Naming

Variables use camelCase. UPPER_SNAKE_CASE for env and config keys. Single-word variable and function names are rarely correct — use compound names (`responseBody`, `statusCode`, not `body`, `code`). Names convey WHAT, not HOW. Function names must specify what is validated (`validateResponseStatus`, not `validate`). Function names MUST NOT embed tool names (`readSubscriptionResponse`, not `parseCurlJsonOutput`).

### Structure

Functions have one responsibility. Ordered top-down: helpers first, main flow last. Every function appears above its first caller. The main flow reads as a sequence of function calls.

### Output and Pipelines

Functions return data via stdout — callers capture with `$(fn)`. Status and diagnostic messages (pass/fail, progress) go to stderr. Heredocs for multi-line output, `printf`/`echo` for single-line. Pipelines max 3 stages — intermediate results in variables.

### Control Flow

Scripts MUST NOT use `else` statements. Use early return or guard clauses instead. The failure path runs first and returns; the success path follows unindented. This produces linear, scannable code.

### Arithmetic Counters

`((count++))` returns exit code 1 when incrementing from 0 (bash treats the pre-increment value as the return code). Under `set -e`, this aborts the script. Always append `|| true`: `((count++)) || true`.

Validate captured output before piping downstream (non-empty, valid JSON). One parser per extraction. Using `jq -e` and `// empty` is sufficient — explicit empty-check guards are optional. Source all environment-specific values from config — never hardcode.

### Curl Defaults

Every `curl` command MUST include timeout flags:
- `--connect-timeout 5` — fail fast on unreachable hosts (5 seconds)
- `--max-time 30` — cap total request time. If the spec has `x-operation-timeout` on an operation, use that value instead. Lower values are acceptable for fire-and-forget cleanup calls.

Every `curl` command MUST capture the HTTP status code separately from the response body. Use `-w "%{http_code}"` with `-o` to a temp file:

```sh
httpStatus=$(curl ${curlInsecure} --connect-timeout 5 --max-time 30 -S -s \
  -o responseBody.tmp -w "%{http_code}" \
  -H "Authorization: Bearer ${authToken}" \
  "${API_BASE_URL}/resources")
responseBody=$(cat responseBody.tmp)
rm -f responseBody.tmp
```

The `-S` flag (show errors) ensures connection failures are visible even with `-s` (silent). Always combine `-s` with `-S`.

When `API_SKIP_TLS_VERIFY=true` in config, prepend `--insecure` to the curl command. Set a variable at the top of the script:

```sh
curlInsecure=""
if [ "${API_SKIP_TLS_VERIFY:-false}" = "true" ]; then
  curlInsecure="--insecure"
fi
```

Then use `${curlInsecure}` in every curl call: `curl ${curlInsecure} --connect-timeout 5 ...`. Do not add `--insecure` when the flag is absent or false.

## Guardrails

### Credential Isolation

Never store credentials in the script.

### Destructive Endpoints Gated

Do not include DELETE operations, secret key rotations, or destructive PUT/PATCH operations (those that overwrite data with no undo or destroy related resources) in the default script flow. Mark them with a `DESTRUCTIVE` prefix and skip unless the user sets `RUN_DESTRUCTIVE=true` in the config:

```sh
DESTRUCTIVE_resetPassword() {
  # skipped by default — set RUN_DESTRUCTIVE=true to run
  ...
}
```

### Error Visibility

Every test failure MUST produce visible output: endpoint name, expected status, received status, and response excerpt. Silent failures are forbidden.
