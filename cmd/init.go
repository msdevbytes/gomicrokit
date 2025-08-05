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
)

func getGoVersion() string {
	ver := runtime.Version() // e.g. "go1.21.3"
	return strings.TrimPrefix(ver, "go")
}

var initCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Initialize a new Go microservice project",
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
		if projectName != "" {
			fmt.Printf("Scaffolding project: %s\n", projectName)
		}

		_ = generator.CreateProjectStructure(projectName, input["docker"] == "y")

		err := generator.RenderTemplates(projectName, generator.TemplateData{
			ProjectName: projectName,
			ModulePath:  modulePath,
			GoVersion:   getGoVersion(),
		})
		if err != nil {
			fmt.Println("❌ Error generating templates:", err)
			return
		}

		projectPath, err := generator.ScaffoldProject(input)
		if err != nil {
			fmt.Println("❌ Error during generation:", err)
			return
		}

		if err := generator.InstallDeps(projectPath); err != nil {
			fmt.Println("❌ Error during dependency installation:", err)
			return
		}
	},
}

func init() {
	initCmd.Flags().StringVar(&db, "db", "mysql", "Database type (postgres, mysql, sqlite)")
	initCmd.Flags().BoolVar(&docker, "docker", true, "Include Dockerfile and Air config")
	initCmd.Flags().StringVar(&module, "module", generator.GetGoModule(), "Go module path (e.g. github.com/user/project)")
}
