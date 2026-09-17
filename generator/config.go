package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const ConfigFileName = ".gmkrc.json"

type ProjectType string

const (
	ProjectTypeREST ProjectType = "rest"
	ProjectTypeGRPC ProjectType = "grpc"
)

type DBType string

const (
	DBMySQL    DBType = "mysql"
	DBPostgres DBType = "postgres"
	DBSQLite   DBType = "sqlite"
	DBNone     DBType = "none"
)

type CacheStore string

const (
	CacheStoreNone   CacheStore = "none"
	CacheStoreRedis  CacheStore = "redis"
	CacheStoreMemory CacheStore = "memory"
)

type FrameworkType string

const (
	FrameworkFiber  FrameworkType = "fiber"
	FrameworkChi    FrameworkType = "chi"
	FrameworkGRPCGo FrameworkType = "grpc-go"
)

type ProjectConfig struct {
	ProjectName string      `json:"project_name"`
	ModulePath  string      `json:"module_path"`
	OutputDir   string      `json:"output_dir,omitempty"`
	Port        string      `json:"port,omitempty"`
	ProjectType ProjectType `json:"project_type"`
	Framework   string      `json:"framework"`
	DB          string      `json:"db"`
	CacheStore  string      `json:"cache_store,omitempty"`
	WithDocker  bool        `json:"with_docker"`
	WithAir     bool        `json:"with_air,omitempty"`
	RESTVersion string      `json:"rest_version"`
	GRPCVersion string      `json:"grpc_version"`
	Features    []string    `json:"features,omitempty"`
}

const (
	DefaultRESTVersion = "latest"
	DefaultGRPCVersion = "latest"
	DefaultPort        = "8000"
)

var supportedFeatures = map[string]struct{}{
	"auth":       {},
	"cache":      {},
	"config":     {},
	"discovery":  {},
	"messaging":  {},
	"metrics":    {},
	"otel":       {},
	"resilience": {},
	"tracing":    {},
}

func NormalizeProjectType(raw string) (ProjectType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(ProjectTypeREST):
		return ProjectTypeREST, nil
	case string(ProjectTypeGRPC):
		return ProjectTypeGRPC, nil
	default:
		return "", fmt.Errorf("invalid project type %q, expected rest or grpc", raw)
	}
}

func NormalizeDB(projectType ProjectType, raw string) (DBType, error) {
	if projectType == ProjectTypeGRPC {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "", string(DBNone):
			return DBNone, nil
		case string(DBMySQL):
			return DBMySQL, nil
		case string(DBPostgres):
			return DBPostgres, nil
		case string(DBSQLite):
			return DBSQLite, nil
		default:
			return "", fmt.Errorf("invalid grpc db %q, expected none, mysql, postgres, or sqlite", raw)
		}
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(DBMySQL):
		return DBMySQL, nil
	case string(DBPostgres):
		return DBPostgres, nil
	case string(DBSQLite):
		return DBSQLite, nil
	default:
		return "", fmt.Errorf("invalid db %q, expected mysql, postgres, or sqlite", raw)
	}
}

func NormalizeCacheStore(features []string, raw string) (CacheStore, error) {
	if !HasFeature(features, "cache") {
		return CacheStoreNone, nil
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(CacheStoreRedis):
		return CacheStoreRedis, nil
	case string(CacheStoreMemory), "inmemory":
		return CacheStoreMemory, nil
	default:
		return "", fmt.Errorf("invalid cache store %q, expected redis or memory", raw)
	}
}

func NormalizeFramework(projectType ProjectType, raw string) (FrameworkType, error) {
	if projectType == ProjectTypeGRPC {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "", string(FrameworkGRPCGo):
			return FrameworkGRPCGo, nil
		default:
			return "", fmt.Errorf("invalid grpc framework %q, expected grpc-go", raw)
		}
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(FrameworkFiber):
		return FrameworkFiber, nil
	case string(FrameworkChi):
		return FrameworkChi, nil
	default:
		return "", fmt.Errorf("invalid rest framework %q, expected fiber or chi", raw)
	}
}

func NormalizePort(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	if p == "" {
		return DefaultPort, nil
	}
	n, err := strconv.Atoi(p)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid port %q, expected an integer between 1 and 65535", raw)
	}
	return strconv.Itoa(n), nil
}

func SaveProjectConfig(projectPath string, cfg ProjectConfig) error {
	if cfg.RESTVersion == "" {
		cfg.RESTVersion = DefaultRESTVersion
	}
	if cfg.GRPCVersion == "" {
		cfg.GRPCVersion = DefaultGRPCVersion
	}
	cfg.OutputDir = ResolveOutputDir(cfg.OutputDir)
	port, err := NormalizePort(cfg.Port)
	if err != nil {
		return err
	}
	cfg.Port = port
	db, err := NormalizeDB(cfg.ProjectType, cfg.DB)
	if err != nil {
		return err
	}
	cfg.DB = string(db)
	framework, err := NormalizeFramework(cfg.ProjectType, cfg.Framework)
	if err != nil {
		return err
	}
	cfg.Framework = string(framework)
	normalizedFeatures, err := NormalizeFeatures(cfg.Features)
	if err != nil {
		return err
	}
	cfg.Features = normalizedFeatures
	cacheStore, err := NormalizeCacheStore(cfg.Features, cfg.CacheStore)
	if err != nil {
		return err
	}
	cfg.CacheStore = string(cacheStore)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal project config: %w", err)
	}

	path := filepath.Join(projectPath, ConfigFileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write project config: %w", err)
	}

	return nil
}

func LoadProjectConfig(root string) (ProjectConfig, error) {
	path := filepath.Join(root, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, err
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ProjectConfig{}, fmt.Errorf("cannot parse %s: %w", ConfigFileName, err)
	}

	normalized, err := NormalizeProjectType(string(cfg.ProjectType))
	if err != nil {
		return ProjectConfig{}, err
	}
	cfg.ProjectType = normalized
	cfg.OutputDir = ResolveOutputDir(cfg.OutputDir)
	port, err := NormalizePort(cfg.Port)
	if err != nil {
		return ProjectConfig{}, err
	}
	cfg.Port = port
	db, err := NormalizeDB(cfg.ProjectType, cfg.DB)
	if err != nil {
		return ProjectConfig{}, err
	}
	cfg.DB = string(db)
	framework, err := NormalizeFramework(cfg.ProjectType, cfg.Framework)
	if err != nil {
		return ProjectConfig{}, err
	}
	cfg.Framework = string(framework)
	if cfg.RESTVersion == "" {
		cfg.RESTVersion = DefaultRESTVersion
	}
	if cfg.GRPCVersion == "" {
		cfg.GRPCVersion = DefaultGRPCVersion
	}
	normalizedFeatures, err := NormalizeFeatures(cfg.Features)
	if err != nil {
		return ProjectConfig{}, err
	}
	cfg.Features = normalizedFeatures
	cacheStore, err := NormalizeCacheStore(cfg.Features, cfg.CacheStore)
	if err != nil {
		return ProjectConfig{}, err
	}
	cfg.CacheStore = string(cacheStore)

	return cfg, nil
}

func DetectProjectType(root string) ProjectType {
	cfg, err := LoadProjectConfig(root)
	if err != nil {
		return ProjectTypeREST
	}
	if cfg.ProjectType == "" {
		return ProjectTypeREST
	}
	return cfg.ProjectType
}

func DetectProjectConfig(root string) ProjectConfig {
	cfg, err := LoadProjectConfig(root)
	if err == nil {
		return cfg
	}
	return ProjectConfig{
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBMySQL),
		CacheStore:  string(CacheStoreRedis),
		WithAir:     false,
		OutputDir:   ".",
		Port:        DefaultPort,
		RESTVersion: DefaultRESTVersion,
		GRPCVersion: DefaultGRPCVersion,
		Features:    []string{},
	}
}

func ResolveOutputDir(dir string) string {
	trimmed := strings.TrimSpace(dir)
	if trimmed == "" {
		return "."
	}
	return trimmed
}

func ResolveProjectPath(cfg ProjectConfig) string {
	return filepath.Join(ResolveOutputDir(cfg.OutputDir), cfg.ProjectName)
}

// ValidateProjectName rejects empty names and path-like values.
func ValidateProjectName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if trimmed == "." || trimmed == ".." {
		return fmt.Errorf("project name %q is not allowed", trimmed)
	}
	if strings.ContainsAny(trimmed, `/\`) {
		return fmt.Errorf("project name must be a single directory name, not a path")
	}
	if filepath.Base(trimmed) != trimmed {
		return fmt.Errorf("project name must be a single directory name, not a path")
	}
	return nil
}

// ProjectTargetExists reports whether outputDir/projectName already exists.
func ProjectTargetExists(outputDir, projectName string) (bool, string, error) {
	if err := ValidateProjectName(projectName); err != nil {
		return false, "", err
	}
	path := ResolveProjectPath(ProjectConfig{
		OutputDir:   outputDir,
		ProjectName: strings.TrimSpace(projectName),
	})
	_, err := os.Stat(path)
	if err == nil {
		return true, path, nil
	}
	if os.IsNotExist(err) {
		return false, path, nil
	}
	return false, path, err
}

func NormalizeFeatures(features []string) ([]string, error) {
	if len(features) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{}, len(features))
	out := make([]string, 0, len(features))
	for _, raw := range features {
		feature := strings.ToLower(strings.TrimSpace(raw))
		if feature == "" {
			continue
		}
		if _, ok := supportedFeatures[feature]; !ok {
			return nil, fmt.Errorf("unsupported feature %q (supported: auth, cache, config, discovery, messaging, metrics, otel, resilience, tracing)", feature)
		}
		if _, exists := seen[feature]; exists {
			continue
		}
		seen[feature] = struct{}{}
		out = append(out, feature)
	}
	return out, nil
}

func ParseFeaturesCSV(input string) ([]string, error) {
	if strings.TrimSpace(input) == "" {
		return []string{}, nil
	}
	return NormalizeFeatures(strings.Split(input, ","))
}

func FeaturesCSV(features []string) string {
	return strings.Join(features, ",")
}

func HasFeature(features []string, target string) bool {
	for _, f := range features {
		if f == target {
			return true
		}
	}
	return false
}
