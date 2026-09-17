package cmd

import (
	"fmt"

	"github.com/msdevbytes/gomicrokit/generator"
	"github.com/spf13/cobra"
)

var keyForce bool

var keyGenerateCmd = &cobra.Command{
	Use:   "key:generate",
	Short: "Generate a Stripe-style API_KEY and write it to .env",
	Long: `Generate a cryptographically random alphanumeric API_KEY and save it to the project's .env.

Keys look like Stripe secrets: svc_test_… (or svc_live_… when APP_ENV is prod/production/live).
The random part is 43 base62 characters (~256 bits) from crypto/rand.

Empty keys and the scaffold placeholder (change-me-in-prod) are replaced by default.
An existing real key is left unchanged unless --force is set.

Examples:
  gmk key:generate
  gmk key:generate --force`,
	Run: func(cmd *cobra.Command, args []string) {
		key, err := generator.GenerateProjectAPIKey(".", keyForce)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		fmt.Println("✅ API_KEY saved to .env")
		fmt.Println(key)
	},
}

func init() {
	keyGenerateCmd.Flags().BoolVar(&keyForce, "force", false, "Replace an existing API_KEY")
	rootCmd.AddCommand(keyGenerateCmd)
}
