package cmd

import (
	"github.com/msdevbytes/gomicrokit/ui"
	"github.com/spf13/cobra"
)

var (
	serviceName string
	forceDelete bool
)

var removeServiceCmd = &cobra.Command{
	Use:   "remove:service",
	Short: "Interactively remove a generated service",
	Run: func(cmd *cobra.Command, args []string) {
		ui.RunRemoveFlow()
	},
}

func init() {
	removeServiceCmd.Flags().StringVar(&serviceName, "name", "", "Name of the service to remove")
	removeServiceCmd.Flags().BoolVar(&forceDelete, "force", false, "Force delete even after freshness period")
	rootCmd.AddCommand(removeServiceCmd)
}
