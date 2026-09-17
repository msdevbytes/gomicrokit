package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func scaffoldDirs(cfg ProjectConfig) []string {
	base := []string{
		"cmd",
		"internal/bootstrap",
		"internal/config",
		"internal/model",
		"internal/service",
		"pkg/logger",
		"test/unit",
	}
	if HasFeature(cfg.Features, "cache") {
		base = append(base, "internal/cache")
	}
	if HasFeature(cfg.Features, "messaging") {
		base = append(base, "internal/messaging")
	}
	if HasFeature(cfg.Features, "resilience") {
		base = append(base, "internal/resilience")
	}
	if HasFeature(cfg.Features, "discovery") {
		base = append(base, "internal/discovery")
	}
	if HasFeature(cfg.Features, "tracing") || HasFeature(cfg.Features, "otel") {
		base = append(base, "pkg/telemetry")
	}
	if cfg.ProjectType == ProjectTypeREST && HasFeature(cfg.Features, "auth") {
		base = append(base, "internal/middleware")
	}

	if cfg.ProjectType == ProjectTypeGRPC {
		if HasFeature(cfg.Features, "metrics") || HasFeature(cfg.Features, "auth") || HasFeature(cfg.Features, "tracing") {
			base = append(base, "internal/server/grpc/interceptors")
		}
		if DBType(cfg.DB) != DBNone {
			base = append(base, "internal/db")
		}
		return append(base,
			"internal/server/grpc",
			"api/proto",
		)
	}

	return append(base,
		"internal/api",
		"internal/db",
		"internal/dto",
		"internal/handler",
		"internal/repository",
		"test/mocks",
	)
}

func scaffoldFiles(cfg ProjectConfig) []string {
	files := []string{
		".env",
		"README.md",
		"go.mod",
		".gitignore",
		"cmd/main.go",
		"internal/bootstrap/app.go",
		"internal/service/container.go",
		"pkg/logger/logger.go",
	}

	if cfg.WithDocker {
		files = append(files, "Dockerfile", "docker-compose.yml", ".dockerignore", ".air.docker.toml")
		if HasFeature(cfg.Features, "otel") {
			files = append(files, "otel-collector.yaml")
		}
	}
	if cfg.WithAir {
		files = append(files, ".air.toml")
	}
	if HasFeature(cfg.Features, "cache") {
		if cfg.CacheStore == string(CacheStoreMemory) {
			files = append(files, "internal/cache/memory.go")
		} else {
			files = append(files, "internal/cache/redis.go")
		}
	}
	if HasFeature(cfg.Features, "config") {
		files = append(files, "internal/config/app_config.go")
	}
	if HasFeature(cfg.Features, "discovery") {
		files = append(files, "internal/discovery/consul.go")
	}
	if HasFeature(cfg.Features, "messaging") {
		files = append(files, "internal/messaging/nats.go")
	}
	if HasFeature(cfg.Features, "resilience") {
		files = append(files, "internal/resilience/policy.go")
	}
	if HasFeature(cfg.Features, "otel") {
		files = append(files, "pkg/telemetry/otel.go")
	}
	if HasFeature(cfg.Features, "tracing") {
		files = append(files, "pkg/telemetry/tracing.go")
	}
	if cfg.ProjectType == ProjectTypeREST && HasFeature(cfg.Features, "auth") {
		files = append(files, "internal/middleware/auth.go")
	}

	if cfg.ProjectType == ProjectTypeGRPC {
		files = append(files,
			"api/proto/health.proto",
			"internal/server/grpc/server.go",
		)
		if DBType(cfg.DB) != DBNone {
			files = append(files,
				"internal/config/db_config.go",
				"internal/db/db.go",
				"internal/db/migrations.go",
			)
		}
		if HasFeature(cfg.Features, "metrics") {
			files = append(files, "internal/server/grpc/interceptors/metrics.go")
		}
		if HasFeature(cfg.Features, "auth") {
			files = append(files, "internal/server/grpc/interceptors/auth.go")
		}
		if HasFeature(cfg.Features, "tracing") {
			files = append(files, "internal/server/grpc/interceptors/tracing.go")
		}
		return files
	}

	files = append(files,
		"internal/api/router.go",
		"internal/config/db_config.go",
		"internal/config/pagination_config.go",
		"internal/handler/default.go",
		"internal/handler/response.go",
		"internal/db/db.go",
		"internal/db/migrations.go",
		"internal/dto/pagination.go",
		"internal/model/base.go",
	)
	if HasFeature(cfg.Features, "metrics") {
		files = append(files, "internal/handler/metrics.go")
	}
	return files
}

func ScaffoldProject(cfg ProjectConfig) (projectPath string, err error) {
	return ResolveProjectPath(cfg), nil
}

// GenerateProjectFrom is kept for compatibility with the UI loading model.
// It validates selected values and confirms the generation intent.
func GenerateProjectFrom(values map[string]string) error {
	pt, err := NormalizeProjectType(values["protocol"])
	if err != nil {
		return err
	}

	fmt.Printf("Generating project: %s\n", values["project"])
	fmt.Printf("Generating module: %s\n", values["module"])
	fmt.Printf("Using framework: %s\n", values["framework"])
	fmt.Printf("Project type: %s\n", pt)
	fmt.Printf("Using DB: %s (GORM: %s)\n", values["db"], values["gorm"])
	fmt.Printf("Dockerfile: %s\n", values["docker"])
	fmt.Printf("Features: %s\n", values["features"])
	return nil
}

func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveVersion(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return "latest"
	}
	return v
}

func installPackages(projectPath string, packages []string, version string) error {
	for _, pkg := range packages {
		target := pkg + "@" + resolveVersion(version)
		if err := runCommand(projectPath, "go", "get", target); err != nil {
			return fmt.Errorf("failed to install %s: %w", target, err)
		}
	}
	return nil
}

func InstallDeps(projectPath string, cfg ProjectConfig) error {
	cfg.Features, _ = NormalizeFeatures(cfg.Features)
	if cfg.ProjectType == ProjectTypeGRPC {
		packages := []string{
			"github.com/joho/godotenv",
			"google.golang.org/grpc",
		}
		if HasFeature(cfg.Features, "cache") && cfg.CacheStore != string(CacheStoreMemory) {
			packages = append(packages, "github.com/redis/go-redis/v9")
		}
		switch DBType(cfg.DB) {
		case DBPostgres:
			packages = append(packages, "gorm.io/gorm", "gorm.io/driver/postgres")
		case DBSQLite:
			packages = append(packages, "gorm.io/gorm", "github.com/glebarez/sqlite")
		case DBMySQL:
			packages = append(packages, "gorm.io/gorm", "gorm.io/driver/mysql")
		}
		if HasFeature(cfg.Features, "discovery") {
			packages = append(packages, "github.com/hashicorp/consul/api")
		}
		if HasFeature(cfg.Features, "messaging") {
			packages = append(packages, "github.com/nats-io/nats.go")
		}
		if HasFeature(cfg.Features, "resilience") {
			packages = append(packages, "github.com/sony/gobreaker")
		}
		if HasFeature(cfg.Features, "otel") {
			packages = append(packages,
				"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
				"go.opentelemetry.io/otel",
				"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc",
				"go.opentelemetry.io/otel/exporters/stdout/stdouttrace",
				"go.opentelemetry.io/otel/sdk",
			)
		}
		if err := installPackages(projectPath, packages, cfg.GRPCVersion); err != nil {
			return err
		}
	} else {
		packages := []string{
			"github.com/gofrs/uuid/v5",
			"github.com/joho/godotenv",
			"gorm.io/gorm",
		}
		if cfg.Framework == string(FrameworkChi) {
			packages = append(packages, "github.com/go-chi/chi/v5")
		} else {
			packages = append(packages, "github.com/gofiber/fiber/v2")
		}
		switch DBType(cfg.DB) {
		case DBPostgres:
			packages = append(packages, "gorm.io/driver/postgres")
		case DBSQLite:
			packages = append(packages, "github.com/glebarez/sqlite")
		default:
			packages = append(packages, "gorm.io/driver/mysql")
		}
		if HasFeature(cfg.Features, "cache") && cfg.CacheStore != string(CacheStoreMemory) {
			packages = append(packages, "github.com/redis/go-redis/v9")
		}
		if HasFeature(cfg.Features, "discovery") {
			packages = append(packages, "github.com/hashicorp/consul/api")
		}
		if HasFeature(cfg.Features, "messaging") {
			packages = append(packages, "github.com/nats-io/nats.go")
		}
		if HasFeature(cfg.Features, "resilience") {
			packages = append(packages, "github.com/sony/gobreaker")
		}
		if HasFeature(cfg.Features, "otel") {
			packages = append(packages,
				"go.opentelemetry.io/otel",
				"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc",
				"go.opentelemetry.io/otel/exporters/stdout/stdouttrace",
				"go.opentelemetry.io/otel/sdk",
			)
		}
		if err := installPackages(projectPath, packages, cfg.RESTVersion); err != nil {
			return err
		}
	}

	return runCommand(projectPath, "go", "mod", "tidy")
}

// CreateProjectStructure creates folders and base files for the new project
func CreateProjectStructure(cfg ProjectConfig) error {
	if cfg.ProjectType == "" {
		cfg.ProjectType = ProjectTypeREST
	}
	cfg.OutputDir = ResolveOutputDir(cfg.OutputDir)
	if cfg.ProjectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	projectPath := ResolveProjectPath(cfg)

	// Check if directory already exists
	if info, err := os.Stat(projectPath); err == nil {
		if info.IsDir() {
			return fmt.Errorf("directory '%s' already exists. Choose a different name or delete the existing directory", projectPath)
		}
		return fmt.Errorf("a file named '%s' already exists", projectPath)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}
	if err := os.Mkdir(projectPath, 0755); err != nil {
		return fmt.Errorf("cannot create project directory: %w", err)
	}

	// create folders
	for _, dir := range scaffoldDirs(cfg) {
		path := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// create placeholder files
	for _, file := range scaffoldFiles(cfg) {
		path := filepath.Join(projectPath, file)
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Project structure created under %s\n", projectPath)
	return nil
}
