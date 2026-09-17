# Commands

Run `gmk <command> --help` for the same text as the CLI.

Root examples from `gmk` itself:

```text
gmk new                    # Interactive project creation
gmk new myapp --dry-run    # Preview what would be created
gmk make:service           # Interactive service generation
gmk run                    # Start Compose watch, Air, or go run
gmk key:generate           # Write a Stripe-style API_KEY to .env
gmk update                 # Install the latest published gmk
gmk version                # Show version info
```

## `gmk new [project-name]`

Scaffold a microservice directory under `--output-dir` (default `.`). The folder name is the project name.

`gmk new` with no argument uses the same wizard as `gmk new myapp`. It asks for the project name first and will not continue if that directory already exists.

| Flag | Default | Description |
|------|---------|-------------|
| `--module` | project name | Go module path |
| `--type` | `rest` | `rest` or `grpc` |
| `--framework` | Fiber for REST, `grpc-go` for gRPC | REST: `fiber` or `chi` |
| `--db` | `mysql` | REST: `mysql`, `postgres`, `sqlite`. gRPC: also `none`. Interactive gRPC defaults to `none`; the flag does not. |
| `--port` | `8000` | App listen port written to `.env` and Compose |
| `--docker` | `true` | Dockerfile, `docker-compose.yml`, `.dockerignore`, `.air.docker.toml` |
| `--air` | `false` | Host Air config (`.air.toml`; binary `tmp/main.exe`) |
| `--features` | empty | Comma-separated list (see [features.md](features.md)) |
| `--cache-store` | `redis` | `redis` or `memory` when `cache` is enabled |
| `--rest-version` | `latest` | Fiber/Chi (and related) module version |
| `--grpc-version` | `latest` | gRPC module version |
| `--output-dir` | `.` | Parent folder for the new project directory |
| `--save-config` | `true` | Write `.gmkrc.json` |
| `--non-interactive` | `false` | Use flags only; project name is required |
| `--dry-run` | `false` | Print planned files; write nothing |

`--docker=false` skips Compose and the Dockerfile. `gmk run` then uses Air or `go run cmd/main.go`.

## `gmk run`

Run **inside** a generated project (or a subdirectory). Reads `.gmkrc.json` and starts the matching process.

| Situation | What runs |
|-----------|-----------|
| Docker Compose present and Docker is running | `docker compose up --watch --build` (foreground; live reload) |
| `--detach` / `-d` | `docker compose up -d --build` (no file watch) |
| `--no-watch` | `docker compose up --build` (foreground, no watch) |
| `--host`, or Docker missing/down | Optional `docker compose up -d` for **infra** services only, then `air` or `go run cmd/main.go` |
| No Docker, `--air` was selected | `air` (falls back to `go run` if Air is not on PATH) |
| No Docker, no Air | `go run cmd/main.go` |

`--watch` is **not** combined with `-d` by default: Compose Watch needs the CLI in the foreground. `.env` is passed as `--env-file` when that file exists.

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Print the command(s), do not start |
| `--host` | `false` | Run the app on the host; start selected Compose dependencies only |
| `--detach` / `-d` | `false` | Detached Compose up, no watch |
| `--no-watch` | `false` | Foreground Compose up without watch |

Alias: `gmk start`.

## `gmk make:service`

Run **inside** a generated project. Detects REST vs gRPC from `.gmkrc.json` (or `--type`).

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | (wizard) | Service name; must be a valid Go identifier (`user`, `Product`) |
| `--type` | auto | `rest` or `grpc` |
| `--framework` | auto | REST handler style: `fiber` or `chi` |
| `--force` | `false` | Overwrite existing generated files |
| `--dry-run` | `false` | Preview paths only |

REST creates model, repository, service, handler, DTO, and a DTO unit test, then registers the handler on the router and in `internal/service/container.go`. Routes land at `API_ROUTE_VERSION` + `/<names>` (for `--name user` that is `/api/v1/users`). Fiber uses `/:id`; Chi uses `/{id}`.

gRPC creates `api/proto/<name>.proto`, a service stub, and a gRPC server registration (`// gmk:register-services`).

Without `--name`, an interactive wizard asks for the name.

## `gmk remove:service`

Run **inside** a generated project. Deletes files recorded in `.gen_history.json` and unregisters them.

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | (wizard) | Service to remove |
| `--force` | `false` | Required if the service was generated more than **1 minute** ago |

`.gen_history.json` is gitignored in the generated project.

## `gmk key:generate`

Run **inside** a generated project. Writes a Stripe-style secret to `API_KEY` in `.env`:

`svc_test_` + 43 alphanumeric characters (or `svc_live_` when `APP_ENV` is `prod`, `production`, or `live`). Entropy comes from `crypto/rand` (~256 bits).

| Flag | Default | Description |
|------|---------|-------------|
| `--force` | `false` | Replace a key that is already set |

Empty `API_KEY` and the placeholder `change-me-in-prod` are replaced without `--force`. Missing `.env` is an error.

## `gmk version`

Prints kit version, git commit, build date, Go version, and OS/arch. Tagged `v1.4.1` prints `gmk v1.4.1`. Git commit and build date stay `unknown` unless you pass ldflags (see the root README).

## `gmk update`

Installs the latest published `gmk` with `go install github.com/msdevbytes/gomicrokit/cmd/gmk@latest` (or a pin). Requires Go on PATH. Alias: `gmk self-update`.

| Flag | Default | Description |
|------|---------|-------------|
| `--check` | `false` | Print current vs latest; do not install |
| `--to` | latest from the module proxy | Module version to install (`v1.4.1`, `latest`) |

Local `dev` builds (if you set `cmd.Version` back to `dev`) are never treated as up to date, so `gmk update` will install the published module.

`gmk update` installs the **latest git tag** the Go module proxy can see (for example `v1.4.1`). Untagged commits on `master` are not picked up until you tag and push that tag. The install path is `…/cmd/gmk` so the binary is named `gmk`, not `gomicrokit`.
