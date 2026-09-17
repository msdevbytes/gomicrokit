package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gmk",
	Short: "Scaffold Go microservices with clean architecture",
	Long: `GMK (GoMicroKit) is an interactive CLI tool for generating scalable Go microservice
boilerplates with Repository pattern, Service layer, and REST handlers.

Examples:
  gmk new                    # Interactive project creation
  gmk new myapp --dry-run    # Preview what would be created
  gmk make:service           # Interactive service generation
  gmk run                    # Start the generated project
  gmk key:generate           # Write a secure API_KEY to .env
  gmk update                 # Install the latest gmk
  gmk version                # Show version info`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serviceCmd)
}
