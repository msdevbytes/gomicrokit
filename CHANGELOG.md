# Changelog

All notable changes to GMK (GoMicroKit) will be documented in this file.

## [Unreleased]

## [1.4.1] - 2026-09-17

### Fixed
- `go install` and `gmk update` install the `gmk` binary (`github.com/msdevbytes/gomicrokit/cmd/gmk`). Installing the module root produced `gomicrokit` and left an older `gmk` on PATH.

## [1.4.0] - 2026-09-17

### Added
- Interactive `gmk new` wizard (survey) for protocol, framework, database, cache, Docker, Air, port, and optional features
- REST **Chi** alongside Fiber; **gRPC** projects (`grpc-go`) with health, reflection, and optional interceptors
- Database choices: MySQL, Postgres, SQLite (gRPC also supports `none`)
- Optional features: `auth`, `cache`, `config`, `discovery`, `messaging`, `metrics`, `otel`, `resilience`, `tracing`
- Feature-aware `docker-compose.yml` (only selected deps: MySQL/Postgres, Redis, NATS, Consul, OTel collector) plus `.dockerignore` and `otel-collector.yaml` when Docker + OpenTelemetry are enabled
- Compose `app` live-reloads via Air and `docker compose up --watch` (no source bind mount)
- Host Air config via `gmk new --air` (`.air.toml`)
- `gmk run` (`gmk start`) starts Compose watch, host Air, or `go run` from `.gmkrc.json` (`--host`, `--detach`, `--no-watch`, `--dry-run`)
- `gmk key:generate` writes a Stripe-style alphanumeric `API_KEY` (`svc_test_…` / `svc_live_…`) into `.env` (`--force` replaces an existing key)
- `gmk update` / `gmk self-update` installs the latest tagged kit via `go install` (`--check`, `--to`)
- `--output-dir`, `--port` (default `8000`), `--non-interactive`, `--save-config` on `gmk new`
- Developer docs: README plus `docs/` (getting started, commands, features, Docker, generated project, troubleshooting)
- MIT `LICENSE`
- Kit tests under `test/`

### Changed
- `gmk new` with or without a name uses the same wizard and rejects a project directory that already exists
- `--docker` (default on) writes Compose files, not only a Dockerfile
- Generated `.env` database defaults match the selected engine so Compose can boot without extra edits
- Database host publish port uses `DB_PORT_FORWARD` (falls back to 3306/5432); `DB_PORT` remains the engine port
- Missing Redis, NATS, or Consul logs a warning and continues instead of crashing the process
- Discovery stays off on the host (`DISCOVERY_PROVIDER=none`) unless Compose sets `consul`

## [1.3.0] - 2026-02-02

### Added
- Short CLI command name `gmk` (previously `gomicrokit`) for easier typing
- `version` command to display version, git commit, build date, and Go version
- `--dry-run` flag for `new` command to preview project structure
- Skip notification when files already exist during service generation
- Proper error messages with context throughout the codebase
- Unit tests for generator package (25.7% coverage)
- Unicode support in service name validation
- Underscore support in service names

### Changed
- Repository template now includes full CRUD implementation with pagination
- Service template now includes business logic methods
- Handler template now includes real REST implementations with validation
- DTO template now includes Create/Update input structs
- Model template now includes sample Name field and TableName method
- Test template now includes actual JSON serialization tests
- Improved error handling - functions return errors instead of calling os.Exit()
- Better UX with "next steps" guide after project creation

### Fixed
- Duplicate `writeHistory` call in service generation
- Redundant empty check in RemoveService
- Time message inconsistency ("5 minutes" vs "1 minute")
- File handle leak in template rendering (defer in loop)
- Non-deterministic file generation order (now sorted)
- Module path incorrectly lowercased
- Duplicate `getGoModule` function consolidated
- Handler template using strings instead of errors for `errorResponse()`
- Model template missing `Name` field referenced by handler and DTO

### Removed
- Unsafe `must()` function that called os.Exit()

## [0.1.0] - Initial Release

### Added
- Interactive project creation with Bubble Tea TUI
- Service generation with `make:service` command
- Service removal with `remove:service` command
- Fiber framework support
- MySQL database support
- GORM ORM integration
- Docker support with Dockerfile generation
- Repository pattern architecture
- Service layer architecture
- Generation history tracking with `.gen_history.json`
