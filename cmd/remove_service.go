package cmd

import (
	"fmt"

	"github.com/msdevbytes/gomicrokit/generator"
	"github.com/msdevbytes/gomicrokit/ui"
	"github.com/spf13/cobra"
)

var (
	rmServiceName string
	rmForceDelete bool
)

var removeServiceCmd = &cobra.Command{
	Use:   "remove:service",
	Short: "Remove a previously generated service",
	Long: `Remove a service and its generated files (model, repository, service,
handler, DTOs). Also cleans up container and router registrations.

Examples:
  gmk remove:service                    # Interactive mode
  gmk remove:service --name user        # Remove 'user' service
  gmk remove:service --name user --force # Bypass time check`,
	Run: func(cmd *cobra.Command, args []string) {
		// If --name flag provided, use non-interactive mode
		if rmServiceName != "" {
			if err := generator.RemoveService(rmServiceName, rmForceDelete); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		} else {
			// Interactive mode
			ui.RunRemoveFlow()
		}
	},
}

func init() {
	removeServiceCmd.Flags().StringVar(&rmServiceName, "name", "", "Name of the service to remove")
	removeServiceCmd.Flags().BoolVar(&rmForceDelete, "force", false, "Force delete even after freshness period")
	rootCmd.AddCommand(removeServiceCmd)
}
