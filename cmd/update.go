package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	GMKInstallModule = "github.com/msdevbytes/gomicrokit/cmd/gmk"
	GMKProxyLatest   = "https://proxy.golang.org/github.com/msdevbytes/gomicrokit/@latest"
)

var (
	updateCheck bool
	updateTo    string

	UpdateHTTPClient = &http.Client{Timeout: 15 * time.Second}
	GoLookPath       = exec.LookPath
	RunGoInstall     = defaultGoInstall
	FetchLatest      = FetchLatestFromProxy
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"self-update"},
	Short:   "Update gmk to the latest published version",
	Long: `Install the latest gmk using Go modules (same as a fresh install).

Requires Go on PATH.

Examples:
  gmk update
  gmk update --check
  gmk update --to v1.4.1`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunUpdate(updateCheck, updateTo); err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Show current vs latest without installing")
	updateCmd.Flags().StringVar(&updateTo, "to", "", "Install a specific module version (e.g. v1.4.1)")
	rootCmd.AddCommand(updateCmd)
}

func RunUpdate(checkOnly bool, pin string) error {
	current := strings.TrimSpace(Version)
	fmt.Printf("Current: gmk %s\n", current)

	target := strings.TrimSpace(pin)
	latest := ""
	if target == "" {
		var err error
		latest, err = FetchLatest()
		if err != nil {
			if checkOnly {
				return fmt.Errorf("cannot check latest version: %w", err)
			}
			fmt.Printf("⚠️  Could not query module proxy (%v); installing @latest\n", err)
			target = "latest"
		} else {
			fmt.Printf("Latest:  %s\n", latest)
			target = latest
		}
	} else {
		fmt.Printf("Requested: %s\n", target)
	}

	if checkOnly {
		if target == "latest" {
			return nil
		}
		if SameKitVersion(current, target) {
			fmt.Println("Already up to date.")
			return nil
		}
		fmt.Println("Run gmk update to install the latest kit.")
		return nil
	}

	if target != "latest" && SameKitVersion(current, target) {
		fmt.Println("Already up to date.")
		return nil
	}

	if _, err := GoLookPath("go"); err != nil {
		mod := GMKInstallModule + "@" + target
		return fmt.Errorf("Go is not on PATH. Install Go, then run:\n  go install %s", mod)
	}

	mod := GMKInstallModule + "@" + target
	fmt.Printf("Installing %s ...\n", mod)
	if err := RunGoInstall(mod); err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("%w\nOn Windows, close other gmk processes and retry (the running binary cannot be overwritten)", err)
		}
		return err
	}
	fmt.Println("✅ Updated. Run gmk version to confirm.")
	return nil
}

func SameKitVersion(current, latest string) bool {
	c := normalizeKitVersion(current)
	l := normalizeKitVersion(latest)
	if c == "" || c == "dev" || c == "unknown" {
		return false
	}
	return c == l
}

func normalizeKitVersion(v string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(v)), "v")
}

func FetchLatestFromProxy() (string, error) {
	req, err := http.NewRequest(http.MethodGet, GMKProxyLatest, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := UpdateHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("module proxy returned %s", resp.Status)
	}
	return ParseModuleLatest(body)
}

func ParseModuleLatest(body []byte) (string, error) {
	var info struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", fmt.Errorf("cannot parse module proxy response: %w", err)
	}
	if strings.TrimSpace(info.Version) == "" {
		return "", fmt.Errorf("module proxy returned an empty version")
	}
	return info.Version, nil
}

func defaultGoInstall(mod string) error {
	cmd := exec.Command("go", "install", mod)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}
