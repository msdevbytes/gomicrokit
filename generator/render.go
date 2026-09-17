package generator

import (
	"embed"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*
var tmplFS embed.FS

type TemplateData struct {
	ProjectName     string
	ModulePath      string
	GoVersion       string
	Port            string
	ProjectType     string
	Framework       string
	DBType          string
	DBName          string
	CacheStore      string
	WithDocker      bool
	HasAuth         bool
	HasCache        bool
	HasConfig       bool
	HasDiscovery    bool
	HasMessaging    bool
	HasMetrics      bool
	HasOTel         bool
	HasResilience   bool
	HasTracing      bool
	HasRedis        bool
	HasMySQL        bool
	HasPostgres     bool
	HasComposeInfra bool
	WithAir         bool
}

// RenderTemplates renders all embedded templates to the project folder
func RenderTemplates(projectPath string, data TemplateData, cfg ProjectConfig) error {
	cfg.Features, _ = NormalizeFeatures(cfg.Features)
	if cfg.Port == "" {
		cfg.Port = DefaultPort
	}
	if strings.TrimSpace(data.GoVersion) == "" {
		data.GoVersion = "1.24"
	}
	data.Port = cfg.Port
	data.Framework = cfg.Framework
	data.DBType = cfg.DB
	data.DBName = SanitizeDBName(cfg.ProjectName)
	data.WithDocker = cfg.WithDocker
	data.WithAir = cfg.WithAir
	data.HasAuth = HasFeature(cfg.Features, "auth")
	data.HasCache = HasFeature(cfg.Features, "cache")
	data.HasConfig = HasFeature(cfg.Features, "config")
	data.HasDiscovery = HasFeature(cfg.Features, "discovery")
	data.HasMessaging = HasFeature(cfg.Features, "messaging")
	data.HasMetrics = HasFeature(cfg.Features, "metrics")
	data.HasOTel = HasFeature(cfg.Features, "otel")
	data.HasResilience = HasFeature(cfg.Features, "resilience")
	data.HasTracing = HasFeature(cfg.Features, "tracing")
	if data.HasCache {
		if cfg.CacheStore == "" {
			data.CacheStore = string(CacheStoreRedis)
		} else {
			data.CacheStore = cfg.CacheStore
		}
	} else {
		data.CacheStore = string(CacheStoreNone)
	}
	data.HasRedis = data.HasCache && data.CacheStore == string(CacheStoreRedis)
	data.HasMySQL = cfg.DB == string(DBMySQL)
	data.HasPostgres = cfg.DB == string(DBPostgres)
	data.HasComposeInfra = data.HasMySQL || data.HasPostgres || data.HasRedis || data.HasMessaging || data.HasDiscovery || data.HasOTel

	files := map[string]string{
		"README.md":                     "README.tmpl",
		"cmd/main.go":                   "main.tmpl",
		"go.mod":                        "go.mod.tmpl",
		".gitignore":                    "gitignore.tmpl",
		".env":                          "env.tmpl",
		"internal/bootstrap/app.go":     "internal/bootstrap/app.tmpl",
		"internal/service/container.go": "internal/service/container.tmpl",
		"pkg/logger/logger.go":          "pkg/logger/logger.tmpl",
	}

	if cfg.WithDocker {
		files["Dockerfile"] = "Dockerfile.tmpl"
		files["docker-compose.yml"] = "docker-compose.yml.tmpl"
		files[".dockerignore"] = "dockerignore.tmpl"
		files[".air.docker.toml"] = "air.docker.toml.tmpl"
		if data.HasOTel {
			files["otel-collector.yaml"] = "otel-collector.yaml.tmpl"
		}
	}
	if cfg.WithAir {
		files[".air.toml"] = "air.toml.tmpl"
	}

	if cfg.ProjectType == ProjectTypeGRPC {
		files["api/proto/health.proto"] = "grpc/health.proto.tmpl"
		files["internal/server/grpc/server.go"] = "grpc/server.tmpl"
		if cfg.DB != string(DBNone) {
			files["internal/config/db_config.go"] = "internal/config/db_config.tmpl"
			files["internal/db/db.go"] = "internal/db/db.tmpl"
			files["internal/db/migrations.go"] = "internal/db/migrations.tmpl"
		}
		if data.HasMetrics {
			files["internal/server/grpc/interceptors/metrics.go"] = "grpc/interceptors/metrics.tmpl"
		}
		if data.HasAuth {
			files["internal/server/grpc/interceptors/auth.go"] = "grpc/interceptors/auth.tmpl"
		}
		if data.HasTracing {
			files["internal/server/grpc/interceptors/tracing.go"] = "grpc/interceptors/tracing.tmpl"
		}
	} else {
		if cfg.Framework == string(FrameworkChi) {
			files["internal/api/router.go"] = "internal/api/router_chi.tmpl"
			files["internal/bootstrap/app.go"] = "internal/bootstrap/app_chi.tmpl"
			files["internal/handler/default.go"] = "internal/handler/default_chi.tmpl"
			files["internal/handler/response.go"] = "internal/handler/response_chi.tmpl"
			if data.HasMetrics {
				files["internal/handler/metrics.go"] = "internal/handler/metrics_chi.tmpl"
			}
			if data.HasAuth {
				files["internal/middleware/auth.go"] = "internal/middleware/auth_chi.tmpl"
			}
		} else {
			files["internal/api/router.go"] = "internal/api/router.tmpl"
			files["internal/bootstrap/app.go"] = "internal/bootstrap/app.tmpl"
			files["internal/handler/default.go"] = "internal/handler/default.tmpl"
			files["internal/handler/response.go"] = "internal/handler/response.tmpl"
			if data.HasMetrics {
				files["internal/handler/metrics.go"] = "internal/handler/metrics.tmpl"
			}
			if data.HasAuth {
				files["internal/middleware/auth.go"] = "internal/middleware/auth.tmpl"
			}
		}
		files["internal/config/db_config.go"] = "internal/config/db_config.tmpl"
		files["internal/config/pagination_config.go"] = "internal/config/pagination_config.tmpl"
		files["internal/db/db.go"] = "internal/db/db.tmpl"
		files["internal/db/migrations.go"] = "internal/db/migrations.tmpl"
		files["internal/dto/pagination.go"] = "internal/dto/pagination.tmpl"
		files["internal/model/base.go"] = "internal/model/base.tmpl"
	}
	if data.HasCache {
		if data.CacheStore == string(CacheStoreMemory) {
			files["internal/cache/memory.go"] = "internal/cache/memory.tmpl"
		} else {
			files["internal/cache/redis.go"] = "internal/cache/redis.tmpl"
		}
	}
	if data.HasConfig {
		files["internal/config/app_config.go"] = "internal/config/app_config.tmpl"
	}
	if HasFeature(cfg.Features, "discovery") {
		files["internal/discovery/consul.go"] = "internal/discovery/consul.tmpl"
	}
	if data.HasMessaging {
		files["internal/messaging/nats.go"] = "internal/messaging/nats.tmpl"
	}
	if data.HasResilience {
		files["internal/resilience/policy.go"] = "internal/resilience/policy.tmpl"
	}
	if data.HasOTel {
		files["pkg/telemetry/otel.go"] = "pkg/telemetry/otel.tmpl"
	}
	if data.HasTracing {
		files["pkg/telemetry/tracing.go"] = "pkg/telemetry/tracing.tmpl"
	}

	// Sort the output file names for deterministic rendering order
	var outFiles []string
	for out := range files {
		outFiles = append(outFiles, out)
	}
	sort.Strings(outFiles)

	for _, out := range outFiles {
		tmpl := files[out]
		if err := renderTemplateToFile(projectPath, out, tmpl, data); err != nil {
			return err
		}
	}

	return nil
}

// renderTemplateToFile renders a single template to a file, ensuring proper cleanup
func renderTemplateToFile(projectPath, out, tmpl string, data TemplateData) error {
	t, err := template.ParseFS(tmplFS, "templates/"+tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(projectPath, out))
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, data)
}

func SanitizeDBName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteByte('_')
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "appdb"
	}
	if unicode.IsDigit(rune(out[0])) {
		return "db_" + out
	}
	return out
}
