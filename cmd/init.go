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
	db     string
	module string
	docker bool
	dryRun bool
)

func getGoVersion() string {
	ver := runtime.Version() // e.g. "go1.21.3"
	return strings.TrimPrefix(ver, "go")
}

var initCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Initialize a new Go microservice project",
	Long: `Create a new Go microservice project with clean architecture.

Includes: Fiber REST framework, GORM ORM, MySQL support, Docker setup,
Repository pattern, Service layer, and organized project structure.

Examples:
  gmk new                                    # Interactive mode
  gmk new myapp                              # Create 'myapp' project
  gmk new myapp --module github.com/me/myapp # With custom module
  gmk new myapp --dry-run                    # Preview only`,
	Run: func(cmd *cobra.Command, args []string) {
		var input map[string]string

		if len(args) < 1 {
			input = ui.RunInteractive()
		} else {
			input = map[string]string{
				"project":   args[0],
				"db":        db,
				"framework": "fiber", // optional CLI flag support later
				"gorm":      "y",
				"docker":    fmt.Sprintf("%v", docker),
			}
			if module == "" {
				module = args[0]
			}
		}

		projectName := input["project"]
		modulePath := input["module"]
		if modulePath == "" {
			modulePath = cases.Lower(language.Und).String(projectName)
		}

		if projectName == "" {
			fmt.Println("❌ Project name is required")
			return
		}

		// Dry run mode - show what would be created
		if dryRun {
			fmt.Println("🔍 Dry run mode - showing what would be created:")
			fmt.Printf("\n📁 Project: %s\n", projectName)
			fmt.Printf("📦 Module:  %s\n", modulePath)
			fmt.Printf("🐳 Docker:  %v\n", input["docker"] == "y")
			fmt.Printf("🗄️  DB:      %s\n", input["db"])
			fmt.Println("\n📂 Directories that would be created:")
			dirs := []string{
				"cmd", "internal/bootstrap", "internal/api", "internal/config",
				"internal/db", "internal/dto", "internal/handler", "internal/model",
				"internal/repository", "internal/service", "pkg/logger", "test/unit",
			}
			for _, d := range dirs {
				fmt.Printf("   - %s/%s\n", projectName, d)
			}
			fmt.Println("\n📄 Files that would be created:")
			files := []string{
				"cmd/main.go", "go.mod", ".env", ".gitignore",
				"internal/bootstrap/app.go", "internal/api/router.go",
				"internal/service/container.go", "pkg/logger/logger.go",
			}
			for _, f := range files {
				fmt.Printf("   - %s/%s\n", projectName, f)
			}
			return
		}

		fmt.Printf("🚀 Scaffolding project: %s\n", projectName)

		if err := generator.CreateProjectStructure(projectName, input["docker"] == "y"); err != nil {
			fmt.Printf("❌ Error creating project structure: %v\n", err)
			return
		}

		err := generator.RenderTemplates(projectName, generator.TemplateData{
			ProjectName: projectName,
			ModulePath:  modulePath,
			GoVersion:   getGoVersion(),
		})
		if err != nil {
			fmt.Printf("❌ Error generating templates: %v\n", err)
			return
		}

		projectPath, err := generator.ScaffoldProject(input)
		if err != nil {
			fmt.Printf("❌ Error during generation: %v\n", err)
			return
		}

		fmt.Println("📦 Installing dependencies...")
		if err := generator.InstallDeps(projectPath); err != nil {
			fmt.Printf("❌ Error during dependency installation: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Project '%s' created successfully!\n", projectName)
		fmt.Printf("\n📋 Next steps:\n")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Printf("   go run cmd/main.go\n")
	},
}

func init() {
	initCmd.Flags().StringVar(&db, "db", "mysql", "Database type (postgres, mysql, sqlite)")
	initCmd.Flags().BoolVar(&docker, "docker", true, "Include Dockerfile and Air config")
	initCmd.Flags().StringVar(&module, "module", "", "Go module path (e.g. github.com/user/project)")
	initCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be created without making changes")
}
