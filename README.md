# GMK (GoMicroKit)

CLI that scaffolds a production-shaped Go microservice: **REST (Fiber or Chi)** or **gRPC**, optional database, and opt-in infra (cache, auth, NATS, Consul, OpenTelemetry).

You pick the stack once. The generated app is meant to run locally with **`gmk run`**, then grow with `gmk make:service`.

```text
gmk new myservice  →  gmk run  →  http://localhost:8000
                           ↓
                    gmk make:service --name user
```

## Requirements

- **Go 1.25+** to build this kit (`go.mod`). Generated modules stamp the Go version of the `gmk` binary.
- **Docker Engine + Compose v2** for the generated stack and live reload
- Optional: [Air](https://github.com/air-verse/air) on the host (`gmk new --air`)

## Install

```bash
go install github.com/msdevbytes/gomicrokit@latest
```

Later:

```bash
gmk update              # install latest
gmk update --check      # show current vs latest without installing
```

From this repo:

```bash
git clone https://github.com/msdevbytes/gomicrokit.git
cd gomicrokit
go build -o gmk .
```

On Windows the binary is `gmk.exe`. Confirm with `gmk version`.

## Five-minute path

```bash
gmk new myservice
cd myservice
gmk key:generate          # only if you enabled auth
gmk run
```

Interactive `gmk new` asks for protocol, framework, database, cache, Docker, Air, and extra features. `--docker` defaults to **on** and writes Compose files that include **only** the services you selected.

Open the app at `http://localhost:8000` (or the port you chose). REST projects expose `/healthz` and `/readyz`.

Host-side run instead of the Compose `app` container:

```bash
gmk run --host
```

That starts only the dependency containers your stack needs, then `air` or `go run cmd/main.go`. `.env` still points at localhost.

## Recipes

```bash
# REST + Fiber + MySQL (non-interactive defaults)
gmk new myservice --non-interactive --module github.com/you/myservice

# REST + Chi + Postgres
gmk new myservice --non-interactive --framework chi --db postgres --module github.com/you/myservice

# REST + auth + Redis + NATS
gmk new myservice --non-interactive --features auth,cache,messaging --cache-store redis

# gRPC with no database (wizard default). Flag default for --db is still mysql.
gmk new myservice --non-interactive --type grpc --db none

# Preview files, write nothing
gmk new myservice --dry-run
```

`--non-interactive` requires a project name. Without extra flags that means REST, Fiber, MySQL, Docker on, port `8000`, no extra features.

## What you choose at generate time

| Choice | Options |
|--------|---------|
| Protocol | `rest` or `grpc` |
| REST framework | `fiber` (default) or `chi` |
| Database | REST: `mysql`, `postgres`, `sqlite`. gRPC: those plus `none` |
| Cache | off, or `redis` / `memory` (wizard step, or `--features cache --cache-store redis`) |
| Port | default `8000` (`--port`) |
| Docker | Dockerfile + feature-aware Compose (default on) |
| Features | `auth`, `cache`, `config`, `discovery`, `messaging`, `metrics`, `otel`, `resilience`, `tracing` |

Each generated project also gets a **README** with run commands for that stack.

## Commands

| Command | Where to run it | What it does |
|---------|-----------------|--------------|
| `gmk new [name]` | anywhere | Scaffold a new project |
| `gmk run` | inside the project | Start Compose watch, host Air, or `go run` from `.gmkrc.json` (alias `gmk start`) |
| `gmk make:service` | inside the project | Add a domain service (REST CRUD or gRPC stub) |
| `gmk remove:service` | inside the project | Remove a service generated less than 1 minute ago (`--force` after that) |
| `gmk key:generate` | inside the project | Write a Stripe-style `svc_test_…` `API_KEY` into `.env` (`--force` to rotate) |
| `gmk update` | anywhere | Install the latest kit (`--check` to preview, `--to v1.4.0` to pin) |
| `gmk version` | anywhere | Print kit version |

Full flags: [docs/commands.md](docs/commands.md).

## Documentation

| Guide | Contents |
|-------|----------|
| [Getting started](docs/getting-started.md) | Wizard steps, recipes, first curl |
| [Commands](docs/commands.md) | Every command and flag |
| [Features](docs/features.md) | What each `--features` value wires in |
| [Docker & live reload](docs/docker.md) | Compose, Watch, ports, `.env` |
| [Generated project](docs/generated-project.md) | Layout, routes, environment |
| [Troubleshooting](docs/troubleshooting.md) | Ports, Air loops, `gmk run` / `gmk update`, refused connections |

## Contributing to the kit

Kit tests live in `test/` (package `gmk_test`).

```bash
go test ./test
go test ./...
go build -o gmk .
go run . new demo --dry-run --non-interactive
```

Embed version data (use this when tagging a release so `gmk version` is not `dev`):

```bash
go build -ldflags "\
  -X 'github.com/msdevbytes/gomicrokit/cmd.Version=v1.4.0' \
  -X 'github.com/msdevbytes/gomicrokit/cmd.GitCommit=$(git rev-parse HEAD)' \
  -X 'github.com/msdevbytes/gomicrokit/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
  -o gmk .
```

Tagged source sets `cmd.Version` (this release is `v1.4.0`). `go install github.com/msdevbytes/gomicrokit@v1.4.0` therefore prints that version. Git commit and build date stay `unknown` unless you pass the ldflags above.

To publish a **new** version that `gmk update` and `go install @latest` can see:

1. Move `[Unreleased]` in `CHANGELOG.md` to a dated section and set `cmd.Version`.
2. `go test ./...`
3. Push `master`.
4. `git tag v1.5.0 && git push origin v1.5.0`
5. Create a GitHub Release from that tag.

## License

MIT. See [LICENSE](LICENSE). Built by [msdevbytes](https://github.com/msdevbytes).
