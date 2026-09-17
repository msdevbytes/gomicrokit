package cmd

import (
	"fmt"

	"github.com/msdevbytes/gomicrokit/generator"
	"github.com/msdevbytes/gomicrokit/ui"
	"github.com/spf13/cobra"
)

var (
	svcName      string
	svcForce     bool
	svcDryRun    bool
	svcType      string
	svcFramework string
)

var serviceCmd = &cobra.Command{
	Use:   "make:service",
	Short: "Generate a service with repo, handler, model, etc.",
	Long: `Generate a complete service with model, repository, service layer,
handler, and DTOs following the repository pattern.

Examples:
  gmk make:service                      # Interactive mode
  gmk make:service --name user          # Generate 'user' service
  gmk make:service --name user --framework chi # Generate chi-compatible handler/route
  gmk make:service --name user --type grpc # Generate grpc service artifacts
  gmk make:service --name user --force  # Overwrite existing files
  gmk make:service --name user --dry-run # Preview only`,
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		var force, dryRun bool
		var modulePath string
		currentCfg := generator.DetectProjectConfig(".")
		projectKind := currentCfg.ProjectType
		serviceFramework := currentCfg.Framework
		if svcType != "" {
			override, err := generator.NormalizeProjectType(svcType)
			if err != nil {
				fmt.Printf("❌ %v\n", err)
				return
			}
			projectKind = override
		}
		if svcFramework != "" {
			fw, err := generator.NormalizeFramework(projectKind, svcFramework)
			if err != nil {
				fmt.Printf("❌ %v\n", err)
				return
			}
			serviceFramework = string(fw)
		}

		// If --name flag provided, use non-interactive mode
		if svcName != "" {
			name = svcName
			force = svcForce
			dryRun = svcDryRun
			modulePath = generator.GetGoModule()
		} else {
			// Interactive mode
			input := ui.RunServiceWizard()
			name = input.Name
			force = input.Force
			dryRun = input.DryRun
			modulePath = input.ModulePath
		}

		err := generator.GenerateService(generator.ServiceOptions{
			Name:        name,
			ModulePath:  modulePath,
			Force:       force,
			DryRun:      dryRun,
			ProjectType: projectKind,
			Framework:   serviceFramework,
		})
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	},
}

func init() {
	serviceCmd.Flags().StringVar(&svcName, "name", "", "Service name (e.g., user, product)")
	serviceCmd.Flags().BoolVar(&svcForce, "force", false, "Overwrite existing files")
	serviceCmd.Flags().BoolVar(&svcDryRun, "dry-run", false, "Preview generated code without writing files")
	serviceCmd.Flags().StringVar(&svcType, "type", "", "Override project type for service generation (rest or grpc)")
	serviceCmd.Flags().StringVar(&svcFramework, "framework", "", "Override rest framework for service generation (fiber or chi)")
}
