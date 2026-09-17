package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/msdevbytes/gomicrokit/generator"
	"github.com/msdevbytes/gomicrokit/ui"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	db             string
	module         string
	docker         bool
	dryRun         bool
	projectType    string
	saveConfig     bool
	restVersion    string
	grpcVersion    string
	features       string
	cacheStore     string
	port           string
	framework      string
	outputDir      string
	air            bool
	nonInteractive bool
)

func getGoVersion() string {
	ver := runtime.Version() // e.g. "go1.21.3"
	return strings.TrimPrefix(ver, "go")
}

var initCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Initialize a new Go microservice project",
	Long: `Create a new Go microservice project with clean architecture.

Includes: selectable REST framework (Fiber or Chi), gRPC support, GORM ORM,
database options (MySQL/Postgres/SQLite), Docker setup,
Repository pattern, Service layer, and organized project structure.

Examples:
  gmk new                                    # Interactive mode
  gmk new myapp                              # Interactive mode with prefilled project name
  gmk new myapp --non-interactive            # Skip prompts
  gmk new myapp --module github.com/me/myapp # With custom module
  gmk new myapp --framework chi --db postgres # REST + Chi + Postgres
  gmk new myapp --type grpc                  # Generate a gRPC project
  gmk new myapp --features auth,metrics      # Enable optional features
  gmk new myapp --features config,resilience,messaging,otel,discovery
  gmk new myapp --port 8080                  # Set generated app port
  gmk new myapp --air                        # Include Air auto-reload config
  gmk new myapp --dry-run                    # Preview only`,
	Run: func(cmd *cobra.Command, args []string) {
		var input map[string]string

		if nonInteractive {
			if len(args) < 1 {
				fmt.Println("❌ project name is required in --non-interactive mode")
				return
			}
			if !ensureProjectNameAvailable(args[0], outputDir) {
				return
			}
			input = map[string]string{
				"project":      args[0],
				"module":       module,
				"db":           db,
				"framework":    framework,
				"protocol":     projectType,
				"gorm":         "y",
				"docker":       fmt.Sprintf("%v", docker),
				"rest_version": restVersion,
				"grpc_version": grpcVersion,
				"features":     features,
				"cache_store":  cacheStore,
				"port":         port,
				"output_dir":   outputDir,
				"air":          boolToYN(air),
				"save_config":  boolToYN(saveConfig),
			}
		} else {
			prefill := map[string]string{
				"module":       module,
				"db":           db,
				"framework":    framework,
				"protocol":     projectType,
				"gorm":         "y",
				"docker":       boolToYN(docker),
				"rest_version": restVersion,
				"grpc_version": grpcVersion,
				"features":     features,
				"cache_store":  cacheStore,
				"port":         port,
				"output_dir":   outputDir,
				"air":          boolToYN(air),
				"save_config":  boolToYN(saveConfig),
			}
			if len(args) > 0 {
				if !ensureProjectNameAvailable(args[0], outputDir) {
					return
				}
				prefill["project"] = args[0]
				if module == "" {
					prefill["module"] = args[0]
				}
			}
			input = ui.RunInteractiveMenu(prefill)
		}
		if strings.TrimSpace(input["framework"]) == "" {
			if strings.EqualFold(input["protocol"], "grpc") {
				input["framework"] = string(generator.FrameworkGRPCGo)
			} else {
				input["framework"] = string(generator.FrameworkFiber)
			}
		}

		projectName := input["project"]
		modulePath := input["module"]
		if modulePath == "" {
			modulePath = cases.Lower(language.Und).String(projectName)
		}
		selectedOutputDir := outputDir
		if v := strings.TrimSpace(input["output_dir"]); v != "" {
			selectedOutputDir = v
		}
		selectedSaveConfig := saveConfig
		if v := strings.ToLower(strings.TrimSpace(input["save_config"])); v == "y" || v == "n" {
			selectedSaveConfig = v == "y"
		}
		selectedWithAir := air
		if v := strings.ToLower(strings.TrimSpace(input["air"])); v == "y" || v == "n" {
			selectedWithAir = v == "y"
		}
		pt, err := generator.NormalizeProjectType(input["protocol"])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		if projectName == "" {
			fmt.Println("❌ Project name is required")
			return
		}
		if !ensureProjectNameAvailable(projectName, selectedOutputDir) {
			return
		}

		cfg := generator.ProjectConfig{
			ProjectName: projectName,
			ModulePath:  modulePath,
			OutputDir:   selectedOutputDir,
			Port:        input["port"],
			ProjectType: pt,
			Framework:   input["framework"],
			DB:          input["db"],
			WithDocker:  input["docker"] == "y" || input["docker"] == "true",
			WithAir:     selectedWithAir,
			RESTVersion: input["rest_version"],
			GRPCVersion: input["grpc_version"],
			CacheStore:  input["cache_store"],
		}
		dbKind, err := generator.NormalizeDB(cfg.ProjectType, cfg.DB)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		cfg.DB = string(dbKind)
		fwKind, err := generator.NormalizeFramework(cfg.ProjectType, cfg.Framework)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		cfg.Framework = string(fwKind)
		parsedFeatures, err := generator.ParseFeaturesCSV(input["features"])
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		cfg.Features = parsedFeatures
		cacheKind, err := generator.NormalizeCacheStore(cfg.Features, cfg.CacheStore)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		cfg.CacheStore = string(cacheKind)
		portValue, err := generator.NormalizePort(cfg.Port)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		cfg.Port = portValue
		if cfg.RESTVersion == "" {
			cfg.RESTVersion = generator.DefaultRESTVersion
		}
		if cfg.GRPCVersion == "" {
			cfg.GRPCVersion = generator.DefaultGRPCVersion
		}

		// Dry run mode - show what would be created
		if dryRun {
			targetPath := generator.ResolveProjectPath(cfg)
			fmt.Println("🔍 Dry run mode - showing what would be created:")
			fmt.Printf("\n📁 Project: %s\n", projectName)
			fmt.Printf("📍 Target path: %s\n", targetPath)
			fmt.Printf("📦 Module:  %s\n", modulePath)
			fmt.Printf("🛰️  Type:    %s\n", cfg.ProjectType)
			fmt.Printf("🔌 Port:    %s\n", cfg.Port)
			fmt.Printf("🐳 Docker:  %v\n", cfg.WithDocker)
			fmt.Printf("🗄️  DB:      %s\n", cfg.DB)
			fmt.Printf("🧱 Framework: %s\n", cfg.Framework)
			fmt.Printf("🗃️  Cache:   %s\n", cfg.CacheStore)
			fmt.Printf("🔁 Air reload: %v\n", cfg.WithAir)
			fmt.Printf("🧩 REST dep version: %s\n", cfg.RESTVersion)
			fmt.Printf("🧩 gRPC dep version: %s\n", cfg.GRPCVersion)
			if len(cfg.Features) == 0 {
				fmt.Printf("🧩 Features: none\n")
			} else {
				fmt.Printf("🧩 Features: %s\n", generator.FeaturesCSV(cfg.Features))
			}
			fmt.Println("\n📂 Directories that would be created depend on selected type.")
			var dirs []string
			if cfg.ProjectType == generator.ProjectTypeGRPC {
				dirs = []string{
					"cmd", "internal/bootstrap", "internal/config",
					"internal/model", "internal/service", "internal/server/grpc",
					"api/proto", "pkg/logger", "test/unit",
				}
				if cfg.DB != string(generator.DBNone) {
					dirs = append(dirs, "internal/db")
				}
			} else {
				dirs = []string{
					"cmd", "internal/bootstrap", "internal/api", "internal/config",
					"internal/db", "internal/dto", "internal/handler", "internal/model",
					"internal/repository", "internal/service", "pkg/logger", "test/unit",
				}
			}
			if cfg.ProjectType == generator.ProjectTypeREST && generator.HasFeature(cfg.Features, "auth") {
				dirs = append(dirs, "internal/middleware")
			}
			if generator.HasFeature(cfg.Features, "cache") {
				dirs = append(dirs, "internal/cache")
			}
			if generator.HasFeature(cfg.Features, "discovery") {
				dirs = append(dirs, "internal/discovery")
			}
			if generator.HasFeature(cfg.Features, "messaging") {
				dirs = append(dirs, "internal/messaging")
			}
			if generator.HasFeature(cfg.Features, "resilience") {
				dirs = append(dirs, "internal/resilience")
			}
			if generator.HasFeature(cfg.Features, "tracing") {
				dirs = append(dirs, "pkg/telemetry")
			}
			if generator.HasFeature(cfg.Features, "otel") && !generator.HasFeature(cfg.Features, "tracing") {
				dirs = append(dirs, "pkg/telemetry")
			}
			if cfg.ProjectType == generator.ProjectTypeGRPC && (generator.HasFeature(cfg.Features, "metrics") || generator.HasFeature(cfg.Features, "auth") || generator.HasFeature(cfg.Features, "tracing")) {
				dirs = append(dirs, "internal/server/grpc/interceptors")
			}
			for _, d := range dirs {
				fmt.Printf("   - %s/%s\n", targetPath, d)
			}
			fmt.Println("\n📄 Files that would be created:")
			files := []string{"README.md", "cmd/main.go", "go.mod", ".env", ".gitignore", ".gmkrc.json", "internal/bootstrap/app.go", "internal/service/container.go", "pkg/logger/logger.go"}
			if cfg.ProjectType == generator.ProjectTypeGRPC {
				files = append(files, "api/proto/health.proto", "internal/server/grpc/server.go")
				if cfg.DB != string(generator.DBNone) {
					files = append(files, "internal/config/db_config.go", "internal/db/db.go", "internal/db/migrations.go")
				}
			} else {
				files = append(files, "internal/api/router.go", "internal/config/db_config.go", "internal/db/db.go")
			}
			if generator.HasFeature(cfg.Features, "metrics") {
				if cfg.ProjectType == generator.ProjectTypeGRPC {
					files = append(files, "internal/server/grpc/interceptors/metrics.go")
				} else {
					files = append(files, "internal/handler/metrics.go")
				}
			}
			if cfg.ProjectType == generator.ProjectTypeREST && generator.HasFeature(cfg.Features, "auth") {
				files = append(files, "internal/middleware/auth.go")
			}
			if generator.HasFeature(cfg.Features, "cache") {
				if cfg.CacheStore == string(generator.CacheStoreMemory) {
					files = append(files, "internal/cache/memory.go")
				} else {
					files = append(files, "internal/cache/redis.go")
				}
			}
			if generator.HasFeature(cfg.Features, "config") {
				files = append(files, "internal/config/app_config.go")
			}
			if generator.HasFeature(cfg.Features, "discovery") {
				files = append(files, "internal/discovery/consul.go")
			}
			if generator.HasFeature(cfg.Features, "messaging") {
				files = append(files, "internal/messaging/nats.go")
			}
			if generator.HasFeature(cfg.Features, "resilience") {
				files = append(files, "internal/resilience/policy.go")
			}
			if generator.HasFeature(cfg.Features, "otel") {
				files = append(files, "pkg/telemetry/otel.go")
			}
			if generator.HasFeature(cfg.Features, "tracing") {
				if cfg.ProjectType == generator.ProjectTypeGRPC {
					files = append(files, "internal/server/grpc/interceptors/tracing.go")
				}
				files = append(files, "pkg/telemetry/tracing.go")
			}
			if cfg.WithDocker {
				files = append(files, "Dockerfile", "docker-compose.yml", ".dockerignore", ".air.docker.toml")
				if generator.HasFeature(cfg.Features, "otel") {
					files = append(files, "otel-collector.yaml")
				}
			}
			if cfg.WithAir {
				files = append(files, ".air.toml")
			}
			for _, f := range files {
				fmt.Printf("   - %s/%s\n", targetPath, f)
			}
			return
		}

		fmt.Printf("🚀 Scaffolding project: %s\n", projectName)

		if err := generator.CreateProjectStructure(cfg); err != nil {
			fmt.Printf("❌ Error creating project structure: %v\n", err)
			return
		}

		projectPath := generator.ResolveProjectPath(cfg)
		err = generator.RenderTemplates(projectPath, generator.TemplateData{
			ProjectName: projectName,
			ModulePath:  modulePath,
			GoVersion:   getGoVersion(),
			Port:        cfg.Port,
			ProjectType: string(cfg.ProjectType),
			Framework:   cfg.Framework,
			DBType:      cfg.DB,
		}, cfg)
		if err != nil {
			fmt.Printf("❌ Error generating templates: %v\n", err)
			return
		}

		projectPath, err = generator.ScaffoldProject(cfg)
		if err != nil {
			fmt.Printf("❌ Error during generation: %v\n", err)
			return
		}

		if selectedSaveConfig {
			if err := generator.SaveProjectConfig(projectPath, cfg); err != nil {
				fmt.Printf("⚠️ Warning: project created but failed to save %s: %v\n", generator.ConfigFileName, err)
			}
		}

		fmt.Println("📦 Installing dependencies...")
		if err := generator.InstallDeps(projectPath, cfg); err != nil {
			fmt.Printf("❌ Error during dependency installation: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Project '%s' created successfully!\n", projectName)
		fmt.Printf("\n📋 Next steps:\n")
		fmt.Printf("   cd %s\n", projectPath)
		if generator.HasFeature(cfg.Features, "auth") {
			fmt.Printf("   gmk key:generate\n")
		}
		fmt.Printf("   gmk run\n")
	},
}

func init() {
	initCmd.Flags().StringVar(&db, "db", "mysql", "Database type. REST: mysql|postgres|sqlite. gRPC: none|mysql|postgres|sqlite")
	initCmd.Flags().BoolVar(&docker, "docker", true, "Include Dockerfile, docker-compose.yml, and related Docker files")
	initCmd.Flags().StringVar(&module, "module", "", "Go module path (e.g. github.com/user/project)")
	initCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be created without making changes")
	initCmd.Flags().StringVar(&projectType, "type", "rest", "Microservice type (rest or grpc)")
	initCmd.Flags().StringVar(&framework, "framework", "", "Framework: rest supports fiber|chi, grpc supports grpc-go")
	initCmd.Flags().BoolVar(&saveConfig, "save-config", true, "Save selected options to .gmkrc.json")
	initCmd.Flags().StringVar(&restVersion, "rest-version", generator.DefaultRESTVersion, "REST dependency version (e.g. latest, v2.52.9)")
	initCmd.Flags().StringVar(&grpcVersion, "grpc-version", generator.DefaultGRPCVersion, "gRPC dependency version (e.g. latest, v1.75.1)")
	initCmd.Flags().StringVar(&features, "features", "", "Comma-separated optional features: auth,cache,config,discovery,messaging,metrics,otel,resilience,tracing")
	initCmd.Flags().StringVar(&cacheStore, "cache-store", "redis", "Cache store when cache feature is enabled (redis or memory)")
	initCmd.Flags().StringVar(&port, "port", generator.DefaultPort, "Application port for generated .env")
	initCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output folder where the project directory will be created")
	initCmd.Flags().BoolVar(&air, "air", false, "Include Air auto-reload config (.air.toml)")
	initCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Skip prompts and use flags/arguments directly")
}

func boolToYN(v bool) string {
	if v {
		return "y"
	}
	return "n"
}

func ensureProjectNameAvailable(name, dir string) bool {
	exists, path, err := generator.ProjectTargetExists(dir, name)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return false
	}
	if exists {
		fmt.Printf("❌ directory '%s' already exists. Choose a different name or delete the existing directory\n", path)
		return false
	}
	return true
}
