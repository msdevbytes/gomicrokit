package generator

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	RunModeComposeWatch  = "compose-watch"
	RunModeComposeUp     = "compose-up"
	RunModeComposeDetach = "compose-detach"
	RunModeHost          = "host"
)

type RunOptions struct {
	Host    bool
	Detach  bool
	NoWatch bool
}

type RunStep struct {
	Label string
	Args  []string
}

type RunPlan struct {
	Dir      string
	Mode     string
	Config   ProjectConfig
	Steps    []RunStep
	Warnings []string
	Hint     string
}

var (
	RunLookPath = exec.LookPath
	DockerReady = DefaultDockerReady
	RunExec     = DefaultRunExec
)

func DefaultDockerReady() bool {
	if _, err := RunLookPath("docker"); err != nil {
		return false
	}
	cmd := exec.Command("docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func DefaultRunExec(dir string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func ExecuteRunPlan(plan RunPlan) error {
	for _, step := range plan.Steps {
		if err := RunExec(plan.Dir, step.Args); err != nil {
			return fmt.Errorf("%s: %w", step.Label, err)
		}
	}
	return nil
}

func FindProjectRoot(start string) (string, error) {
	dir, err := filepath.Abs(strings.TrimSpace(start))
	if err != nil {
		return "", err
	}
	cur := dir
	for {
		if fileExists(filepath.Join(cur, ConfigFileName)) {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	if fileExists(filepath.Join(dir, "cmd", "main.go")) {
		return dir, nil
	}
	return "", fmt.Errorf("not a generated GMK project (no %s or cmd/main.go). Run this from the project directory", ConfigFileName)
}

func PlanProjectRun(start string, opts RunOptions) (RunPlan, error) {
	root, err := FindProjectRoot(start)
	if err != nil {
		return RunPlan{}, err
	}
	if !fileExists(filepath.Join(root, "cmd", "main.go")) {
		return RunPlan{}, fmt.Errorf("not a generated GMK project (missing cmd/main.go)")
	}

	cfg, cfgErr := LoadProjectConfig(root)
	if cfgErr != nil {
		cfg = inferRunConfig(root)
	}

	hasCompose := composeFileExists(root)
	if cfg.WithDocker && !hasCompose {
		cfg.WithDocker = false
	}

	wantDocker := !opts.Host && (cfg.WithDocker || hasCompose)
	dockerOK := wantDocker && DockerReady()

	if opts.Detach && !wantDocker {
		return RunPlan{}, fmt.Errorf("--detach only applies when the project has Docker Compose")
	}
	if opts.NoWatch && !wantDocker {
		return RunPlan{}, fmt.Errorf("--no-watch only applies when the project has Docker Compose")
	}

	plan := RunPlan{
		Dir:    root,
		Config: cfg,
		Hint:   runHint(cfg),
	}

	if wantDocker && !dockerOK {
		plan.Warnings = append(plan.Warnings, "Docker is not available; starting the app on the host instead")
	}

	if wantDocker && dockerOK {
		args := composePrefix(root)
		switch {
		case opts.Detach:
			plan.Mode = RunModeComposeDetach
			args = append(args, "up", "-d", "--build")
			plan.Steps = []RunStep{{Label: "compose detach", Args: args}}
		case opts.NoWatch:
			plan.Mode = RunModeComposeUp
			args = append(args, "up", "--build")
			plan.Steps = []RunStep{{Label: "compose up", Args: args}}
		default:
			plan.Mode = RunModeComposeWatch
			args = append(args, "up", "--watch", "--build")
			plan.Steps = []RunStep{{Label: "compose watch", Args: args}}
		}
		return plan, nil
	}

	plan.Mode = RunModeHost
	if hasCompose && DockerReady() {
		if infra := ComposeInfraServices(cfg); len(infra) > 0 {
			args := append(composePrefix(root), "up", "-d")
			args = append(args, infra...)
			plan.Steps = append(plan.Steps, RunStep{Label: "compose infra", Args: args})
		}
	}

	app, warn, err := hostAppCommand(cfg)
	if err != nil {
		return RunPlan{}, err
	}
	if warn != "" {
		plan.Warnings = append(plan.Warnings, warn)
	}
	plan.Steps = append(plan.Steps, RunStep{Label: "app", Args: app})
	return plan, nil
}

func ComposeInfraServices(cfg ProjectConfig) []string {
	var svcs []string
	switch strings.ToLower(strings.TrimSpace(cfg.DB)) {
	case string(DBMySQL):
		svcs = append(svcs, "mysql")
	case string(DBPostgres):
		svcs = append(svcs, "postgres")
	}
	if HasFeature(cfg.Features, "cache") && strings.EqualFold(cfg.CacheStore, string(CacheStoreRedis)) {
		svcs = append(svcs, "redis")
	}
	if HasFeature(cfg.Features, "messaging") {
		svcs = append(svcs, "nats")
	}
	if HasFeature(cfg.Features, "discovery") {
		svcs = append(svcs, "consul")
	}
	if HasFeature(cfg.Features, "otel") {
		svcs = append(svcs, "otel-collector")
	}
	return svcs
}

func inferRunConfig(root string) ProjectConfig {
	cfg := DetectProjectConfig(root)
	cfg.WithDocker = composeFileExists(root)
	cfg.WithAir = fileExists(filepath.Join(root, ".air.toml"))
	if cfg.Port == "" {
		cfg.Port = DefaultPort
	}
	return cfg
}

func hostAppCommand(cfg ProjectConfig) ([]string, string, error) {
	if cfg.WithAir {
		if _, err := RunLookPath("air"); err == nil {
			return []string{"air"}, "", nil
		}
		if _, err := RunLookPath("go"); err != nil {
			return nil, "", fmt.Errorf("air is not on PATH, and Go is not on PATH. Install Air (https://github.com/air-verse/air) or Go")
		}
		return []string{"go", "run", "cmd/main.go"}, "air is not on PATH; using go run cmd/main.go", nil
	}
	if _, err := RunLookPath("go"); err != nil {
		return nil, "", fmt.Errorf("Go is not on PATH. Install Go, then retry")
	}
	return []string{"go", "run", "cmd/main.go"}, "", nil
}

func composePrefix(root string) []string {
	args := []string{"docker", "compose"}
	if fileExists(filepath.Join(root, ".env")) {
		args = append(args, "--env-file", ".env")
	}
	return args
}

func composeFileExists(root string) bool {
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		if fileExists(filepath.Join(root, name)) {
			return true
		}
	}
	return false
}

func runHint(cfg ProjectConfig) string {
	port := strings.TrimSpace(cfg.Port)
	if port == "" {
		port = DefaultPort
	}
	if cfg.ProjectType == ProjectTypeGRPC {
		return fmt.Sprintf("gRPC listening on %s", port)
	}
	return fmt.Sprintf("curl http://localhost:%s/healthz", port)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FormatRunCommand(args []string) string {
	return strings.Join(args, " ")
}
