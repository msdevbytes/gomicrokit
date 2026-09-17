# Docker and live reload

`--docker` (default **true**) writes:

- `Dockerfile` — `builder` (static binary), `dev` (Air), `final` (runtime image)
- `docker-compose.yml` — `app` plus **only** selected dependencies
- `.dockerignore`
- `.air.docker.toml` — Air config used **inside** the `dev` container (Linux `tmp/main`)

Host Air (`.air.toml` with `tmp/main.exe`) is separate: `gmk new --air`. Compose live reload does **not** need `--air`.

## Start the stack

From the generated project:

```bash
gmk run
```

That is `docker compose up --watch --build` when Compose files exist and Docker is running. `gmk run --detach` is `docker compose up -d --build` (no file watch). Equivalent manual command:

```bash
docker compose up --watch --build
```

Detached, no file watch:

```bash
docker compose up -d --build
```

Infra only (app runs on the host with `gmk run --host`, or `go run` / `air`):

```bash
docker compose up -d mysql redis nats consul otel-collector
```

Use the service names that actually exist in your file. SQLite or memory cache will not add `mysql` / `redis`.

## How code updates work

The `app` service builds Dockerfile target **`dev`** and runs Air. Source is **not** bind-mounted (`.:/src` was removed because it restarted Air in a loop).

Updates use Compose Watch `sync` into `/src/cmd`, `/src/internal`, `/src/pkg`, and `/src/api` (gRPC only). Air rebuilds when those Go files change.

| Host change | Compose Watch action |
|-------------|----------------------|
| `cmd/`, `internal/`, `pkg/`, `api/` | `sync` into the container, then Air rebuilds |
| `go.mod` / `go.sum` | image **rebuild** |
| `.env` | container **restart** (reloads `env_file`) |
| `.air.docker.toml` | image **rebuild** |

Without `--watch`, the container keeps the source copied at **image build** time.

The only app volume is `go-mod-cache` (`/go/pkg/mod`).

## Ports and `.env`

`.env` is for **host** processes (`127.0.0.1` / `localhost`). The Compose `app` service overrides hosts to Docker DNS (`mysql`, `redis`, `nats`, `consul`, `otel-collector`).

| Variable | Meaning |
|----------|---------|
| `PORT` | App listen port (mapped as `PORT:PORT`) |
| `DB_PORT` | Database engine port **inside** the network (`3306` / `5432`) |
| `DB_PORT_FORWARD` | Host port published for MySQL/Postgres. If unset, Compose uses 3306 or 5432 |
| `REDIS_PORT` | Host publish for Redis (default 6379) |
| `NATS_URL` | Host: `nats://127.0.0.1:4222`. In Compose app: `nats://nats:4222` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Host collector `localhost:4317` when Docker+otel; empty means stdout exporter. Compose app uses `otel-collector:4317`. |

If host port 3306 is already taken, set `DB_PORT_FORWARD=3310` and keep `DB_PORT=3306`.

## Production image

```bash
docker compose build --target final
```

`final` is a static binary on Alpine, not Air. The default Compose `app` service always uses `target: dev`.

## Troubleshooting

See [troubleshooting.md](troubleshooting.md) for Air restart loops, missing `app` containers, `gmk run` falling back to the host, and connection refused on `go run`.
