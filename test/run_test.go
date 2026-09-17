package gmk_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/msdevbytes/gomicrokit/generator"
)

func TestPlanProjectRunComposeWatch(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "docksvc",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBMySQL),
		Port:        "8088",
		WithDocker:  true,
	}, true, false, true)

	restore := stubRunDeps(t, true, map[string]bool{"docker": true})
	defer restore()

	plan, err := PlanProjectRun(root, RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != RunModeComposeWatch {
		t.Fatalf("mode = %s", plan.Mode)
	}
	got := FormatRunCommand(plan.Steps[0].Args)
	if got != "docker compose --env-file .env up --watch --build" {
		t.Fatalf("cmd = %q", got)
	}
	if !strings.Contains(plan.Hint, "8088") {
		t.Fatalf("hint = %q", plan.Hint)
	}
}

func TestPlanProjectRunComposeDetachAndNoWatch(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "docksvc",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBMySQL),
		WithDocker:  true,
	}, true, false, false)

	restore := stubRunDeps(t, true, map[string]bool{"docker": true})
	defer restore()

	plan, err := PlanProjectRun(root, RunOptions{Detach: true})
	if err != nil {
		t.Fatal(err)
	}
	if FormatRunCommand(plan.Steps[0].Args) != "docker compose up -d --build" {
		t.Fatalf("detach cmd = %q", FormatRunCommand(plan.Steps[0].Args))
	}

	plan, err = PlanProjectRun(root, RunOptions{NoWatch: true})
	if err != nil {
		t.Fatal(err)
	}
	if FormatRunCommand(plan.Steps[0].Args) != "docker compose up --build" {
		t.Fatalf("no-watch cmd = %q", FormatRunCommand(plan.Steps[0].Args))
	}
}

func TestPlanProjectRunDockerUnavailableFallsBackToAir(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "airsvc",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkChi),
		DB:          string(DBSQLite),
		WithDocker:  true,
		WithAir:     true,
	}, true, true, false)

	restore := stubRunDeps(t, false, map[string]bool{"air": true, "go": true})
	defer restore()

	plan, err := PlanProjectRun(root, RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != RunModeHost {
		t.Fatalf("mode = %s", plan.Mode)
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected docker fallback warning")
	}
	if FormatRunCommand(plan.Steps[0].Args) != "air" {
		t.Fatalf("cmd = %q", FormatRunCommand(plan.Steps[0].Args))
	}
}

func TestPlanProjectRunHostWithInfra(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "hostsvc",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBPostgres),
		CacheStore:  string(CacheStoreRedis),
		WithDocker:  true,
		WithAir:     true,
		Features:    []string{"cache", "messaging", "otel"},
	}, true, true, false)

	restore := stubRunDeps(t, true, map[string]bool{"docker": true, "air": true, "go": true})
	defer restore()

	plan, err := PlanProjectRun(root, RunOptions{Host: true})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != RunModeHost {
		t.Fatalf("mode = %s", plan.Mode)
	}
	if len(plan.Steps) != 2 {
		t.Fatalf("steps = %d", len(plan.Steps))
	}
	infra := FormatRunCommand(plan.Steps[0].Args)
	if !strings.HasPrefix(infra, "docker compose up -d ") {
		t.Fatalf("infra = %q", infra)
	}
	for _, svc := range []string{"postgres", "redis", "nats", "otel-collector"} {
		if !strings.Contains(infra, svc) {
			t.Errorf("infra missing %s: %s", svc, infra)
		}
	}
	if strings.Contains(infra, " app") || strings.HasSuffix(infra, " app") {
		t.Errorf("host mode should not start the app container: %s", infra)
	}
	if FormatRunCommand(plan.Steps[1].Args) != "air" {
		t.Fatalf("app = %q", FormatRunCommand(plan.Steps[1].Args))
	}
}

func TestPlanProjectRunNoDockerGoRun(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "gosvc",
		ProjectType: ProjectTypeGRPC,
		Framework:   string(FrameworkGRPCGo),
		DB:          string(DBNone),
		Port:        "9000",
		WithDocker:  false,
		WithAir:     false,
	}, false, false, false)

	restore := stubRunDeps(t, false, map[string]bool{"go": true})
	defer restore()

	plan, err := PlanProjectRun(root, RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if FormatRunCommand(plan.Steps[0].Args) != "go run cmd/main.go" {
		t.Fatalf("cmd = %q", FormatRunCommand(plan.Steps[0].Args))
	}
	if plan.Hint != "gRPC listening on 9000" {
		t.Fatalf("hint = %q", plan.Hint)
	}
}

func TestPlanProjectRunAirMissingFallsBackToGo(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "noair",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBSQLite),
		WithDocker:  false,
		WithAir:     true,
	}, false, true, false)

	restore := stubRunDeps(t, false, map[string]bool{"go": true})
	defer restore()

	plan, err := PlanProjectRun(root, RunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if FormatRunCommand(plan.Steps[0].Args) != "go run cmd/main.go" {
		t.Fatalf("cmd = %q", FormatRunCommand(plan.Steps[0].Args))
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected air fallback warning")
	}
}

func TestPlanProjectRunDetachWithoutDocker(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "nodock",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBSQLite),
		WithDocker:  false,
	}, false, false, false)

	restore := stubRunDeps(t, false, map[string]bool{"go": true})
	defer restore()

	if _, err := PlanProjectRun(root, RunOptions{Detach: true}); err == nil {
		t.Fatal("expected error")
	}
}

func TestFindProjectRootWalksUp(t *testing.T) {
	root := writeRunProject(t, ProjectConfig{
		ProjectName: "nested",
		ProjectType: ProjectTypeREST,
		Framework:   string(FrameworkFiber),
		DB:          string(DBSQLite),
		WithDocker:  false,
	}, false, false, false)
	nested := filepath.Join(root, "internal", "handler")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	got, err := FindProjectRoot(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("root = %q want %q", got, root)
	}
}

func TestPlanProjectRunMissingProject(t *testing.T) {
	dir := t.TempDir()
	if _, err := PlanProjectRun(dir, RunOptions{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestComposeInfraServices(t *testing.T) {
	got := ComposeInfraServices(ProjectConfig{
		DB:         string(DBMySQL),
		CacheStore: string(CacheStoreRedis),
		Features:   []string{"cache", "discovery"},
	})
	want := []string{"mysql", "redis", "consul"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v want %v", got, want)
	}
	if len(ComposeInfraServices(ProjectConfig{DB: string(DBSQLite)})) != 0 {
		t.Fatal("sqlite should have no infra services")
	}
}

func writeRunProject(t *testing.T, cfg ProjectConfig, compose, air, env bool) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SaveProjectConfig(root, cfg); err != nil {
		t.Fatal(err)
	}
	if compose {
		if err := os.WriteFile(filepath.Join(root, "docker-compose.yml"), []byte("services:\n  app:\n    image: example\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if air {
		if err := os.WriteFile(filepath.Join(root, ".air.toml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if env {
		if err := os.WriteFile(filepath.Join(root, ".env"), []byte("PORT=8000\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func stubRunDeps(t *testing.T, dockerReady bool, bins map[string]bool) func() {
	t.Helper()
	origReady := DockerReady
	origLook := RunLookPath
	DockerReady = func() bool { return dockerReady }
	RunLookPath = func(file string) (string, error) {
		if bins[file] {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}
	return func() {
		DockerReady = origReady
		RunLookPath = origLook
	}
}
