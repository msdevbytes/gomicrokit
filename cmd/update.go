package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	GMKInstallModule = "github.com/msdevbytes/gomicrokit/cmd/gmk"
	GMKProxyLatest   = "https://proxy.golang.org/github.com/msdevbytes/gomicrokit/@latest"
	GMKGitHubLatest  = "https://api.github.com/repos/msdevbytes/gomicrokit/releases/latest"
)

var (
	updateCheck bool
	updateTo    string

	UpdateHTTPClient = &http.Client{Timeout: 15 * time.Second}
	GoLookPath       = exec.LookPath
	RunGoInstall     = defaultGoInstall
	FetchLatest      = FetchLatestPublished
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
  gmk update --to v1.4.2`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunUpdate(updateCheck, updateTo); err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Show current vs latest without installing")
	updateCmd.Flags().StringVar(&updateTo, "to", "", "Install a specific module version (e.g. v1.4.2)")
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
		switch CompareKitVersion(current, target) {
		case 0:
			fmt.Println("Already up to date.")
		case 1:
			fmt.Println("Already up to date (newer than the module proxy listing).")
		default:
			fmt.Println("Run gmk update to install the latest kit.")
		}
		return nil
	}

	if target != "latest" && CompareKitVersion(current, target) >= 0 {
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
	_, _, _, ok := parseKitVersion(current)
	if !ok {
		return false
	}
	return CompareKitVersion(current, latest) == 0
}

// CompareKitVersion returns -1 if current is older, 0 if equal, 1 if newer.
// Non-semver values (dev, unknown, empty) compare as older than a real version.
func CompareKitVersion(current, latest string) int {
	cm, ci, cp, cok := parseKitVersion(current)
	lm, li, lp, lok := parseKitVersion(latest)
	if !cok && !lok {
		return 0
	}
	if !cok {
		return -1
	}
	if !lok {
		return 1
	}
	switch {
	case cm != lm:
		return cmpInt(cm, lm)
	case ci != li:
		return cmpInt(ci, li)
	default:
		return cmpInt(cp, lp)
	}
}

func parseKitVersion(v string) (major, minor, patch int, ok bool) {
	s := normalizeKitVersion(v)
	if s == "" || s == "dev" || s == "unknown" {
		return 0, 0, 0, false
	}
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) == 0 || parts[0] == "" {
		return 0, 0, 0, false
	}
	nums := [3]int{}
	for i := 0; i < len(parts) && i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return 0, 0, 0, false
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], true
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func normalizeKitVersion(v string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(v)), "v")
}

func FetchLatestPublished() (string, error) {
	var versions []string
	if v, err := FetchLatestFromGitHub(); err == nil {
		versions = append(versions, v)
	}
	if v, err := FetchLatestFromProxy(); err == nil {
		versions = append(versions, v)
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("cannot determine latest version from GitHub or the module proxy")
	}
	best := versions[0]
	for _, v := range versions[1:] {
		if CompareKitVersion(v, best) > 0 {
			best = v
		}
	}
	return best, nil
}

func FetchLatestFromGitHub() (string, error) {
	req, err := http.NewRequest(http.MethodGet, GMKGitHubLatest, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
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
		return "", fmt.Errorf("GitHub releases returned %s", resp.Status)
	}
	var info struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", fmt.Errorf("cannot parse GitHub release: %w", err)
	}
	tag := strings.TrimSpace(info.TagName)
	if tag == "" {
		return "", fmt.Errorf("GitHub release has an empty tag")
	}
	return tag, nil
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
