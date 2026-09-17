package gmk_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/msdevbytes/gomicrokit/generator"
)

func TestNormalizeProjectType(t *testing.T) {
	tests := []struct {
		in      string
		want    ProjectType
		wantErr bool
	}{
		{in: "", want: ProjectTypeREST},
		{in: "rest", want: ProjectTypeREST},
		{in: "grpc", want: ProjectTypeGRPC},
		{in: "REST", want: ProjectTypeREST},
		{in: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		got, err := NormalizeProjectType(tt.in)
		if tt.wantErr && err == nil {
			t.Fatalf("NormalizeProjectType(%q) expected error", tt.in)
		}
		if !tt.wantErr && got != tt.want {
			t.Fatalf("NormalizeProjectType(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSaveLoadProjectConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := ProjectConfig{
		ProjectName: "svc",
		ModulePath:  "github.com/acme/svc",
		OutputDir:   "tmp/out",
		Port:        "9001",
		ProjectType: ProjectTypeGRPC,
		Framework:   "grpc-go",
		DB:          "postgres",
		CacheStore:  "memory",
		WithDocker:  true,
		WithAir:     true,
		Features:    []string{"metrics", "auth", "cache"},
	}

	if err := SaveProjectConfig(tmpDir, cfg); err != nil {
		t.Fatalf("SaveProjectConfig() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, ConfigFileName)); err != nil {
		t.Fatalf("config file not written: %v", err)
	}

	got, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadProjectConfig() error = %v", err)
	}

	if got.ProjectType != ProjectTypeGRPC {
		t.Fatalf("loaded project type = %q, want %q", got.ProjectType, ProjectTypeGRPC)
	}
	if got.ModulePath != cfg.ModulePath {
		t.Fatalf("loaded module = %q, want %q", got.ModulePath, cfg.ModulePath)
	}
	if got.OutputDir != "tmp/out" {
		t.Fatalf("loaded output dir = %q, want %q", got.OutputDir, "tmp/out")
	}
	if got.Port != "9001" {
		t.Fatalf("loaded port = %q, want %q", got.Port, "9001")
	}
	if !got.WithAir {
		t.Fatalf("loaded with_air = %v, want true", got.WithAir)
	}
	if len(got.Features) != 3 {
		t.Fatalf("loaded features length = %d, want 3", len(got.Features))
	}
	if got.DB != "postgres" {
		t.Fatalf("loaded db = %q, want %q", got.DB, "postgres")
	}
	if got.CacheStore != "memory" {
		t.Fatalf("loaded cache store = %q, want %q", got.CacheStore, "memory")
	}
}

func TestParseFeaturesCSV(t *testing.T) {
	got, err := ParseFeaturesCSV("auth,metrics,auth")
	if err != nil {
		t.Fatalf("ParseFeaturesCSV() error = %v", err)
	}
	if FeaturesCSV(got) != "auth,metrics" {
		t.Fatalf("features csv = %q, want %q", FeaturesCSV(got), "auth,metrics")
	}

	if _, err := ParseFeaturesCSV("unknown"); err == nil {
		t.Fatalf("expected error for unknown feature")
	}

	got, err = ParseFeaturesCSV("discovery,auth")
	if err != nil {
		t.Fatalf("ParseFeaturesCSV(discovery,auth) error = %v", err)
	}
	if FeaturesCSV(got) != "discovery,auth" {
		t.Fatalf("features csv = %q, want %q", FeaturesCSV(got), "discovery,auth")
	}

	got, err = ParseFeaturesCSV("config,resilience,messaging,otel")
	if err != nil {
		t.Fatalf("ParseFeaturesCSV(config,resilience,messaging,otel) error = %v", err)
	}
	if FeaturesCSV(got) != "config,resilience,messaging,otel" {
		t.Fatalf("features csv = %q, want %q", FeaturesCSV(got), "config,resilience,messaging,otel")
	}
}

func TestNormalizeDB(t *testing.T) {
	tests := []struct {
		name        string
		projectType ProjectType
		db          string
		want        DBType
		wantErr     bool
	}{
		{name: "grpc default none", projectType: ProjectTypeGRPC, db: "", want: DBNone},
		{name: "grpc mysql", projectType: ProjectTypeGRPC, db: "mysql", want: DBMySQL},
		{name: "grpc postgres", projectType: ProjectTypeGRPC, db: "postgres", want: DBPostgres},
		{name: "grpc sqlite", projectType: ProjectTypeGRPC, db: "sqlite", want: DBSQLite},
		{name: "grpc invalid", projectType: ProjectTypeGRPC, db: "mongo", wantErr: true},
		{name: "rest default mysql", projectType: ProjectTypeREST, db: "", want: DBMySQL},
		{name: "rest postgres", projectType: ProjectTypeREST, db: "postgres", want: DBPostgres},
		{name: "rest sqlite", projectType: ProjectTypeREST, db: "sqlite", want: DBSQLite},
		{name: "rest invalid", projectType: ProjectTypeREST, db: "mongo", wantErr: true},
	}

	for _, tt := range tests {
		got, err := NormalizeDB(tt.projectType, tt.db)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", tt.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: got %q want %q", tt.name, got, tt.want)
		}
	}
}

func TestNormalizeCacheStore(t *testing.T) {
	tests := []struct {
		name     string
		features []string
		input    string
		want     CacheStore
		wantErr  bool
	}{
		{name: "cache disabled forces none", features: []string{"metrics"}, input: "redis", want: CacheStoreNone},
		{name: "cache enabled default redis", features: []string{"cache"}, input: "", want: CacheStoreRedis},
		{name: "cache enabled redis", features: []string{"cache"}, input: "redis", want: CacheStoreRedis},
		{name: "cache enabled memory", features: []string{"cache"}, input: "memory", want: CacheStoreMemory},
		{name: "cache enabled invalid", features: []string{"cache"}, input: "memcached", wantErr: true},
	}

	for _, tt := range tests {
		got, err := NormalizeCacheStore(tt.features, tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", tt.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: got %q want %q", tt.name, got, tt.want)
		}
	}
}

func TestNormalizeFramework(t *testing.T) {
	tests := []struct {
		name        string
		projectType ProjectType
		framework   string
		want        FrameworkType
		wantErr     bool
	}{
		{name: "rest default fiber", projectType: ProjectTypeREST, framework: "", want: FrameworkFiber},
		{name: "rest chi", projectType: ProjectTypeREST, framework: "chi", want: FrameworkChi},
		{name: "rest invalid", projectType: ProjectTypeREST, framework: "echo", wantErr: true},
		{name: "grpc default", projectType: ProjectTypeGRPC, framework: "", want: FrameworkGRPCGo},
		{name: "grpc valid", projectType: ProjectTypeGRPC, framework: "grpc-go", want: FrameworkGRPCGo},
		{name: "grpc invalid", projectType: ProjectTypeGRPC, framework: "fiber", wantErr: true},
	}

	for _, tt := range tests {
		got, err := NormalizeFramework(tt.projectType, tt.framework)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", tt.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: got %q want %q", tt.name, got, tt.want)
		}
	}
}

func TestNormalizePort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default empty", input: "", want: "8000"},
		{name: "valid", input: "3000", want: "3000"},
		{name: "trims spaces", input: " 9090 ", want: "9090"},
		{name: "too low", input: "0", wantErr: true},
		{name: "too high", input: "70000", wantErr: true},
		{name: "not number", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		got, err := NormalizePort(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", tt.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: got %q want %q", tt.name, got, tt.want)
		}
	}
}

func TestValidateProjectName(t *testing.T) {
	if err := ValidateProjectName("myservice"); err != nil {
		t.Fatalf("valid name: %v", err)
	}
	for _, name := range []string{"", ".", "..", "foo/bar", `foo\bar`} {
		if err := ValidateProjectName(name); err == nil {
			t.Fatalf("ValidateProjectName(%q) expected error", name)
		}
	}
}

func TestProjectTargetExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-exists-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	exists, _, err := ProjectTargetExists(tmpDir, "freshsvc")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected freshsvc to be available")
	}

	if err := os.Mkdir(filepath.Join(tmpDir, "takensvc"), 0755); err != nil {
		t.Fatal(err)
	}
	exists, path, err := ProjectTargetExists(tmpDir, "takensvc")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected takensvc to exist")
	}
	if path != filepath.Join(tmpDir, "takensvc") {
		t.Fatalf("path = %q", path)
	}

	if _, _, err := ProjectTargetExists(tmpDir, "../escape"); err == nil {
		t.Fatal("expected path-like name to error")
	}
}
