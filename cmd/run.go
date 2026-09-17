package cmd

import (
	"fmt"
	"strings"

	"github.com/msdevbytes/gomicrokit/generator"
	"github.com/spf13/cobra"
)

var (
	runDryRun  bool
	runHost    bool
	runDetach  bool
	runNoWatch bool
)

var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"start"},
	Short:   "Start the generated project from .gmkrc.json",
	Long: `Start this project the way it was generated.

From a generated project directory (or any subdirectory):

  Docker Compose present and Docker is running:
    docker compose up --watch --build

  Docker unavailable, or --host:
    optional infra containers, then air (if configured) or go run cmd/main.go

Watch stays in the foreground so file sync works. Use --detach for
docker compose up -d --build (no live reload).

Examples:
  gmk run
  gmk run --dry-run
  gmk run --host
  gmk run --detach`,
	Run: func(cmd *cobra.Command, args []string) {
		plan, err := generator.PlanProjectRun(".", generator.RunOptions{
			Host:    runHost,
			Detach:  runDetach,
			NoWatch: runNoWatch,
		})
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		printRunPlan(plan)
		if runDryRun {
			return
		}
		if err := generator.ExecuteRunPlan(plan); err != nil {
			fmt.Printf("❌ %v\n", err)
		}
	},
}

func printRunPlan(plan generator.RunPlan) {
	cfg := plan.Config
	name := strings.TrimSpace(cfg.ProjectName)
	if name == "" {
		name = plan.Dir
	}
	kind := string(cfg.ProjectType)
	if kind == "" {
		kind = string(generator.ProjectTypeREST)
	}
	meta := kind
	if cfg.Framework != "" && cfg.ProjectType != generator.ProjectTypeGRPC {
		meta = kind + "/" + cfg.Framework
	}
	port := strings.TrimSpace(cfg.Port)
	if port == "" {
		port = generator.DefaultPort
	}
	fmt.Printf("▶ %s (%s) port %s\n", name, meta, port)
	fmt.Printf("Mode: %s\n", runModeLabel(plan.Mode))
	for _, warning := range plan.Warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}
	for _, step := range plan.Steps {
		fmt.Printf("$ %s\n", generator.FormatRunCommand(step.Args))
	}
	if plan.Hint != "" {
		fmt.Println(plan.Hint)
	}
}

func runModeLabel(mode string) string {
	switch mode {
	case generator.RunModeComposeWatch:
		return "Docker Compose watch"
	case generator.RunModeComposeUp:
		return "Docker Compose up"
	case generator.RunModeComposeDetach:
		return "Docker Compose detached"
	case generator.RunModeHost:
		return "host"
	default:
		return mode
	}
}

func init() {
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Print the command that would run, then exit")
	runCmd.Flags().BoolVar(&runHost, "host", false, "Run Air or go run on the host (start Compose infra only)")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "docker compose up -d --build (no file watch)")
	runCmd.Flags().BoolVar(&runNoWatch, "no-watch", false, "docker compose up --build in the foreground without watch")
	rootCmd.AddCommand(runCmd)
}
