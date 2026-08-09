# DevKit

Turn any CLI tool into a visual dashboard. Manage dev services, scaffold projects, view errors visually, and more — all from one CLI.

## Dashboard

`dev serve` opens the DevKit web dashboard at `http://localhost:8080` — a visual
control center for every DevKit tool: service status, project commands, and live
output in one dark-themed UI.

![DevKit Dashboard](docs/devkit-dashboard.png)

## Install

```bash
# From source
git clone git@github.com:Gopal-student21/devkit.git
cd devkit
go build -o ~/bin/dev ./cmd/dev
```

Requires Go 1.23+

## Commands

### Core
| Command | Description |
|---------|-------------|
| `dev up [services...]` | Start Docker containers (postgres, redis, mysql, mongo, minio) |
| `dev stop [services...]` | Stop and remove containers |
| `dev status` | Show status of all services |
| `dev logs [service]` | View container logs |
| `dev env` | Generate .env file from running services |

### Project
| Command | Description |
|---------|-------------|
| `dev init` | Initialize DevKit in an existing project |
| `dev new <name> [template]` | Scaffold a new project (node, python, go) |
| `dev detect` | Auto-detect project stack |
| `dev config` | Show current configuration |

### Database
| Command | Description |
|---------|-------------|
| `dev db shell` | Open database shell (psql, mongosh, mysql) |
| `dev db url` | Print database connection URL |
| `dev migrate up` | Run database migrations |
| `dev seed` | Load test data into database |

### Quality
| Command | Description |
|---------|-------------|
| `dev test` | Run project test suite |
| `dev qa test` | Run QA tests |
| `dev qa api` | Test API endpoints |
| `dev qa security` | Security vulnerability scan |
| `dev qa performance` | Performance benchmark |
| `dev review` | Static code analysis |
| `dev verify` | Verification layer for AI-generated code (security + smell + AI attribution) |
| `dev contract init` | Create API contract (OpenAPI 3.0) |
| `dev contract test` | Test API matches contract (schema validation, CI-ready) |
| `dev contract mock` | Start mock server |
| `dev contract types` | Generate TypeScript types from contract |
| `dev contract validate` | Validate contract file |
| `dev contract seed` | Generate FK-aware mock seed data (no orphaned refs) |

### Tools
| Command | Description |
|---------|-------------|
| `dev errview` | Visual error viewer for any language |
| `dev serve` | Start web dashboard at localhost:8080 |

## errview

Visual error viewer that parses compiler/linter output and shows errors with code context, line numbers, and pointers.

```bash
# Pipe any compiler/linter
go build ./... 2>&1 | dev errview
npx tsc --noEmit 2>&1 | dev errview
eslint src/ 2>&1 | dev errview
cargo build 2>&1 | dev errview

# Read from file
dev errview errors.txt

# Compact mode (one line per error)
dev errview -c errors.txt
```

**Supported formats:** Go, TypeScript, ESLint, Python, Rust, GCC/Clang, and any `file:line:col` format.

### Example output

```
━━━ Error Summary ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1 error(s)  1 file(s) affected

▸ ./main.go
  ✗ ERROR line 5:2
  undefined: foo
  ┌─
    2 │
    3 │ func main() {
    4 │     x := 10
  5 │     foo(x)
    │       ^ undefined: foo
    6 │ }
  └─
```

## Quick Start

```bash
# Start Postgres + Redis
dev up postgres redis

# Generate .env with connection strings
dev env

# Check what's running
dev status

# Open database shell
dev db shell

# View errors visually
go build ./... 2>&1 | dev errview

# Stop everything
dev stop
```

## verify

Verification layer for AI-generated code — catches security issues, code smells, and AI-author markers before merge. Acts as a CI gate.

```bash
dev verify              # verify staged changes
dev verify --all        # verify all uncommitted changes
dev verify --file x.ts  # verify a specific file
dev verify --tests      # also run the project test suite (verify before merge)
dev verify --html r.html   # write a visual HTML report
dev verify --open       # write + open the visual report in your browser
dev verify --ci         # JSON output for CI pipelines
dev verify --strict     # exit non-zero on any finding
```

Checks cover: hardcoded secrets & credential values (sk_/ghp_/AKIA/JWT), `eval`/command execution, SQL injection, XSS (`innerHTML`), TS `any`/non-null assertions, empty catches, ignored errors, debug logs, TODOs, magic numbers — plus **AI-author attribution**.

### AI attribution

`dev verify` answers "how much of this change looks AI-authored?" per file and per corpus. It combines several weak signals into a per-line attribution: explicit generated markers, **echo comments** (comments that restate the very next code token), **essay-style docblocks** (3+ long prose comment lines), type widening (`any`/non-null) and swallowed errors. Each file gets an AI% with a None/Low/Moderate/High signal, surfaced in the terminal report, the JSON output, and the HTML report.

### Test gate

`dev verify --tests` detects the stack (Go/Node/Python/Rust/Java) and runs the project suite (`go test ./...`, `npm test`, `pytest`, `cargo test`, `mvn test`) with a 90s timeout. Failed tests set the verdict to **BLOCK** — so verification is a real pre-merge gate, not just static scanning.

## contract

Local-first API contract / mock / verify loop. One `api.yaml` is the single source of truth.

```bash
# Create the contract
dev contract init

# Validate the contract file
dev contract validate

# Test the real API against the contract (schema validation, exit 1 on failure)
dev contract test --url http://localhost:3000
dev contract test --strict        # extra fields also fail

# Start a mock server from the contract (frontend dev, no backend needed)
dev contract mock --port 4000

# Generate TypeScript types from the contract
dev contract types

# Generate FK-aware mock seed data — foreign keys point at records that exist
dev contract seed                     # 5 records/schema -> seed.json
dev contract seed --count 20          # 20 records per schema
dev contract seed --format sql        # INSERT statements -> seed.sql
dev contract seed --seed 7 --out f.json   # deterministic
```

`dev contract test` validates each endpoint's JSON response against the declared
schema: required fields, types, enums, email format, and `$ref` resolution.
`dev contract seed` detects relationships via `$ref` or naming conventions
(`userId`/`order_id`) and generates **referentially-consistent** records — no
orphaned foreign keys, unlike naive fake-data generators.

## Changelog

### v0.9.0 (2026-08-09)
- **`dev verify` hardening** — the O2 verification/hardening layer for AI-generated code:
  - **AI attribution engine** — per-file & per-corpus AI% from generated markers, echo comments, essay-style docblocks, type widening and swallowed errors (None/Low/Moderate/High)
  - **`--tests` test gate** — auto-detects the stack and runs the project suite (`go test`/`npm test`/`pytest`/`cargo test`/`mvn test`); failed tests block the merge verdict
  - **`--html` / `--open`** — standalone dark-themed visual report with verdict banner, risk/AI chips, per-file findings and an AI% bar

### v0.8.0 (2026-08-09)
- **New `dev contract seed`** — FK-aware, deterministic mock-data generation (JSON or SQL) that respects `$ref` and `userId`/`order_id` naming conventions so foreign keys always reference existing records
- **Rewritten `dev contract test`** — real schema validation (required fields, types, enums, email format, `$ref`) instead of status-code-only curl checks; `--strict` mode and exit 1 for CI

### v0.7.0 (2026-08-08)
- **New `dev verify`** — verification/hardening layer for AI-generated code: security scans, code-smell detection, and AI-author attribution with a visual risk report + `--ci` JSON mode for pre-merge gates

### v0.6.1 (2026-08-08)
- **Fixed** `contract types` generating empty interfaces — now emits real TypeScript from `components.schemas` (optionals via `required`, enums, `anyOf` unions, arrays, nested objects, `$ref` resolution)
- **Fixed** `contract mock` server crash (`%!d(MISSING)` port) — mock now resolves `$ref` schemas and generates realistic values (email, date-time, URLs)
- **Fixed** mock array/object responses returning `{}`

### v0.6.0 (2026-07)
- `errview` — visual error viewer for any language
- API contract testing, QA automation, code review

### v0.5.0
- API contract testing, QA automation, code review

### v0.4.0
- Universal CLI Dashboard

## License

MIT
