# Generated project

Layout depends on protocol and features. Docker files appear when `--docker` is true (the default). A project **README.md** is always rendered with run commands for the selected stack.

From the project directory: `gmk run` (see [commands.md](commands.md#gmk-run)).

## REST (Fiber or Chi)

```
myservice/
├── cmd/main.go
├── internal/
│   ├── api/router.go          # routes; /healthz and /readyz
│   ├── bootstrap/app.go
│   ├── config/                # db + pagination; app_config.go if config feature
│   ├── db/
│   ├── dto/
│   ├── handler/               # default.go plus per-service handlers
│   ├── middleware/            # if auth
│   ├── model/
│   ├── repository/
│   └── service/container.go
├── pkg/logger/
├── test/unit/  test/mocks/
├── .env
├── .gmkrc.json                # saved generate options
├── .gen_history.json          # make:service history (gitignored)
├── Dockerfile
├── docker-compose.yml
├── .air.docker.toml
├── README.md
└── go.mod
```

Chi uses the `*_chi` templates for router, handlers, and auth middleware.

Built-in HTTP routes (before any `make:service`):

| Method | Path | Notes |
|--------|------|--------|
| `GET` | `/` | Welcome payload |
| `GET` | `/healthz` | Liveness |
| `GET` | `/readyz` | Readiness |
| `GET` | `/api/v1/` | Default handler (`API_ROUTE_VERSION`) |
| `GET` | `/metrics` | Only if `metrics` feature |

Auth, when selected, wraps the `/api/v1` group, not `/healthz`.

## gRPC

```
myservice/
├── cmd/main.go
├── api/proto/health.proto     # plus per-service protos from make:service
├── internal/
│   ├── bootstrap/app.go
│   ├── server/grpc/server.go
│   ├── server/grpc/interceptors/   # auth / metrics / tracing if selected
│   ├── service/
│   └── db/                    # omitted when --db none
└── ...
```

The process registers `HealthService` from `health.proto`, the standard gRPC health server, and reflection. You still need `protoc` (and plugins) in your own workflow to compile new `.proto` files; the kit drops the proto and a server stub.

## Adding a service

From the project root:

```bash
gmk make:service --name user
```

REST registers `GET/POST /api/v1/users` and id routes (`/:id` on Fiber, `/{id}` on Chi). List accepts `page` and `limit` query params (default 1 and 10). gRPC registers a server under `// gmk:register-services`.

Undo shortly after generate:

```bash
gmk remove:service --name user
```

After one minute, pass `--force`.

## Environment

`.env` is gitignored in the generated `.gitignore`. Typical keys:

| Key | Role |
|-----|------|
| `PORT` / `HOST` | HTTP/gRPC listen (Compose sets `HOST=0.0.0.0`) |
| `APP_ENV` | `dev` by default; affects migrate safety |
| `API_ROUTE_VERSION` | REST prefix, default `/api/v1` |
| `DB_*` | GORM connection; SQLite uses `DB_NAME` as the file path |
| `DB_PORT_FORWARD` | Host publish only (Compose); not used by the Go DSN |
| `REDIS_*` | Redis cache |
| `NATS_URL` | Messaging |
| `DISCOVERY_PROVIDER` | `none` or `consul` |
| `API_KEY` / `AUTH_HEADER` | Auth (`x-api-key` by default). `gmk key:generate` writes `svc_test_…` or `svc_live_…` |
| `ENABLE_*` | Records which features were generated (not a runtime off-switch; see [features.md](features.md)) |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown (default `10s`) |
| `FORCE_MIGRATE` | `yes` drops tables before AutoMigrate when `APP_ENV` is not `dev` and not `prod`. In `dev`, GORM AutoMigrate is skipped. |

Host `.env` uses localhost so `go run` can reach published Compose ports. The `app` service overrides `DB_HOST`, `REDIS_HOST`, `NATS_URL`, `DISCOVERY_PROVIDER`, `OTEL_EXPORTER_OTLP_ENDPOINT`, and similar to Docker service names.

## Config file

`.gmkrc.json` stores `project_name`, `module_path`, `project_type`, `framework`, `db`, `cache_store`, `features`, `port`, `with_docker`, `with_air`, and dependency versions so later generators and `gmk run` stay consistent.
