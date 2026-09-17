# Troubleshooting

## `unknown command "update"` (or missing `run` / `key:generate`)

`go install github.com/msdevbytes/gomicrokit@latest` installs a binary named **`gomicrokit`**, not `gmk`. An older `gmk` (v1.3.0) can stay first on PATH.

From v1.4.1 install the `gmk` package:

```bash
go install github.com/msdevbytes/gomicrokit/cmd/gmk@latest
gmk version   # expect v1.4.1 or newer
```

On Windows, `where gmk` should point at `%USERPROFILE%\go\bin\gmk.exe` (or your `GOBIN`). If `gmk version` is still old, that path is a leftover binary.

## `gmk update` cannot overwrite the binary (Windows)

Windows will not replace `gmk.exe` while it is running. Close other terminals using `gmk`, then run `gmk update` again, or:

```bash
go install github.com/msdevbytes/gomicrokit/cmd/gmk@latest
```

`gmk update` also needs `go` on PATH (same as install).

If `--check` still shows an older tag after you pushed new commits, the Go module proxy has not seen a **new git tag** yet. Tag and push the next version (for example `v1.5.0`), then wait a minute for the proxy.

## `gmk new` still shows the old arrow-key confirmation screen

Rebuild the CLI from this repo (`go install ./cmd/gmk`) and run that binary. An older `gmk` on `PATH` still uses the Bubble Tea flow when no name is given.

## Directory already exists

`gmk new myservice` fails immediately if `./myservice` (or `--output-dir/myservice`) is already there. `gmk new` with no name asks again until the name is free. Delete the folder or pick another name.

## Windows wizard looks broken or hangs

Set `GMK_FORCE_BASIC_MENU=1` and re-run `gmk new`. Numbered prompts replace the survey selector.

## `gmk run` starts `go run` / Air instead of Compose

`gmk run` only uses Compose when Docker CLI is on PATH **and** `docker info` succeeds. Otherwise it prints a warning and starts the host process. Install Docker Desktop (or Engine + Compose v2), start the daemon, and retry. `gmk run --dry-run` shows the command without starting anything.

`--detach` (`-d`) is `docker compose up -d --build` and does **not** watch files. Live reload needs the default foreground `gmk run`.

## `docker compose` cannot bind 3306 (or 5432)

Another MySQL/Postgres is already on the host. In `.env` set `DB_PORT_FORWARD=3310` (or any free port) and keep `DB_PORT=3306` (engine port inside Docker). Recreate:

```bash
docker compose down
docker compose up --watch --build
```

The Go DSN uses `DB_PORT`, not `DB_PORT_FORWARD`.

## Air rebuilds forever, logs show `Syncing service "app"`

Do **not** add a `.:/src` bind mount. Watch already syncs `cmd/`, `internal/`, `pkg/` (and `api/` for gRPC). Recreate with the generated compose file:

```bash
docker compose down
docker compose up --watch --build
```

`app` uses `restart: on-failure` so an Air exit of 0 does not loop.

## Redis / NATS / Consul “connection refused” on `go run`

Those processes are not on localhost until you start the matching Compose services. Either:

```bash
gmk run
```

or `gmk run --host` (infra containers + Air/`go run` on the host). Missing Redis/NATS/Consul should **warn** and continue, not crash.

## App container missing in Docker Desktop

`app` is a default service (not a Compose profile). Use `gmk run` or `docker compose up --watch --build`, not infra-only `up mysql redis …`.

## JSON span / trace lines on stdout (especially gRPC)

Empty `OTEL_EXPORTER_OTLP_ENDPOINT` means the OTel SDK exports to **stdout**. Docker+otel writes `localhost:4317` in host `.env`; the Compose `app` service points at `otel-collector:4317`. Leave the endpoint empty only if you want console traces.

`4317` is not assumed unless you selected **both** Docker and `otel`.

## `gmk key:generate` fails with missing `.env`

Run it from the generated project root (the folder that contains `.env`). The command does not create `.env`.

## `gmk remove:service` asks for `--force`

Services older than one minute need `--force`. That window is intentional so accidental deletes of older code need an explicit override.

## `gmk new --non-interactive --type grpc` still scaffolds MySQL

CLI `--db` defaults to `mysql` for every protocol. Pass `--db none` to match the interactive gRPC default.

## Live reload does not pick up edits

1. Confirm you used `--watch` (`docker compose up --watch --build`).
2. Edit files under `cmd/`, `internal/`, or `pkg/` (or `api/` for gRPC). Other paths are not synced.
3. `go.mod` / `go.sum` changes rebuild the image; they do not hot-patch the running binary.

## Health checks fail from the host

REST listens on `HOST` from `.env` (`localhost`) for `go run`, and `0.0.0.0` inside Compose. Curl `http://localhost:<PORT>/healthz` using the port you generated (default `8000`). gRPC has no HTTP `/healthz`; use the generated `HealthService` proto and gRPC health.
