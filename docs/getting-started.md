# Getting started

## Create a project

From any directory:

```bash
gmk new
gmk new myservice
```

Both use the **same** interactive wizard. Without a name, the first prompt is the project name. If `<output-dir>/<name>` already exists, GMK asks for a different name instead of continuing (or exits when the name was passed on the command line). To skip prompts:

```bash
gmk new myservice --non-interactive --module github.com/you/myservice
```

`--non-interactive` requires a project name. Defaults then are REST, Fiber, MySQL, Docker on, port `8000`, no extra features.

The project directory is `<output-dir>/<project-name>` (`--output-dir` defaults to `.`). After generate, `gmk` prints `cd` to that path.

### Wizard prompts

1. Project name (required; rejected if that folder already exists), module path, port (default `8000`)
2. Protocol: REST or gRPC
3. REST: Fiber or Chi, then MySQL / Postgres / SQLite  
   gRPC: database including **None** (wizard default). `--db` on the CLI still defaults to `mysql` unless you pass `--db none`.
4. Cache layer yes/no, then Redis or in-memory
5. Docker (Dockerfile + Compose) — default yes
6. Host Air config (`.air.toml`) — default no; Compose already uses Air inside the `dev` image
7. Optional features (space to toggle)
8. Output directory and whether to save `.gmkrc.json`

You get a summary and can generate, restart the wizard, or cancel.

Selections are saved to `.gmkrc.json` unless you pass `--save-config=false`. Later `gmk run`, `make:service`, and `remove:service` read that file.

### Useful flags

```bash
gmk new myservice --type grpc --db none
gmk new myservice --framework chi --db postgres --port 8080
gmk new myservice --features auth,cache,messaging --cache-store redis
gmk new myservice --output-dir sandbox/generated
gmk new myservice --dry-run
```

## Windows terminals

The wizard uses [survey](https://github.com/AlecAivazis/survey). If the fancy selector fails, numbered prompts still work. To force the numbered UI:

```bash
set GMK_FORCE_BASIC_MENU=1
gmk new myservice
```

PowerShell: `$env:GMK_FORCE_BASIC_MENU=1`.

## First run of a generated app

1. `cd` into the project (the path printed by `gmk new`)
2. If auth is enabled: `gmk key:generate` (writes `svc_test_…`, replacing `change-me-in-prod`)
3. Start the stack: `gmk run` (or `gmk start`). Compose watch when Docker is available; otherwise Air or `go run`. `gmk run --dry-run` prints the command. `gmk run --host` runs the app on the host and starts only infra containers.
4. REST health check:

```bash
curl http://localhost:8000/healthz
curl http://localhost:8000/readyz
```

The generated `README.md` repeats these commands for the stack you chose.

To develop with `go run` or host Air instead of the Compose `app` service, use `gmk run --host`. That starts **only** the dependency containers that exist for your stack, then runs the binary on the host. `.env` points at `localhost` for that mode. See [docker.md](docker.md).

## Add a domain

From the project root:

```bash
gmk make:service --name user
```

REST registers CRUD on `/api/v1/users` (plural of the service name):

```bash
curl http://localhost:8000/api/v1/users
curl -X POST http://localhost:8000/api/v1/users -H "Content-Type: application/json" -d "{}"
```

If auth is on, send the header from `AUTH_HEADER` (default `x-api-key`) with the `API_KEY` from `.env` (`svc_test_…` after `gmk key:generate`).

gRPC adds `api/proto/user.proto` and a server stub. You still need `protoc` in your own workflow to compile new protos.

Undo shortly after generate:

```bash
gmk remove:service --name user
```

After one minute, pass `--force`.

## Next

- Keep the kit current: `gmk update` (or `gmk update --check` first)
- Feature list: [features.md](features.md)
- Command reference: [commands.md](commands.md)
- Ports, Watch, Air: [docker.md](docker.md)
- If something fails: [troubleshooting.md](troubleshooting.md)
