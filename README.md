# GMK (GoMicroKit)

A CLI tool for scaffolding Go microservices with clean architecture patterns.

## Features

- Interactive project setup with Bubble Tea TUI
- Repository pattern with GORM
- Service layer architecture
- REST API handlers (Fiber)
- Docker support
- Auto-generated tests
- Service generation and removal

## Installation

```bash
go install github.com/msdevbytes/gomicrokit@latest
```

Or build from source:

```bash
git clone https://github.com/msdevbytes/gomicrokit.git
cd gomicrokit
go build -o gmk .
```

## Quick Start

### Create a New Project

Interactive mode:
```bash
gmk new
```

Non-interactive:
```bash
gmk new myproject --module github.com/user/myproject --db mysql --docker
```

Preview without creating files:
```bash
gmk new myproject --dry-run
```

### Generate a Service

```bash
gmk make:service
```

Or non-interactive:
```bash
gmk make:service --name user --force
```

Preview generated code:
```bash
gmk make:service --name user --dry-run
```

### Remove a Service

```bash
gmk remove:service
```

Or with force (bypass time check):
```bash
gmk remove:service --name user --force
```

### Check Version

```bash
gmk version
```

## Commands

| Command | Description |
|---------|-------------|
| `new [name]` | Create a new microservice project |
| `make:service` | Generate a new service (model, repo, handler, etc.) |
| `remove:service` | Remove a previously generated service |
| `version` | Print version information |

## Flags

### `new` command

| Flag | Default | Description |
|------|---------|-------------|
| `--module` | | Go module path (e.g., github.com/user/project) |
| `--db` | mysql | Database type (mysql, postgres, sqlite) |
| `--docker` | true | Include Dockerfile |
| `--dry-run` | false | Preview what would be created |

### `make:service` command

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | | Service name (e.g., user, product) |
| `--force` | false | Overwrite existing files |
| `--dry-run` | false | Preview generated code |

## Generated Project Structure

```
myproject/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── router.go        # Route definitions
│   ├── bootstrap/
│   │   └── app.go           # App initialization
│   ├── config/
│   │   ├── db_config.go     # Database configuration
│   │   └── pagination_config.go
│   ├── db/
│   │   ├── db.go            # Database connection
│   │   └── migrations.go    # Auto-migrations
│   ├── dto/
│   │   └── pagination.go    # Shared DTOs
│   ├── handler/
│   │   ├── default.go       # Default handlers
│   │   └── response.go      # Response helpers
│   ├── model/
│   │   └── base.go          # Base model with UUID
│   ├── repository/          # Data access layer
│   └── service/
│       └── container.go     # Dependency injection
├── pkg/
│   └── logger/
│       └── logger.go        # Logging utilities
├── test/
│   ├── mocks/
│   └── unit/
├── .env                     # Environment variables
├── .gitignore
├── Dockerfile
├── go.mod
└── .gen_history.json        # Service generation history
```

## Generated Service Files

When you run `gmk make:service --name user`, these files are created:

```
internal/
├── model/user_model.go          # GORM model
├── repository/user_repository.go # Repository interface + implementation
├── service/user_service.go       # Business logic layer
├── handler/user_handler.go       # REST handlers
├── dto/user_dto.go              # Request/Response DTOs
test/
└── unit/dto/user_input_test.go  # DTO tests
```

### Generated Code Examples

**Repository** - Full CRUD with pagination:
```go
type UserRepository interface {
    FindAll(page, limit int) ([]model.User, int64, error)
    FindByID(id string) (*model.User, error)
    Create(entity *model.User) error
    Update(entity *model.User) error
    Delete(id string) error
}
```

**Handler** - REST endpoints with validation:
```go
func (h *UserHandler) Register(router fiber.Router) {
    router.Get("/", h.list)      // GET /users?page=1&limit=10
    router.Post("/", h.create)   // POST /users
    router.Get("/:id", h.get)    // GET /users/:id
    router.Put("/:id", h.update) // PUT /users/:id
    router.Delete("/:id", h.delete) // DELETE /users/:id
}
```

## Environment Variables

```env
APP_NAME="My Service"
APP_ENV=dev
PORT=8000
API_ROUTE_VERSION=/api/v1
FORCE_MIGRATE=no
DB_USER=root
DB_PASSWORD=
DB_HOST=127.0.0.1
DB_NAME=mydb
DB_PORT=3306
```

## Building with Version Info

```bash
go build -ldflags "\
  -X 'github.com/msdevbytes/gomicrokit/cmd.Version=1.0.0' \
  -X 'github.com/msdevbytes/gomicrokit/cmd.GitCommit=$(git rev-parse HEAD)' \
  -X 'github.com/msdevbytes/gomicrokit/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
  -o gmk .
```

## Development

```bash
# Run tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Build
go build -o gmk .

# Run locally
go run . new myproject --dry-run
```

## License

MIT

## Credits

Built by [msdevbytes](https://github.com/msdevbytes)
