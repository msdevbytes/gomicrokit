package cmd

import (
	"fmt"

	"github.com/msdevbytes/gomicrokit/generator"
	"github.com/msdevbytes/gomicrokit/ui"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:     "make:service",
	Short:   "Generate a service with repo, handler, model, etc.",
	Example: "gomicrokit make:service --name=rsvp",
	Run: func(cmd *cobra.Command, args []string) {
		input := ui.RunServiceWizard()

		err := generator.GenerateService(generator.ServiceOptions{
			Name:       input.Name,
			ModulePath: input.ModulePath,
			Force:      input.Force,
			DryRun:     input.DryRun,
		})
		if err != nil {
			fmt.Println("❌ Error:", err)
		}
	},
}
