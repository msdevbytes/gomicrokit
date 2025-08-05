package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var defaultDirs = []string{
	"cmd",
	"internal/bootstrap",
	"internal/api",
	"internal/config",
	"internal/db",
	"internal/dto",
	"internal/handler",
	"internal/model",
	"internal/repository",
	"internal/routes",
	"internal/service",
	"pkg",
	"pkg/utils",
	"pkg/logger",
	"test/mocks",
	"test/unit",
}

var defaultFiles = []string{
	".env",
	".air.toml",
	"README.md",
	"Dockerfile",
	"go.mod",                               // template generated
	".gitignore",                           // template generated
	"cmd/main.go",                          // template generated
	"internal/bootstrap/app.go",            // template generated
	"internal/api/router.go",               // template generated
	"internal/config/db_config.go",         // template generated
	"internal/config/pagination_config.go", // template generated
	"internal/handler/default.go",          // template generated
	"internal/handler/response.go",         // template generated
	"internal/db/db.go",                    // template generated
	"internal/db/migrations.go",            // template generated
	"internal/dto/pagination.go",           // template generated
	"internal/model/base.go",               // template generated
	"internal/service/container.go",        // template generated
	"pkg/logger/logger.go",                 // template generated
}

func ScaffoldProject(values map[string]string) (projectPath string, err error) {
	// generate project files and folders...
	return "./" + values["project"], nil
}

func GenerateProjectFrom(values map[string]string) error {
	project := values["project"]
	module := values["module"]
	framework := values["framework"]
	db := values["db"]
	gorm := values["gorm"]
	docker := values["docker"]

	// 👇 Replace with your actual logic later
	fmt.Printf("Generating project: %s\n", project)
	fmt.Printf("Generating module: %s\n", module)
	fmt.Printf("Using framework: %s\n", framework)
	fmt.Printf("Using DB: %s (GORM: %s)\n", db, gorm)
	fmt.Printf("Dockerfile: %s\n", docker)
	fmt.Printf("Pressing Enter to confirm...\n")

	return nil
}

func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InstallDeps(projectPath string) error {
	return runCommand(projectPath, "go", "mod", "tidy")
}

// CreateProjectStructure creates folders and base files for the new project
func CreateProjectStructure(projectName string, withDocker bool) error {
	if err := os.Mkdir(projectName, 0755); err != nil {
		return err
	}

	// create folders
	for _, dir := range defaultDirs {
		path := filepath.Join(projectName, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// create placeholder files
	for _, file := range defaultFiles {
		if !withDocker && file == "Dockerfile" {
			continue
		}
		path := filepath.Join(projectName, file)
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Project structure created under ./%s\n", projectName)
	return nil
}
