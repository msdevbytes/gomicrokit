# Features

Pass extras as `--features auth,cache,messaging` (comma-separated, case-insensitive). The interactive wizard also has a dedicated cache prompt (yes/no, then Redis vs memory). Cache is **not** listed in the multi-select catalog; the cache question is what enables it.

Missing Redis, NATS, or Consul does **not** crash the process: bootstrap logs a warning and continues. Discovery stays off until you set `DISCOVERY_PROVIDER=consul` (Compose `app` sets that automatically when discovery was selected).

| ID | What gets generated | Runtime / Compose |
|----|---------------------|-------------------|
| `auth` | REST middleware or gRPC interceptor; `API_KEY` / `AUTH_HEADER` | Guard compares the header to `API_KEY`. Generate a Stripe-style key with `gmk key:generate` (`svc_test_…`, or `svc_live_…` when `APP_ENV` is prod). |
| `cache` | `internal/cache/redis.go` or `memory.go` | Redis Compose service only if `--cache-store redis` (default). Memory cache needs no container. Uses `CACHE_STORE`, not an `ENABLE_CACHE` flag. |
| `config` | `internal/config/app_config.go` | Typed `PORT`, `APP_NAME`, shutdown timeout. |
| `discovery` | Consul registrar | Compose `consul`. Host `.env` still has `DISCOVERY_PROVIDER=none` so `go run` is safe; Compose `app` sets `consul`. |
| `messaging` | NATS client | Compose `nats` (`NATS_URL`). |
| `metrics` | HTTP handler (`GET /metrics`) or gRPC interceptor | No extra container. |
| `otel` | `pkg/telemetry/otel.go`, `otel-collector.yaml` when Docker is on | Compose `otel-collector`. Empty `OTEL_EXPORTER_OTLP_ENDPOINT` uses **stdout**; Docker+otel writes `localhost:4317` in host `.env` and `otel-collector:4317` on the Compose `app` service. |
| `resilience` | Circuit breaker + timeout helper | No extra container. |
| `tracing` | Light request-trace helpers; gRPC tracing interceptor when type is gRPC | Often paired with `otel`. |

Unknown feature IDs fail at generate time.

## Cache store

```bash
gmk new app --features cache --cache-store redis
gmk new app --features cache --cache-store memory
```

Invalid stores error at generate time. Without the `cache` feature, store is `none`. `--cache-store` is ignored unless `cache` is on.

## Flags in `.env`

Selected features write `ENABLE_<FEATURE>=yes` into `.env` (`auth`, `config`, `discovery`, `messaging`, `metrics`, `otel`, `resilience`, `tracing`). Those keys record the generate-time choice. Wiring is compiled in; flipping `ENABLE_AUTH=no` does not remove middleware.

Runtime knobs that **do** change behavior without regenerating:

| Knob | Effect |
|------|--------|
| `DISCOVERY_PROVIDER` | `none` skips Consul; Compose `app` sets `consul` when discovery was selected |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | empty → stdout traces; set → OTLP gRPC |
| `API_KEY` | empty disables the API-key check (middleware still mounted) |
| `CACHE_STORE` | written at generate time (`redis`, `memory`, or `none`) |
