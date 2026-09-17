package gmk_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/msdevbytes/gomicrokit/generator"
)

func TestSanitizeDBName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{in: "myrestsvc", want: "myrestsvc"},
		{in: "My-REST_svc", want: "my_rest_svc"},
		{in: "9lives", want: "db_9lives"},
		{in: "   ", want: "appdb"},
	}
	for _, tt := range tests {
		if got := SanitizeDBName(tt.in); got != tt.want {
			t.Fatalf("SanitizeDBName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRenderDockerComposeFeatureGating(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-compose-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(originalDir) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	cfg := ProjectConfig{
		ProjectName: "stacksvc",
		ModulePath:  "github.com/acme/stacksvc",
		Port:        "8088",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBMySQL),
		CacheStore:  string(CacheStoreRedis),
		WithDocker:  true,
		Features:    []string{"cache", "messaging", "discovery", "otel"},
	}

	if err := CreateProjectStructure(cfg); err != nil {
		t.Fatalf("CreateProjectStructure: %v", err)
	}
	if err := RenderTemplates("stacksvc", TemplateData{
		ProjectName: cfg.ProjectName,
		ModulePath:  cfg.ModulePath,
		GoVersion:   "1.24",
		ProjectType: string(cfg.ProjectType),
		Framework:   cfg.Framework,
		DBType:      cfg.DB,
	}, cfg); err != nil {
		t.Fatalf("RenderTemplates: %v", err)
	}

	compose, err := os.ReadFile(filepath.Join("stacksvc", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	body := string(compose)
	for _, want := range []string{
		"  mysql:",
		"  redis:",
		"  nats:",
		"  consul:",
		"  otel-collector:",
		"  app:",
		"DB_HOST: mysql",
		"REDIS_HOST: redis",
		"NATS_URL: nats://nats:4222",
		"CONSUL_ADDRESS: consul:8500",
		"OTEL_EXPORTER_OTLP_ENDPOINT: otel-collector:4317",
		`"8088:8088"`,
		"${DB_PORT_FORWARD:-3306}:3306",
		"target: dev",
		"action: sync",
		".air.docker.toml",
		"go-mod-cache",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("compose missing %q", want)
		}
	}
	if strings.Contains(body, ".:/src") {
		t.Error("bind-mounting the project root with compose watch causes a restart loop")
	}
	for _, notWant := range []string{"  postgres:", "  sqlite:"} {
		if strings.Contains(body, notWant) {
			t.Errorf("compose unexpectedly contains %q", notWant)
		}
	}

	if _, err := os.Stat(filepath.Join("stacksvc", "otel-collector.yaml")); err != nil {
		t.Fatalf("otel-collector.yaml: %v", err)
	}

	dockerfile, err := os.ReadFile(filepath.Join("stacksvc", "Dockerfile"))
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}
	df := string(dockerfile)
	if strings.Contains(df, "{{.Port}}") {
		t.Error("rendered Dockerfile still contains {{.Port}}")
	}
	if !strings.Contains(df, "ARG PORT=8088") {
		t.Error("rendered Dockerfile should bake selected port into ARG PORT")
	}
	if !strings.Contains(df, "AS dev") {
		t.Error("Dockerfile should include a live-reload dev stage")
	}
	if _, err := os.Stat(filepath.Join("stacksvc", ".air.docker.toml")); err != nil {
		t.Fatalf(".air.docker.toml: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join("stacksvc", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	readmeBody := string(readme)
	for _, want := range []string{"# stacksvc", "gmk run", "docker compose up --watch --build", "gmk make:service"} {
		if !strings.Contains(readmeBody, want) {
			t.Errorf("generated README missing %q", want)
		}
	}
	if !strings.Contains(body, `PORT: "8088"`) {
		t.Error("compose should pass selected PORT as a build arg")
	}

	envBody, err := os.ReadFile(filepath.Join("stacksvc", ".env"))
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	env := string(envBody)
	if !strings.Contains(env, "REDIS_HOST=localhost") {
		t.Error(".env should keep localhost Redis for host-side go run")
	}
	if !strings.Contains(env, "NATS_URL=nats://127.0.0.1:4222") {
		t.Error(".env should keep localhost NATS for host-side go run")
	}
	if strings.Contains(env, "#DB_USER=") {
		t.Error(".env commented out database credentials")
	}
	if !strings.Contains(env, "DB_USER=root") {
		t.Error(".env should set uncommented MySQL credentials")
	}
	if !strings.Contains(env, "DB_PORT_FORWARD=3306") {
		t.Error(".env should set DB_PORT_FORWARD for host port publishing")
	}
	if strings.Contains(env, "# observabilityOTEL") || strings.Contains(env, "#OTEL_EXPORTER_OTLP_ENDPOINT") {
		t.Error(".env commented out OTEL endpoint")
	}
	if !strings.Contains(env, "OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317") {
		t.Error(".env should point OTLP at published collector when docker+otel")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Log("docker not on PATH; skipping compose config validation")
		return
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Logf("docker daemon unavailable; skipping compose config validation: %v", err)
		return
	}
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "--env-file", ".env", "config")
	cmd.Dir = "stacksvc"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker compose config failed: %v\n%s", err, out)
	}
}

func TestRenderDockerComposeOmitsUnselectedServices(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-compose-sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(originalDir) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	cfg := ProjectConfig{
		ProjectName: "litesvc",
		ModulePath:  "github.com/acme/litesvc",
		Port:        "8000",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkChi),
		DB:          string(DBSQLite),
		CacheStore:  string(CacheStoreMemory),
		WithDocker:  true,
		Features:    []string{"cache"},
	}

	if err := CreateProjectStructure(cfg); err != nil {
		t.Fatalf("CreateProjectStructure: %v", err)
	}
	if err := RenderTemplates("litesvc", TemplateData{
		ProjectName: cfg.ProjectName,
		ModulePath:  cfg.ModulePath,
		GoVersion:   "1.24",
		ProjectType: string(cfg.ProjectType),
		Framework:   cfg.Framework,
		DBType:      cfg.DB,
	}, cfg); err != nil {
		t.Fatalf("RenderTemplates: %v", err)
	}

	compose, err := os.ReadFile(filepath.Join("litesvc", "docker-compose.yml"))
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	body := string(compose)
	for _, notWant := range []string{"  mysql:", "  postgres:", "  redis:", "  nats:", "  consul:", "  otel-collector:", `profiles: ["app"]`} {
		if strings.Contains(body, notWant) {
			t.Errorf("sqlite/memory compose unexpectedly contains %q", notWant)
		}
	}
	if !strings.Contains(body, "  app:") {
		t.Error("compose should still include the app service")
	}
	if !strings.Contains(body, "target: dev") || !strings.Contains(body, "action: sync") {
		t.Error("sqlite compose should still enable live reload watch")
	}
	if !strings.Contains(body, "go-mod-cache:") {
		t.Error("compose should declare go-mod-cache volume")
	}
	if _, err := os.Stat(filepath.Join("litesvc", "otel-collector.yaml")); !os.IsNotExist(err) {
		t.Error("otel-collector.yaml should not be generated without otel")
	}
}
