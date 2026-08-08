# DevKit

Turn any CLI tool into a visual dashboard. Manage dev services, scaffold projects, view errors visually, and more — all from one CLI.

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
| `dev contract init` | Create API contract (OpenAPI 3.0) |
| `dev contract test` | Test API matches contract |
| `dev contract mock` | Start mock server |
| `dev contract types` | Generate TypeScript types from contract |
| `dev contract validate` | Validate contract file |

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

## Changelog

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
